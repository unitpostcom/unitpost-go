package unitpost

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func q(params map[string]string) url.Values {
	if len(params) == 0 {
		return nil
	}
	out := url.Values{}
	for k, v := range params {
		if v != "" {
			out.Set(k, v)
		}
	}
	return out
}

// optParams returns the optional params map of a variadic List, or nil. It
// keeps List(ctx) compiling where a list gained paging params later.
func optParams(params []map[string]string) map[string]string {
	if len(params) > 0 {
		return params[0]
	}
	return nil
}

// Sms is the SMS channel (paid plans): send/list messages, plus the
// channel-scoped Templates, Campaigns and Topics resources and the SMS setup
// (MessagingPrograms, TestDestinations, ConsentSetups, Billing, Registrations,
// Quotes). Carrier purchases go through Quotes (a price, then a confirm with
// the opt-in sms:purchase scope); releasing a number, identity documents and
// toll-free verification stay in the dashboard.
type Sms struct {
	http              *httpClient
	Templates         *SmsTemplates
	Campaigns         *SmsCampaigns
	Topics            *SmsTopics
	MessagingPrograms *SmsMessagingPrograms
	TestDestinations  *SmsTestDestinations
	ConsentSetups     *SmsConsentSetups
	Billing           *SmsBilling
	Registrations     *SmsRegistrations
	Quotes            *SmsQuotes
}

func newSms(h *httpClient) *Sms {
	return &Sms{
		http:              h,
		Templates:         &SmsTemplates{http: h},
		Campaigns:         &SmsCampaigns{http: h},
		Topics:            &SmsTopics{http: h},
		MessagingPrograms: &SmsMessagingPrograms{http: h},
		TestDestinations:  &SmsTestDestinations{http: h},
		ConsentSetups:     &SmsConsentSetups{http: h},
		Billing:           &SmsBilling{http: h},
		Registrations:     &SmsRegistrations{http: h},
		Quotes:            &SmsQuotes{http: h},
	}
}

// SmsTemplates are plain-text SMS bodies with {{variable}} placeholders.
type SmsTemplates struct{ http *httpClient }

func (x *SmsTemplates) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/templates", q(params), nil, "")
}

// ListAll iterates every row of /sms/templates across pages (see Pager).
func (x *SmsTemplates) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/sms/templates", params)
}
func (x *SmsTemplates) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms/templates", nil, body, "")
}
func (x *SmsTemplates) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/templates/"+enc(id), nil, nil, "")
}
func (x *SmsTemplates) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/templates/"+enc(id), nil, body, "")
}
func (x *SmsTemplates) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/sms/templates/"+enc(id), nil, nil, "")
}

// Usage says what uses the template, and whether a change would be accepted
// now: its live campaigns and automations, queued texts, and can_edit_content
// / can_unpublish with the reason when not. Read it before an edit or a
// delete. Needs templates:read.
func (x *SmsTemplates) Usage(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/templates/"+enc(id)+"/usage", nil, nil, "")
}

// TestSend sends ONE test text of the template to body["to"], which must be a
// simulator number (+14254147755 delivers, +14254147167 fails) or a verified
// number from Sms.TestDestinations.List (any other is a 403). Every
// {{variable}} needs a value in "variables" (422). It goes from "sender_id",
// or is routed by the destination's country. It is a real text, billed like
// any. Needs sms:send. Pass WithIdempotencyKey so a retry never texts twice;
// without it the SDK mints a key per call.
func (x *SmsTemplates) TestSend(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/templates/"+enc(id)+"/test-send", nil, body, idempotencyKeyFrom(opts))
}

// SmsCampaigns send a published SMS template to a segment. Create makes a
// draft (omit sender_id for automatic routing); Send sends now or at
// scheduled_at (pass nil for now); Delete archives.
type SmsCampaigns struct{ http *httpClient }

func (x *SmsCampaigns) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/campaigns", q(params), nil, "")
}

// ListAll iterates every row of /sms/campaigns across pages (see Pager).
func (x *SmsCampaigns) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/sms/campaigns", params)
}
func (x *SmsCampaigns) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns", nil, body, "")
}
func (x *SmsCampaigns) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/campaigns/"+enc(id), nil, nil, "")
}
func (x *SmsCampaigns) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/campaigns/"+enc(id), nil, body, "")
}
func (x *SmsCampaigns) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/sms/campaigns/"+enc(id), nil, nil, "")
}
func (x *SmsCampaigns) Send(ctx context.Context, id string, body any) (any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/send", nil, body, "")
}
func (x *SmsCampaigns) Cancel(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/cancel", nil, nil, "")
}
func (x *SmsCampaigns) Pause(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/pause", nil, nil, "")
}
func (x *SmsCampaigns) Resume(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/resume", nil, nil, "")
}

// Unschedule moves a scheduled campaign back to draft so it can be edited; it
// clears its send time.
func (x *SmsCampaigns) Unschedule(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/unschedule", nil, nil, "")
}

// Duplicate copies a campaign into a new draft (201). body is optional
// (map[string]any{"name": "..."}, or nil). Pass WithIdempotencyKey so a retry
// returns the same copy; without it the SDK mints a key per call.
func (x *SmsCampaigns) Duplicate(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/duplicate", nil, body, idempotencyKeyFrom(opts))
}
func (x *SmsCampaigns) Reschedule(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms/campaigns/"+enc(id)+"/reschedule", nil, body, "")
}
func (x *SmsCampaigns) Validate(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/campaigns/"+enc(id)+"/validate", nil, nil, "")
}

// ListRecipients lists how each recipient fared: the carrier's answer for a
// text that went out (delivered, sent, failed, ...), or why none was sent
// (skipped, with reason_code). Newest first, cursor-paginated on id ("limit",
// "after", "before"); "q" matches digits of the number. skipped_by_reason
// counts skipped recipients over the whole campaign.
func (x *SmsCampaigns) ListRecipients(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/campaigns/"+enc(id)+"/recipients", q(params), nil, "")
}

// SmsTopics are SMS subscription topics (independent from email topics);
// Delete archives.
type SmsTopics struct{ http *httpClient }

func (x *SmsTopics) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/topics", q(params), nil, "")
}

// ListAll iterates every row of /sms/topics across pages (see Pager).
func (x *SmsTopics) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/sms/topics", params)
}
func (x *SmsTopics) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms/topics", nil, body, "")
}
func (x *SmsTopics) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/topics/"+enc(id), nil, nil, "")
}
func (x *SmsTopics) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/topics/"+enc(id), nil, body, "")
}
func (x *SmsTopics) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/sms/topics/"+enc(id), nil, nil, "")
}

// OptOut opts many contacts out of this SMS topic as a background job (202,
// the job): {"contact_ids": [...]} or {"emails": [...]} (up to 1,000) or
// {"segment_id": "..."} (every member). Follow it with Jobs.Get. Pass
// WithIdempotencyKey so a retry answers the same job; without it the SDK
// mints a key per call. There is no bulk opt-in.
func (x *SmsTopics) OptOut(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/topics/"+enc(id)+"/opt-out", nil, body, idempotencyKeyFrom(opts))
}

// SmsMessagingPrograms are what a sender registration says your texts are
// for, how people opt in, the help and stop replies and sample messages. A
// registration copies the program when it is submitted, so a later edit never
// changes one. Reads need sms:read; the writes need sms:manage, which FULL
// keys have and SENDING keys don't. Submitting a registration is a paid order
// through Sms.Quotes.
type SmsMessagingPrograms struct{ http *httpClient }

