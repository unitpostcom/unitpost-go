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
