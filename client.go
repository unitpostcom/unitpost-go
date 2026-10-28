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
	// Sms is the SMS channel. Available on paid plans: on the free plan Send
	// returns a 403 with code limit_reached; reads keep working.
	Sms           *Sms
	Contacts      *Contacts
	ContactFields *ContactFields
	Segments     *Segments
	BrandKits    *BrandKits
	// LibraryImages is the image library templates embed (/img/… paths).
	LibraryImages *LibraryImages
	// Branding is the email footer and unsubscribe page look; its writes need
	// branding:write (on no key preset).
	Branding *Branding
	// Workspace is who is calling (Whoami, any credential), the activity log
	// (ListActivity, needs activity:read, on no key preset), and the
	// workspace's name and image; the writes need branding:write (on no key
	// preset).
	Workspace    *Workspace
	Webhooks     *Webhooks
	ApiKeys      *APIKeys
	Logs         *Logs // the API request log (List); needs apikeys:read, on no key preset
	Suppressions *Suppressions
	Jobs         *Jobs
	Usage        *Usage
}

// New reads UNITPOST_API_KEY from the environment.
func New() *Client { return NewWithOptions(Options{}) }

// NewWithOptions constructs a client with explicit options.
func NewWithOptions(opts Options) *Client {
	h := newHTTP(opts)
	return &Client{
		Email:         newEmail(h),
		Sms:           newSms(h),
		Contacts:      &Contacts{http: h},
		ContactFields: &ContactFields{http: h},
		Segments:      &Segments{http: h},
		BrandKits:     &BrandKits{http: h},
		LibraryImages: &LibraryImages{http: h},
		Branding:      &Branding{http: h},
		Workspace:     &Workspace{http: h},
		Webhooks:      &Webhooks{http: h},
		ApiKeys:       &APIKeys{http: h},
		Logs:          &Logs{http: h},
		Suppressions:  &Suppressions{http: h},
		Jobs:          &Jobs{http: h},
		Usage:         &Usage{http: h},
	}
}
