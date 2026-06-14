package proxmox

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RetryOption configures the retry behaviour set via WithRetry.
type RetryOption func(*retryPolicy)

// retryPolicy is the internal configuration for the retry RoundTripper.
type retryPolicy struct {
	maxAttempts    int
	initialBackoff time.Duration
	maxBackoff     time.Duration
	condition      func(*http.Response, error) bool
	// now and sleep are overridable for tests.
	now   func() time.Time
	sleep func(context.Context, time.Duration) error
	// rand is the jitter source; guarded by mu.
	mu   sync.Mutex
	rand *rand.Rand
}

const (
	defaultRetryMaxAttempts    = 3
	defaultRetryInitialBackoff = 200 * time.Millisecond
	defaultRetryMaxBackoff     = 5 * time.Second
)

// defaultRetryPolicy returns the policy used when WithRetry is called with no
// options. The shape — three attempts, 200ms-5s exponential backoff with full
// jitter, retries on idempotent verbs plus buffered-body POST, Retry-After
// honored on 429 — is documented on WithRetry.
func defaultRetryPolicy() *retryPolicy {
	return &retryPolicy{
		maxAttempts:    defaultRetryMaxAttempts,
		initialBackoff: defaultRetryInitialBackoff,
		maxBackoff:     defaultRetryMaxBackoff,
		condition:      defaultRetryCondition,
		now:            time.Now,
		sleep:          contextSleep,
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // backoff jitter is not security-sensitive
	}
}

// defaultRetryCondition retries on network errors, HTTP 502, 503, 504, and 429.
// 4xx responses other than 429 are surfaced immediately to the caller.
func defaultRetryCondition(res *http.Response, err error) bool {
	if err != nil {
		// Network-level error — connection refused, EOF, DNS, timeout, etc.
		// Context-cancellation errors are filtered out separately by the
		// caller so we don't retry a user-requested cancellation.
		return true
	}
	if res == nil {
		return false
	}
	switch res.StatusCode {
	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusTooManyRequests:
		return true
	}
	return false
}

// WithRetryMax sets the maximum number of attempts (including the first).
// Default 3. Values less than 1 are clamped to 1 (no retry).
func WithRetryMax(n int) RetryOption {
	return func(p *retryPolicy) {
		if n < 1 {
			n = 1
		}
		p.maxAttempts = n
	}
}

// WithRetryBackoff overrides the exponential backoff bounds. initial is the
// first backoff window (full-jitter sampled in [0, initial)); the window doubles
// per attempt and is capped at max. Defaults: 200ms initial, 5s max.
func WithRetryBackoff(initial, max time.Duration) RetryOption {
	return func(p *retryPolicy) {
		if initial > 0 {
			p.initialBackoff = initial
		}
		if max > 0 {
			p.maxBackoff = max
		}
	}
}

// WithRetryCondition replaces the predicate that decides whether a response or
// error should trigger another attempt. The function is called with the result
// of the inner RoundTripper; exactly one of res / err is non-nil. The default
// retries on net errors and HTTP 502, 503, 504, 429.
func WithRetryCondition(fn func(*http.Response, error) bool) RetryOption {
	return func(p *retryPolicy) {
		if fn != nil {
			p.condition = fn
		}
	}
}

// WithRetry installs a RoundTripper wrapper that retries transient failures
// on the underlying transport.
//
// Default policy: max 3 attempts, exponential backoff with full jitter from
// 200ms to 5s, retries on network errors and HTTP 502, 503, 504, 429. The
// Retry-After header on 429 or 503 overrides the computed backoff (capped at
// the configured max). Only idempotent verbs (GET, PUT, DELETE) and POST with
// a fully-buffered body are retried; in this client request bodies are always
// []byte, so POST is rewindable.
//
// Cumulative timeout is bounded by the request context; the wrapper respects
// ctx.Done() between attempts and returns the context error as soon as
// cancellation is observed.
//
// WithRetry composes with the other transport-touching options
// (WithInsecureSkipVerify, WithRootCAs, WithClientCertificate, WithProxy,
// WithProxyFromEnvironment, WithRequestInterceptor). It wraps whichever
// transport the client currently has when the option runs; if a subsequent
// WithHTTPClient replaces the client, the retry wrapper is rewrapped onto the
// new client's transport so the caller's intent is preserved.
func WithRetry(opts ...RetryOption) Option {
	policy := defaultRetryPolicy()
	for _, o := range opts {
		o(policy)
	}
	return func(c *Client) {
		c.retryPolicy = policy
	}
}

// installRetryWrapper installs the retry RoundTripper on top of whatever
// transport the client ended up with after every option func ran. Called
// from finalizeOptions so the wrapper survives WithHTTPClient regardless of
// option order.
//
// Unlike the TLS / proxy options, we don't need to mutate the underlying
// transport — wrapping it as the .base of a retryRoundTripper is read-only —
// so we don't promote / Clone the default. If c.httpClient.Transport is nil
// we just point the wrapper at http.DefaultTransport directly. (Test
// harnesses that swap DefaultTransport for an interceptor — gock — work fine
// because we wrap whatever the global currently is.)
func (c *Client) installRetryWrapper() {
	if c.retryPolicy == nil {
		return
	}
	c.ensureOwnHTTPClient()
	base := c.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	c.httpClient.Transport = &retryRoundTripper{
		base:   base,
		policy: c.retryPolicy,
	}
}