// List lists your programs, newest first (cursor-paginated). params is
// optional: "limit", "after", "before", and "archived" ("true" includes
// archived ones).
func (x *SmsMessagingPrograms) List(ctx context.Context, params ...map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/messaging-programs", q(optParams(params)), nil, "")
}

// Create creates a program (201; "name" is required). Each field is checked
// by the carriers' rules: lengths, help and stop replies that are a sentence
// and not just a keyword, links to a real web address. Pass
// WithIdempotencyKey so a retry never makes a second one; without it the SDK
// mints a key per call.
func (x *SmsMessagingPrograms) Create(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/messaging-programs", nil, body, idempotencyKeyFrom(opts))
}

// Get returns one program, archived ones too.
func (x *SmsMessagingPrograms) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/messaging-programs/"+enc(id), nil, nil, "")
}

// Update edits a program: only the fields you send change, nil clears one,
// and the result must still pass the carriers' rules. "revision" moves on a
// material change. An archived program is a 409: Duplicate it to keep
// editing.
func (x *SmsMessagingPrograms) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/messaging-programs/"+enc(id), nil, body, "")
}

// Delete archives a program. Registrations that copied it keep their copy.
func (x *SmsMessagingPrograms) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/sms/messaging-programs/"+enc(id), nil, nil, "")
}

// Duplicate copies a program (an archived one too) into a new editable one
// named "… (copy)" at revision 1 (201). It sends no body. Pass
// WithIdempotencyKey so a retry returns the same copy; without it the SDK
// mints a key per call.
func (x *SmsMessagingPrograms) Duplicate(ctx context.Context, id string, opts ...RequestOption) (any, error) {
	return x.duplicate(ctx, id, nil, opts...)
}

// duplicate is Duplicate's request. body is always nil (the route takes none
// and nil sends nothing); it is a parameter so the call has the shape of every
// other idempotent send, which sdk:check reads.
func (x *SmsMessagingPrograms) duplicate(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/messaging-programs/"+enc(id)+"/duplicate", nil, body, idempotencyKeyFrom(opts))
}

// SmsTestDestinations are the numbers a template test may text besides the
// simulator, "pending" until the code texted to one is entered, then
// "verified". Adding and verifying one stays in the dashboard (Phone numbers →
// Testing), because it proves the phone is yours. List needs sms:read; Delete
// needs sms:manage.
type SmsTestDestinations struct{ http *httpClient }

// List lists your test numbers, newest first (cursor-paginated). params is
// optional: "limit", "after", "before".
func (x *SmsTestDestinations) List(ctx context.Context, params ...map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/test-destinations", q(optParams(params)), nil, "")
}

// Delete removes a test number; template tests can no longer text it.
func (x *SmsTestDestinations) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/sms/test-destinations/"+enc(id), nil, nil, "")
}

// SmsConsentSetups are where and how people opt in to your texts: the form or
// page, and the brand and purpose it discloses. A messaging program points at
// one with "consent_setup_id". Needs sms:read.
type SmsConsentSetups struct{ http *httpClient }

// List lists your consent setups except archived ones, newest first
// (cursor-paginated). params is optional: "limit", "after", "before".
func (x *SmsConsentSetups) List(ctx context.Context, params ...map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/consent-setups", q(optParams(params)), nil, "")
}

// SmsBilling is what the workspace pays for SMS. Needs sms:read; buying goes
// through Sms.Quotes.
type SmsBilling struct{ http *httpClient }

// ListResources lists what the workspace pays for SMS every month: each
// number and 10DLC campaign, its monthly price, status and next charge, and
// for a past-due number the date it is released ("grace_until"). Newest
// first (cursor-paginated); params is optional: "limit", "after", "before".
func (x *SmsBilling) ListResources(ctx context.Context, params ...map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/billing/resources", q(optParams(params)), nil, "")
}

// SmsQuotes are carrier purchases: register a 10DLC brand, vet it, register a
// campaign, order a number. Each is two calls: Create prices the order and
// binds what it buys, Confirm places it at the price you showed.
type SmsQuotes struct{ http *httpClient }

// Create prices an order (201). The quote binds the action and what it buys,
// and returns "due_now_cents", "monthly_cents", "line_items" and a single-use
// "confirm_token" (only in this response). It lasts 10 minutes and charges
// nothing. Needs sms:manage.
func (x *SmsQuotes) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms/quotes", nil, body, "")
}

// Get returns one quote: its price, status and, once confirmed, what the
// order made. Needs sms:read.
func (x *SmsQuotes) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/quotes/"+enc(id), nil, nil, "")
}

// Confirm places the quoted order and charges it. Show the person the quote's
// price and get their yes first. Needs sms:purchase, which a key or app has
// only when a member who holds it ticked it, and the owner's monthly cap for SMS
// purchases by API. A changed price or content answers 409: quote again. body
// is {"confirm_token": "sqct_…"}. Without WithIdempotencyKey the SDK mints a
// key per call so a retry can't buy twice; pass your own to replay the same
// answer later.
func (x *SmsQuotes) Confirm(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/quotes/"+enc(id)+"/confirm", nil, body, idempotencyKeyFrom(opts))
}

// SmsRegistrations are carrier registrations: 10DLC brand and campaign
// registrations and toll-free verifications, with their review state and the
// carrier's feedback. Submitting one is a paid order through Sms.Quotes;
// toll-free verification and removing one stay in the dashboard. Needs
// sms:read.
type SmsRegistrations struct{ http *httpClient }

// List lists your registrations, newest first (cursor-paginated). params is
// optional: "limit", "after", "before". A 10DLC campaign also carries the
// number it backs ("sender") and one "setup_state".
func (x *SmsRegistrations) List(ctx context.Context, params ...map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/registrations", q(optParams(params)), nil, "")
}

// Send sends one SMS. Pass WithIdempotencyKey to make a retried send safe across
// your own process restarts; without it the SDK mints a key per call.
func (x *Sms) Send(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms", nil, body, idempotencyKeyFrom(opts))
}
func (x *Sms) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/"+enc(id), nil, nil, "")
}

// Update cancels a queued/scheduled SMS ({"cancel": true}, credit refunded)
// or moves a scheduled one ({"scheduled_at": "<ISO 8601>"}).
func (x *Sms) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/"+enc(id), nil, body, "")
}

// List lists SMS messages (cursor-paginated, newest first). params: "limit",
// "after", "before", "direction", "status", "from", "to", "participant",
// "contact_id", "campaign_id", "created_after", "created_before", "q" (the
// message text; 3 or more digits also match either number; at most 200
// characters), "type" ("transactional" or "marketing") and "sender_id" (a
// number's "snum_…" id: its texts, both directions). A "q", "type" or
// "sender_id" value the list can't apply is ignored and named in the Unitpost-Warning
// response header; Unitpost-Strict: true refuses it instead.
func (x *Sms) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms", q(params), nil, "")
}

// ListAll iterates every row of /sms across pages (see Pager).
func (x *Sms) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/sms", params)
}

