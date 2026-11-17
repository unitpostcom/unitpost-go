package unitpost

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type seen struct {
	method, path, rawPath, auth, idem, query string
}

func orgServer(t *testing.T, reply func(n int) string) (*httptest.Server, *[]seen) {
	t.Helper()
	var mu sync.Mutex
	var calls []seen
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, seen{r.Method, r.URL.Path, r.URL.EscapedPath(), r.Header.Get("Authorization"), r.Header.Get("Idempotency-Key"), r.URL.RawQuery})
		n := len(calls)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if reply != nil {
			// An httptest stub answering the SDK's own requests with canned JSON: no browser, no HTML.
			io.WriteString(w, reply(n)) // nosemgrep: go.lang.security.audit.xss.no-io-writestring-to-responsewriter.no-io-writestring-to-responsewriter
			return
		}
		io.WriteString(w, `{"id":"x"}`)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func newOrg(srv *httptest.Server, id string) *OrganizationClient {
	return NewOrganizationWithOptions(OrganizationOptions{APIKey: "ok_live_test", OrganizationID: id, BaseURL: srv.URL, HTTPClient: srv.Client()})
}

func recoverErr(f func()) (e *Error) {
	defer func() {
		if r := recover(); r != nil {
			e, _ = r.(*Error)
		}
	}()
	f()
	return nil
}

func TestOrganizationSendsOrgKeyAndFillsID(t *testing.T) {
	srv, calls := orgServer(t, nil)
	org := newOrg(srv, "org_a/b")
	if org.OrganizationID != "org_a/b" {
		t.Fatalf("id=%q", org.OrganizationID)
	}
	if _, err := org.Workspaces.Get(context.Background(), "ws 1"); err != nil {
		t.Fatal(err)
	}
	c := (*calls)[0]
	if c.auth != "Bearer ok_live_test" {
		t.Fatalf("auth=%q", c.auth)
	}
	if c.rawPath != "/api/v1/organizations/org_a%2Fb/workspaces/ws%201" {
		t.Fatalf("path=%q", c.rawPath)
	}
}

func TestOrganizationEnvFallbacks(t *testing.T) {
	t.Setenv("UNITPOST_ORG_KEY", "ok_live_env")
	t.Setenv("UNITPOST_ORGANIZATION_ID", "org_env")
	t.Setenv("UNITPOST_API_KEY", "pk_live_wrong")
	srv, calls := orgServer(t, nil)
	t.Setenv("UNITPOST_BASE_URL", srv.URL)
	org := NewOrganization()
	if _, err := org.Usage.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	c := (*calls)[0]
	if c.auth != "Bearer ok_live_env" || c.path != "/api/v1/organizations/org_env/usage" {
		t.Fatalf("got %+v", c)
	}
}

func TestOrganizationMissingKeyAndID(t *testing.T) {
	t.Setenv("UNITPOST_ORG_KEY", "")
	t.Setenv("UNITPOST_ORGANIZATION_ID", "")
	t.Setenv("UNITPOST_API_KEY", "pk_live_only")
	e := recoverErr(func() { NewOrganizationWithOptions(OrganizationOptions{OrganizationID: "org_1"}) })
	if e == nil || e.Code != "missing_api_key" || !strings.Contains(e.Message, "UNITPOST_ORG_KEY") || !strings.Contains(e.Message, "Organization → API") {
		t.Fatalf("err=%+v", e)
	}
	e = recoverErr(func() { NewOrganizationWithOptions(OrganizationOptions{APIKey: "ok_live_x"}) })
	if e == nil || e.Code != "missing_organization_id" || !strings.Contains(e.Message, "UNITPOST_ORGANIZATION_ID") || !strings.Contains(e.Message, "org_") {
		t.Fatalf("err=%+v", e)
	}
}

func TestOrganizationCreateIdempotencyKey(t *testing.T) {
	srv, calls := orgServer(t, nil)
	org := newOrg(srv, "org_1")
	ctx := context.Background()
	org.Workspaces.Create(ctx, map[string]any{"name": "A"}, WithIdempotencyKey("mine"))
	org.Workspaces.Create(ctx, map[string]any{"name": "B"})
	org.PlanQuotes.Confirm(ctx, "ws_1", "pq_1", map[string]any{"confirm_token": "t"}, WithIdempotencyKey("pay"))
	org.PlanQuotes.Confirm(ctx, "ws_1", "pq_1", map[string]any{"confirm_token": "t"})
	got := []string{(*calls)[0].idem, (*calls)[1].idem, (*calls)[2].idem, (*calls)[3].idem}
	if got[0] != "mine" || got[2] != "pay" || got[1] == "" || got[3] == "" || got[1] == got[3] {
		t.Fatalf("keys=%v", got)
	}
}