// retryRoundTripper is the http.RoundTripper wrapper installed by WithRetry.
type retryRoundTripper struct {
	base   http.RoundTripper
	policy *retryPolicy
}

// RoundTrip implements http.RoundTripper. It re-issues the request up to
// policy.maxAttempts times when policy.condition returns true. Between attempts
// it sleeps for an exponentially-growing, full-jitter window (or for the
// duration specified by a Retry-After header, when present and larger).
func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Snapshot the body once. http.Request.Body is consumed on each
	// RoundTrip, so without a rewind source we can only attempt once for
	// requests with bodies.
	rewind, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}

	idempotent := isIdempotent(req.Method) || rewind != nil

	var (
		lastRes *http.Response
		lastErr error
	)

	for attempt := 0; attempt < r.policy.maxAttempts; attempt++ {
		// Reset the body before every attempt after the first.
		if attempt > 0 && rewind != nil {
			req.Body = rewind()
		}

		res, err := r.base.RoundTrip(req)

		// Honor user-initiated cancellation immediately — never retry it.
		if err != nil && ctxErr(ctx) != nil {
			return nil, ctx.Err()
		}

		lastRes, lastErr = res, err

		if !r.policy.condition(res, err) {
			return res, err
		}

		// Don't retry non-idempotent verbs without a rewindable body.
		if !idempotent {
			return res, err
		}

		// No more attempts left — return whatever we got.
		if attempt == r.policy.maxAttempts-1 {
			return res, err
		}

		// Compute the backoff window. Retry-After (if present) overrides.
		delay := r.computeBackoff(attempt, res)

		// Drain and close the body so the underlying connection can be
		// reused on the next attempt.
		drainResponse(res)

		if err := r.policy.sleep(ctx, delay); err != nil {
			return nil, err
		}
	}

	return lastRes, lastErr
}

// computeBackoff returns the delay before the next attempt. It honors the
// Retry-After header on 429 and 503 responses, capping at policy.maxBackoff
// when the header's value is larger than the configured cap. Otherwise it
// returns a full-jitter sample in [0, min(maxBackoff, initial * 2^attempt)).
func (r *retryRoundTripper) computeBackoff(attempt int, res *http.Response) time.Duration {
	if res != nil {
		if d, ok := parseRetryAfter(res.Header.Get("Retry-After"), r.policy.now()); ok {
			if d > r.policy.maxBackoff {
				d = r.policy.maxBackoff
			}
			if d < 0 {
				d = 0
			}
			return d
		}
	}

	// Full jitter: backoff = rand.Int63n(min(cap, base * 2^attempt))
	window := r.policy.initialBackoff
	for i := 0; i < attempt && window < r.policy.maxBackoff; i++ {
		window *= 2
	}
	if window > r.policy.maxBackoff {
		window = r.policy.maxBackoff
	}
	if window <= 0 {
		return 0
	}

	r.policy.mu.Lock()
	defer r.policy.mu.Unlock()
	return time.Duration(r.policy.rand.Int63n(int64(window)))
}

// parseRetryAfter parses the Retry-After header value, which RFC 7231 defines
// as either delta-seconds or an HTTP-date. Returns the resulting duration and
// true on success.
func parseRetryAfter(v string, now time.Time) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := t.Sub(now)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}

// isIdempotent reports whether the HTTP method is safe to replay without a
// rewindable body.
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete,
		http.MethodOptions, http.MethodTrace:
		return true
	}
	return false
}

// snapshotBody returns a function that produces a fresh ReadCloser for the
// request body on each call, suitable for replay across retries. It returns
// (nil, nil) when the request has no body to replay.
//
// Preference order:
//  1. If the caller set req.GetBody, use it directly — that's the http stdlib's
//     own contract for replayable bodies.
//  2. Otherwise, fully buffer req.Body into memory and serve a fresh
//     bytes.Reader on each call. In go-proxmox the Req() helper always passes
//     a []byte, so the buffer cost is the same []byte we already had.
//
// If a body exists and neither option applies (Body is non-nil and there's no
// GetBody) we still buffer it. The caller's original Body is closed after the
// snapshot.
func snapshotBody(req *http.Request) (func() io.ReadCloser, error) {
	if req.GetBody != nil {
		return func() io.ReadCloser {
			rc, err := req.GetBody()
			if err != nil {
				return io.NopCloser(strings.NewReader(""))
			}
			return rc
		}, nil
	}
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	_ = req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(buf))
	// Also install GetBody so any other layer that wants to replay can do so.
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(buf)), nil
	}
	return func() io.ReadCloser {
		return io.NopCloser(bytes.NewReader(buf))
	}, nil
}

// drainResponse drains and closes a non-nil response body so the underlying
// HTTP/1.1 connection can be reused for the next attempt.
func drainResponse(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()
}

// ctxErr returns ctx.Err() — split out so the retry loop can be explicit
// about treating context cancellation as a terminal condition.
func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// contextSleep sleeps for d, returning early with ctx.Err() if ctx is canceled.
func contextSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctxErr(ctx)
	}
	if ctx == nil {
		time.Sleep(d)
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
