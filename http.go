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
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unitpostcom/unitpost-go/generated"
)

const (
	SDKVersion     = "0.3.0"
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
			return nil, &Error{Code: "encode_error", Message: err.Error()}, 0
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, payload)
	if err != nil {
		return nil, &Error{Code: "network_error", Message: err.Error()}, 0
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
		return nil, &Error{Code: "network_error", Message: err.Error()}, 0
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var retryAfter time.Duration
	if v := res.Header.Get("Retry-After"); v != "" {
		if secs, convErr := strconv.Atoi(v); convErr == nil {
			retryAfter = time.Duration(secs) * time.Second
		}
	}
	if res.StatusCode == http.StatusNoContent {
		return nil, nil, 0
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
	return parsed, nil, 0
}

func (c *httpClient) backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
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
