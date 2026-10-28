package unitpost

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unitpostcom/unitpost-go/generated"
)

const (
	SDKVersion     = "0.4.0"
	DefaultBaseURL = "https://www.unitpost.com"
)

// Error is a structured API error. Returned as the error value of resource
// methods — never a panic for an API-level failure.
type Error struct {
	Code      string
	Message   string
	Status    int
	RequestID string
	Details   []map[string]any
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

type httpClient struct {
	apiKey         string
	baseURL        string
	http           *http.Client
	maxRetries     int
	retryBaseDelay time.Duration
	maxRetryDelay  time.Duration
}

func newHTTP(opts Options) *httpClient {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("UNITPOST_API_KEY")
	}
	if apiKey == "" {
		panic(&Error{Code: "missing_api_key", Message: "No API key provided. Pass APIKey or set UNITPOST_API_KEY."})
	}
	base := opts.BaseURL
	if base == "" {
		base = os.Getenv("UNITPOST_BASE_URL")
	}
	if base == "" {
		base = DefaultBaseURL
	}
	base = strings.TrimRight(base, "/")
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	maxRetries := 2
	if opts.MaxRetriesSet {
		maxRetries = opts.MaxRetries
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}
	return &httpClient{
		apiKey:         apiKey,
		baseURL:        base,
		http:           hc,
		maxRetries:     maxRetries,
		retryBaseDelay: 500 * time.Millisecond,
		maxRetryDelay:  20 * time.Second,
	}
}

func (c *httpClient) request(ctx context.Context, method, path string, query url.Values, body any, idempotencyKey string) (any, error) {
	retrySafe := method == http.MethodGet || idempotencyKey != ""
	var lastErr error
	for attempt := 0; ; attempt++ {
		data, err, retryAfter := c.attempt(ctx, method, path, query, body, idempotencyKey)
		if err == nil {
			return data, nil
		}
		lastErr = err
		ue, ok := err.(*Error)
		retryable := ok && (ue.Status == 0 || ue.Status == 429 || (ue.Status >= 500 && ue.Status <= 599))
		if !retryable || !retrySafe || attempt >= c.maxRetries {
			return nil, lastErr
		}
		delay := c.backoff(attempt, retryAfter)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
}

func (c *httpClient) attempt(ctx context.Context, method, path string, query url.Values, body any, idempotencyKey string) (any, error, time.Duration) {
	u := c.baseURL + "/api/" + generated.APIVersion + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var payload io.Reader
	if body != nil && method != http.MethodGet {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{Code: "encode_error", Message: err.Error()}, noRetryAfter
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, payload)
	if err != nil {
		return nil, &Error{Code: "network_error", Message: err.Error()}, noRetryAfter
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "unitpost-go/"+SDKVersion)
	req.Header.Set("Accept", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, &Error{Code: "network_error", Message: err.Error()}, noRetryAfter
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	retryAfter := noRetryAfter
	if d, ok := parseRetryAfter(res.Header.Get("Retry-After"), time.Now()); ok {
		retryAfter = d
	}
	if res.StatusCode == http.StatusNoContent {
		return nil, nil, noRetryAfter
	}
	var parsed any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		code, msg := "http_error", fmt.Sprintf("Request failed with status %d.", res.StatusCode)
		if m, ok := parsed.(map[string]any); ok {
			if e, ok := m["error"].(map[string]any); ok {
				if s, ok := e["code"].(string); ok {
					code = s
				}
				if s, ok := e["message"].(string); ok {
					msg = s
				}
			}
		}
		return nil, &Error{Code: code, Message: msg, Status: res.StatusCode, RequestID: res.Header.Get("X-Request-Id")}, retryAfter
	}
	return parsed, nil, noRetryAfter
}

// noRetryAfter marks a response that carried no usable Retry-After header.
// It is negative so that "Retry-After: 0" (retry now) stays distinguishable.
const noRetryAfter time.Duration = -1

// parseRetryAfter reads a Retry-After header: delta-seconds ("3", "1.5") or an
// HTTP date ("Wed, 21 Oct 2026 07:28:00 GMT"). It returns how long to wait from
// now, and false when the header is absent or unreadable. A date in the past
// means retry now. backoff clamps the result to maxRetryDelay.
func parseRetryAfter(v string, now time.Time) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.ParseFloat(v, 64); err == nil {
		if math.IsNaN(secs) || math.IsInf(secs, 0) {
			return 0, false
		}
		if secs <= 0 {
			return 0, true
		}
		if secs >= float64(math.MaxInt64)/float64(time.Second) {
			return time.Duration(math.MaxInt64), true
		}
		return time.Duration(secs * float64(time.Second)), true
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := t.Sub(now); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

// backoff is the wait before the next retry: the server's Retry-After when it
// sent one, else exponential backoff (retryBaseDelay * 2^attempt) with full
// jitter. Both are capped at maxRetryDelay.
func (c *httpClient) backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter >= 0 {
		if retryAfter > c.maxRetryDelay {
			return c.maxRetryDelay
		}
		return retryAfter
	}
	ceiling := c.retryBaseDelay * time.Duration(1<<attempt)
	if ceiling > c.maxRetryDelay {
		ceiling = c.maxRetryDelay
	}
	if ceiling <= 0 {
		return 0
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ceiling
	}
	n := binary.BigEndian.Uint64(b[:]) % uint64(ceiling+1)
	return time.Duration(n)
}

func enc(s string) string { return url.PathEscape(s) }

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
