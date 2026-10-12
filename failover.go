package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// EndpointStrategy selects the order in which a client configured with
// WithFallbackEndpoints tries its endpoints.
type EndpointStrategy int

const (
	// EndpointPriority (the default) sends every request to the first
	// healthy endpoint in configured order: the primary URL given to
	// NewClient, then each fallback in the order it was added. Traffic
	// moves to a fallback only when the endpoints ahead of it are failing,
	// and moves back to the primary once its cooldown has expired and a
	// request to it succeeds.
	EndpointPriority EndpointStrategy = iota

	// EndpointRoundRobin spreads requests across the healthy endpoints in
	// rotation. Failover within a request and health tracking work the
	// same as with EndpointPriority.
	EndpointRoundRobin
)

const (
	defaultEndpointCooldown    = 30 * time.Second
	defaultEndpointMaxCooldown = 5 * time.Minute
)

// Proxmox-specific status codes. pveproxy answers 595 when the entry node
// cannot connect to the node that should serve the request, and 596 when
// that proxied connection fails or times out (including TLS errors). Neither
// says anything is wrong with the request itself.
const (
	statusPVEConnectError = 595
	statusPVEProxyError   = 596
)

// WithFallbackEndpoints adds further API base URLs for the same Proxmox
// cluster, e.g. "https://pve2:8006/api2/json". Every node in a cluster serves
// the full cluster-wide API and forwards node-specific calls to the node that
// owns them, so any node can answer any request. The URL passed to NewClient
// stays the primary; the fallbacks are tried, in the order given, when it is
// unreachable. Repeated calls append. Without this option the client talks to
// the single NewClient URL exactly as before.
//
// Failover is deliberately conservative:
//
//   - It is triggered by transport errors (DNS, dial, connection refused,
//     TLS handshake, a timeout or reset before a response) and by HTTP 502,
//     503, 504, 595 and 596.
//   - It is never triggered by a 4xx or by any other 5xx: those are answers
//     from Proxmox about the request, and another node would give the same
//     answer.
//   - Only GET, HEAD and OPTIONS requests fail over after the request may
//     have reached a server. POST, PUT, DELETE and PATCH fail over only when
//     the request provably never left the client: the connection or TLS
//     handshake failed before the request headers were written. A POST that
//     received any response, or failed after it was written, is returned to
//     the caller as-is, so a create or a start is never executed twice.
//   - Requests whose body can't be replayed (no GetBody, e.g. UploadReader
//     with a plain io.Reader) are sent to the currently preferred endpoint
//     only and do not fail over.
//   - Context cancellation and deadlines stop failover immediately; the
//     context bounds the whole request across every endpoint tried.
//
// An endpoint that fails with a transport error or 502/503/504 is marked
// unhealthy for a cooldown (WithEndpointCooldown; 30s doubling to at most 5m
// per consecutive failure) and is skipped while other endpoints are healthy.
// When every endpoint is cooling down they are all still tried, in order,
// rather than failing without trying. 595 and 596 fail the request over but
// don't mark the endpoint unhealthy, because they usually mean the node the
// request is about is down, not the node that answered.
//
// Composition with WithRetry: failover sits inside the retry wrapper. Each
// retry attempt tries up to every endpoint once, with no backoff between
// endpoints; the retry backoff applies between rounds. The worst case is
// therefore (retry attempts) x (number of endpoints) transport attempts per
// call. WithRetry's own policy for re-sending a POST is unchanged by this
// option.
//
// Only requests addressed to the primary base URL are redirected; an
// absolute URL to any other host passes through untouched, which is how a
// caller pins a request to one node. The websocket helpers (TermWebSocket,
// VNCWebSocket) dial the currently preferred endpoint once and never switch
// a live session. See the README section "Multiple endpoints and failover"
// for the authentication and pinning details.
//
// All endpoints must present certificates the client trusts (the cluster CA
// via WithRootCAFile verifies every node), and a tls.Config.ServerName set on
// a custom transport applies to every endpoint, so leave it empty.
func WithFallbackEndpoints(urls ...string) Option {
	return func(c *Client) {
		c.fallbackEndpoints = append(c.fallbackEndpoints, urls...)
	}
}

// WithEndpointStrategy selects how a client with fallback endpoints orders
// them. The default is EndpointPriority. Has no effect without
// WithFallbackEndpoints.
func WithEndpointStrategy(s EndpointStrategy) Option {
	return func(c *Client) {
		c.endpointStrategy = s
	}
}

// WithEndpointCooldown sets how long a failed endpoint is skipped. The first
// failure benches it for base; each consecutive failure doubles that, up to
// max. A success resets it. Defaults: 30s base, 5m max. Non-positive values
// keep the default. Has no effect without WithFallbackEndpoints.
func WithEndpointCooldown(base, max time.Duration) Option {
	return func(c *Client) {
		if base > 0 {
			c.endpointCooldown = base
		}
		if max > 0 {
			c.endpointMaxCooldown = max
		}
	}
}

