package unitpost

import (
	"context"
	"net/http"
	"os"
	"time"
)

// The organization client (ADR-0171): the second client, for an ORGANIZATION
// key (ok_live_…). A platform that gives each of its customers a Unitpost
// workspace uses it to create, close and restore those workspaces, mint their
// keys, manage the organization's webhook endpoints, read usage per workspace,
// change a workspace's plan and manage which workspaces share the
// organization's shared plan.
//
//	org := unitpost.NewOrganization() // UNITPOST_ORG_KEY + UNITPOST_ORGANIZATION_ID
//	ws, err := org.Workspaces.Create(ctx, map[string]any{"name": "Acme", "external_id": "cust_42"})
//
// One credential per client: Client never calls an organization route, and
// OrganizationClient never calls a workspace route. To work inside one of the
// workspaces, mint its key with WorkspaceKeys.Create and make a Client with
// it. Same transport, retries and errors as Client.

// OrganizationOptions configure an OrganizationClient.
type OrganizationOptions struct {
	// APIKey is the organization key (ok_live_…). Defaults to UNITPOST_ORG_KEY.
	APIKey string
	// OrganizationID is the organization's id (org_…). Defaults to UNITPOST_ORGANIZATION_ID.
	OrganizationID string
	BaseURL        string
	Timeout        time.Duration
	HTTPClient     *http.Client
	MaxRetries     int
	MaxRetriesSet  bool
}

// OrganizationClient is the Unitpost organization-key client. Construct once,
// reuse everywhere.
type OrganizationClient struct {
	// OrganizationID is the organization every call acts on (org_…).
	OrganizationID string
	// Workspaces are the organization's workspaces: List, ListAll, Create, Get, Update, Close, Restore.
	Workspaces *OrganizationWorkspaces
	// WorkspaceKeys mints keys for one of the organization's workspaces.
	WorkspaceKeys *OrganizationWorkspaceKeys
	// Webhooks are the organization's webhook endpoints: events from every
	// workspace it owns, each with workspace_id and external_id.
	Webhooks *OrganizationWebhooks
	// Usage is usage per workspace for a billing period (List, ListAll).
	Usage *OrganizationUsage
	// PlanQuotes change the plan of one workspace the organization pays for: Create, Get, Confirm.
	PlanQuotes *OrganizationPlanQuotes
	// SharedPlan is the organization's shared plan: Get, and ListWorkspaces,
	// AddWorkspace, UpdateWorkspace, RemoveWorkspace for the workspaces on it.
	SharedPlan *OrganizationSharedPlan
}

// NewOrganization reads UNITPOST_ORG_KEY and UNITPOST_ORGANIZATION_ID from the environment.
func NewOrganization() *OrganizationClient { return NewOrganizationWithOptions(OrganizationOptions{}) }

// NewOrganizationWithOptions constructs an organization client with explicit
// options. Like New it panics with a *Error when the key is missing
// (code "missing_api_key"), and with code "missing_organization_id" when the
// organization id is. The organization key is never read from UNITPOST_API_KEY.
func NewOrganizationWithOptions(opts OrganizationOptions) *OrganizationClient {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("UNITPOST_ORG_KEY")
	}
	if apiKey == "" {
		panic(&Error{Code: "missing_api_key", Message: "No organization key provided. Pass APIKey or set UNITPOST_ORG_KEY. An Owner or Admin makes one in the dashboard under Organization → API."})
	}
	orgID := opts.OrganizationID
	if orgID == "" {
		orgID = os.Getenv("UNITPOST_ORGANIZATION_ID")
	}
	if orgID == "" {
		panic(&Error{Code: "missing_organization_id", Message: "No organization id provided. Pass OrganizationID (org_…) or set UNITPOST_ORGANIZATION_ID."})
	}
	h := newHTTP(Options{
		APIKey:        apiKey,
		BaseURL:       opts.BaseURL,
		Timeout:       opts.Timeout,
		HTTPClient:    opts.HTTPClient,
		MaxRetries:    opts.MaxRetries,
		MaxRetriesSet: opts.MaxRetriesSet,
	})
	return &OrganizationClient{
		OrganizationID: orgID,
		Workspaces:     &OrganizationWorkspaces{http: h, organizationID: orgID},
		WorkspaceKeys:  &OrganizationWorkspaceKeys{http: h, organizationID: orgID},
		Webhooks:       &OrganizationWebhooks{http: h, organizationID: orgID},
		Usage:          &OrganizationUsage{http: h, organizationID: orgID},
		PlanQuotes:     &OrganizationPlanQuotes{http: h, organizationID: orgID},
		SharedPlan:     &OrganizationSharedPlan{http: h, organizationID: orgID},
	}
}

// OrganizationWorkspaces are the organization's workspaces.
type OrganizationWorkspaces struct {
	http           *httpClient
	organizationID string
}

// List returns the organization's workspaces, newest first (cursor-paginated).
// params: "limit", "after", "before", "external_id". Needs organizations:read.
func (x *OrganizationWorkspaces) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/workspaces", q(params), nil, "")
}