// Brands lists SMS brands and their setup status. A brand is the business
// identity carriers register before they approve a phone number, so this is what
// to poll to answer "is SMS ready yet?": status says whether you or the carrier
// is the blocker, and missing lists exactly what is still required. Create and
// edit a brand with CreateBrand / UpdateBrand (free); registering it with
// carriers is a paid order: Sms.Quotes.Create prices it, Sms.Quotes.Confirm
// places it. Needs sms:read.
func (x *Sms) Brands(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/brands", q(params), nil, "")
}

// CreateBrand creates a brand (201): the business identity carriers register
// before they approve a number. A draft may be partial (only "name" is
// required); "missing" says what is still needed. Free, with no carrier call;
// one brand per website domain (409). The tax id is write-only: stored
// encrypted, read back as "tax_id_on_file". Registering the brand with
// carriers is a paid order through Sms.Quotes. Needs sms:manage.
// Pass WithIdempotencyKey so a retry never makes a second brand; without it
// the SDK mints a key per call.
func (x *Sms) CreateBrand(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/sms/brands", nil, body, idempotencyKeyFrom(opts))
}

// GetBrand returns one brand with every field you can edit, its "status" and
// "missing" list. Needs sms:read.
func (x *Sms) GetBrand(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/brands/"+enc(id), nil, nil, "")
}

// UpdateBrand edits a brand: only the fields you send change, nil clears an
// optional one; send {"archived": ...} alone to archive or restore it. The
// tax id is write-only. A locked brand ("editable": false: carriers are
// reviewing or approved it, or it has numbers) refuses carrier-field edits
// with 409 brand_locked, "details" naming each field, except those a
// correction request opened. Needs sms:manage.
func (x *Sms) UpdateBrand(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/sms/brands/"+enc(id), nil, body, "")
}

// Numbers lists SMS phone numbers. Only numbers with status "active" can send. A
// ten_dlc number has no phone_number until carriers approve it, and a simulator
// number only reaches verified test destinations. Order a number through
// Sms.Quotes; releasing one stays in the dashboard.
func (x *Sms) Numbers(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/numbers", q(params), nil, "")
}

// Stats returns SMS delivery stats over a rolling window ("days": 1–365,
// default 30): queued, sent, delivered, failed, suppressed, canceled, inbound
// and segments billed, with delivery, failure and suppression rates. No opens
// or clicks on SMS.
func (x *Sms) Stats(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/stats", q(params), nil, "")
}

// SmsConsent reads a contact's SMS consent state and immutable consent history.
func (x *Sms) SmsConsent(ctx context.Context, contactID string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(contactID)+"/sms-consent", nil, nil, "")
}

// RecordSmsConsent records an SMS consent change (opt-in or opt-out) for a contact.
func (x *Sms) RecordSmsConsent(ctx context.Context, contactID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/"+enc(contactID)+"/sms-consent", nil, body, "")
}

// RequestSmsOptIn texts the contact a request to reply YES to confirm
// marketing texts. body is optional (map[string]any{"phone": "+1..."}, or nil).
// It sends a real text, so pass WithIdempotencyKey to make a retry safe;
// without it the SDK mints a key per call.
func (x *Sms) RequestSmsOptIn(ctx context.Context, contactID string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/"+enc(contactID)+"/sms-opt-in-request", nil, body, idempotencyKeyFrom(opts))
}

// Email is the channel root: send, batch, templates, campaigns, topics,
// domains and inbox conversations (Threads).
type Email struct {
	http      *httpClient
	Topics    *Topics
	Campaigns *Campaigns
	Templates *Templates
	Domains   *Domains
	Threads   *EmailThreads
}

func newEmail(h *httpClient) *Email {
	return &Email{http: h, Topics: &Topics{http: h}, Campaigns: &Campaigns{http: h}, Templates: &Templates{http: h}, Domains: &Domains{http: h}, Threads: &EmailThreads{http: h}}
}

// Send sends or schedules one transactional email. Pass WithIdempotencyKey to
// make a retried send safe across your own process restarts; without it the SDK
// mints a key per call.
func (x *Email) Send(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email", nil, body, idempotencyKeyFrom(opts))
}

// Batch sends up to 100 emails in one request; WithIdempotencyKey as for Send.
func (x *Email) Batch(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/batch", nil, body, idempotencyKeyFrom(opts))
}
func (x *Email) GetBatch(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/batches/"+enc(id), nil, nil, "")
}
func (x *Email) CancelBatch(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/batches/"+enc(id)+"/cancel", nil, nil, "")
}

// ListBatches lists your batch sends, newest first (cursor-paginated), as the
// dashboard's Batches tab lists them. Each row is the same live rollup
// GetBatch returns: how many emails it holds, when it's scheduled, a count per
// delivery status and how many are still cancelable. params: "limit",
// "after", "before".
func (x *Email) ListBatches(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/batches", q(params), nil, "")
}

// List lists sent and scheduled emails (cursor-paginated, newest first).
// params: "limit", "after", "before", "status", "to", "contact_id",
// "campaign_id", "batch_id", "created_after", "created_before", "q" (an exact
// recipient address, or part of the subject or the sender; at most 200
// characters), "source" ("api", "smtp", "test", "batch", "campaign" or
// "reply") and "outcome" ("delivered", "opened", "clicked", "unsubscribed",
// "bounced", "complained", "failed" or "pending"; with "campaign_id": who
// opened or clicked that campaign). A "q", "source" or "outcome" value the
// list can't apply is ignored and named in the Unitpost-Warning response
// header; Unitpost-Strict: true refuses it instead.
func (x *Email) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email", q(params), nil, "")
}

// ListAll iterates every row of /email across pages (see Pager).
func (x *Email) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/email", params)
}
func (x *Email) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/"+enc(id), nil, nil, "")
}

// ListEvents lists one sent email's delivery events, oldest first, as the
// dashboard's email detail shows the timeline. Each event carries its detail
// in "data" under the webhook's names (data.bounce.diagnostic_code,
// data.click.link, data.complaint.feedback_type, ...). Cursor-paginated:
// params "limit", "after", "before" ("after" / "before" take an emev_ id).
// Needs emails:read.
func (x *Email) ListEvents(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/"+enc(id)+"/events", q(params), nil, "")
}
func (x *Email) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/"+enc(id), nil, body, "")
}

// Retry sends a failed email again on the same id, as the dashboard's Retry
// does. Only a one-off send whose failure was temporary is retried: a campaign
// recipient, a template test, a suppressed or rejected send, or a content
// error is a 409 saying why (send a new email instead). It passes the same
// checks as a new send: the sending domain, your plan and quota. It sends no
// body; follow the email with Get. It sends real email, so pass
// WithIdempotencyKey to make a retried call safe across your own process
// restarts; without it the SDK mints a key per call.
func (x *Email) Retry(ctx context.Context, id string, opts ...RequestOption) (any, error) {
	return x.retry(ctx, id, nil, opts...)
}

// retry is Retry's request. body is always nil (the route takes none and nil
// sends nothing); it is a parameter so the call has the shape of every other
// idempotent send, which sdk:check reads.
func (x *Email) retry(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/"+enc(id)+"/retry", nil, body, idempotencyKeyFrom(opts))
}