func TestOrganizationSharedPlanIdempotencyKey(t *testing.T) {
	srv, calls := orgServer(t, nil)
	org := newOrg(srv, "org_1")
	ctx := context.Background()
	org.SharedPlan.AddWorkspace(ctx, "ws_1", map[string]any{"cap_cents": 500}, WithIdempotencyKey("add"))
	org.SharedPlan.AddWorkspace(ctx, "ws_1", nil)
	org.SharedPlan.UpdateWorkspace(ctx, "ws_1", map[string]any{"cap_cents": nil})
	org.SharedPlan.RemoveWorkspace(ctx, "ws_1", WithIdempotencyKey("remove"))
	org.SharedPlan.RemoveWorkspace(ctx, "ws_1")
	var got []string
	for _, c := range *calls {
		got = append(got, c.idem)
	}
	if got[0] != "add" || got[1] == "" || got[2] != "" || got[3] != "remove" || got[4] == "" || got[1] == got[4] {
		t.Fatalf("keys=%v", got)
	}
}

func TestOrganizationRoutes(t *testing.T) {
	srv, calls := orgServer(t, nil)
	org := newOrg(srv, "org_1")
	ctx := context.Background()
	org.Workspaces.List(ctx, map[string]string{"external_id": "cust_1"})
	org.Workspaces.Create(ctx, map[string]any{})
	org.Workspaces.Get(ctx, "ws_1")
	org.Workspaces.Update(ctx, "ws_1", map[string]any{})
	org.Workspaces.Close(ctx, "ws_1")
	org.Workspaces.Restore(ctx, "ws_1")
	org.WorkspaceKeys.Create(ctx, "ws_1", map[string]any{})
	org.Webhooks.List(ctx, nil)
	org.Webhooks.Create(ctx, map[string]any{})
	org.Webhooks.Get(ctx, "wh_1")
	org.Webhooks.Update(ctx, "wh_1", map[string]any{})
	org.Webhooks.Delete(ctx, "wh_1")
	org.Webhooks.RotateSecret(ctx, "wh_1")
	org.Webhooks.Test(ctx, "wh_1")
	org.Webhooks.ListDeliveries(ctx, "wh_1", map[string]string{"workspace_id": "ws_1"})
	org.Webhooks.ReplayDelivery(ctx, "wh_1", "whd_1")
	org.Webhooks.Flush(ctx, "wh_1", nil)
	org.Usage.List(ctx, map[string]string{"period": "current"})
	org.PlanQuotes.Create(ctx, "ws_1", map[string]any{})
	org.PlanQuotes.Get(ctx, "ws_1", "pq_1")
	org.PlanQuotes.Confirm(ctx, "ws_1", "pq_1", map[string]any{})
	org.SharedPlan.Get(ctx)
	org.SharedPlan.ListWorkspaces(ctx)
	org.SharedPlan.AddWorkspace(ctx, "ws_1", nil)
	org.SharedPlan.UpdateWorkspace(ctx, "ws_1", map[string]any{"cap_cents": nil})
	org.SharedPlan.RemoveWorkspace(ctx, "ws_1")
	want := []string{
		"GET /organizations/org_1/workspaces",
		"POST /organizations/org_1/workspaces",
		"GET /organizations/org_1/workspaces/ws_1",
		"PATCH /organizations/org_1/workspaces/ws_1",
		"POST /organizations/org_1/workspaces/ws_1/close",
		"POST /organizations/org_1/workspaces/ws_1/restore",
		"POST /organizations/org_1/workspaces/ws_1/api-keys",
		"GET /organizations/org_1/webhooks",
		"POST /organizations/org_1/webhooks",
		"GET /organizations/org_1/webhooks/wh_1",
		"PATCH /organizations/org_1/webhooks/wh_1",
		"DELETE /organizations/org_1/webhooks/wh_1",
		"POST /organizations/org_1/webhooks/wh_1/rotate-secret",
		"POST /organizations/org_1/webhooks/wh_1/test",
		"GET /organizations/org_1/webhooks/wh_1/deliveries",
		"POST /organizations/org_1/webhooks/wh_1/deliveries/whd_1/replay",
		"POST /organizations/org_1/webhooks/wh_1/flush",
		"GET /organizations/org_1/usage",
		"POST /organizations/org_1/workspaces/ws_1/plan-quotes",
		"GET /organizations/org_1/workspaces/ws_1/plan-quotes/pq_1",
		"POST /organizations/org_1/workspaces/ws_1/plan-quotes/pq_1/confirm",
		"GET /organizations/org_1/shared-plan",
		"GET /organizations/org_1/shared-plan/workspaces",
		"PUT /organizations/org_1/shared-plan/workspaces/ws_1",
		"PATCH /organizations/org_1/shared-plan/workspaces/ws_1",
		"DELETE /organizations/org_1/shared-plan/workspaces/ws_1",
	}
	var got []string
	for _, c := range *calls {
		got = append(got, c.method+" "+strings.TrimPrefix(c.path, "/api/v1"))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("routes:\n got %v\nwant %v", got, want)
	}
	if (*calls)[0].query != "external_id=cust_1" || (*calls)[14].query != "workspace_id=ws_1" {
		t.Fatalf("query: %q %q", (*calls)[0].query, (*calls)[14].query)
	}
}

