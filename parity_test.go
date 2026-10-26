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