// Stats returns deliverability/engagement aggregates over a rolling window
// ("days": 1–365, default 30) or a period: "from" and "to", ISO 8601 with Z
// or an offset. Add "grain" ("5m", "1h" or "1d", with "from"/"to") and "tz"
// (an IANA zone) for "series", the send volume per bucket. "domain_id"
// (dom_…) counts one sending domain. A value the API can't apply is ignored
// and named in the Unitpost-Warning response header; Unitpost-Strict: true
// refuses it instead.
func (x *Email) Stats(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/stats", q(params), nil, "")
}
func (x *Email) ReceivedList(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/received", q(params), nil, "")
}
func (x *Email) ReceivedGet(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/received/"+enc(id), nil, nil, "")
}
func (x *Email) ReceivedAttachmentUrl(ctx context.Context, id, attachmentID string) (any, error) {
	return x.http.request(ctx, "GET", "/email/received/"+enc(id)+"/attachments/"+enc(attachmentID), nil, nil, "")
}

// ReceivedRaw returns a signed URL to download the received message's
// original MIME (.eml), valid for about 5 minutes ("expires_at"), with its
// "size_bytes". A body over 256 KB isn't inlined on ReceivedGet (its "text"
// and "html" are null), so this is how to read it. A copy of one of your own
// sends ("sent_by_us") is a 403. Needs emails:read.
func (x *Email) ReceivedRaw(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/received/"+enc(id)+"/raw", nil, nil, "")
}

// Contacts manages workspace contacts.
type Contacts struct{ http *httpClient }

func (x *Contacts) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts", q(params), nil, "")
}

// ListAll iterates every row of /contacts across pages (see Pager).
func (x *Contacts) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/contacts", params)
}
func (x *Contacts) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts", nil, body, "")
}
func (x *Contacts) Get(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(idOrEmail), nil, nil, "")
}
func (x *Contacts) Update(ctx context.Context, idOrEmail string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/contacts/"+enc(idOrEmail), nil, body, "")
}
func (x *Contacts) Delete(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "DELETE", "/contacts/"+enc(idOrEmail), nil, nil, "")
}
func (x *Contacts) Import(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/imports", nil, body, "")
}
func (x *Contacts) ListImports(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/imports", q(params), nil, "")
}
func (x *Contacts) GetImport(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/imports/"+enc(id), nil, nil, "")
}
func (x *Contacts) CancelImport(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/imports/"+enc(id)+"/cancel", nil, nil, "")
}

// RetryImport re-runs a failed import from its stored rows (202); rows already
// added are skipped.
func (x *Contacts) RetryImport(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/imports/"+enc(id)+"/retry", nil, nil, "")
}

// Addresses lists the contact's email addresses and phone numbers, with each
// one's status (one page, never has_more).
func (x *Contacts) Addresses(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(idOrEmail)+"/addresses", nil, nil, "")
}

// Activity lists the contact's most recent emails and texts, newest first
// ("limit" 1–100, default 25; no cursor).
func (x *Contacts) Activity(ctx context.Context, idOrEmail string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(idOrEmail)+"/activity", q(params), nil, "")
}

// ListSegments lists the segments the contact is in, with its per-segment
// opt-out (not paginated). To add the contact to a static segment or remove
// it, use Segments.AddMember / Segments.RemoveMember.
func (x *Contacts) ListSegments(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(idOrEmail)+"/segments", nil, nil, "")
}

// Export exports contacts to a CSV file (202), as the Contacts page's Export
// does. body is optional: the filters "q", "status", "phone", "segment_id"
// and "segment_subscription" (as on List) keep the matching contacts; nil
// exports every contact. It runs as a background job: follow it with
// Jobs.Get, then fetch the file with Jobs.Download. One export runs at a
// time: another while it runs is a 409 "job_in_progress". "segment_id" also
// needs segments:read.
func (x *Contacts) Export(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/export", nil, body, "")
}

// ContactFields are custom contact attributes.
type ContactFields struct{ http *httpClient }

func (x *ContactFields) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contact-fields", q(params), nil, "")
}

// ListAll iterates every row of /contact-fields across pages (see Pager).
func (x *ContactFields) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/contact-fields", params)
}
func (x *ContactFields) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contact-fields", nil, body, "")
}
func (x *ContactFields) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/contact-fields/"+enc(id), nil, nil, "")
}
func (x *ContactFields) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/contact-fields/"+enc(id), nil, body, "")
}
func (x *ContactFields) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/contact-fields/"+enc(id), nil, nil, "")
}
func (x *ContactFields) Rename(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contact-fields/"+enc(id)+"/rename", nil, body, "")
}

// Segments are contact groups.
type Jobs struct{ http *httpClient }

// List lists your background jobs, newest first (cursor-paginated), as
// Settings → Bulk operations lists them: the kinds an API operation starts
// that your key may read. params: "status" (QUEUED, RUNNING, COMPLETED,
// FAILED, CANCELED), "kind" ("contacts.export", "contact_field.delete", …),
// "limit", "after", "before".
func (x *Jobs) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/jobs", q(params), nil, "")
}

// Get retrieves a background job: the 202 answer of an operation too large for
// one request. Poll until status is COMPLETED, FAILED or CANCELED.
func (x *Jobs) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/jobs/"+enc(id), nil, nil, "")
}

// Cancel stops a queued or running job. A running job stops after the part in
// flight; the work already done stays. A kind that can't stop half-way (a
// custom-field rename) is a 409 saying why. It sends no body.
func (x *Jobs) Cancel(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/jobs/"+enc(id)+"/cancel", nil, nil, "")
}

// Retry resumes a FAILED job from where it stopped (202): nothing it already
// processed runs again. Any other status is a 409. It sends no body.
func (x *Jobs) Retry(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/jobs/"+enc(id)+"/retry", nil, nil, "")
}

// Download returns a finished job's file (a contacts export's CSV) as a signed
// "url" that expires in 5 minutes: fetch it right away, and call Download
// again for a fresh one. A job that isn't COMPLETED, or a kind with no file,
// is a 409.
func (x *Jobs) Download(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/jobs/"+enc(id)+"/download", nil, nil, "")
}

type Segments struct{ http *httpClient }

func (x *Segments) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/segments", q(params), nil, "")
}

// ListAll iterates every row of /segments across pages (see Pager).
func (x *Segments) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/segments", params)
}
func (x *Segments) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/segments", nil, body, "")
}
func (x *Segments) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/segments/"+enc(id), nil, nil, "")
}
func (x *Segments) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/segments/"+enc(id), nil, body, "")
}
func (x *Segments) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/segments/"+enc(id), nil, nil, "")
}
func (x *Segments) ListMembers(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/segments/"+enc(id)+"/contacts", q(params), nil, "")
}
func (x *Segments) AddMember(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/segments/"+enc(id)+"/contacts", nil, body, "")
}
func (x *Segments) RemoveMember(ctx context.Context, id, contact string) (any, error) {
	return x.http.request(ctx, "DELETE", "/segments/"+enc(id)+"/contacts/"+enc(contact), nil, nil, "")
}

// UpdateMember opts a member (by id or email) out of this segment's sends
// without removing it (map[string]any{"subscribed": false}), or back in, as
// the members table's toggle does. It stays a member. A dynamic segment has no
// per-segment opt-out (422).
func (x *Segments) UpdateMember(ctx context.Context, id, contact string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/segments/"+enc(id)+"/contacts/"+enc(contact), nil, body, "")
}

// Preview says how many contacts a saved segment reaches: its members, the
// subscribed set a campaign would mail (0 means a campaign would send
// nothing), and a sample (which also needs contacts:read).
func (x *Segments) Preview(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/segments/"+enc(id)+"/preview", nil, nil, "")
}

