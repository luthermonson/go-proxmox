package proxmox

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// togglingVersionServer is reachable only while up is true. While it is false
// the handler hijacks and drops the connection so the client sees a transport
// failure (a node that is down); once flipped true it answers /version and
// counts the hit, so a probe or request that lands on it is observable.
func togglingVersionServer(t *testing.T, up *atomic.Bool, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !up.Load() {
			if hj, ok := w.(http.Hijacker); ok {
				if conn, _, err := hj.Hijack(); err == nil {
					_ = conn.Close()
				}
			}
			return
		}
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, versionJSON)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// downState reports whether endpoint idx is in a failing streak the breaker has
// not yet cleared. Unlike healthy(), which flips back lazily once a cooldown
// window elapses, benchedTil stays set until an actual success, so this is a
// stable signal to assert on.
func downState(f *failoverRoundTripper, idx int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return !f.endpoints[idx].benchedTil.IsZero()
}

// recovered reports whether endpoint idx has been cleared by a success (request
// or probe): no pending bench, not evicted, failure count reset.
func recovered(f *failoverRoundTripper, idx int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	ep := f.endpoints[idx]
	return ep.benchedTil.IsZero() && !ep.evicted && ep.failures == 0
}

// isEvicted reports whether endpoint idx is soft-evicted.
func isEvicted(f *failoverRoundTripper, idx int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.endpoints[idx].evicted
}

// TestFailover_HealthCheckRestoresRecoveredEndpoint verifies that with
// WithEndpointHealthCheck the background supervisor brings a recovered endpoint
// back into rotation on its own, with no user request hitting it, and that the
// goroutine then exits (probing clears) once everything is healthy.
func TestFailover_HealthCheckRestoresRecoveredEndpoint(t *testing.T) {
	var up atomic.Bool // endpoint 2 starts down
	var hits1, hits2 atomic.Int64
	srv1 := countingVersionServer(t, &hits1)
	srv2 := togglingVersionServer(t, &up, &hits2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(srv2.URL),
		WithEndpointStrategy(EndpointRoundRobin),
		WithEndpointFailureThreshold(1),
		WithEndpointCooldown(20*time.Millisecond, 40*time.Millisecond),
		WithEndpointHealthCheck(ctx))
	f := client.failover
	require.NotNil(t, f)

	// A few requests bench the dead endpoint; every request still succeeds via
	// the live primary.
	for range 4 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	require.Eventually(t, func() bool { return downState(f, 1) }, time.Second, 10*time.Millisecond,
		"the dead endpoint should be benched")
	require.Eventually(t, func() bool { return f.probing.Load() }, time.Second, 5*time.Millisecond,
		"the health-check supervisor should be running while an endpoint is down")

	// Bring the endpoint back. With NO further user requests, the poller must
	// probe and restore it — recovered() only holds after a successful probe.
	up.Store(true)
	require.Eventually(t, func() bool { return recovered(f, 1) }, 2*time.Second, 10*time.Millisecond,
		"the recovered endpoint should return to rotation via the background probe")
	assert.Positive(t, hits2.Load(), "the supervisor must have probed the recovered endpoint")

	// Once everything is healthy the goroutine exits on its own.
	require.Eventually(t, func() bool { return !f.probing.Load() }, 2*time.Second, 10*time.Millisecond,
		"the supervisor should self-exit once all endpoints are healthy")
}

// TestFailover_HealthCheckEvictsAndRestores verifies WithEndpointMaxDowntime:
// an endpoint down past the max is soft-evicted, the last live endpoint is
// never evicted, and a recovered evicted endpoint is restored.
func TestFailover_HealthCheckEvictsAndRestores(t *testing.T) {
	var up atomic.Bool // endpoint 2 starts down
	var hits1, hits2 atomic.Int64
	srv1 := countingVersionServer(t, &hits1)
	srv2 := togglingVersionServer(t, &up, &hits2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(srv2.URL),
		WithEndpointStrategy(EndpointRoundRobin),
		WithEndpointFailureThreshold(1),
		WithEndpointCooldown(10*time.Millisecond, 20*time.Millisecond),
		WithEndpointHealthCheck(ctx),
		WithEndpointMaxDowntime(50*time.Millisecond))
	f := client.failover
	require.NotNil(t, f)

	for range 4 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool { return isEvicted(f, 1) }, 2*time.Second, 10*time.Millisecond,
		"an endpoint down past the max downtime should be soft-evicted")
	assert.False(t, isEvicted(f, 0), "the last live endpoint must never be evicted")

	// Recovery restores an evicted endpoint.
	up.Store(true)
	require.Eventually(t, func() bool { return recovered(f, 1) }, 2*time.Second, 10*time.Millisecond,
		"a recovered evicted endpoint should be restored to rotation")
}

// TestFailover_NoHealthCheckByDefault verifies the opt-in contract: without
// WithEndpointHealthCheck no background goroutine runs, even with a benched
// endpoint — the breaker stays lazy exactly as before.
func TestFailover_NoHealthCheckByDefault(t *testing.T) {
	var hits1 atomic.Int64
	srv1 := countingVersionServer(t, &hits1)
	dead := deadEndpoint(t)

	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(dead),
		WithEndpointStrategy(EndpointRoundRobin),
		WithEndpointFailureThreshold(1),
		WithEndpointCooldown(10*time.Millisecond, 20*time.Millisecond))
	f := client.failover
	require.NotNil(t, f)

	for range 4 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	require.Eventually(t, func() bool { return downState(f, 1) }, time.Second, 10*time.Millisecond)
	assert.False(t, f.probing.Load(), "no supervisor goroutine should run without WithEndpointHealthCheck")
	time.Sleep(60 * time.Millisecond)
	assert.False(t, f.probing.Load(), "still no supervisor goroutine after the cooldown elapses")
}

// TestFailover_HealthCheckStopsOnContextCancel verifies the supervisor stops
// when the health-check context is cancelled, so a client whose endpoint stays
// down does not leak a goroutine past its lifetime.
func TestFailover_HealthCheckStopsOnContextCancel(t *testing.T) {
	var up atomic.Bool // stays down for the whole test
	var hits1, hits2 atomic.Int64
	srv1 := countingVersionServer(t, &hits1)
	srv2 := togglingVersionServer(t, &up, &hits2)

	ctx, cancel := context.WithCancel(context.Background())
	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(srv2.URL),
		WithEndpointStrategy(EndpointRoundRobin),
		WithEndpointFailureThreshold(1),
		WithEndpointCooldown(10*time.Millisecond, 20*time.Millisecond),
		WithEndpointHealthCheck(ctx))
	f := client.failover
	require.NotNil(t, f)

	for range 4 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	require.Eventually(t, func() bool { return f.probing.Load() }, time.Second, 5*time.Millisecond,
		"the supervisor should be running while the endpoint is down")

	cancel()
	require.Eventually(t, func() bool { return !f.probing.Load() }, 2*time.Second, 10*time.Millisecond,
		"the supervisor should stop after the health-check context is cancelled")
}
