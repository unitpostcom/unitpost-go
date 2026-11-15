package unitpost

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallerIdempotencyKey(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		io.WriteString(w, `{"id":"x"}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	ctx := context.Background()
	body := map[string]any{"to": "b@test.com"}
	if _, err := c.Email.Send(ctx, body, WithIdempotencyKey("order-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Email.Batch(ctx, []any{body}, WithIdempotencyKey("batch-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.Send(ctx, body, WithIdempotencyKey("sms-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Email.Send(ctx, body); err != nil {
		t.Fatal(err)
	}
	if keys[0] != "order-1" || keys[1] != "batch-1" || keys[2] != "sms-1" {
		t.Fatalf("keys=%v", keys)
	}
	if keys[3] == "" || keys[3] == "order-1" {
		t.Fatalf("auto key=%q", keys[3])
	}
}

func TestListAllFollowsCursor(t *testing.T) {
	var afters []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		after := r.URL.Query().Get("after")
		afters = append(afters, after)
		if r.URL.Query().Get("status") != "active" {
			t.Errorf("filter dropped: %v", r.URL.Query())
		}
		if after == "" {
			io.WriteString(w, `{"object":"list","has_more":true,"data":[{"id":"c1"},{"id":"c2"}]}`)
			return
		}
		io.WriteString(w, `{"object":"list","has_more":false,"data":[{"id":"c3"}]}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	p := c.Contacts.ListAll(map[string]string{"status": "active"})
	var ids []string
	for p.Next(context.Background()) {
		ids = append(ids, p.Current()["id"].(string))
	}
	if err := p.Err(); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[2] != "c3" {
		t.Fatalf("ids=%v", ids)
	}
	if len(afters) != 2 || afters[1] != "c2" {
		t.Fatalf("afters=%v", afters)
	}
}

func TestListAllSurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		io.WriteString(w, `{"error":{"code":"forbidden","message":"no"}}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	p := c.Webhooks.ListAll(nil)
	if p.Next(context.Background()) {
		t.Fatal("expected no rows")
	}
	if ue, ok := p.Err().(*Error); !ok || ue.Code != "forbidden" {
		t.Fatalf("err=%v", p.Err())
	}
}

func TestListAllRefusesBefore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request may be sent: %s", r.URL)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	p := c.Contacts.ListAll(map[string]string{"before": "con_9"})
	if p.Next(context.Background()) {
		t.Fatal("expected no rows")
	}
	if p.Err() != ErrForwardOnly {
		t.Fatalf("err=%v", p.Err())
	}
}

// ADR-0160: releasing a number and removing / replacing a registration take
// the caller's key, or mint one.
func TestSmsLifecycleWritesTakeAKey(t *testing.T) {
	type hit struct{ method, path, key, body string }
	var hits []hit
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		hits = append(hits, hit{r.Method, r.URL.Path, r.Header.Get("Idempotency-Key"), string(b)})
		io.WriteString(w, `{"id":"x"}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	ctx := context.Background()
	if _, err := c.Sms.ReleaseNumber(ctx, "snum_1", map[string]any{"force": true}, WithIdempotencyKey("rel-1")); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.Registrations.Remove(ctx, "sreg_1", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.Registrations.Replace(ctx, "sreg_2", nil); err != nil {
		t.Fatal(err)
	}
	want := []hit{
		{"POST", "/api/v1/sms/numbers/snum_1/release", "rel-1", `{"force":true}`},
		{"POST", "/api/v1/sms/registrations/sreg_1/remove", "", `{}`},
		{"POST", "/api/v1/sms/registrations/sreg_2/replace", "", ``},
	}
	for i, w := range want {
		h := hits[i]
		if h.method != w.method || h.path != w.path || h.body != w.body {
			t.Fatalf("hit %d = %+v, want %+v", i, h, w)
		}
		if w.key != "" && h.key != w.key {
			t.Fatalf("hit %d key=%q, want %q", i, h.key, w.key)
		}
		if h.key == "" {
			t.Fatalf("hit %d has no Idempotency-Key", i)
		}
	}
}

// Two-way SMS: adding a texter as a contact and text-to-join keywords. The
// routes answer a repeat with 200, so the writes carry no Idempotency-Key.
func TestSmsAddContactAndJoinKeywords(t *testing.T) {
	type hit struct{ method, path, body string }
	var hits []hit
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		hits = append(hits, hit{r.Method, r.URL.Path, string(b)})
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		io.WriteString(w, `{"id":"x"}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	ctx := context.Background()
	if _, err := c.Sms.AddContact(ctx, "sms_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.ListJoinKeywords(ctx, "snum_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.AddJoinKeyword(ctx, "snum_1", map[string]any{"keyword": "JOIN"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Sms.RemoveJoinKeyword(ctx, "snum_1", "sjk_1"); err != nil {
		t.Fatal(err)
	}
	want := []hit{
		{"POST", "/api/v1/sms/sms_1/contact", `{}`},
		{"GET", "/api/v1/sms/numbers/snum_1/join-keywords", ``},
		{"POST", "/api/v1/sms/numbers/snum_1/join-keywords", `{"keyword":"JOIN"}`},
		{"DELETE", "/api/v1/sms/numbers/snum_1/join-keywords/sjk_1", ``},
	}
	if len(hits) != len(want) {
		t.Fatalf("hits = %+v", hits)
	}
	for i, w := range want {
		if hits[i] != w {
			t.Fatalf("hit %d = %+v, want %+v", i, hits[i], w)
		}
	}
}