// PreviewFilter counts an unsaved dynamic filter (map[string]any{"filter":
// ...}) before you create the segment, as the segment dialog's live count
// does: how many contacts match, how many of them a campaign would mail, and a
// sample (which also needs contacts:read). Nothing is saved.
func (x *Segments) PreviewFilter(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/segments/preview", nil, body, "")
}

// AllContacts returns the workspace's managed "All subscribed contacts"
// segment, the dashboard's "Everyone subscribed" campaign target: made the
// first time you ask (201), the same segment after that (200). It sends no
// body.
func (x *Segments) AllContacts(ctx context.Context) (any, error) {
	return x.http.request(ctx, "POST", "/segments/all-contacts", nil, nil, "")
}

// Topics are email subscription topics.
type Topics struct{ http *httpClient }

func (x *Topics) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/topics", q(params), nil, "")
}

// ListAll iterates every row of /email/topics across pages (see Pager).
func (x *Topics) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/email/topics", params)
}
func (x *Topics) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/topics", nil, body, "")
}
func (x *Topics) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/topics/"+enc(id), nil, nil, "")
}
func (x *Topics) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/topics/"+enc(id), nil, body, "")
}
func (x *Topics) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/topics/"+enc(id), nil, nil, "")
}

// OptOut opts many contacts out of this topic as a background job (202, the
// job): {"contact_ids": [...]} or {"emails": [...]} (up to 1,000, each a
// contact of this workspace) or {"segment_id": "..."} (every member, however
// many). Each is opted out as SetTopic would, with its webhook only when it
// was subscribed. Follow it with Jobs.Get. Pass WithIdempotencyKey so a retry
// answers the same job; without it the SDK mints a key per call. There is no
// bulk opt-in.
func (x *Topics) OptOut(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/topics/"+enc(id)+"/opt-out", nil, body, idempotencyKeyFrom(opts))
}
func (x *Topics) ListTopics(ctx context.Context, contactID string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(contactID)+"/topics", q(params), nil, "")
}

func (x *Topics) SetTopic(ctx context.Context, contactID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/"+enc(contactID)+"/topics", nil, body, "")
}

// UpdateTopics changes a contact's preferences for several topics in one call
// (map[string]any{"topics": []map[string]any{{"topic_id": "top_...",
// "subscribed": false}}}, up to 50, each topic once); topics left out keep
// theirs. Every topic is checked before any is set (an unknown one is a 404
// and nothing changes). It answers the resolved preferences in the order sent
// ({"object": "list", "has_more": false, "data": [...]}). Idempotent.
func (x *Topics) UpdateTopics(ctx context.Context, contactID string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/contacts/"+enc(contactID)+"/topics", nil, body, "")
}

// Campaigns are marketing broadcasts.
type Campaigns struct{ http *httpClient }

func (x *Campaigns) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns", q(params), nil, "")
}

// ListAll iterates every row of /email/campaigns across pages (see Pager).
func (x *Campaigns) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/email/campaigns", params)
}
func (x *Campaigns) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns", nil, body, "")
}
func (x *Campaigns) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id), nil, nil, "")
}
func (x *Campaigns) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/campaigns/"+enc(id), nil, body, "")
}
func (x *Campaigns) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/campaigns/"+enc(id), nil, nil, "")
}
// Send a draft: it arms its saved scheduled_at while that is still ahead,
// otherwise it sends NOW. SendWith takes a body to choose the time.
func (x *Campaigns) Send(ctx context.Context, id string) (any, error) {
	return x.SendWith(ctx, id, nil)
}

// SendWith sends a draft with a body: map[string]any{"scheduled_at":
// "<ISO 8601>"} schedules it, {"scheduled_at": nil} sends now, nil = Send.
func (x *Campaigns) SendWith(ctx context.Context, id string, body any) (any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/send", nil, body, "")
}
func (x *Campaigns) Cancel(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/cancel", nil, nil, "")
}
func (x *Campaigns) Pause(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/pause", nil, nil, "")
}
func (x *Campaigns) Resume(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/resume", nil, nil, "")
}

// Unschedule moves a scheduled campaign back to draft so it can be edited; it
// clears its send time.
func (x *Campaigns) Unschedule(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/unschedule", nil, nil, "")
}

// Duplicate copies a campaign into a new draft (201). body is optional
// (map[string]any{"name": "..."}, or nil). Pass WithIdempotencyKey so a retry
// returns the same copy; without it the SDK mints a key per call.
func (x *Campaigns) Duplicate(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/duplicate", nil, body, idempotencyKeyFrom(opts))
}
func (x *Campaigns) Reschedule(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/reschedule", nil, body, "")
}
func (x *Campaigns) Validate(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id)+"/validate", nil, nil, "")
}

// Stats returns the campaign's results: delivered, unique opens and clicks,
// unsubscribes, bounces, complaints and their rates. Every rate is null until
// something was sent; open and click rates are null when that tracking was off.
func (x *Campaigns) Stats(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id)+"/stats", nil, nil, "")
}

// ListStats returns Stats for up to 100 campaigns in one read (1–100 ids). An
// id that isn't one of this workspace's email campaigns is left out of data,
// never an error; has_more is always false.
func (x *Campaigns) ListStats(ctx context.Context, ids []string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/stats", url.Values{"ids": {strings.Join(ids, ",")}}, nil, "")
}

// ListRecipients lists who the campaign would reach if sent now, member by
// member: sendable, missing_data, suppressed or excluded, with the reason.
// params: "status" (default all), "limit" (1–100, default 50), "after". Pass
// the page's next_cursor as "after" for the next page (rows carry no id). A
// page can hold fewer rows than limit while has_more is true.
func (x *Campaigns) ListRecipients(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id)+"/recipients", q(params), nil, "")
}

// ListAffected lists, after an automatic quality pause, the recipients whose
// message hard-bounced or drew a spam complaint. They're already suppressed;
// this is for cleaning the source list. Up to 10,000; has_more says it was
// cut there.
func (x *Campaigns) ListAffected(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id)+"/quality/affected", nil, nil, "")
}

// RemoveAffected takes, after an automatic quality pause, the flagged contacts
// out of the campaign's static segment (needs segments:write too; a filter
// segment is a 409). A repeat removes nothing. It doesn't resume the campaign.
func (x *Campaigns) RemoveAffected(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/quality/remove-affected", nil, nil, "")
}

// RequestReview asks for the review a campaign needs before it can resume
// after a repeat quality pause (Resume answers 409 review_required until
// then). A repeat answers already_requested: true and notifies no one.
func (x *Campaigns) RequestReview(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/quality/request-review", nil, nil, "")
}

// Templates are reusable email designs.
type Templates struct{ http *httpClient }

func (x *Templates) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/templates", q(params), nil, "")
}

// ListAll iterates every row of /email/templates across pages (see Pager).
func (x *Templates) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/email/templates", params)
}
func (x *Templates) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/templates", nil, body, "")
}
func (x *Templates) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/templates/"+enc(id), nil, nil, "")
}
func (x *Templates) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/templates/"+enc(id), nil, body, "")
}
func (x *Templates) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/templates/"+enc(id), nil, nil, "")
}

// TestSend sends ONE test of the template to body["to"] with sample variables.
// Once a sending domain is verified it may go to anyone; until then only to a
// member of your workspace.
func (x *Templates) TestSend(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/templates/"+enc(id)+"/test-send", nil, body, "")
}

