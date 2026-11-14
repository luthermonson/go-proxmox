package proxmox

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

// newTestFailover builds a failoverRoundTripper directly for unit tests, with a
// caller-controlled clock and circuit-breaker tunables.
func newTestFailover(now func() time.Time, threshold int, base, max time.Duration, bases ...string) *failoverRoundTripper {
	f := &failoverRoundTripper{
		base:         http.DefaultTransport,
		log:          &LeveledLogger{Level: LevelError},
		now:          now,
		threshold:    threshold,
		cooldownBase: base,
		cooldownMax:  max,
		endpoints:    make([]*endpoint, len(bases)),
	}
	for i, b := range bases {
		f.endpoints[i] = &endpoint{base: b}
	}
	return f
}

const versionJSON = `{"data":{"release":"9.1","version":"9.1-1","repoid":"9a1b2c3d"}}`

// countingVersionServer returns an httptest.Server that answers GET/POST/etc.
// to /version with a fixed payload and increments hits on every request.
func countingVersionServer(t *testing.T, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, versionJSON)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// deadEndpoint spins up a server, grabs its URL, then closes it so that
// connections to that URL are refused — a reliable "node is down" address.
func deadEndpoint(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()
	return url
}

// hangupAfterReadServer accepts the request (so the client finishes writing its
// headers and body) and then drops the connection without sending any HTTP
// response, simulating a node that dies mid-request after delivery.
func hangupAfterReadServer(t *testing.T, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			return
		}
		_ = conn.Close()
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestFailover_GetOnDeadPrimary verifies a GET fails over from an unreachable
// primary to a healthy fallback, and that the fallback becomes the current
// endpoint afterwards.
func TestFailover_GetOnDeadPrimary(t *testing.T) {
	var liveHits atomic.Int64
	live := countingVersionServer(t, &liveHits)
	dead := deadEndpoint(t)

	client := NewClient(dead, WithAdditionalEndpoints(live.URL))
	require.NotNil(t, client.failover, "failover wrapper should be installed with two endpoints")

	ver, err := client.Version(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ver)
	assert.Equal(t, "9.1-1", ver.Version)
	assert.Equal(t, int64(1), liveHits.Load(), "fallback should have served exactly one request")

	// The live endpoint is now the current/sticky one.
	assert.Equal(t, live.URL, client.failover.currentBase())

	// A second request goes straight to the now-current endpoint.
	_, err = client.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), liveHits.Load())
}

