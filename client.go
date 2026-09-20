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
	// Sms is the SMS channel (beta, behind the launch gate). Wired in its GA
	// shape; while a workspace's gate is off every call returns the same 404
	// the REST surface does.
	Sms           *Sms
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
		Sms:           &Sms{http: h},
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