// Render returns the "subject", "html" and "text" a send of this template
// would carry, with your footer and branding, filled from "variables" (and
// from a contact's fields with "contact_id"). body is optional:
// map[string]any{"variables": ..., "subject": ..., "contact_id": ...}, or nil
// to render with no variables. It sends and stores nothing, so it takes no
// idempotency key. "missing_variables" lists the variables left without a
// value: they render as a literal {{token}}, and a real send would refuse
// them. Needs templates:read, and has its own per-workspace rate limit.
func (x *Templates) Render(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/templates/"+enc(id)+"/render", nil, body, "")
}

// Duplicate copies a template into a new draft (201): its name (with
// " (copy)", or your "name"), subject, editor and body carry over; its status,
// versions and usage don't. body is optional (map[string]any{"name": "..."},
// or nil). Pass WithIdempotencyKey so a retry returns the same copy; without
// it the SDK mints a key per call.
func (x *Templates) Duplicate(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/templates/"+enc(id)+"/duplicate", nil, body, idempotencyKeyFrom(opts))
}

// FromLibrary adds one of the gallery's ready-made emails
// (unitpost.com/templates) to your workspace as a draft Design template
// (201). body: map[string]any{"slug": "...", "name": "..."} ("name" is
// optional); the ones marked for paid plans answer 403 limit_reached on the
// free plan. Pass WithIdempotencyKey so a retry returns the same template;
// without it the SDK mints a key per call.
func (x *Templates) FromLibrary(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/templates/from-library", nil, body, idempotencyKeyFrom(opts))
}

// Usage says what uses the template, and whether a change would be accepted
// now: the campaigns that reference it, its live campaigns and automations,
// queued sends, and can_edit_content / can_unpublish with the reason when
// not. Read it before an edit or a delete: a content edit is refused while a
// live campaign renders the template, and a delete while any campaign uses
// it.
func (x *Templates) Usage(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/templates/"+enc(id)+"/usage", nil, nil, "")
}

// ListVersions lists the template's history, newest first: a version is
// added on every publish and on every content save while it's published.
// params: "limit" (1–100, default 20), "after". It pages by version number:
// pass the last row's version (an integer, written as a string) as "after",
// not an id. There is no "before".
func (x *Templates) ListVersions(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/templates/"+enc(id)+"/versions", q(params), nil, "")
}

// RestoreVersion puts a version's subject and body back, as the editor's
// History → Restore does. It is an update, so a system template, or one a
// live campaign renders, is refused with a 409 naming why (Usage says so up
// front). On a published template the restore adds a new version; history is
// never rewritten. It sends no body.
func (x *Templates) RestoreVersion(ctx context.Context, id string, version int) (any, error) {
	v := strconv.Itoa(version)
	return x.http.request(ctx, "POST", "/email/templates/"+enc(id)+"/versions/"+enc(v)+"/restore", nil, nil, "")
}

// EmailThreads are inbox conversations (/email/threads, thr_...): each received
// message and the replies and forwards sent from it are one conversation, as
// Emails → Received shows them.
type EmailThreads struct{ http *httpClient }

// List lists every conversation (Spam included), newest activity first.
// params: "q" (searches the subject, the sender and the opening words),
// "limit" (1–100, default 50), "after". Pass the page's next_cursor as "after"
// for the next page (not the last row's id).
func (x *EmailThreads) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/threads", q(params), nil, "")
}

// Get returns one conversation, oldest message first: what you received (with
// bodies) and what you sent from it (status and recipients, no body).
func (x *EmailThreads) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/threads/"+enc(id), nil, nil, "")
}

// Delete deletes a conversation for good: its received messages are erased;
// replies and forwards sent from it stay on your sent emails. It can't be
// undone.
func (x *EmailThreads) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/threads/"+enc(id), nil, nil, "")
}

// Reply replies, replies all or forwards one received message of the
// conversation (received_email_id): your text/html go on top and the original
// is quoted under them. It sends now and marks the conversation read. It sends
// real email, so pass WithIdempotencyKey to make a retry safe; without it the
// SDK mints a key per call.
func (x *EmailThreads) Reply(ctx context.Context, id string, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/email/threads/"+enc(id)+"/reply", nil, body, idempotencyKeyFrom(opts))
}

// BrandKits are Brand Kit profiles (voice, colors, pinned /img/{id}
// graphics), and the imports that fill them from a website or the brand's own
// words. The reads (List, Get, GetImport) need templates:read; the writes and
// starting, accepting or canceling an import need branding:write, which no key
// preset grants: tick it on the key.
type BrandKits struct{ http *httpClient }

func (x *BrandKits) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/brand-kits", q(params), nil, "")
}
func (x *BrandKits) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/brand-kits/"+enc(id), nil, nil, "")
}

// Create creates an empty profile (201): body is {"name": "...",
// "make_default": true}; "name" is unique in the workspace. The first profile,
// or one with "make_default", becomes the default. At most five profiles (403
// limit_reached). Fill it with Update or Import.
func (x *BrandKits) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits", nil, body, "")
}

// Update edits a profile: only the fields you send change, a list replaces the
// list, and nil or "" clears a text field. "graphics" pins library images by
// "image_id" (img_…). "sync_visual" on the default profile also sets the email
// branding's accent from its colors.
func (x *BrandKits) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/brand-kits/"+enc(id), nil, body, "")
}

// Delete deletes a profile. When it was the default, the oldest remaining
// profile becomes the default. The last profile can't be deleted (409): Reset
// it instead. Its pinned images stay in the library.
func (x *BrandKits) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/brand-kits/"+enc(id), nil, nil, "")
}

// Reset clears a profile's content (voice, colors, graphics, pages), keeping
// its name and whether it's the default. It sends no body.
func (x *BrandKits) Reset(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits/"+enc(id)+"/reset", nil, nil, "")
}

// SetDefault makes a profile the default: the one agents write with unless
// told otherwise. It sends no body.
func (x *BrandKits) SetDefault(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits/"+enc(id)+"/set-default", nil, nil, "")
}

// AddGraphic adds a graphic (201) from a public https image "url" (fetched
// once, no redirects): body is {"url": "...", "role": "logo", "label": "..."}.
// It goes into the image library, counted against its storage, and is pinned
// with its "role" (default "logo"). To pin an image already in the library,
// Update the profile's "graphics".
func (x *BrandKits) AddGraphic(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits/"+enc(id)+"/graphics", nil, body, "")
}

// Import builds a profile from a website ("url", or a bare domain) or the
// brand's own words ("text"). It runs in the background, usually under a
// minute (202): follow it with GetImport. By default the result is applied
// when it lands, to "brand_kit_id", else the default profile, or a new one
// when the default already has content. With "review": true it stops at a
// proposal (preview_ready) for AcceptImport; one review import is open at a
// time, and a new one replaces it.
func (x *BrandKits) Import(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits/imports", nil, body, "")
}

// GetImport retrieves one import: its status, the proposal once it's ready,
// warnings and any error.
func (x *BrandKits) GetImport(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/brand-kits/imports/"+enc(id), nil, nil, "")
}

// AcceptImport applies a review import's proposal (preview_ready) to a
// profile: voice, colors and pages, and its logos and backgrounds brought into
// the image library and pinned. body is optional (nil sends {}):
// "brand_kit_id" applies it to another profile, "name" names it,
// "sync_visual" also sets the email branding's accent. To change other
// fields, accept and then Update.
func (x *BrandKits) AcceptImport(ctx context.Context, id string, body any) (any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return x.http.request(ctx, "POST", "/brand-kits/imports/"+enc(id)+"/accept", nil, body, "")
}