// installFailoverWrapper wraps the client's transport in a failover
// RoundTripper when fallback endpoints are configured. It runs from
// finalizeOptions before installRetryWrapper, so retry ends up outermost.
func (c *Client) installFailoverWrapper() {
	if len(c.fallbackEndpoints) == 0 {
		return
	}
	primary := strings.TrimRight(c.baseURL, "/")
	if _, err := url.Parse(primary); err != nil || primary == "" {
		c.log.Errorf("WithFallbackEndpoints: primary URL %q is not usable; failover disabled", c.baseURL)
		return
	}
	bases := []string{primary}
	seen := map[string]bool{primary: true}
	for _, raw := range c.fallbackEndpoints {
		u := strings.TrimRight(strings.TrimSpace(raw), "/")
		if u == "" || seen[u] {
			continue
		}
		if parsed, err := url.Parse(u); err != nil || parsed.Scheme == "" || parsed.Host == "" {
			c.log.Errorf("WithFallbackEndpoints: ignoring invalid endpoint %q", raw)
			continue
		}
		seen[u] = true
		bases = append(bases, u)
	}
	if len(bases) < 2 {
		return
	}

	c.ensureOwnHTTPClient()
	base := c.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cooldown, maxCooldown := c.endpointCooldown, c.endpointMaxCooldown
	if cooldown <= 0 {
		cooldown = defaultEndpointCooldown
	}
	if maxCooldown <= 0 {
		maxCooldown = defaultEndpointMaxCooldown
	}
	if maxCooldown < cooldown {
		maxCooldown = cooldown
	}
	f := &failoverRoundTripper{
		base:        base,
		endpoints:   make([]*endpoint, len(bases)),
		strategy:    c.endpointStrategy,
		cooldown:    cooldown,
		maxCooldown: maxCooldown,
		log:         c.log,
		now:         time.Now,
	}
	for i, b := range bases {
		f.endpoints[i] = &endpoint{base: b}
	}
	c.failover = f
	c.httpClient.Transport = f
}

// endpointBaseURL is the base URL a new connection that can't fail over
// mid-flight (the websocket helpers) should use: the preferred endpoint when
// failover is configured, else the NewClient URL.
func (c *Client) endpointBaseURL() string {
	if c.failover != nil {
		return c.failover.preferred()
	}
	return c.baseURL
}

// endpoint is one API base URL plus its health state. The state fields are
// guarded by failoverRoundTripper.mu.
type endpoint struct {
	base string

	failures     int       // consecutive failures
	unhealthyTil time.Time // zero when healthy
}

// failoverRoundTripper redirects requests addressed to the primary base URL
// to whichever endpoint the strategy picks, and moves on to the next one when
// an attempt fails in a way that is safe and useful to fail over.
type failoverRoundTripper struct {
	base        http.RoundTripper
	endpoints   []*endpoint // endpoints[0] is the primary; immutable
	strategy    EndpointStrategy
	cooldown    time.Duration
	maxCooldown time.Duration
	log         LeveledLoggerInterface
	now         func() time.Time

	mu sync.Mutex
	rr atomic.Uint64
}

// RoundTrip implements http.RoundTripper.
func (f *failoverRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	suffix, ok := f.suffix(req.URL)
	if !ok {
		// Not addressed to the primary: an absolute URL to another host.
		return f.base.RoundTrip(req)
	}

	ctx := req.Context()
	safe := isSafeMethod(req.Method)
	replayable := req.Body == nil || req.Body == http.NoBody || req.GetBody != nil
	order := f.order()
	if !replayable {
		order = order[:1]
	}

	var (
		res *http.Response
		err error
	)
	for i, ep := range order {
		// Release the previous endpoint's response (nil on the first pass
		// or after a transport error) before moving on.
		drainResponse(res)
		if cerr := ctx.Err(); cerr != nil {
			return nil, cerr
		}

		attempt, perr := f.prepare(req, ep, suffix, i)
		if perr != nil {
			return nil, perr
		}
		var wroteHeaders, gotConn atomic.Bool
		trace := &httptrace.ClientTrace{
			GotConn:      func(httptrace.GotConnInfo) { gotConn.Store(true) },
			WroteHeaders: func() { wroteHeaders.Store(true) },
		}
		attempt = attempt.WithContext(httptrace.WithClientTrace(attempt.Context(), trace))

		res, err = f.base.RoundTrip(attempt)
		if res != nil {
			// Callers (and handleResponse's log redaction) compare
			// res.Request.URL against the primary base URL.
			res.Request = req
		}

		if err != nil && ctx.Err() != nil {
			// The caller's context ended. If it ran out while we were
			// still trying to connect, this endpoint is not answering
			// within the caller's budget: bench it so the next call
			// starts elsewhere instead of hanging on the same dead host.
			if errors.Is(ctx.Err(), context.DeadlineExceeded) && !gotConn.Load() {
				f.markFailed(ep, "no connection before the context deadline")
			}
			return nil, ctx.Err()
		}

		retry, bench, reason := classifyFailover(res, err)
		if !retry {
			if err == nil {
				f.markOK(ep)
			}
			return res, err
		}
		if bench {
			f.markFailed(ep, reason)
		}
		if !safe && (res != nil || wroteHeaders.Load()) {
			// The request reached, or may have reached, a server;
			// replaying it on another node could execute it twice.
			return res, err
		}
		if i < len(order)-1 {
			f.log.Warnf("proxmox endpoint %s failed (%s); failing over to %s", ep.base, reason, order[i+1].base)
		} else {
			f.log.Debugf("proxmox endpoint %s failed (%s); no endpoints left to try", ep.base, reason)
		}
	}
	return res, err
}

