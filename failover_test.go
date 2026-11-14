package proxmox

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// TestFailover_NonIdempotentNotReplayedAfterDelivery verifies a POST that was
// delivered to a node (headers written) but whose connection then dropped is
// NOT replayed on another endpoint — replaying could double-apply the write.
func TestFailover_NonIdempotentNotReplayedAfterDelivery(t *testing.T) {
	var primaryHits, fallbackHits atomic.Int64
	primary := hangupAfterReadServer(t, &primaryHits)
	fallback := countingVersionServer(t, &fallbackHits)

	client := NewClient(primary.URL, WithAdditionalEndpoints(fallback.URL))

	err := client.Req(context.Background(), http.MethodPost, "/cluster/foo", []byte(`{"a":1}`), nil)
	require.Error(t, err, "a POST that hit a node then lost the connection must return an error")
	assert.Equal(t, int64(1), primaryHits.Load(), "primary received the POST once")
	assert.Equal(t, int64(0), fallbackHits.Load(), "a delivered POST must NOT be replayed on the fallback")
}

// TestFailover_NonIdempotentFailsOverBeforeDelivery verifies the converse: a
// POST to an unreachable primary (connection never established, nothing
// written) IS safe to fail over, because the request provably never reached a
// node.
func TestFailover_NonIdempotentFailsOverBeforeDelivery(t *testing.T) {
	var fallbackHits atomic.Int64
	fallback := countingVersionServer(t, &fallbackHits)
	dead := deadEndpoint(t)

	client := NewClient(dead, WithAdditionalEndpoints(fallback.URL))

	err := client.Req(context.Background(), http.MethodPost, "/cluster/foo", []byte(`{"a":1}`), nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), fallbackHits.Load(), "a pre-delivery connection failure is safe to fail over")
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
	assert.Equal(t, []string{TestURI, "http://pve2:8006"}, c2.failover.endpoints)
}

func TestFailoverSuffix(t *testing.T) {
	f := &failoverRoundTripper{endpoints: []string{"https://pve1:8006"}}
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
