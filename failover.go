package proxmox

import (
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

// WithAdditionalEndpoints registers one or more extra API base URLs for the
// same Proxmox cluster, e.g. "https://pve2:8006/api2/json". It is additive:
// the URL passed to NewClient stays the primary endpoint, and the URLs given
// here are failover targets tried, in the order supplied, only when the
// current endpoint is unreachable. Repeated calls append. Without this option
// the client talks to the single NewClient URL exactly as it did before, so
// the option is fully backward compatible.
//
// Any node in a quorate Proxmox cluster serves the whole cluster-wide API and
// transparently forwards node-specific calls to the node that owns them, and
// auth tickets / API tokens are cluster-wide, so a request that failed only
// because one node was down can be answered by any other node — including the
// authentication itself, since a ticket minted by one node is accepted by the
// rest of the cluster. That is what makes failover safe here.
//
// Failover is deliberately conservative:
//
//   - It is triggered ONLY by transport/connection-level failures: DNS
//     resolution failure, dial failure, connection refused, TLS handshake
//     failure, or a timeout/reset before a response is received — i.e. the
//     node is unreachable. This mirrors the err != nil branch of
//     defaultRetryCondition in retry.go.
//   - It is NEVER triggered by a well-formed HTTP response from a reachable
//     node, including any 4xx or 5xx status. Those are real answers about the
//     request; another node would give the same answer, and replaying a write
//     after a node has already processed it could apply the side effect twice.
//     (Transient 5xx retries against the same node remain the job of WithRetry.)
//   - Only GET, HEAD and OPTIONS requests fail over after the request may have
//     reached a server. A POST, PUT, DELETE or PATCH fails over only when the
//     request provably never left the client (no request headers were written
//     before the connection failed); if any bytes went out, the request is
//     returned to the caller as-is so a create/start/delete is never executed
//     twice. This is intentionally narrower than retry.go's isIdempotent,
//     because Proxmox PUT/DELETE calls are not reliably idempotent (a relative
//     disk resize, a delete that spawns a task).
//   - A request whose body cannot be replayed (no GetBody — e.g. UploadReader
//     with a streaming io.Reader) is sent to the current endpoint only and
//     does not fail over.
//   - Context cancellation and deadlines stop failover immediately; the
//     caller's context bounds the whole request across every endpoint tried.
//
// After a request succeeds against an endpoint, that endpoint becomes the
// current one for subsequent requests (sticky last-good); the client does not
// flap back to a previously-dead endpoint until the current one fails. The
// primary is always the starting point for a brand-new client.
//
// Composition with WithRetry: failover sits inside the retry wrapper. Each
// retry attempt tries up to every endpoint once (no backoff between
// endpoints); the retry backoff applies between whole rounds.
//
// Only requests addressed to the current/primary base URL are redirected; an
// absolute URL to any other host passes through untouched, which is how a
// caller can pin a request to one specific node.
//
// All endpoints must present certificates the client trusts (verify the
// cluster CA once via WithRootCAFile / WithRootCAs); a single TLS config
// applies to every endpoint.
func WithAdditionalEndpoints(urls ...string) Option {
	return func(c *Client) {
		c.fallbackEndpoints = append(c.fallbackEndpoints, urls...)
	}
}

// installFailoverWrapper wraps the client's transport in a failover
// RoundTripper when at least one usable fallback endpoint is configured. It is
// called from finalizeOptions before installRetryWrapper, so the retry wrapper
// ends up outermost and failover sits directly on top of the real transport.
func (c *Client) installFailoverWrapper() {
	if len(c.fallbackEndpoints) == 0 {
		return
	}

	// endpoints[0] is the primary, stored verbatim so it matches the string
	// Req builds (c.baseURL + path). Fallbacks have any trailing slash trimmed
	// so concatenation with a leading-slash suffix never doubles up.
	endpoints := []string{c.baseURL}
	seen := map[string]bool{strings.TrimRight(c.baseURL, "/"): true}
	for _, raw := range c.fallbackEndpoints {
		u := strings.TrimRight(strings.TrimSpace(raw), "/")
		if u == "" || seen[u] {
			continue
		}
		parsed, err := url.Parse(u)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			c.log.Errorf("WithAdditionalEndpoints: ignoring invalid endpoint %q", raw)
			continue
		}
		seen[u] = true
		endpoints = append(endpoints, u)
	}
	if len(endpoints) < 2 {
		return
	}

	c.ensureOwnHTTPClient()
	base := c.httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	f := &failoverRoundTripper{
		base:      base,
		endpoints: endpoints,
		log:       c.log,
	}
	c.failover = f
	c.httpClient.Transport = f
}

// endpointBaseURL returns the base URL a connection that cannot fail over
// mid-flight (the websocket helpers) should dial: the current endpoint when
// failover is configured, else the NewClient URL.
func (c *Client) endpointBaseURL() string {
	if c.failover != nil {
		return c.failover.currentBase()
	}
	return c.baseURL
}

