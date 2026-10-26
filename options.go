package unitpost

// RequestOption tunes one call. Pass it as the trailing variadic argument of a
// write that supports it (Email.Send, Email.Batch, Sms.Send).
type RequestOption func(*requestOptions)

type requestOptions struct {
	idempotencyKey string
}

// WithIdempotencyKey sets the Idempotency-Key header. Reuse the same key when you
// retry the same logical send (for example after a crash or timeout in your own
// code): the API replays the first result instead of sending twice. When you
// omit it, the SDK mints a fresh key per call, which still makes the SDK's own
// automatic retries safe.
func WithIdempotencyKey(key string) RequestOption {
	return func(o *requestOptions) { o.idempotencyKey = key }
}

// idempotencyKeyFrom returns the caller's key, or a fresh one when none was given.
func idempotencyKeyFrom(opts []RequestOption) string {
	var o requestOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.idempotencyKey != "" {
		return o.idempotencyKey
	}
	return newID()
}
