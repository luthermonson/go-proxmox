package proxmox

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFailover_RoundRobinSpreadsRequests verifies that under EndpointRoundRobin
// consecutive requests rotate over every healthy endpoint, primary included,
// instead of sticking to the one that last answered.
func TestFailover_RoundRobinSpreadsRequests(t *testing.T) {
	var hits [3]atomic.Int64
	srv1 := countingVersionServer(t, &hits[0])
	srv2 := countingVersionServer(t, &hits[1])
	srv3 := countingVersionServer(t, &hits[2])

	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(srv2.URL, srv3.URL),
		WithEndpointStrategy(EndpointRoundRobin))
	require.NotNil(t, client.failover)

	for range 9 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	for i := range hits {
		assert.Equal(t, int64(3), hits[i].Load(), "endpoint %d should have served a third of the requests", i)
	}
}

// TestFailover_RoundRobinSkipsBenchedEndpoint verifies the rotation leaves a
// benched endpoint out: once the dead node has failed threshold times, the
// remaining requests are shared by the live nodes without dialing it again.
func TestFailover_RoundRobinSkipsBenchedEndpoint(t *testing.T) {
	var hits [2]atomic.Int64
	live1 := countingVersionServer(t, &hits[0])
	live2 := countingVersionServer(t, &hits[1])
	dead := deadEndpoint(t)

	client := NewClient(live1.URL,
		WithAdditionalEndpoints(dead, live2.URL),
		WithEndpointStrategy(EndpointRoundRobin),
		WithEndpointFailureThreshold(1))

	for range 6 {
		_, err := client.Version(context.Background())
		require.NoError(t, err, "a dead endpoint in the rotation must fail over, not fail the request")
	}
	assert.Equal(t, int64(6), hits[0].Load()+hits[1].Load(), "every request was answered by a live node")
	// Requests whose turn lands on the dead endpoint continue to the next live one in
	// cyclic order, so the live node after it serves more than its own share.
	assert.Equal(t, int64(2), hits[0].Load())
	assert.Equal(t, int64(4), hits[1].Load())

	// The dead endpoint is benched: its failure count reached the threshold
	// and it is skipped without being dialed while the others are healthy.
	f := client.failover
	f.mu.Lock()
	defer f.mu.Unlock()
	assert.False(t, f.endpoints[1].healthy(f.now()), "the dead endpoint should be benched")
}

// TestFailover_RoundRobinConcurrent verifies the rotation is safe and still
// even-handed when many goroutines share the client.
func TestFailover_RoundRobinConcurrent(t *testing.T) {
	var hits [3]atomic.Int64
	srv1 := countingVersionServer(t, &hits[0])
	srv2 := countingVersionServer(t, &hits[1])
	srv3 := countingVersionServer(t, &hits[2])

	client := NewClient(srv1.URL,
		WithAdditionalEndpoints(srv2.URL, srv3.URL),
		WithEndpointStrategy(EndpointRoundRobin))

	const n = 30
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			_, err := client.Version(context.Background())
			assert.NoError(t, err)
		})
	}
	wg.Wait()
	for i := range hits {
		assert.Equal(t, int64(n/3), hits[i].Load(), "endpoint %d", i)
	}
}

// TestFailover_DefaultStrategyIsSticky verifies the default is unchanged: with
// no strategy set, every request after the first stays on the primary.
func TestFailover_DefaultStrategyIsSticky(t *testing.T) {
	var primaryHits, otherHits atomic.Int64
	primary := countingVersionServer(t, &primaryHits)
	other := countingVersionServer(t, &otherHits)

	client := NewClient(primary.URL, WithAdditionalEndpoints(other.URL))
	for range 4 {
		_, err := client.Version(context.Background())
		require.NoError(t, err)
	}
	assert.Equal(t, int64(4), primaryHits.Load())
	assert.Equal(t, int64(0), otherHits.Load())
	assert.Equal(t, EndpointSticky, client.failover.strategy)
}

// TestOrder_RoundRobinCyclesFromCursor pins the ordering itself: each call
// starts one endpoint further along, healthy endpoints follow in cyclic order
// and benched ones trail as the last resort.
func TestOrder_RoundRobinCyclesFromCursor(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f := newTestFailover(func() time.Time { return fixed }, 1, time.Minute, time.Minute, "https://a", "https://b", "https://c")
	f.strategy = EndpointRoundRobin

	assert.Equal(t, []int{0, 1, 2}, f.order())
	assert.Equal(t, []int{1, 2, 0}, f.order())
	assert.Equal(t, []int{2, 0, 1}, f.order())
	assert.Equal(t, []int{0, 1, 2}, f.order())

	// Bench b: it drops to the end of every order while cooling down.
	f.markFailed(1, errBoom)
	assert.Equal(t, []int{2, 0, 1}, f.order())
	assert.Equal(t, []int{2, 0, 1}, f.order())
	assert.Equal(t, []int{0, 2, 1}, f.order())
}

func TestEndpointStrategyString(t *testing.T) {
	assert.Equal(t, "sticky", EndpointSticky.String())
	assert.Equal(t, "round-robin", EndpointRoundRobin.String())
	assert.Equal(t, "EndpointStrategy(7)", EndpointStrategy(7).String())
}

func TestParseEndpointStrategy(t *testing.T) {
	for in, want := range map[string]EndpointStrategy{
		"": EndpointSticky, "sticky": EndpointSticky, " Sticky ": EndpointSticky,
		"round-robin": EndpointRoundRobin, "RoundRobin": EndpointRoundRobin, "rr": EndpointRoundRobin,
	} {
		got, err := ParseEndpointStrategy(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err := ParseEndpointStrategy("random")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"random"`)
}
