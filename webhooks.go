package unitpost

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WebhookVerificationError is raised when a delivery's signature is missing,
// invalid, or outside the timestamp tolerance window.
type WebhookVerificationError struct{ Msg string }

func (e *WebhookVerificationError) Error() string { return e.Msg }

// VerifyWebhook checks a svix-compatible signature and returns the parsed event.
func VerifyWebhook(payload string, secret string, headers http.Header, toleranceSeconds int) (any, error) {
	if toleranceSeconds == 0 {
		toleranceSeconds = 300
	}
	id := headers.Get("svix-id")
	timestamp := headers.Get("svix-timestamp")
	signature := headers.Get("svix-signature")
	if id == "" || timestamp == "" || signature == "" {
		return nil, &WebhookVerificationError{Msg: "Missing svix-id, svix-timestamp, or svix-signature header."}
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, &WebhookVerificationError{Msg: "Invalid svix-timestamp header."}
	}
	now := time.Now().Unix()
	if abs(now-ts) > int64(toleranceSeconds) {
		return nil, &WebhookVerificationError{Msg: "Webhook timestamp is outside the tolerance window (possible replay)."}
	}
	key, err := secretKeyBytes(secret)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, key)
	fmt.Fprintf(mac, "%s.%d.%s", id, ts, payload)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	matched := false
	for _, part := range strings.Fields(signature) {
		_, sig, ok := strings.Cut(part, ",")
		if !ok {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(sig)) == 1 {
			matched = true
			break
		}
	}
	if !matched {
		return nil, &WebhookVerificationError{Msg: "Webhook signature verification failed."}
	}
	var parsed any
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return nil, &WebhookVerificationError{Msg: "Webhook payload is not valid JSON."}
	}
	return parsed, nil
}

func secretKeyBytes(secret string) ([]byte, error) {
	raw := strings.TrimPrefix(secret, "whsec_")
	if pad := (4 - len(raw)%4) % 4; pad != 0 {
		raw += strings.Repeat("=", pad)
	}
	b, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		b, err = base64.RawURLEncoding.DecodeString(strings.TrimRight(raw, "="))
	}
	if err != nil {
		return nil, &WebhookVerificationError{Msg: "Invalid webhook secret."}
	}
	return b, nil
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