// CancelImport discards a review import's proposal or stops its crawl. It
// sends no body.
func (x *BrandKits) CancelImport(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/brand-kits/imports/"+enc(id)+"/cancel", nil, nil, "")
}

// Domains are sending domains.
type Domains struct{ http *httpClient }

func (x *Domains) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/domains", q(params), nil, "")
}

// ListAll iterates every row of /email/domains across pages (see Pager).
func (x *Domains) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/email/domains", params)
}
func (x *Domains) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/domains", nil, body, "")
}
func (x *Domains) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/domains/"+enc(id), nil, nil, "")
}
func (x *Domains) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/domains/"+enc(id), nil, nil, "")
}
func (x *Domains) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/domains/"+enc(id), nil, body, "")
}
func (x *Domains) Verify(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/domains/"+enc(id)+"/verify", nil, nil, "")
}

// Diagnose is a read-only DNS diagnosis: per-record answers and ranked findings with fixes.
func (x *Domains) Diagnose(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/domains/"+enc(id)+"/diagnose", nil, nil, "")
}

// Sender addresses (/email/senders, sndr_...): saved From addresses.

// ListSenders lists your saved From addresses, as the domain page lists them
// and the dashboard's From pickers offer them: each a mailbox on one of your
// domains, with its display name and "from", the value to send with. params:
// "domain_id" (one domain's), "limit", "after", "before". A sender can send
// once its domain can.
func (x *Domains) ListSenders(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/senders", q(params), nil, "")
}

// CreateSender saves a From address on one of your domains (201): body
// map[string]any{"address": "...", "display_name": "..."} ("display_name" is
// optional). The domain may still be verifying. A mailbox already saved on
// that domain is a 409, so a retry never saves it twice.
func (x *Domains) CreateSender(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/senders", nil, body, "")
}

// GetSender returns one saved From address.
func (x *Domains) GetSender(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/senders/"+enc(id), nil, nil, "")
}

// UpdateSender changes a sender's display name (map[string]any{"display_name":
// ...}; nil or "" clears it), the only field that changes: a different mailbox
// is a new sender.
func (x *Domains) UpdateSender(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/senders/"+enc(id), nil, body, "")
}

// DeleteSender removes a saved From address from the pickers. Templates and
// campaigns that send from it keep their From.
func (x *Domains) DeleteSender(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/email/senders/"+enc(id), nil, nil, "")
}

// Webhooks manage event endpoints and verify deliveries.
type Webhooks struct{ http *httpClient }

func (x *Webhooks) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/webhooks", q(params), nil, "")
}

// ListAll iterates every row of /webhooks across pages (see Pager).
func (x *Webhooks) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/webhooks", params)
}
func (x *Webhooks) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/webhooks", nil, body, "")
}
func (x *Webhooks) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/webhooks/"+enc(id), nil, nil, "")
}
func (x *Webhooks) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/webhooks/"+enc(id), nil, body, "")
}
func (x *Webhooks) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/webhooks/"+enc(id), nil, nil, "")
}
func (x *Webhooks) Test(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/webhooks/"+enc(id)+"/test", nil, nil, "")
}

// RotateSecret mints a new signing secret, returned once; the old one co-signs for 24 h.
func (x *Webhooks) RotateSecret(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/webhooks/"+enc(id)+"/rotate-secret", nil, nil, "")
}

// ListDeliveries is the endpoint's delivery log, newest first
// (cursor-paginated). params: "limit", "after", "before", "status" (one or
// more of pending, queued, retrying, succeeded, failed, exhausted,
// comma-separated, e.g. "failed,exhausted") and "event_type" (e.g.
// "email.bounced"). The log ignored these filters before, so a value it can't
// apply is ignored and named in the Unitpost-Warning response header;
// Unitpost-Strict: true refuses it instead.
func (x *Webhooks) ListDeliveries(ctx context.Context, id string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/webhooks/"+enc(id)+"/deliveries", q(params), nil, "")
}

// ListAllDeliveries iterates every delivery of the endpoint (see Pager), with
// the same filters as ListDeliveries.
func (x *Webhooks) ListAllDeliveries(id string, params map[string]string) *Pager {
	return newPager(x.http, "/webhooks/"+enc(id)+"/deliveries", params)
}

// ReplayDelivery delivers one event to the endpoint again (202), as the
// dashboard's Replay does. The same event goes out with the same id, so a
// receiver that dedupes on it treats the replay as the delivery it already
// has. The webhook must be enabled: a disabled or suspended one is a 409
// (re-enable it with Update first). Find deliveries with ListDeliveries. It
// sends no body.
func (x *Webhooks) ReplayDelivery(ctx context.Context, id, deliveryID string) (any, error) {
	return x.http.request(ctx, "POST", "/webhooks/"+enc(id)+"/deliveries/"+enc(deliveryID)+"/replay", nil, nil, "")
}

// Flush delivers the backlog (202): the deliveries held while the webhook was
// suspended ("queued"), as the dashboard's "Retry queued" does. body is
// optional: map[string]any{"include_failed": true} also delivers the ones that
// failed for good ("failed", "exhausted"); nil delivers the queued ones only.
// The webhook must be enabled: a disabled or suspended one is a 409 (re-enable
// it with Update first); re-enabling never delivers the backlog on its own.
// "flushed" counts the deliveries handed back.
func (x *Webhooks) Flush(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/webhooks/"+enc(id)+"/flush", nil, body, "")
}
func (x *Webhooks) Verify(payload, secret string, headers http.Header, toleranceSeconds int) (any, error) {
	return VerifyWebhook(payload, secret, headers, toleranceSeconds)
}

// APIKeys manages workspace API keys.
type APIKeys struct{ http *httpClient }

// List returns the workspace's API keys (prefix, scope, capabilities, last
// use; never the secret). With an API key or an MCP sign-in it needs
// apikeys:read, which no key preset includes.
func (x *APIKeys) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/api-keys", q(params), nil, "")
}

// ListAll iterates every row of /api-keys across pages (see Pager).
func (x *APIKeys) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/api-keys", params)
}
func (x *APIKeys) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/api-keys", nil, body, "")
}
func (x *APIKeys) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/api-keys/"+enc(id), nil, nil, "")
}

// Usage is one key's last 24 hours: how many requests it made, how many were
// answered 4xx, 5xx or 429, and when it was last used. Check it before you
// revoke a key (is anything still calling with it?) or when an integration's
// calls fail. Needs apikeys:read, which no key preset includes.
func (x *APIKeys) Usage(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/api-keys/"+enc(id)+"/usage", nil, nil, "")
}

// Logs is the workspace's API request log: one row per call. It needs
// apikeys:read, which no key preset includes.
type Logs struct{ http *httpClient }

// List returns your workspace's API request log, newest first
// (cursor-paginated, rlog_ ids): each call's method, path, status, time taken
// (duration_ms), the API key that made it (api_key_id) and its x-request-id
// (request_id: quote it to support). Never the caller's IP and never a body.
// params: "limit", "after", "before", "status" (a code such as "429", or a
// class: "2xx", "3xx", "4xx" or "5xx"), "api_key_id" (key_…), "method" (GET,
// POST, PUT, PATCH or DELETE), "path_prefix" (e.g. "/email/campaigns", at most
// 200 characters), "request_id" (the x-request-id a response carried),
// "created_after" and "created_before" (ISO 8601 with Z or an offset); an
// unknown key is a 422. Rows are kept 30 days. Needs apikeys:read, which no
// key preset includes, and has its own per-workspace rate limit.
func (x *Logs) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/logs", q(params), nil, "")
}