// ListAll iterates every workspace of the organization across pages (see
// Pager), with the same filters as List. It pages forward only.
func (x *OrganizationWorkspaces) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/organizations/"+enc(x.organizationID)+"/workspaces", params)
}

// Create makes a workspace (201). Its Owner is the key's creator. Pass
// WithIdempotencyKey so a retry never makes a second one; without it the SDK
// mints a key per call. Needs organizations:write.
func (x *OrganizationWorkspaces) Create(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces", nil, body, idempotencyKeyFrom(opts))
}

// Get returns one workspace (ws_…). Needs organizations:read.
func (x *OrganizationWorkspaces) Get(ctx context.Context, workspaceID string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID), nil, nil, "")
}

// Update renames a workspace or sets its external_id or its SMS purchases
// cap. Needs organizations:write.
func (x *OrganizationWorkspaces) Update(ctx context.Context, workspaceID string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID), nil, body, "")
}

// Close closes a workspace: it stops sending at once. It can be restored for
// 30 days (restorable_until); unless restored, it is deleted once those 30
// days are over and 7 days have passed since the end of the billing period it
// closed in, whichever is later (deletes_at). Closing twice changes nothing.
// Needs organizations:write.
func (x *OrganizationWorkspaces) Close(ctx context.Context, workspaceID string) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/close", nil, nil, "")
}

// Restore reopens a closed workspace within its 30 days. Needs organizations:write.
func (x *OrganizationWorkspaces) Restore(ctx context.Context, workspaceID string) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/restore", nil, nil, "")
}

// OrganizationWorkspaceKeys are the keys of the organization's workspaces.
type OrganizationWorkspaceKeys struct {
	http           *httpClient
	organizationID string
}

// Create mints a workspace key (201) for one of the organization's
// workspaces. The answer holds the secret ("key") once; use it with
// NewWithOptions(Options{APIKey: key}). Needs organizations:write.
func (x *OrganizationWorkspaceKeys) Create(ctx context.Context, workspaceID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/api-keys", nil, body, "")
}

// OrganizationWebhooks are the organization's webhook endpoints.
type OrganizationWebhooks struct {
	http           *httpClient
	organizationID string
}

// List returns the organization's webhook endpoints. Needs organizations:read.
func (x *OrganizationWebhooks) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/webhooks", q(params), nil, "")
}

// ListAll iterates every webhook endpoint of the organization across pages
// (see Pager). It pages forward only. Needs organizations:read.
func (x *OrganizationWebhooks) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/organizations/"+enc(x.organizationID)+"/webhooks", params)
}

// Create makes an endpoint. The answer has "signing_secret" exactly once.
// Needs organizations:write.
func (x *OrganizationWebhooks) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/webhooks", nil, body, "")
}

// Get returns one endpoint (wh_…). Needs organizations:read.
func (x *OrganizationWebhooks) Get(ctx context.Context, webhookID string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID), nil, nil, "")
}

// Update changes an endpoint. Needs organizations:write.
func (x *OrganizationWebhooks) Update(ctx context.Context, webhookID string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID), nil, body, "")
}

// Delete removes an endpoint. Needs organizations:write.
func (x *OrganizationWebhooks) Delete(ctx context.Context, webhookID string) (any, error) {
	return x.http.request(ctx, "DELETE", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID), nil, nil, "")
}

// RotateSecret mints a new signing secret; the answer carries it exactly
// once. Needs organizations:write.
func (x *OrganizationWebhooks) RotateSecret(ctx context.Context, webhookID string) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/rotate-secret", nil, nil, "")
}

// Test fires a sample event at the endpoint. Needs organizations:write.
func (x *OrganizationWebhooks) Test(ctx context.Context, webhookID string) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/test", nil, nil, "")
}

// ListDeliveries is the endpoint's delivery log, newest first; each row
// names its workspace_id and external_id. params: "limit", "after",
// "before", "status", "event_type", "workspace_id", "external_id". Needs
// organizations:write (the rows carry each event's data, so
// organizations:read alone is a 403).
func (x *OrganizationWebhooks) ListDeliveries(ctx context.Context, webhookID string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/deliveries", q(params), nil, "")
}

// ListAllDeliveries iterates the endpoint's whole delivery log across pages
// (see Pager), with the same filters as ListDeliveries. It pages forward
// only. Needs organizations:write.
func (x *OrganizationWebhooks) ListAllDeliveries(webhookID string, params map[string]string) *Pager {
	return newPager(x.http, "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/deliveries", params)
}

// ReplayDelivery delivers one event to the endpoint again (202). Needs
// organizations:write.
func (x *OrganizationWebhooks) ReplayDelivery(ctx context.Context, webhookID, deliveryID string) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/deliveries/"+enc(deliveryID)+"/replay", nil, nil, "")
}

// Flush delivers the backlog (202). body is optional:
// map[string]any{"include_failed": true} adds the ones that failed for good;
// nil delivers the queued ones only. Needs organizations:write.
func (x *OrganizationWebhooks) Flush(ctx context.Context, webhookID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/webhooks/"+enc(webhookID)+"/flush", nil, body, "")
}

