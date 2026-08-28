package unitpost

import (
	"net/http"
	"time"
)

// Options configure a Client.
type Options struct {
	APIKey        string
	BaseURL       string
	Timeout       time.Duration
	HTTPClient    *http.Client
	MaxRetries    int
	MaxRetriesSet bool
}

// Client is the Unitpost API client. Construct once, reuse everywhere.
type Client struct {
	Email         *Email
	// NOTE (pre-launch): Sms is intentionally absent while SMS is unpublished.
	Contacts      *Contacts
	ContactFields *ContactFields
	Segments     *Segments
	BrandKits    *BrandKits
	Webhooks     *Webhooks
	ApiKeys      *APIKeys
	Suppressions *Suppressions
	Usage        *Usage
}

// New reads UNITPOST_API_KEY from the environment.
func New() *Client { return NewWithOptions(Options{}) }

// NewWithOptions constructs a client with explicit options.
func NewWithOptions(opts Options) *Client {
	h := newHTTP(opts)
	return &Client{
		Email:         newEmail(h),
		Contacts:      &Contacts{http: h},
		ContactFields: &ContactFields{http: h},
		Segments:      &Segments{http: h},
		BrandKits:     &BrandKits{http: h},
		Webhooks:      &Webhooks{http: h},
		ApiKeys:       &APIKeys{http: h},
		Suppressions:  &Suppressions{http: h},
		Usage:         &Usage{http: h},
	}
}
