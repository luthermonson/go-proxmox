package proxmox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Circuit-breaker defaults. An endpoint is benched only after
// defaultEndpointFailureThreshold consecutive transport failures (so a single
// transient blip never penalises an otherwise-healthy node). The first bench
// lasts defaultEndpointCooldownBase and doubles with each further consecutive
// failure up to defaultEndpointCooldownMax. A success clears the state.
const (
	defaultEndpointFailureThreshold = 3
	defaultEndpointCooldownBase     = 10 * time.Second
	defaultEndpointCooldownMax      = 180 * time.Second
)

// WithAdditionalEndpoints registers one or more extra API base URLs for the
// same Proxmox cluster, e.g. "https://pve2:8006/api2/json". It is additive:
// the URL passed to NewClient stays the primary endpoint, and the URLs given
// here are failover targets tried, in the order supplied, when the current
// endpoint is unreachable or returns a gateway-class error. Repeated calls
// append. Without this option
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
//   - It is triggered by transport/connection-level failures (DNS resolution
//     failure, dial failure, connection refused, TLS handshake failure, or a
//     timeout/reset before a response is received — i.e. the node is
//     unreachable), AND by a gateway-class HTTP status that means the entry
//     node could not itself serve the request: a 502/503/504, or Proxmox's
//     595/596, which pveproxy returns when it answered but could not reach the
//     target node that owns the resource.
//   - It is NOT triggered by any other HTTP response. A 4xx, a 500, or an
//     ordinary 2xx/3xx is a real answer from a reachable node; another node
//     would give the same answer, so failing over would only duplicate work.
//     (Transient 5xx retries against the same node remain the job of WithRetry.)
//   - Benching reflects blame: a 502/503/504 implicates the entry node and
//     feeds the circuit breaker; a 595/596 implicates the target node, not the
//     entry node, so the entry node is skipped for the request but not benched.
//   - A pure connection-establishment failure (DNS, dial, connection refused,
//     TLS handshake) proves the request never left the client, so it fails
//     over for ANY method, including POST/PUT/DELETE. An ambiguous outcome —
//     a failure after request headers were written (a mid-flight timeout or
//     reset), or any gateway-class status (the server received the request and
//     may already have acted) — fails over only for the idempotent methods
//     GET/HEAD/OPTIONS; a write that may have been delivered is returned to the
//     caller as-is so a create/start/delete is never executed twice.
//     "Idempotent" here is intentionally narrower than retry.go's isIdempotent,
//     because Proxmox PUT/DELETE calls are not reliably idempotent (a relative
//     disk resize, a delete that spawns a task).
//   - A request whose body cannot be replayed (no GetBody — e.g. UploadReader
//     with a streaming io.Reader) is sent to the current endpoint only and
//     does not fail over.
//   - Context cancellation and deadlines stop failover immediately; the
//     caller's context bounds the whole request across every endpoint tried. A
//     deadline that expires before an endpoint even connects benches that
//     endpoint, so the next call starts elsewhere instead of hanging on a
//     black-holed node.
//
// Circuit breaker: an endpoint that fails at the transport layer
// WithEndpointFailureThreshold times in a row (default 3) is benched for a
// cooldown window (WithEndpointCooldown; default 10s, doubling per further
// consecutive failure up to 180s) and is skipped without being dialed while
// another endpoint is healthy, so a dead node is not re-probed on every
// request. By default the bench is evaluated lazily on the request path — there
// is no background goroutine — and the endpoint is re-probed once the window
// expires; a success clears the state. WithEndpointHealthCheck adds an opt-in
// background supervisor that probes benched endpoints proactively instead, and
// WithEndpointMaxDowntime soft-evicts one that stays down too long. When every
// endpoint is benched they are all still tried, in order, rather than failing
// without an attempt.
//
// After a request succeeds against an endpoint, that endpoint becomes the
// current one for subsequent requests (sticky last-good); the client does not
// flap back to a previously-dead endpoint until the current one fails. The
// current endpoint and the per-endpoint health are shared across all
// goroutines using the client, so concurrent callers cooperate: once one
// request has benched a dead node the others skip it instead of each
// rediscovering it. The primary is always the starting point for a brand-new
// client.
//
// Composition with WithRetry: failover sits inside the retry wrapper. Each
// retry attempt tries up to every healthy endpoint once (no backoff between
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

