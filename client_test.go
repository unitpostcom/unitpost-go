package unitpost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const goldenSecret = "whsec_dGVzdC1zZWNyZXQtMTIzNDU2Nzg5MA"
const goldenBody = `{"type":"email.delivered","created_at":"2023-11-14T22:13:20.000Z","data":{"id":"em_golden_1"}}`
const goldenSig = "v1,0+/GSmvt3iOjDLxc/D/1BZ4Da3gbNR4VyGP0fGMmE3I="

func TestEmailSend(t *testing.T) {
	var gotUA, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/api/v1/email") {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"em_1","status":"queued"}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	data, err := c.Email.Send(context.Background(), map[string]any{"from": "a@test.com", "to": "b@test.com"})
	if err != nil {
		t.Fatal(err)
	}
	m := data.(map[string]any)
	if m["id"] != "em_1" {
		t.Fatalf("id=%v", m["id"])
	}
	if gotAuth != "Bearer test_key" {
		t.Fatalf("auth=%q", gotAuth)
	}
	if !strings.HasPrefix(gotUA, "unitpost-go/") {
		t.Fatalf("ua=%q", gotUA)
	}
}

func TestEmailSendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		io.WriteString(w, `{"error":{"code":"validation_error","message":"bad from"}}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	_, err := c.Email.Send(context.Background(), map[string]any{"from": "x"})
	ue, ok := err.(*Error)
	if !ok || ue.Code != "validation_error" || ue.Status != 422 {
		t.Fatalf("err=%v", err)
	}
}

func TestSmsListForwardsFilters(t *testing.T) {
	var got map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/api/v1/sms") {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		got = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"object":"list","has_more":false,"data":[]}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	want := map[string]string{
		"direction":      "inbound",
		"status":         "received",
		"from":           "+15551230000",
		"to":             "+18005550100",
		"participant":    "+15551230000",
		"contact_id":     "contact_1",
		"campaign_id":    "smscmp_1",
		"created_after":  "2026-10-01T00:00:00Z",
		"created_before": "2026-10-24T00:00:00Z",
		"limit":          "20",
	}
	params := map[string]string{"after": ""} // empty values are dropped
	for k, v := range want {
		params[k] = v
	}
	if _, err := c.Sms.List(context.Background(), params); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("query=%v", got)
	}
	for k, v := range want {
		if g := got[k]; len(g) != 1 || g[0] != v {
			t.Fatalf("%s=%v, want %q", k, g, v)
		}
	}
}

func TestVerifyWebhookGolden(t *testing.T) {
	h := http.Header{}
	h.Set("svix-id", "msg_2VeXh1EN7xExAmAmQ8gxxYz")
	h.Set("svix-timestamp", "1700000000")
	h.Set("svix-signature", goldenSig)
	event, err := VerifyWebhook(goldenBody, goldenSecret, h, 10_000_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if event.(map[string]any)["type"] != "email.delivered" {
		t.Fatalf("%v", event)
	}
}

func TestVerifyWebhookTampered(t *testing.T) {
	ts := time.Now().Unix()
	h := http.Header{}
	h.Set("svix-id", "msg_1")
	h.Set("svix-timestamp", jsonInt(ts))
	h.Set("svix-signature", "v1,aaaa")
	_, err := VerifyWebhook(`{"type":"email.delivered"}`, goldenSecret, h, 300)
	if err == nil {
		t.Fatal("expected error")
	}
}

func jsonInt(n int64) string {
	b, _ := json.Marshal(n)
	return strings.Trim(string(b), `"`)
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"3", 3 * time.Second, true},
		{"1.5", 1500 * time.Millisecond, true},
		{"0", 0, true},
		{"Mon, 26 Oct 2026 12:00:07 GMT", 7 * time.Second, true},
		{"Monday, 26-Oct-26 12:00:07 GMT", 7 * time.Second, true}, // RFC 850
		{"Mon Oct 26 12:00:07 2026", 7 * time.Second, true},       // asctime
		{"Mon, 26 Oct 2026 11:00:00 GMT", 0, true},                // past date: retry now
		{"", 0, false},
		{"soon", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseRetryAfter(tc.in, now)
		if ok != tc.ok || got != tc.want {
			t.Errorf("parseRetryAfter(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestBackoffCapsRetryAfterAndJitters(t *testing.T) {
	c := &httpClient{retryBaseDelay: 500 * time.Millisecond, maxRetryDelay: 20 * time.Second}
	if got := c.backoff(0, time.Hour); got != 20*time.Second {
		t.Fatalf("Retry-After of an hour waited %v, want the 20s cap", got)
	}
	if got := c.backoff(0, 0); got != 0 {
		t.Fatalf("Retry-After: 0 waited %v, want 0", got)
	}
	if got := c.backoff(0, 3*time.Second); got != 3*time.Second {
		t.Fatalf("Retry-After: 3 waited %v", got)
	}
	for i := 0; i < 50; i++ {
		if got := c.backoff(1, noRetryAfter); got < 0 || got > time.Second {
			t.Fatalf("jittered backoff %v outside [0, 1s]", got)
		}
	}
}

// A 429 whose Retry-After is an HTTP date is honored (clamped to maxRetryDelay),
// not ignored in favor of the jittered backoff.
func TestRetry429HonorsHTTPDateRetryAfter(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{"error":{"code":"rate_limited","message":"slow down"}}`)
			return
		}
		io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()
	h := newHTTP(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	h.retryBaseDelay = time.Millisecond // a jittered backoff would wait at most 1ms
	h.maxRetryDelay = 200 * time.Millisecond
	start := time.Now()
	data, err := h.request(context.Background(), http.MethodGet, "/usage", nil, nil, "")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || data.(map[string]any)["ok"] != true {
		t.Fatalf("calls=%d data=%v", calls, data)
	}
	if elapsed < 200*time.Millisecond || elapsed > 5*time.Second {
		t.Fatalf("waited %v, want the 200ms cap", elapsed)
	}
}