func TestOrganizationListAllFollowsCursor(t *testing.T) {
	srv, calls := orgServer(t, func(n int) string {
		if n == 1 {
			return `{"object":"list","has_more":true,"data":[{"id":"ws_1"},{"id":"ws_2"}]}`
		}
		return `{"object":"list","has_more":false,"data":[{"id":"ws_3"}]}`
	})
	org := newOrg(srv, "org_1")
	p := org.Workspaces.ListAll(map[string]string{"limit": "2", "external_id": "c"})
	var ids []string
	for p.Next(context.Background()) {
		ids = append(ids, p.Current()["id"].(string))
	}
	if p.Err() != nil || strings.Join(ids, ",") != "ws_1,ws_2,ws_3" {
		t.Fatalf("ids=%v err=%v", ids, p.Err())
	}
	if len(*calls) != 2 || !strings.Contains((*calls)[1].query, "after=ws_2") || !strings.Contains((*calls)[1].query, "external_id=c") {
		t.Fatalf("calls=%+v", *calls)
	}
	bad := org.Workspaces.ListAll(map[string]string{"before": "ws_1"})
	if bad.Next(context.Background()) || bad.Err() != ErrForwardOnly {
		t.Fatalf("before not refused: %v", bad.Err())
	}
}

// Usage rows have no `id`: their pager reads `workspace_id` for the cursor.
func TestOrganizationUsageListAllPagesByWorkspaceID(t *testing.T) {
	srv, calls := orgServer(t, func(n int) string {
		switch n {
		case 1:
			return `{"object":"list","has_more":true,"data":[{"object":"workspace_usage","workspace_id":"ws_1"}]}`
		case 2:
			return `{"object":"list","has_more":false,"data":[{"object":"workspace_usage","workspace_id":"ws_2"}]}`
		}
		return `{"object":"list","has_more":false,"data":[{"id":"wh_1"}]}`
	})
	org := newOrg(srv, "org_1")
	p := org.Usage.ListAll(map[string]string{"period": "current"})
	var ids []string
	for p.Next(context.Background()) {
		ids = append(ids, p.Current()["workspace_id"].(string))
	}
	if p.Err() != nil || strings.Join(ids, ",") != "ws_1,ws_2" {
		t.Fatalf("ids=%v err=%v", ids, p.Err())
	}
	if (*calls)[1].path != "/api/v1/organizations/org_1/usage" || !strings.Contains((*calls)[1].query, "after=ws_1") {
		t.Fatalf("calls=%+v", *calls)
	}
	w := org.Webhooks.ListAll(nil)
	if !w.Next(context.Background()) || w.Current()["id"] != "wh_1" || (*calls)[2].path != "/api/v1/organizations/org_1/webhooks" {
		t.Fatalf("webhooks ListAll: %+v %v", *calls, w.Err())
	}
}

func TestWorkspaceClientHasNoOrganizationAccessors(t *testing.T) {
	typ := reflect.TypeOf(Client{})
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if strings.Contains(strings.ToLower(name), "organization") || name == "WorkspaceKeys" || name == "PlanQuotes" {
			t.Errorf("Client has %s", name)
		}
	}
}