// failoverRoundTripper redirects requests addressed to the primary base URL to
// the current endpoint, and moves on to the next endpoint when an attempt
// fails at the transport level in a way that is safe to replay elsewhere.
type failoverRoundTripper struct {
	base      http.RoundTripper
	endpoints []string // endpoints[0] is the primary; immutable after construction
	log       LeveledLoggerInterface

	mu      sync.Mutex
	current int // index into endpoints; the sticky last-good endpoint
}

// RoundTrip implements http.RoundTripper.
func (f *failoverRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	suffix, ok := f.suffix(req.URL)
	if !ok {
		// Absolute URL addressed to some other host: pass through untouched so
		// a caller can pin a request to a specific node.
		return f.base.RoundTrip(req)
	}

	ctx := req.Context()
	safe := isSafeMethod(req.Method)
	replayable := req.Body == nil || req.Body == http.NoBody || req.GetBody != nil

	order := f.order()
	if !replayable {
		// The body can only be read once; stay on the current endpoint.
		order = order[:1]
	}

	var (
		res *http.Response
		err error
	)
	for pos, ep := range order {
		if cerr := ctx.Err(); cerr != nil {
			return nil, cerr
		}

		attempt, perr := f.prepare(req, ep, suffix, pos)
		if perr != nil {
			return nil, perr
		}

		// httptrace tells us whether any request bytes were written. If the
		// connection failed before WroteHeaders fired, the request provably
		// never reached a server and is safe to replay on another endpoint
		// regardless of method.
		var wroteHeaders atomic.Bool
		trace := &httptrace.ClientTrace{
			WroteHeaders: func() { wroteHeaders.Store(true) },
		}
		attempt = attempt.WithContext(httptrace.WithClientTrace(attempt.Context(), trace))

		res, err = f.base.RoundTrip(attempt)
		if err == nil {
			// A reachable node answered (any status). Restore res.Request to
			// the original so handleResponse's /access/ticket URL comparison
			// and log redaction still see the primary base URL. Remember this
			// endpoint as the current one.
			if res != nil {
				res.Request = req
			}
			f.markCurrent(ep)
			return res, nil
		}

		// Honor caller cancellation immediately — never fail over a
		// user-requested cancel/deadline.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// A non-idempotent request that may have reached a server must not be
		// silently replayed on another node.
		if !safe && wroteHeaders.Load() {
			return res, err
		}

		if pos < len(order)-1 {
			f.log.Warnf("proxmox endpoint %s unreachable (%v); failing over to %s",
				f.endpoints[ep], err, f.endpoints[order[pos+1]])
		} else {
			f.log.Debugf("proxmox endpoint %s unreachable (%v); no endpoints left to try",
				f.endpoints[ep], err)
		}
	}
	return res, err
}

// prepare returns the request to send to endpoint index ep. The first attempt
// reuses the caller's body; later attempts take a fresh copy from GetBody.
func (f *failoverRoundTripper) prepare(req *http.Request, ep int, suffix string, pos int) (*http.Request, error) {
	u, err := url.Parse(f.endpoints[ep] + suffix)
	if err != nil {
		return nil, err
	}
	out := req.Clone(req.Context())
	out.URL = u
	out.Host = u.Host
	if pos > 0 && req.GetBody != nil {
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
	primary := f.endpoints[0]
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

// order returns the endpoint indices to try for one request: the current
// (sticky last-good) endpoint first, then the remaining endpoints in
// configured order.
func (f *failoverRoundTripper) order() []int {
	f.mu.Lock()
	cur := f.current
	f.mu.Unlock()

	out := make([]int, 0, len(f.endpoints))
	out = append(out, cur)
	for i := range f.endpoints {
		if i != cur {
			out = append(out, i)
		}
	}
	return out
}

// currentBase returns the base URL of the current (sticky last-good) endpoint.
func (f *failoverRoundTripper) currentBase() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.endpoints[f.current]
}

// markCurrent records ep as the current endpoint for subsequent requests.
func (f *failoverRoundTripper) markCurrent(ep int) {
	f.mu.Lock()
	changed := f.current != ep
	f.current = ep
	f.mu.Unlock()
	if changed {
		f.log.Infof("proxmox: current endpoint is now %s", f.endpoints[ep])
	}
}

// isSafeMethod reports whether replaying the request on another node after it
// may have reached a server is harmless. Narrower than retry.go's isIdempotent
// on purpose: Proxmox PUT and DELETE calls can carry side effects, so failover
// treats them like POST once the request has been delivered.
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// unwrapTransport returns the *http.Transport underneath any RoundTripper
// wrapper this package installs (retry, failover), or nil when the base is a
// custom non-Transport implementation (or nil). Used by the websocket helpers,
// which need the underlying *http.Transport's TLS config and cannot route
// through the wrappers.
func unwrapTransport(rt http.RoundTripper) *http.Transport {
	for {
		switch t := rt.(type) {
		case nil:
			return nil
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