// Suppressions is the suppression list.
type Suppressions struct{ http *httpClient }

// List returns your workspace's suppressed addresses, newest first
// (cursor-paginated). params: "limit", "after", "before", "q" (part of the
// address, at most 200 characters), "reason" (bounce, complaint, unsubscribe,
// manual, import or api), "created_after", "created_before" (ISO 8601 with Z
// or an offset). "scope" is deprecated and ignored. The list ignored these
// filters before, so a value it can't apply is ignored and named in the
// Unitpost-Warning response header; Unitpost-Strict: true refuses it instead.
func (x *Suppressions) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/suppressions", q(params), nil, "")
}

// ListAll iterates every row of /suppressions across pages (see Pager), with
// the same filters as List.
func (x *Suppressions) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/suppressions", params)
}
func (x *Suppressions) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/suppressions", nil, body, "")
}
func (x *Suppressions) Get(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "GET", "/suppressions/"+enc(idOrEmail), nil, nil, "")
}
func (x *Suppressions) Delete(ctx context.Context, idOrEmail string) (any, error) {
	return x.http.request(ctx, "DELETE", "/suppressions/"+enc(idOrEmail), nil, nil, "")
}

// LibraryImages is the workspace's image library: the images a template
// embeds. Each image's "path" (/img/…) is what a template's <Image src> (or
// <img src>) takes; "url" is its public address. Reads need templates:read,
// writes templates:write.
type LibraryImages struct{ http *httpClient }

// List lists the library, newest first (cursor-paginated). params: "limit",
// "after", "before".
func (x *LibraryImages) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/library-images", q(params), nil, "")
}

// ListAll iterates every row of /library-images across pages (see Pager).
func (x *LibraryImages) ListAll(params map[string]string) *Pager {
	return newPager(x.http, "/library-images", params)
}

// Create adds an image (201) from a public https "url" (fetched once, no
// redirects, no private addresses) or base64 "data" with its "content_type":
// PNG, JPEG, GIF or WebP up to 5 MB, counted against the plan's image storage
// (403 limit_reached). Optional "filename" and "alt". Pass WithIdempotencyKey
// so a retried upload stores one image; without it the SDK mints a key per
// call.
func (x *LibraryImages) Create(ctx context.Context, body any, opts ...RequestOption) (any, error) {
	return x.http.request(ctx, "POST", "/library-images", nil, body, idempotencyKeyFrom(opts))
}
func (x *LibraryImages) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/library-images/"+enc(id), nil, nil, "")
}

// Update changes the alt text the image suggests when it's inserted
// (map[string]any{"alt": "..."}; nil or "" clears it).
func (x *LibraryImages) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/library-images/"+enc(id), nil, body, "")
}

// Delete removes an image from the library. While a template, a campaign or
// sent mail uses it, the answer is a 409 resource_in_use naming them
// ("details"); ForceDelete deletes it anyway.
func (x *LibraryImages) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/library-images/"+enc(id), nil, nil, "")
}

// ForceDelete removes an image even while something uses it (?force=true):
// those emails then show a broken image wherever it isn't cached yet.
func (x *LibraryImages) ForceDelete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/library-images/"+enc(id), url.Values{"force": {"true"}}, nil, "")
}

// Branding is the look marketing emails' footer and the hosted unsubscribe
// page share, as Settings → Brand shows it. Reading it needs templates:read;
// the writes need branding:write, which no key preset grants: tick it on the
// key.
type Branding struct{ http *httpClient }

// Get returns the colors, font, company name and postal address, the
// unsubscribe page, the footer and the logo; null means the default.
func (x *Branding) Get(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/branding", nil, nil, "")
}

// Update changes branding: only the fields you send change, and null restores
// a default. Hiding the "Sent with Unitpost" mark needs a paid plan (403
// limit_reached).
func (x *Branding) Update(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/branding", nil, body, "")
}

// SetLogo replaces the logo, from a public https "url" or base64 "data" (PNG,
// JPEG, GIF or WebP, up to 5 MB).
func (x *Branding) SetLogo(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "PUT", "/branding/logo", nil, body, "")
}

// RemoveLogo clears the brand logo (the workspace image shows instead, else
// none). It sends no body and answers the branding.
func (x *Branding) RemoveLogo(ctx context.Context) (any, error) {
	return x.http.request(ctx, "DELETE", "/branding/logo", nil, nil, "")
}

// Workspace is the calling credential (Whoami, which any valid credential may
// read), the activity log (ListActivity, which needs activity:read), and the
// workspace's name and image. Update, SetImage and RemoveImage need
// branding:write. No key preset grants either capability: tick it on the key.
// Deleting the workspace stays in the dashboard.
type Workspace struct{ http *httpClient }

// Whoami says who is calling: the workspace, the credential (an API key or an
// OAuth sign-in, never its secret), its "capabilities" and the workspace's
// rate limit. Call it first to see what the credential may do: a call that
// needs a capability it lacks answers 403 insufficient_scope. It needs no
// capability: any valid credential may read it.
func (x *Workspace) Whoami(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/whoami", nil, nil, "")
}

// ListActivity is the workspace activity log, newest first (cursor-paginated,
// act_ ids): what changed, when, who changed it (actor.name: a teammate,
// System, Ask Unit) and, for a change made over the API, MCP or Ask Unit, how
// (actor.via, with the API key). The rows the dashboard's Activity page and
// each record's timeline show. params: "limit", "after", "before", "subject"
// (contact, segment, topic, field, template, component, campaign, email,
// domain, sender_email, library_image, api_key, webhook, suppression,
// message, handle or workspace), "subject_id" (with "subject": one record's
// changes), "contact_id", "campaign_id", "q" (part of the record's name or the
// actor's, at most 200 characters) and "channel" ("email": without the SMS
// rows); an unknown key is a 422. Needs activity:read, which no key preset
// includes: the log names your teammates.
func (x *Workspace) ListActivity(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/activity", q(params), nil, "")
}

// ListAllActivity iterates every activity row (see Pager), with the same
// filters as ListActivity.
func (x *Workspace) ListAllActivity(params map[string]string) *Pager {
	return newPager(x.http, "/activity", params)
}

// Update renames the workspace (map[string]any{"name": "..."}).
func (x *Workspace) Update(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/workspace", nil, body, "")
}

// SetImage replaces the workspace image (its avatar in the dashboard, and the
// email logo when there's no brand logo), from a public https "url" or base64
// "data" (PNG, JPEG, GIF or WebP, up to 5 MB).
func (x *Workspace) SetImage(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "PUT", "/workspace/image", nil, body, "")
}

// RemoveImage clears the workspace image. It sends no body and answers the
// workspace.
func (x *Workspace) RemoveImage(ctx context.Context) (any, error) {
	return x.http.request(ctx, "DELETE", "/workspace/image", nil, nil, "")
}

// Usage is the billing-period snapshot.
type Usage struct{ http *httpClient }

func (x *Usage) Get(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/usage", nil, nil, "")
}
