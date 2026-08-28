package unitpost

import (
	"context"
	"net/http"
	"net/url"
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

// Sms is the SMS channel (beta).
type Sms struct{ http *httpClient }

func (x *Sms) Send(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/sms", nil, body, newID())
}
func (x *Sms) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/sms/"+enc(id), nil, nil, "")
}
func (x *Sms) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/sms", q(params), nil, "")
}

// Email is the channel root: send, batch, templates, campaigns, topics, domains.
type Email struct {
	http      *httpClient
	Topics    *Topics
	Campaigns *Campaigns
	Templates *Templates
	Domains   *Domains
}

func newEmail(h *httpClient) *Email {
	return &Email{http: h, Topics: &Topics{http: h}, Campaigns: &Campaigns{http: h}, Templates: &Templates{http: h}, Domains: &Domains{http: h}}
}

func (x *Email) Send(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email", nil, body, newID())
}
func (x *Email) Batch(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/batch", nil, body, newID())
}
func (x *Email) GetBatch(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/batches/"+enc(id), nil, nil, "")
}
func (x *Email) CancelBatch(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/batches/"+enc(id)+"/cancel", nil, nil, "")
}
func (x *Email) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email", q(params), nil, "")
}
func (x *Email) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/"+enc(id), nil, nil, "")
}
func (x *Email) Update(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "PATCH", "/email/"+enc(id), nil, body, "")
}
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

// Contacts manages workspace contacts.
type Contacts struct{ http *httpClient }

func (x *Contacts) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts", q(params), nil, "")
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

// ContactFields are custom contact attributes.
type ContactFields struct{ http *httpClient }

func (x *ContactFields) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contact-fields", q(params), nil, "")
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
type Segments struct{ http *httpClient }

func (x *Segments) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/segments", q(params), nil, "")
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

// Topics are email subscription topics.
type Topics struct{ http *httpClient }

func (x *Topics) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/topics", q(params), nil, "")
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
func (x *Topics) ListTopics(ctx context.Context, contactID string, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/contacts/"+enc(contactID)+"/topics", q(params), nil, "")
}
func (x *Topics) SetTopic(ctx context.Context, contactID string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/contacts/"+enc(contactID)+"/topics", nil, body, "")
}

// Campaigns are marketing broadcasts.
type Campaigns struct{ http *httpClient }

func (x *Campaigns) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns", q(params), nil, "")
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
func (x *Campaigns) Send(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/send", nil, nil, "")
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
func (x *Campaigns) Reschedule(ctx context.Context, id string, body any) (any, error) {
	return x.http.request(ctx, "POST", "/email/campaigns/"+enc(id)+"/reschedule", nil, body, "")
}
func (x *Campaigns) Validate(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/email/campaigns/"+enc(id)+"/validate", nil, nil, "")
}

// Templates are reusable email designs.
type Templates struct{ http *httpClient }

func (x *Templates) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/templates", q(params), nil, "")
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

// BrandKits are read-only brand profiles.
type BrandKits struct{ http *httpClient }

func (x *BrandKits) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/brand-kits", q(params), nil, "")
}
func (x *BrandKits) Get(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "GET", "/brand-kits/"+enc(id), nil, nil, "")
}

// Domains are sending domains.
type Domains struct{ http *httpClient }

func (x *Domains) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/email/domains", q(params), nil, "")
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

// Webhooks manage event endpoints and verify deliveries.
type Webhooks struct{ http *httpClient }

func (x *Webhooks) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/webhooks", q(params), nil, "")
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
func (x *Webhooks) Verify(payload, secret string, headers http.Header, toleranceSeconds int) (any, error) {
	return VerifyWebhook(payload, secret, headers, toleranceSeconds)
}

// APIKeys manages workspace API keys.
type APIKeys struct{ http *httpClient }

func (x *APIKeys) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/api-keys", q(params), nil, "")
}
func (x *APIKeys) Create(ctx context.Context, body any) (any, error) {
	return x.http.request(ctx, "POST", "/api-keys", nil, body, "")
}
func (x *APIKeys) Delete(ctx context.Context, id string) (any, error) {
	return x.http.request(ctx, "DELETE", "/api-keys/"+enc(id), nil, nil, "")
}

// Suppressions is the suppression list.
type Suppressions struct{ http *httpClient }

func (x *Suppressions) List(ctx context.Context, params map[string]string) (any, error) {
	return x.http.request(ctx, "GET", "/suppressions", q(params), nil, "")
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

// Usage is the billing-period snapshot.
type Usage struct{ http *httpClient }

func (x *Usage) Get(ctx context.Context) (any, error) {
	return x.http.request(ctx, "GET", "/usage", nil, nil, "")
}