// prepare returns the request to send to ep. The first attempt reuses the
// caller's body; later attempts take a fresh copy from GetBody.
func (f *failoverRoundTripper) prepare(req *http.Request, ep *endpoint, suffix string, attempt int) (*http.Request, error) {
	u, err := url.Parse(ep.base + suffix)
	if err != nil {
		return nil, err
	}
	out := req.Clone(req.Context())
	out.URL = u
	out.Host = u.Host
	if attempt > 0 && req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		out.Body = body
	}
	return out, nil
}

// suffix returns the part of u after the primary base URL, and whether u is
// addressed to the primary at all.
func (f *failoverRoundTripper) suffix(u *url.URL) (string, bool) {
	s := u.String()
	primary := f.endpoints[0].base
	if !strings.HasPrefix(s, primary) {
		return "", false
	}
	rest := s[len(primary):]
	if rest != "" && rest[0] != '/' && rest[0] != '?' {
		// "https://pve1:8006" must not match "https://pve1:80067/...".
		return "", false
	}
	return rest, true
}

// order returns the endpoints to try for one request: healthy endpoints in
// strategy order first, then the ones cooling down, in configured order, as
// a last resort.
func (f *failoverRoundTripper) order() []*endpoint {
	now := f.now()
	f.mu.Lock()
	healthy := make([]*endpoint, 0, len(f.endpoints))
	var benched []*endpoint
	for _, ep := range f.endpoints {
		if ep.unhealthyTil.IsZero() || !now.Before(ep.unhealthyTil) {
			healthy = append(healthy, ep)
		} else {
			benched = append(benched, ep)
		}
	}
	f.mu.Unlock()

	if f.strategy == EndpointRoundRobin && len(healthy) > 1 {
		start := int(f.rr.Add(1)-1) % len(healthy)
		healthy = append(healthy[start:], healthy[:start]...)
	}
	return append(healthy, benched...)
}

// preferred returns the base URL of the endpoint the next request would go
// to first.
func (f *failoverRoundTripper) preferred() string {
	now := f.now()
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, ep := range f.endpoints {
		if ep.unhealthyTil.IsZero() || !now.Before(ep.unhealthyTil) {
			return ep.base
		}
	}
	return f.endpoints[0].base
}

func (f *failoverRoundTripper) markFailed(ep *endpoint, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	ep.failures++
	d := f.cooldown
	for i := 1; i < ep.failures && d < f.maxCooldown; i++ {
		d *= 2
	}
	if d > f.maxCooldown {
		d = f.maxCooldown
	}
	ep.unhealthyTil = f.now().Add(d)
	f.log.Debugf("proxmox endpoint %s marked unhealthy for %s after %d consecutive failure(s): %s", ep.base, d, ep.failures, reason)
}

func (f *failoverRoundTripper) markOK(ep *endpoint) {
	f.mu.Lock()
	wasDown := ep.failures > 0
	ep.failures = 0
	ep.unhealthyTil = time.Time{}
	f.mu.Unlock()
	if wasDown {
		f.log.Infof("proxmox endpoint %s is answering again", ep.base)
	}
}

// classifyFailover decides what a single attempt's outcome means for
// failover: whether another endpoint should be tried, whether this endpoint
// should be benched, and a short reason for the logs. It looks only at the
// raw transport result, never at the decoded body, so it is independent of
// how handleResponse turns responses into errors.
func classifyFailover(res *http.Response, err error) (retry, bench bool, reason string) {
	if err != nil {
		return true, true, err.Error()
	}
	if res == nil {
		return false, false, ""
	}
	switch res.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true, true, res.Status
	case statusPVEConnectError, statusPVEProxyError:
		return true, false, res.Status
	}
	return false, false, ""
}

// isSafeMethod reports whether replaying the request on another node after
// it may have reached a server is harmless. This is narrower than
// isIdempotent in retry.go: Proxmox PUT and DELETE calls are not reliably
// idempotent (a relative disk resize, a delete that starts a task), so
// failover treats them like POST.
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// unwrapTransport returns the *http.Transport underneath any wrapper this
// package installs (retry, failover), or nil when there is none.
func unwrapTransport(rt http.RoundTripper) *http.Transport {
	for {
		switch t := rt.(type) {
		case *http.Transport:
			return t
		case *retryRoundTripper:
			rt = t.base
		case *failoverRoundTripper:
			rt = t.base
		default:
			return nil
		}
	}
}