// EndpointStrategy selects the endpoint a request is sent to first when more
// than one is configured with WithAdditionalEndpoints. Failover, the circuit
// breaker and the replay rules above apply the same way under every strategy;
// the strategy only decides where a request starts.
type EndpointStrategy int

const (
	// EndpointSticky stays on the endpoint that last answered until it fails
	// (sticky last-good). The default, and the behaviour before strategies
	// existed: a brand-new client starts on the primary and only moves when
	// the current endpoint is unreachable or returns a gateway-class status.
	EndpointSticky EndpointStrategy = iota

	// EndpointRoundRobin starts each request on the next healthy endpoint in
	// turn, so API load spreads over every node instead of pinning the one
	// behind the NewClient URL, and the proxy hop a pinned node adds for
	// every other node's resources is spread with it. Benched endpoints are
	// skipped in the rotation and come back once their cooldown ends. Any
	// node answers any request (pveproxy forwards node-specific calls to the
	// owner), so correctness does not depend on which node is hit; a listing
	// read from a non-owning node may lag the owner by a moment, as it does
	// for any request that node serves.
	EndpointRoundRobin
)

// String implements fmt.Stringer.
func (s EndpointStrategy) String() string {
	switch s {
	case EndpointSticky:
		return "sticky"
	case EndpointRoundRobin:
		return "round-robin"
	default:
		return fmt.Sprintf("EndpointStrategy(%d)", int(s))
	}
}

// ParseEndpointStrategy returns the strategy named by s: "sticky" or
// "round-robin" (also "roundrobin" and "rr"), ignoring case. It is the
// counterpart of EndpointStrategy.String for configuration files and
// environment variables.
func ParseEndpointStrategy(s string) (EndpointStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "sticky":
		return EndpointSticky, nil
	case "round-robin", "roundrobin", "rr":
		return EndpointRoundRobin, nil
	default:
		return EndpointSticky, fmt.Errorf("unknown endpoint strategy %q (want sticky or round-robin)", s)
	}
}

// WithEndpointStrategy sets how the client chooses the endpoint a request
// starts on; see EndpointStrategy. The default is EndpointSticky. Has no
// effect without WithAdditionalEndpoints.
func WithEndpointStrategy(s EndpointStrategy) Option {
	return func(c *Client) {
		c.endpointStrategy = s
	}
}

// WithEndpointFailureThreshold sets how many consecutive transport failures an
// endpoint must accumulate before the circuit breaker benches it. The default
// is 3, which keeps a single transient blip from penalising a healthy node.
// Values below 1 are clamped to 1. Has no effect without
// WithAdditionalEndpoints.
func WithEndpointFailureThreshold(n int) Option {
	return func(c *Client) {
		if n < 1 {
			n = 1
		}
		c.endpointFailureThreshold = n
	}
}

// WithEndpointCooldown overrides the circuit-breaker cooldown bounds. base is
// the first bench window once the failure threshold is crossed; it doubles
// with each further consecutive failure and is capped at max. Defaults: 10s
// base, 180s max. Non-positive values keep the default. Has no effect without
// WithAdditionalEndpoints.
func WithEndpointCooldown(base, max time.Duration) Option {
	return func(c *Client) {
		if base > 0 {
			c.endpointCooldownBase = base
		}
		if max > 0 {
			c.endpointCooldownMax = max
		}
	}
}

// WithEndpointHealthCheck turns on active health checking for the failover
// endpoints. By default a benched endpoint is only re-probed lazily, when a
// request happens to reach it; with this option a single background goroutine
// probes benched endpoints on the circuit breaker's backoff schedule and
// returns them to rotation the moment they answer, off the request critical
// path. The goroutine runs only while at least one endpoint is down and stops
// once they have all recovered, so nothing runs while the cluster is healthy;
// it is also stopped when ctx is cancelled, so pass a context tied to the
// client's lifetime (one you cancel on shutdown). A nil ctx leaves health
// checking off. Has no effect without WithAdditionalEndpoints.
func WithEndpointHealthCheck(ctx context.Context) Option {
	return func(c *Client) {
		c.endpointHealthCtx = ctx
	}
}

