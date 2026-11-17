package unitpost

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpendCapsGetAndUpdate(t *testing.T) {
	var methods, paths []string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		paths = append(paths, r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"object":"spend_caps","sms_purchases":{"cap_cents":5000}}`)
	}))
	defer srv.Close()
	c := NewWithOptions(Options{APIKey: "test_key", BaseURL: srv.URL, HTTPClient: srv.Client()})
	if _, err := c.SpendCaps.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	data, err := c.SpendCaps.Update(context.Background(), map[string]any{"sms_purchases": map[string]any{"cap_cents": 5000}})
	if err != nil {
		t.Fatal(err)
	}
	if data.(map[string]any)["object"] != "spend_caps" {
		t.Fatalf("data=%v", data)
	}
	if strings.Join(methods, ",") != "GET,PATCH" {
		t.Fatalf("methods=%v", methods)
	}
	for _, p := range paths {
		if !strings.HasSuffix(p, "/api/v1/spend-caps") {
			t.Fatalf("path=%q", p)
		}
	}
	var sent map[string]map[string]float64
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil || sent["sms_purchases"]["cap_cents"] != 5000 {
		t.Fatalf("body=%q err=%v", gotBody, err)
	}
}