// Verify checks a delivery's signature and returns the parsed event (the
// top-level VerifyWebhook).
func (x *OrganizationWebhooks) Verify(payload, secret string, headers http.Header, toleranceSeconds int) (any, error) {
	return VerifyWebhook(payload, secret, headers, toleranceSeconds)
}

// OrganizationUsage is usage per workspace.
type OrganizationUsage struct {
	http           *httpClient
	organizationID string
}

// List returns usage per workspace (workspace_usage rows), each for that
// workspace's own billing period containing "period". Paged by workspace.
// params: "period" ("current", or an ISO 8601 instant), "external_id",
// "limit", "after", "before". Needs organizations:read.
func (x *OrganizationUsage) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/usage", q(params), nil, "")
}

// ListAll iterates the usage row of every workspace across pages (see Pager),
// with the same filters as List, paging by the last row's workspace_id. It
// pages forward only. Needs organizations:read.
func (x *OrganizationUsage) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/organizations/"+enc(x.organizationID)+"/usage", params).byCursor("workspace_id")
}

// OrganizationPlanQuotes change the plan of one workspace.
type OrganizationPlanQuotes struct {
	http           *httpClient
	organizationID string
}

// Create prices a plan change for one workspace the organization pays for
// (201). It charges nothing; the quote lasts 10 minutes and holds a
// single-use "confirm_token" (only in this answer). Its "pending_change"
// ({"plan", "tier_index", "effective_at"} or nil) is a downgrade already
// scheduled for the period end; with one, a quote for the current plan and
// tier is a "noop" whose confirm cancels it. Needs organizations:billing.
func (x *OrganizationPlanQuotes) Create(ctx context.Context, workspaceID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/plan-quotes", nil, body, "")
}

// Get returns one quote (pq_…). Needs organizations:read.
func (x *OrganizationPlanQuotes) Get(ctx context.Context, workspaceID, quoteID string) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/plan-quotes/"+enc(quoteID), nil, nil, "")
}

// Confirm applies the quoted change and charges it. body is
// {"confirm_token": "…"}. Pass WithIdempotencyKey so a retry never charges
// twice; without it the SDK mints a key per call. Needs organizations:billing.
func (x *OrganizationPlanQuotes) Confirm(ctx context.Context, workspaceID, quoteID string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/organizations/"+enc(x.organizationID)+"/workspaces/"+enc(workspaceID)+"/plan-quotes/"+enc(quoteID)+"/confirm", nil, body, idempotencyKeyFrom(opts))
}

// OrganizationSharedPlan is the organization's shared plan: one plan whose
// allowance its workspaces share. Buying, upgrading and ending it are done in
// the dashboard.
type OrganizationSharedPlan struct {
	http           *httpClient
	organizationID string
}

// Get returns the organization's shared plan: whether it's turned on
// ("approved"), its "status" ("none" until one is bought), plan, price and
// period, and this period's "used_cents", "overage_cents", "paused". Needs
// organizations:read.
func (x *OrganizationSharedPlan) Get(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/shared-plan", nil, nil, "")
}

// ListWorkspaces returns every workspace on the shared plan, with its cap and
// what it drew this period (one page). Needs organizations:read.
func (x *OrganizationSharedPlan) ListWorkspaces(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/organizations/"+enc(x.organizationID)+"/shared-plan/workspaces", nil, nil, "")
}

// AddWorkspace puts a workspace on the shared plan (it moves onto the
// organization's card), or sets its cap when it's already on it. body is
// optional: map[string]any{"cap_cents": 500}; nil sends {}. Refusals carry the
// reason in Error.Details[0].Code (own_plan, limit_reached, …). Pass
// WithIdempotencyKey so a retry answers what the first call did; without it
// the SDK mints a key per call. Needs organizations:billing.
func (x *OrganizationSharedPlan) AddWorkspace(ctx context.Context, workspaceID string, body any, opts ...RequestOption) (any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return x.http.request(ctx, "PUT", "/organizations/"+enc(x.organizationID)+"/shared-plan/workspaces/"+enc(workspaceID), nil, body, idempotencyKeyFrom(opts))
}

// UpdateWorkspace sets a workspace's cap: map[string]any{"cap_cents": 500},
// or {"cap_cents": nil} to remove it. Needs organizations:billing.
func (x *OrganizationSharedPlan) UpdateWorkspace(ctx context.Context, workspaceID string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/organizations/"+enc(x.organizationID)+"/shared-plan/workspaces/"+enc(workspaceID), nil, body, "")
}

// RemoveWorkspace takes a workspace off the shared plan: "drops_to" is "free"
// or "needs_plan"; one not on it answers "left": false. Pass
// WithIdempotencyKey to choose the key; without it the SDK mints one per
// call. Needs organizations:billing.
func (x *OrganizationSharedPlan) RemoveWorkspace(ctx context.Context, workspaceID string, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "DELETE", "/organizations/"+enc(x.organizationID)+"/shared-plan/workspaces/"+enc(workspaceID), nil, nil, idempotencyKeyFrom(opts))
}