// TestFailover_NoFailoverOn500 verifies a well-formed HTTP response (500) from a
// reachable primary is returned to the caller and never fails over, even for a
// GET — a 500 is a real answer, not an unreachable node.
func TestFailover_NoFailoverOn500(t *testing.T) {
	var primaryHits, fallbackHits atomic.Int64
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryHits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"data":null}`)
	}))
	t.Cleanup(primary.Close)
	fallback := countingVersionServer(t, &fallbackHits)

	client := NewClient(primary.URL, WithAdditionalEndpoints(fallback.URL))

	_, err := client.Version(context.Background())
	require.Error(t, err, "a 500 must surface as an error, not succeed via failover")
	assert.Equal(t, int64(1), primaryHits.Load(), "primary answered once")
	assert.Equal(t, int64(0), fallbackHits.Load(), "a 500 is a real answer; must NOT fail over")
	// Current endpoint stays the primary (it answered).
	assert.Equal(t, primary.URL, client.failover.currentBase())
}

// TestFailover_GetFailsOverAfterDelivery verifies a GET does fail over even
// after the request reached a node (headers written) and the connection then
// dropped — a GET has no side effect to double-apply.
func TestFailover_GetFailsOverAfterDelivery(t *testing.T) {
	var primaryHits, fallbackHits atomic.Int64
	primary := hangupAfterReadServer(t, &primaryHits)
	fallback := countingVersionServer(t, &fallbackHits)

	client := NewClient(primary.URL, WithAdditionalEndpoints(fallback.URL))

	ver, err := client.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "9.1-1", ver.Version)
	assert.Equal(t, int64(1), primaryHits.Load())
	assert.Equal(t, int64(1), fallbackHits.Load(), "a GET is safe to fail over after delivery")
}

// TestFailover_AllEndpointsDead verifies the last transport error is returned
// when no endpoint is reachable.
func TestFailover_AllEndpointsDead(t *testing.T) {
	d1 := deadEndpoint(t)
	d2 := deadEndpoint(t)

	client := NewClient(d1, WithAdditionalEndpoints(d2))
	_, err := client.Version(context.Background())
	require.Error(t, err)
}

// TestFailover_ContextCancelStopsFailover verifies a cancelled context aborts
// immediately and does not sweep further endpoints.
func TestFailover_ContextCancelStopsFailover(t *testing.T) {
	var fallbackHits atomic.Int64
	fallback := countingVersionServer(t, &fallbackHits)
	dead := deadEndpoint(t)

	client := NewClient(dead, WithAdditionalEndpoints(fallback.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Version(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int64(0), fallbackHits.Load(), "a cancelled context must not reach any endpoint")
}

// TestNewClient_SingleEndpointUnchanged verifies the single-endpoint client is
// byte-for-byte the old behaviour: no failover wrapper, transport untouched.
func TestNewClient_SingleEndpointUnchanged(t *testing.T) {
	c := NewClient(TestURI)
	assert.Nil(t, c.failover, "a single-endpoint client must not install a failover wrapper")
	assert.Same(t, http.DefaultClient, c.httpClient, "single-endpoint client keeps the default http client")
	assert.Equal(t, TestURI, c.endpointBaseURL())
}

// TestWithAdditionalEndpoints_InvalidAndDuplicateIgnored verifies invalid or
// duplicate fallbacks are dropped; with no usable fallback left, no wrapper is
// installed and behaviour is unchanged.
func TestWithAdditionalEndpoints_InvalidAndDuplicateIgnored(t *testing.T) {
	// Only invalid + a duplicate of the primary -> nothing usable -> no wrapper.
	c := NewClient(TestURI, WithAdditionalEndpoints("", "not a url", TestURI))
	assert.Nil(t, c.failover, "no usable fallback should leave failover disabled")

	// A trailing slash on a fallback is normalised and still counts.
	c2 := NewClient(TestURI, WithAdditionalEndpoints("http://pve2:8006/"))
	require.NotNil(t, c2.failover)
	assert.Equal(t, []string{TestURI, "http://pve2:8006"}, endpointBases(c2.failover))
}

// endpointBases returns the ordered base URLs configured on a failover tripper.
func endpointBases(f *failoverRoundTripper) []string {
	out := make([]string, len(f.endpoints))
	for i, ep := range f.endpoints {
		out[i] = ep.base
	}
	return out
}

func TestFailoverSuffix(t *testing.T) {
	f := &failoverRoundTripper{endpoints: []*endpoint{{base: "https://pve1:8006"}}}
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://pve1:8006/nodes/x", "/nodes/x", true},
		{"https://pve1:8006", "", true},
		{"https://pve1:8006?foo=bar", "?foo=bar", true},
		{"https://pve1:80067/nodes/x", "", false},
		{"https://other:8006/nodes/x", "", false},
	}
	for _, tc := range cases {
		u, err := url.Parse(tc.in)
		require.NoError(t, err)
		got, ok := f.suffix(u)
		assert.Equal(t, tc.ok, ok, tc.in)
		assert.Equal(t, tc.want, got, tc.in)
	}
}

func TestIsSafeMethod(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		assert.True(t, isSafeMethod(m), m)
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		assert.False(t, isSafeMethod(m), m)
	}
}

func TestUnwrapTransport(t *testing.T) {
	assert.Nil(t, unwrapTransport(nil))

	base := &http.Transport{}
	assert.Same(t, base, unwrapTransport(base))

	// retry -> failover -> transport should unwrap all the way down.
	f := &failoverRoundTripper{base: base}
	r := &retryRoundTripper{base: f}
	assert.Same(t, base, unwrapTransport(r))
}

// --- #1 circuit breaker / cooldown -----------------------------------------

// TestFailover_CircuitBreakerThresholdAndReprobe verifies: a sub-threshold blip
// does not bench an endpoint; crossing the threshold benches it (pushed to the
// last-resort slot); it is re-probed after the window; and a success clears it.
func TestFailover_CircuitBreakerThresholdAndReprobe(t *testing.T) {
	origin := time.Unix(1_000_000, 0)
	clk := origin
	f := newTestFailover(func() time.Time { return clk }, 3, 10*time.Second, 180*time.Second,
		"http://a", "http://b")

	// Two failures (< threshold of 3) must not bench endpoint 0.
	f.markFailed(0, errBoom)
	f.markFailed(0, errBoom)
	assert.Equal(t, []int{0, 1}, f.order(), "a blip below the threshold must not bench")

	// Third consecutive failure crosses the threshold -> benched, tried last.
	f.markFailed(0, errBoom)
	assert.Equal(t, []int{1, 0}, f.order(), "endpoint benched after threshold is tried last")

	// Still benched inside the 10s window.
	clk = origin.Add(9 * time.Second)
	assert.Equal(t, []int{1, 0}, f.order())

	// Re-probed once the window expires.
	clk = origin.Add(11 * time.Second)
	assert.Equal(t, []int{0, 1}, f.order(), "endpoint re-probed after cooldown")

	// A success clears the failure state entirely.
	clk = origin
	f.markFailed(0, errBoom) // bench again
	require.Equal(t, []int{1, 0}, f.order())
	f.markOK(0)
	assert.Equal(t, []int{0, 1}, f.order(), "a success clears the bench")
}

// TestFailover_CooldownBackoffGrows verifies the bench window grows
// exponentially from base, capped at max.
func TestFailover_CooldownBackoffGrows(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	f := newTestFailover(func() time.Time { return now }, 1, 10*time.Second, 40*time.Second,
		"http://a", "http://b")

	var got []time.Duration
	for i := 0; i < 4; i++ {
		f.markFailed(0, errBoom)
		f.mu.Lock()
		got = append(got, f.endpoints[0].benchedTil.Sub(now))
		f.mu.Unlock()
	}
	// threshold 1, failures 1..4 -> 10s, 20s, 40s (cap), 40s (cap)
	assert.Equal(t, []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second, 40 * time.Second}, got)
}

// TestFailover_BenchedEndpointSkippedWithoutDialing verifies a benched endpoint
// is skipped without being dialed during its cooldown, even when it is the
// current endpoint, and is re-probed after the window expires.
func TestFailover_BenchedEndpointSkippedWithoutDialing(t *testing.T) {
	var deadHits, liveHits atomic.Int64
	dead := hangupAfterReadServer(t, &deadHits)
	live := countingVersionServer(t, &liveHits)

	client := NewClient(dead.URL,
		WithAdditionalEndpoints(live.URL),
		WithEndpointFailureThreshold(1),
		WithEndpointCooldown(30*time.Second, 30*time.Second),
	)
	require.NotNil(t, client.failover)

	var clk atomic.Int64
	clk.Store(time.Now().UnixNano())
	client.failover.now = func() time.Time { return time.Unix(0, clk.Load()) }

	// Trip the breaker: the GET hits the dead endpoint once, benches it (threshold
	// 1), then fails over to the live one.
	_, err := client.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), deadHits.Load())

	// Even forced back to current, the benched dead endpoint is not dialed again
	// during the cooldown window.
	for i := 0; i < 5; i++ {
		client.failover.markCurrent(0)
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	assert.Equal(t, int64(1), deadHits.Load(), "benched endpoint must not be dialed during cooldown")

	// After the window expires it is re-probed (dialed once more).
	clk.Add(int64(31 * time.Second))
	client.failover.markCurrent(0)
	_, err = client.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), deadHits.Load(), "benched endpoint must be re-probed after the window")
}

// --- #2 shared cursor across concurrent requests ---------------------------

// TestFailover_SharedCursorNoThunderingHerd verifies the sticky current
// endpoint is shared state: once one request has moved it to the healthy node,
// concurrent callers all start there and none re-dials the dead endpoint.
func TestFailover_SharedCursorNoThunderingHerd(t *testing.T) {
	var deadHits, liveHits atomic.Int64
	dead := hangupAfterReadServer(t, &deadHits)
	live := countingVersionServer(t, &liveHits)

	client := NewClient(dead.URL, WithAdditionalEndpoints(live.URL))

	// Warm up once so the shared cursor moves to the live endpoint.
	_, err := client.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), deadHits.Load())
	require.Equal(t, live.URL, client.failover.currentBase())

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			// Use a per-goroutine destination (not Version, which writes the
			// shared c.version field) so this test isolates failover state.
			var v Version
			assert.NoError(t, client.Get(context.Background(), "/version", &v))
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(1), deadHits.Load(),
		"concurrent requests must share the cursor and not re-dial the dead endpoint")
	assert.Equal(t, int64(1+n), liveHits.Load())
}

// --- #3 connect-error vs ambiguous-error idempotency split -----------------

func TestIsConnectError(t *testing.T) {
	assert.False(t, isConnectError(nil))
	assert.False(t, isConnectError(errBoom))

	// Dial failures (connection refused, unreachable, dial timeout).
	assert.True(t, isConnectError(&net.OpError{Op: "dial", Err: errors.New("connect: connection refused")}))
	// A mid-flight error after the connection was established is NOT a connect error.
	assert.False(t, isConnectError(&net.OpError{Op: "read", Err: errors.New("connection reset by peer")}))

	// DNS resolution failure.
	assert.True(t, isConnectError(&net.DNSError{Err: "no such host", Name: "pve.invalid"}))

	// connection refused from the syscall layer, bare and wrapped.
	assert.True(t, isConnectError(syscall.ECONNREFUSED))
	assert.True(t, isConnectError(&net.OpError{Op: "dial", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}))

	// TLS / certificate verification failures happen before the request is sent.
	assert.True(t, isConnectError(x509.UnknownAuthorityError{}))

	// Wrapping is unwrapped by errors.As / errors.Is.
	assert.True(t, isConnectError(fmt.Errorf("get %q: %w", "https://x", &net.OpError{Op: "dial", Err: errBoom})))
}

// TestFailover_ConnectRefusedPostFailsOver verifies a POST to an endpoint whose
// connection is refused (pure connect error, nothing written) fails over —
// strictly more capable than before, and safe because the write never left.
func TestFailover_ConnectRefusedPostFailsOver(t *testing.T) {
	var fallbackHits atomic.Int64
	fallback := countingVersionServer(t, &fallbackHits)
	dead := deadEndpoint(t)

	client := NewClient(dead, WithAdditionalEndpoints(fallback.URL))

	err := client.Req(context.Background(), http.MethodPost, "/cluster/foo", []byte(`{"a":1}`), nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), fallbackHits.Load(),
		"a connection-refused POST never left the client and is safe to fail over")
}

// TestFailover_HeadersWrittenPostNotReplayed verifies a POST that reached a node
// (headers written) and then lost the connection is NOT replayed — the server
// may already have acted.
func TestFailover_HeadersWrittenPostNotReplayed(t *testing.T) {
	var primaryHits, fallbackHits atomic.Int64
	primary := hangupAfterReadServer(t, &primaryHits)
	fallback := countingVersionServer(t, &fallbackHits)

	client := NewClient(primary.URL, WithAdditionalEndpoints(fallback.URL))

	err := client.Req(context.Background(), http.MethodPost, "/cluster/foo", []byte(`{"a":1}`), nil)
	require.Error(t, err, "a delivered POST that then failed must surface the error")
	assert.Equal(t, int64(1), primaryHits.Load(), "primary received the POST once")
	assert.Equal(t, int64(0), fallbackHits.Load(), "a maybe-delivered POST must NOT be replayed")
}