// WithEndpointMaxDowntime soft-evicts an endpoint that has been continuously
// unreachable for at least d: it is dropped from the normal rotation and
// probed only at the cooldown maximum, but it is never removed for good — a
// probe that finally answers restores it — and the last endpoint still eligible
// is never evicted, so the client can always reach the cluster. Requires
// WithEndpointHealthCheck (eviction relies on the active probe to bring a node
// back); d <= 0 disables eviction. Has no effect without WithAdditionalEndpoints.
func WithEndpointMaxDowntime(d time.Duration) Option {
	return func(c *Client) {
		c.endpointMaxDowntime = d
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
	bases := []string{c.baseURL}
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

	threshold := c.endpointFailureThreshold
	if threshold < 1 {
		threshold = defaultEndpointFailureThreshold
	}
	cooldownBase := c.endpointCooldownBase
	if cooldownBase <= 0 {
		cooldownBase = defaultEndpointCooldownBase
	}
	cooldownMax := c.endpointCooldownMax
	if cooldownMax <= 0 {
		cooldownMax = defaultEndpointCooldownMax
	}
	if cooldownMax < cooldownBase {
		cooldownMax = cooldownBase
	}

	f := &failoverRoundTripper{
		base:         base,
		log:          c.log,
		now:          time.Now,
		strategy:     c.endpointStrategy,
		threshold:    threshold,
		cooldownBase: cooldownBase,
		cooldownMax:  cooldownMax,
		healthCtx:    c.endpointHealthCtx,
		maxDowntime:  c.endpointMaxDowntime,
		endpoints:    make([]*endpoint, len(bases)),
	}
	for i, b := range bases {
		f.endpoints[i] = &endpoint{base: b}
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

// endpoint is one API base URL plus its circuit-breaker health state. The
// state fields are guarded by failoverRoundTripper.mu.
type endpoint struct {
	base string

	failures   int       // consecutive transport failures
	benchedTil time.Time // zero when healthy; otherwise skip until this time
	downSince  time.Time // first failure of the current down streak; zero when healthy
	evicted    bool      // soft-evicted after exceeding max downtime; healthy() stays false until a probe clears it
}

// failoverRoundTripper redirects requests addressed to the primary base URL to
// the current endpoint, moves on to the next healthy endpoint when an attempt
// fails at the transport level, or returns a gateway-class status, in a way
// that is safe to replay elsewhere, and benches an endpoint that keeps failing
// so it is skipped fast. All of its
// mutable state (the sticky current cursor and the per-endpoint health) is
// shared across goroutines so concurrent callers cooperate.
type failoverRoundTripper struct {
	base         http.RoundTripper
	log          LeveledLoggerInterface
	now          func() time.Time
	threshold    int
	cooldownBase time.Duration
	cooldownMax  time.Duration

	mu        sync.Mutex
	endpoints []*endpoint // endpoints[0] is the primary; slice is immutable, elements are guarded by mu
	current   int         // index into endpoints; the sticky last-good endpoint

	// strategy decides where a request starts (see EndpointStrategy); cursor
	// is the round-robin position, advanced once per request.
	strategy EndpointStrategy
	cursor   atomic.Uint64

	// Active health checking (WithEndpointHealthCheck). healthCtx is nil when
	// disabled; while set, a single supervisor goroutine — guarded by probing,
	// started on demand and self-exiting when every endpoint is healthy —
	// probes down endpoints and restores them off the request path. maxDowntime
	// soft-evicts an endpoint that stays down that long (WithEndpointMaxDowntime).
	healthCtx   context.Context
	maxDowntime time.Duration
	probing     atomic.Bool
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
	for pos, idx := range order {
		if cerr := ctx.Err(); cerr != nil {
			return nil, cerr
		}

		attempt, perr := f.prepare(req, idx, suffix, pos)
		if perr != nil {
			return nil, perr
		}

		// httptrace tells us whether any request bytes were written. If the
		// connection failed before WroteHeaders fired, the request provably
		// never reached a server and is safe to replay on another endpoint
		// regardless of method.
		var wroteHeaders, gotConn atomic.Bool
		trace := &httptrace.ClientTrace{
			GotConn:      func(httptrace.GotConnInfo) { gotConn.Store(true) },
			WroteHeaders: func() { wroteHeaders.Store(true) },
		}
		attempt = attempt.WithContext(httptrace.WithClientTrace(attempt.Context(), trace))

		res, err = f.base.RoundTrip(attempt)
		if err == nil {
			// A reachable node answered. A gateway-class status means this
			// entry node could not itself serve or forward the request (a bad
			// gateway, or Proxmox's 595/596 "can't reach the target node"), so
			// another endpoint may still succeed.
			retry, bench, reason := classifyStatus(res)
			if !retry {
				// A genuine answer (any non-gateway status, including 4xx/500).
				// Restore res.Request to the original so handleResponse's
				// /access/ticket URL comparison and log redaction still see the
				// primary base URL. Clear this endpoint's failure state and make
				// it the current one.
				if res != nil {
					res.Request = req
				}
				f.markOK(idx)
				f.markCurrent(idx)
				return res, nil
			}
			// A 502/503/504 implicates the endpoint itself, so feed the breaker;
			// 595/596 mean this entry node couldn't reach the target node, so
			// fail over without penalising an otherwise-healthy entry node.
			if bench {
				f.markFailed(idx, fmt.Errorf("endpoint returned %s", reason))
			}
			// The request reached a server and it responded, so a non-idempotent
			// call may already have taken effect; only safe methods may be
			// replayed on another node.
			if !safe {
				if res != nil {
					res.Request = req
				}
				return res, nil
			}
			// Discard the gateway response before trying the next node so the
			// connection can return to the pool, and present the failure as an
			// error so the retry loop (and a final return) treats it uniformly.
			drainResponse(res)
			res, err = nil, fmt.Errorf("proxmox endpoint %s returned %s", f.endpoints[idx].base, reason)
			if pos < len(order)-1 {
				f.log.Warnf("proxmox endpoint %s returned %s; failing over to %s",
					f.endpoints[idx].base, reason, f.endpoints[order[pos+1]].base)
			} else {
				f.log.Debugf("proxmox endpoint %s returned %s; no endpoints left to try",
					f.endpoints[idx].base, reason)
			}
			continue
		}

		// The caller's context ended during the attempt. A user-requested cancel
		// is honored without touching endpoint health, but a deadline that
		// expired before we ever established a connection provably implicates
		// this endpoint (a black-holed node): bench it so a caller deadline
		// shorter than the dial timeout can't pin every future request to the
		// dead endpoint forever.
		if cerr := ctx.Err(); cerr != nil {
			if errors.Is(cerr, context.DeadlineExceeded) && !gotConn.Load() {
				f.markFailed(idx, fmt.Errorf("no connection before the context deadline: %w", err))
			}
			return nil, cerr
		}

		// Record the transport failure in the shared breaker state so this and
		// concurrent requests skip the dead endpoint once the threshold is hit.
		f.markFailed(idx, err)

		// Decide whether this request may continue to another endpoint. A
		// request that provably never reached a server — a pure connection
		// error, or no request headers written — is safe to replay for ANY
		// method. Otherwise the server may already have acted, so only the
		// idempotent methods continue.
		neverDelivered := isConnectError(err) || !wroteHeaders.Load()
		if !safe && !neverDelivered {
			return res, err
		}

		if pos < len(order)-1 {
			f.log.Warnf("proxmox endpoint %s unreachable (%v); failing over to %s",
				f.endpoints[idx].base, err, f.endpoints[order[pos+1]].base)
		} else {
			f.log.Debugf("proxmox endpoint %s unreachable (%v); no endpoints left to try",
				f.endpoints[idx].base, err)
		}
	}
	return res, err
}

// prepare returns the request to send to endpoint index idx. The first attempt
// reuses the caller's body; later attempts take a fresh copy from GetBody.
func (f *failoverRoundTripper) prepare(req *http.Request, idx int, suffix string, pos int) (*http.Request, error) {
	u, err := url.Parse(f.endpoints[idx].base + suffix)
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

// order returns the endpoint indices to try for one request: the healthy
// endpoints first, then any benched endpoints as a last resort (so a request
// still goes out when everything is cooling down). Which healthy endpoint
// leads depends on the strategy: under EndpointSticky it is the current
// (sticky last-good) endpoint, followed by the rest in configured order;
// under EndpointRoundRobin each call advances a cursor over the endpoints and
// the healthy ones follow in cyclic order from there.
func (f *failoverRoundTripper) order() []int {
	now := f.now()
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.strategy == EndpointRoundRobin {
		n := len(f.endpoints)
		start := int(f.cursor.Add(1)-1) % n
		healthy := make([]int, 0, n)
		var benched []int
		for k := range n {
			i := (start + k) % n
			if f.endpoints[i].healthy(now) {
				healthy = append(healthy, i)
			} else {
				benched = append(benched, i)
			}
		}
		return append(healthy, benched...)
	}

	cur := f.current
	healthy := make([]int, 0, len(f.endpoints))
	var benched []int
	// Put the current endpoint first among the healthy ones.
	if f.endpoints[cur].healthy(now) {
		healthy = append(healthy, cur)
	}
	for i, ep := range f.endpoints {
		if i == cur {
			continue
		}
		if ep.healthy(now) {
			healthy = append(healthy, i)
		} else {
			benched = append(benched, i)
		}
	}
	if !f.endpoints[cur].healthy(now) {
		// current is benched: it still belongs in the last-resort list.
		benched = append([]int{cur}, benched...)
	}
	return append(healthy, benched...)
}

// healthy reports whether the endpoint may be dialed at time now. A
// soft-evicted endpoint is never healthy until a health probe restores it, so
// it stays out of the normal rotation (it remains a last-resort target in
// order()).
func (e *endpoint) healthy(now time.Time) bool {
	if e.evicted {
		return false
	}
	return e.benchedTil.IsZero() || !now.Before(e.benchedTil)
}

// currentBase returns the base URL of the current (sticky last-good) endpoint.
func (f *failoverRoundTripper) currentBase() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.endpoints[f.current].base
}

// markCurrent records idx as the current endpoint: the one subsequent requests
// start on under EndpointSticky, and the one connections that cannot fail over
// mid-flight (see endpointBaseURL) dial under every strategy. Under round robin
// the current endpoint changes with nearly every request, so it is not logged.
func (f *failoverRoundTripper) markCurrent(idx int) {
	f.mu.Lock()
	changed := f.current != idx
	f.current = idx
	f.mu.Unlock()
	if changed && f.strategy != EndpointRoundRobin {
		f.log.Infof("proxmox: current endpoint is now %s", f.endpoints[idx].base)
	}
}

// markOK clears an endpoint's failure state after a successful request or
// health probe, returning it to rotation (including from soft-eviction).
func (f *failoverRoundTripper) markOK(idx int) {
	f.mu.Lock()
	ep := f.endpoints[idx]
	wasEvicted := ep.evicted
	wasDown := ep.failures > 0 || !ep.benchedTil.IsZero() || wasEvicted
	ep.failures = 0
	ep.benchedTil = time.Time{}
	ep.downSince = time.Time{}
	ep.evicted = false
	f.mu.Unlock()
	switch {
	case wasEvicted:
		f.log.Infof("proxmox endpoint %s recovered and is back in rotation", ep.base)
	case wasDown:
		f.log.Infof("proxmox endpoint %s is answering again", ep.base)
	}
}

// markFailed records a transport failure for an endpoint and benches it once
// the consecutive-failure threshold is reached, with an exponentially growing
// cooldown capped at cooldownMax.
func (f *failoverRoundTripper) markFailed(idx int, cause error) {
	now := f.now()
	f.mu.Lock()
	ep := f.endpoints[idx]
	ep.failures++
	if ep.downSince.IsZero() {
		ep.downSince = now
	}
	var benched time.Duration
	if ep.failures >= f.threshold {
		benched = f.cooldownBase
		for i := f.threshold; i < ep.failures && benched < f.cooldownMax; i++ {
			benched *= 2
		}
		if benched > f.cooldownMax {
			benched = f.cooldownMax
		}
		ep.benchedTil = now.Add(benched)
	}
	// Soft-evict an endpoint that has stayed down past maxDowntime: it leaves
	// the rotation (healthy() is false until a probe restores it) and is then
	// probed only at the cooldown maximum. The last endpoint still eligible is
	// never evicted, so the client can always reach the cluster. Eviction only
	// applies with the active poller (WithEndpointHealthCheck) to bring it back.
	evicted := false
	if f.healthCtx != nil && f.maxDowntime > 0 && !ep.evicted &&
		now.Sub(ep.downSince) >= f.maxDowntime && f.nonEvictedCountLocked() > 1 {
		ep.evicted = true
		evicted = true
	}
	failures := ep.failures
	f.mu.Unlock()

	if benched > 0 {
		f.log.Warnf("proxmox endpoint %s benched for %s after %d consecutive failure(s): %v",
			ep.base, benched, failures, cause)
	}
	if evicted {
		f.log.Warnf("proxmox endpoint %s evicted from rotation after being down for at least %s; it will be probed at the %s cooldown maximum until it answers",
			ep.base, f.maxDowntime, f.cooldownMax)
	}
	// A newly or still-benched endpoint has work for the health-check supervisor;
	// start it if active health checking is on and it is not already running.
	if benched > 0 {
		f.maybeStartHealthCheck()
	}
}

// nonEvictedCountLocked returns how many endpoints are not soft-evicted. The
// caller must hold f.mu.
func (f *failoverRoundTripper) nonEvictedCountLocked() int {
	n := 0
	for _, ep := range f.endpoints {
		if !ep.evicted {
			n++
		}
	}
	return n
}

// maybeStartHealthCheck launches the background health-check supervisor when
// active health checking is enabled (WithEndpointHealthCheck) and it is not
// already running. It is called whenever an endpoint is benched; the goroutine
// exits on its own once every endpoint is healthy again, so no goroutine runs
// while the cluster is up.
func (f *failoverRoundTripper) maybeStartHealthCheck() {
	if f.healthCtx == nil || f.healthCtx.Err() != nil {
		return
	}
	if f.probing.CompareAndSwap(false, true) {
		go f.healthCheckLoop()
	}
}

// healthCheckLoop probes down endpoints on the circuit breaker's own backoff
// schedule and returns each to rotation as soon as it answers. It runs only
// while at least one endpoint is down and stops when all have recovered or the
// health context is cancelled.
func (f *failoverRoundTripper) healthCheckLoop() {
	for {
		wait, any := f.nextProbe(f.now())
		if !any {
			// Nothing to probe. Release the running flag, then re-check to close
			// the race with a concurrent bench that set it just after our check:
			// if work reappeared, reclaim the flag and keep going, otherwise a
			// fresh bench will start a new loop.
			f.probing.Store(false)
			if _, stillAny := f.nextProbe(f.now()); !stillAny {
				return
			}
			if !f.probing.CompareAndSwap(false, true) {
				return // another goroutine took over
			}
			continue
		}
		if wait > 0 {
			select {
			case <-f.healthCtx.Done():
				f.probing.Store(false)
				return
			case <-time.After(wait):
			}
		} else if f.healthCtx.Err() != nil {
			f.probing.Store(false)
			return
		}
		f.probeDown()
	}
}

// nextProbe reports how long until the earliest benched endpoint is due to be
// re-probed, and whether any endpoint is down at all. A zero wait with any true
// means at least one down endpoint is due to be probed now. Holds f.mu.
func (f *failoverRoundTripper) nextProbe(now time.Time) (time.Duration, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var (
		any  bool
		next time.Time
	)
	for _, ep := range f.endpoints {
		// benchedTil stays non-zero from the first bench until a success clears
		// it, so it — not healthy(), which flips back lazily once the cooldown
		// window passes — is the signal that an endpoint is still down and owed
		// a probe. The probe is due when its bench window has elapsed.
		if ep.benchedTil.IsZero() {
			continue
		}
		any = true
		if !ep.benchedTil.After(now) {
			return 0, true
		}
		if next.IsZero() || ep.benchedTil.Before(next) {
			next = ep.benchedTil
		}
	}
	if !any {
		return 0, false
	}
	return next.Sub(now), true
}

// probeDown probes every endpoint whose bench window has elapsed, restoring the
// ones that answer and re-benching (and eventually evicting) the ones that do
// not.
func (f *failoverRoundTripper) probeDown() {
	now := f.now()
	f.mu.Lock()
	var toProbe []int
	for i, ep := range f.endpoints {
		// Probe every endpoint whose bench window has elapsed but which a
		// success has not yet cleared (benchedTil still set); see nextProbe.
		if !ep.benchedTil.IsZero() && !ep.benchedTil.After(now) {
			toProbe = append(toProbe, i)
		}
	}
	f.mu.Unlock()

	for _, idx := range toProbe {
		if f.healthCtx.Err() != nil {
			return
		}
		if err := f.probe(idx); err != nil {
			f.markFailed(idx, fmt.Errorf("health probe: %w", err))
		} else {
			f.markOK(idx)
		}
	}
}

// probe issues a lightweight GET /version to the endpoint and reports whether
// it is reachable. Any HTTP response — an auth challenge included — means the
// node answered; only a transport error counts as still-down, which is exactly
// what benched the endpoint in the first place. It dials f.base directly so the
// probe neither recurses through failover nor disturbs the sticky cursor.
func (f *failoverRoundTripper) probe(idx int) error {
	ctx, cancel := context.WithTimeout(f.healthCtx, f.probeTimeout())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.endpoints[idx].base+"/version", nil)
	if err != nil {
		return err
	}
	res, err := f.base.RoundTrip(req)
	if err != nil {
		return err
	}
	drainResponse(res)
	return nil
}

// probeTimeout bounds a single health probe so it cannot outlast the window it
// is refreshing.
func (f *failoverRoundTripper) probeTimeout() time.Duration {
	d := f.cooldownBase
	if d <= 0 || d > 15*time.Second {
		d = 15 * time.Second
	}
	if d < time.Second {
		d = time.Second
	}
	return d
}

// isSafeMethod reports whether replaying the request on another node after it
// may have reached a server is harmless. Narrower than retry.go's isIdempotent
// on purpose: Proxmox PUT and DELETE calls can carry side effects, so failover
// treats them like POST once the request may have been delivered.
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// Proxmox-specific gateway status codes returned by pveproxy when the entry
// node you connected to could not reach the node that owns the resource. They
// mean "try a different entry node", not "this node is unhealthy".
const (
	statusPVEConnectError = 595 // pveproxy: connection to the target node failed
	statusPVEProxyError   = 596 // pveproxy: proxy handshake to the target node failed
)

// classifyStatus decides whether a response that arrived without a transport
// error is in fact a gateway-class failure that should fail over to another
// endpoint, and whether that endpoint should be benched for it. A 502/503/504
// implicates the entry node itself (bench it); Proxmox's 595/596 mean the entry
// node could not reach the *target* node, so fail over without benching the
// entry node. Every other status (2xx/3xx/4xx, and 500) is a real answer.
func classifyStatus(res *http.Response) (retry, bench bool, reason string) {
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

// isConnectError reports whether err is a pure connection-establishment
// failure — DNS resolution failure, dial failure (connection refused,
// unreachable, dial timeout), or a TLS handshake/certificate failure. All of
// these happen before any request bytes are written, so the request provably
// never reached a server and is safe to replay on another endpoint regardless
// of HTTP method. A mid-flight error after headers were written is NOT a
// connect error and is handled by the WroteHeaders check in RoundTrip.
func isConnectError(err error) bool {
	if err == nil {
		return false
	}

	// Dial failures surface as *net.OpError with Op == "dial" (connection
	// refused, no route to host, dial timeout all included).
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return true
	}

	// DNS resolution failure.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// Connection refused, surfaced from the syscall layer (belt-and-suspenders
	// alongside the Op == "dial" check above, across platforms).
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	// TLS handshake / certificate verification failures happen during
	// connection setup, before the request is written.
	var recordErr tls.RecordHeaderError
	if errors.As(err, &recordErr) {
		return true
	}
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		return true
	}
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return true
	}
	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return true
	}
	var certInvalid x509.CertificateInvalidError
	return errors.As(err, &certInvalid)
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
