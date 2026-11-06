package proxmox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNode is one Proxmox node in a test cluster: a real httptest server
// with a swappable handler and a request counter.
type fakeNode struct {
	srv  *httptest.Server
	hits atomic.Int64

	mu      sync.Mutex
	handler http.HandlerFunc
}

func newFakeNode(t *testing.T) *fakeNode {
	t.Helper()
	n := &fakeNode{}
	n.handler = okHandler
	n.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.hits.Add(1)
		n.mu.Lock()
		h := n.handler
		n.mu.Unlock()
		h(w, r)
	}))
	t.Cleanup(n.srv.Close)
	return n
}

func (n *fakeNode) base() string { return n.srv.URL + "/api2/json" }

func (n *fakeNode) set(h http.HandlerFunc) {
	n.mu.Lock()
	n.handler = h
	n.mu.Unlock()
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"data":{"version":"9.0","release":"9.0","repoid":"x"}}`)
}

func statusHandler(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, _ = io.WriteString(w, `{"data":null}`)
	}
}

// hangUpHandler reads the request and then drops the connection without a
// response: the request reached the server, the client sees a transport error.
func hangUpHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	hj, ok := w.(http.Hijacker)
	if !ok {
		panic("hijack unsupported")
	}
	conn, _, err := hj.Hijack()
	if err == nil {
		_ = conn.Close()
	}
}

// dialGate makes dials to selected hosts fail with ECONNREFUSED before any
// byte is written, which is what a powered-off or rebooting node looks like.
type dialGate struct {
	mu   sync.Mutex
	dead map[string]bool
}

func (g *dialGate) kill(n *fakeNode) { g.setDead(n, true) }

func (g *dialGate) setDead(n *fakeNode, dead bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.dead == nil {
		g.dead = map[string]bool{}
	}
	g.dead[n.srv.Listener.Addr().String()] = dead
}

func (g *dialGate) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	g.mu.Lock()
	dead := g.dead[addr]
	g.mu.Unlock()
	if dead {
		return nil, &net.OpError{Op: "dial", Net: network, Err: syscall.ECONNREFUSED}
	}
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

// newFailoverClient builds a client whose primary is nodes[0] and whose
// fallbacks are the rest, on a private transport (so gock's global
// interception can't interfere) and with a controllable clock.
func newFailoverClient(t *testing.T, gate *dialGate, nodes []*fakeNode, opts ...Option) (*Client, *fakeClock) {
	t.Helper()
	var fallbacks []string
	for _, n := range nodes[1:] {
		fallbacks = append(fallbacks, n.base())
	}
	// No keep-alives, so a node killed by the gate can't be reached over a
	// connection left idle from an earlier request.
	tr := &http.Transport{DialContext: gate.dial, DisableKeepAlives: true}
	t.Cleanup(tr.CloseIdleConnections)
	all := append([]Option{
		WithHTTPClient(&http.Client{Transport: tr}),
		WithFallbackEndpoints(fallbacks...),
	}, opts...)
	c := NewClient(nodes[0].base(), all...)
	require.NotNil(t, c.failover, "failover wrapper not installed")
	clock := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	c.failover.now = clock.now
	return c, clock
}

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

func TestFailover_PrimaryDownUsesFallback(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	v, err := c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "9.0", v.Version)
	assert.Equal(t, int64(1), b.hits.Load())

	// The primary is now cooling down, so the next call goes straight to b.
	assert.Equal(t, b.base(), c.endpointBaseURL())
	_, err = c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), b.hits.Load())
}

func TestFailover_RecoversAfterCooldown(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	a.set(statusHandler(http.StatusServiceUnavailable))
	c, clock := newFailoverClient(t, gate, []*fakeNode{a, b}, WithEndpointCooldown(10*time.Second, time.Minute))

	_, err := c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), a.hits.Load())
	assert.Equal(t, int64(1), b.hits.Load())

	// Inside the cooldown the primary is skipped.
	a.set(okHandler)
	clock.advance(5 * time.Second)
	_, err = c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), a.hits.Load())
	assert.Equal(t, int64(2), b.hits.Load())

	// After it, the primary is tried first again and keeps the traffic.
	clock.advance(6 * time.Second)
	for i := 0; i < 2; i++ {
		_, err = c.Version(context.Background())
		require.NoError(t, err)
	}
	assert.Equal(t, int64(3), a.hits.Load())
	assert.Equal(t, int64(2), b.hits.Load())
	assert.Equal(t, a.base(), c.endpointBaseURL())
}

func TestFailover_CooldownBacksOff(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	c, clock := newFailoverClient(t, gate, []*fakeNode{a, b}, WithEndpointCooldown(10*time.Second, 25*time.Second))

	cooldown := func() time.Duration {
		c.failover.mu.Lock()
		defer c.failover.mu.Unlock()
		return c.failover.endpoints[0].unhealthyTil.Sub(clock.now())
	}

	_, err := c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cooldown())

	clock.advance(10 * time.Second)
	_, err = c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 20*time.Second, cooldown())

	clock.advance(20 * time.Second)
	_, err = c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 25*time.Second, cooldown(), "capped at max")
}

func TestFailover_PostWithResponseIsNotReplayed(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	a.set(statusHandler(http.StatusServiceUnavailable))
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	// handleResponse doesn't currently turn a bare 503 into an error, so
	// assert on what reached each node rather than on the returned error.
	_ = c.Post(context.Background(), "/nodes/node1/qemu", map[string]int{"vmid": 101}, nil)
	assert.Equal(t, int64(1), a.hits.Load())
	assert.Equal(t, int64(0), b.hits.Load(), "a POST that got a response must not be sent to another node")
}

func TestFailover_PostLostAfterWriteIsNotReplayed(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			gate := &dialGate{}
			a, b := newFakeNode(t), newFakeNode(t)
			a.set(hangUpHandler)
			c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

			err := c.Req(context.Background(), method, "/nodes/node1/qemu/101/status/start", []byte(`{}`), nil)
			require.Error(t, err)
			assert.Equal(t, int64(1), a.hits.Load())
			assert.Equal(t, int64(0), b.hits.Load(), "a %s that reached a server must not be replayed", method)
		})
	}

	t.Run("GET", func(t *testing.T) {
		gate := &dialGate{}
		a, b := newFakeNode(t), newFakeNode(t)
		a.set(hangUpHandler)
		c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

		_, err := c.Version(context.Background())
		require.NoError(t, err, "the same failure on a GET is safe to replay")
		assert.Equal(t, int64(1), a.hits.Load())
		assert.Equal(t, int64(1), b.hits.Load())
	})
}

func TestFailover_PostThatNeverConnectedFailsOver(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	var gotBody string
	b.set(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = io.WriteString(w, `{"data":"UPID:node1:00000001:00000001:00000001:qmcreate:101:root@pam:"}`)
	})
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	var upid string
	err := c.Post(context.Background(), "/nodes/node1/qemu", map[string]int{"vmid": 101}, &upid)
	require.NoError(t, err)
	assert.Equal(t, int64(1), b.hits.Load())
	assert.Equal(t, `{"vmid":101}`, gotBody, "body is replayed intact")
	assert.Contains(t, upid, "qmcreate")
}

func TestFailover_ClientErrorsDoNotFailOver(t *testing.T) {
	for _, code := range []int{
		http.StatusBadRequest, http.StatusNotFound, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusNotImplemented,
	} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			gate := &dialGate{}
			a, b := newFakeNode(t), newFakeNode(t)
			a.set(statusHandler(code))
			c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

			_ = c.Get(context.Background(), "/nodes", nil)
			assert.Equal(t, int64(1), a.hits.Load())
			assert.Equal(t, int64(0), b.hits.Load())
			assert.Equal(t, a.base(), c.endpointBaseURL(), "primary must not be benched")
		})
	}
}

func TestFailover_ProxmoxProxyStatusesFailOver(t *testing.T) {
	for _, code := range []int{595, 596, http.StatusBadGateway, http.StatusGatewayTimeout} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			gate := &dialGate{}
			a, b := newFakeNode(t), newFakeNode(t)
			a.set(statusHandler(code))
			c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

			_, err := c.Version(context.Background())
			require.NoError(t, err)
			assert.Equal(t, int64(1), a.hits.Load())
			assert.Equal(t, int64(1), b.hits.Load())

			// 595/596 usually mean the target node is down, not the entry
			// node, so they don't bench the endpoint; gateway errors do.
			benched := c.endpointBaseURL() != a.base()
			assert.Equal(t, code < 595, benched)
		})
	}
}

func TestFailover_AllEndpointsDownStillTriesEach(t *testing.T) {
	gate := &dialGate{}
	a, b, d := newFakeNode(t), newFakeNode(t), newFakeNode(t)
	for _, n := range []*fakeNode{a, b, d} {
		n.set(statusHandler(http.StatusServiceUnavailable))
	}
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b, d})

	for i := 1; i <= 2; i++ {
		_ = c.Get(context.Background(), "/nodes", nil)
		for _, n := range []*fakeNode{a, b, d} {
			assert.Equal(t, int64(i), n.hits.Load())
		}
	}

	// One node comes back while all are still benched: it is found.
	b.set(okHandler)
	_, err := c.Version(context.Background())
	require.NoError(t, err)
}

func TestFailover_ComposesWithRetry(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	a.set(statusHandler(http.StatusServiceUnavailable))
	b.set(statusHandler(http.StatusServiceUnavailable))
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b}, WithRetry(WithRetryMax(3)))
	rt, ok := c.httpClient.Transport.(*retryRoundTripper)
	require.True(t, ok, "retry must wrap failover, got %T", c.httpClient.Transport)
	_, ok = rt.base.(*failoverRoundTripper)
	require.True(t, ok)
	rt.policy.sleep = fastNoopSleep

	_ = c.Get(context.Background(), "/nodes", nil)
	// 3 retry rounds x 2 endpoints.
	assert.Equal(t, int64(3), a.hits.Load())
	assert.Equal(t, int64(3), b.hits.Load())
}

func TestFailover_RespectsContext(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.Get(ctx, "/version", nil)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, int64(0), b.hits.Load())
}

func TestFailover_DeadlineWhileConnectingBenchesEndpoint(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	tr := &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == a.srv.Listener.Addr().String() {
			<-ctx.Done() // a black-holed host: the SYN is never answered
			return nil, ctx.Err()
		}
		return gate.dial(ctx, network, addr)
	}}
	t.Cleanup(tr.CloseIdleConnections)
	c := NewClient(a.base(), WithHTTPClient(&http.Client{Transport: tr}), WithFallbackEndpoints(b.base()))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := c.Get(ctx, "/version", nil)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	// The next call doesn't hang on the same host.
	_, err = c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), b.hits.Load())
}

func TestFailover_RoundRobin(t *testing.T) {
	gate := &dialGate{}
	nodes := []*fakeNode{newFakeNode(t), newFakeNode(t), newFakeNode(t)}
	c, _ := newFailoverClient(t, gate, nodes, WithEndpointStrategy(EndpointRoundRobin))

	for i := 0; i < 9; i++ {
		_, err := c.Version(context.Background())
		require.NoError(t, err)
	}
	for _, n := range nodes {
		assert.Equal(t, int64(3), n.hits.Load())
	}

	// A dead node drops out of the rotation without failing calls.
	gate.kill(nodes[1])
	for i := 0; i < 6; i++ {
		_, err := c.Version(context.Background())
		require.NoError(t, err)
	}
	assert.Equal(t, int64(3), nodes[1].hits.Load())
	assert.Equal(t, int64(12), nodes[0].hits.Load()+nodes[2].hits.Load())
}

func TestFailover_ConcurrentUse(t *testing.T) {
	gate := &dialGate{}
	nodes := []*fakeNode{newFakeNode(t), newFakeNode(t), newFakeNode(t)}
	gate.kill(nodes[0])
	nodes[1].set(statusHandler(596))
	c, clock := newFailoverClient(t, gate, nodes, WithEndpointStrategy(EndpointRoundRobin))

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 4; j++ {
				if i%10 == 0 {
					clock.advance(time.Minute)
					gate.setDead(nodes[0], j%2 == 0)
				}
				// Decode into a local: Client.Version writes the shared
				// c.version field and is not itself safe for concurrent use.
				var v Version
				if err := c.Get(context.Background(), "/version", &v); err != nil {
					errs <- err
				}
				_ = c.endpointBaseURL()
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFailover_CredentialsAuthAcrossNodes(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	b.set(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/access/ticket") {
			_, _ = io.WriteString(w, `{"data":{"username":"root@pam","ticket":"PVE:root@pam:T","CSRFPreventionToken":"CSRF"}}`)
			return
		}
		if r.Header.Get("Cookie") != "PVEAuthCookie=PVE:root@pam:T" || r.Header.Get("CSRFPreventionToken") != "CSRF" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		okHandler(w, r)
	})
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b},
		WithCredentials(&Credentials{Username: "root@pam", Password: "secret"}))

	v, err := c.Version(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "9.0", v.Version)
	assert.Equal(t, "PVE:root@pam:T", c.Session().Ticket)
}

func TestFailover_AbsoluteURLsAreNotRedirected(t *testing.T) {
	gate := &dialGate{}
	a, b, other := newFakeNode(t), newFakeNode(t), newFakeNode(t)
	other.set(statusHandler(http.StatusServiceUnavailable))
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	_ = c.Get(context.Background(), other.base()+"/version", nil)
	assert.Equal(t, int64(1), other.hits.Load())
	assert.Equal(t, int64(0), a.hits.Load()+b.hits.Load())
}

func TestFailover_NonReplayableBodyIsPinned(t *testing.T) {
	gate := &dialGate{}
	a, b := newFakeNode(t), newFakeNode(t)
	gate.kill(a)
	c, _ := newFailoverClient(t, gate, []*fakeNode{a, b})

	err := c.UploadReader("/nodes/node1/storage/local/upload", map[string]string{"content": "iso"},
		"x.iso", strings.NewReader("data"), 4, nil)
	require.Error(t, err)
	assert.Equal(t, int64(0), b.hits.Load())
}

func TestFailover_Options(t *testing.T) {
	t.Run("no fallbacks leaves the transport alone", func(t *testing.T) {
		c := NewClient("https://pve1:8006/api2/json")
		assert.Nil(t, c.failover)
		assert.Equal(t, "https://pve1:8006/api2/json", c.endpointBaseURL())
	})
	t.Run("duplicates and blanks are dropped", func(t *testing.T) {
		c := NewClient("https://pve1:8006/api2/json/",
			WithFallbackEndpoints("https://pve1:8006/api2/json", " ", "https://pve2:8006/api2/json/"),
			WithFallbackEndpoints("https://pve2:8006/api2/json", "not a url"))
		require.NotNil(t, c.failover)
		var bases []string
		for _, ep := range c.failover.endpoints {
			bases = append(bases, ep.base)
		}
		assert.Equal(t, []string{"https://pve1:8006/api2/json", "https://pve2:8006/api2/json"}, bases)
	})
	t.Run("only the primary given is a no-op", func(t *testing.T) {
		c := NewClient("https://pve1:8006", WithFallbackEndpoints("https://pve1:8006/"))
		assert.Nil(t, c.failover)
	})
	t.Run("suffix matching respects host boundaries", func(t *testing.T) {
		c := NewClient("https://pve1:8006", WithFallbackEndpoints("https://pve2:8006"))
		require.NotNil(t, c.failover)
		req, _ := http.NewRequest(http.MethodGet, "https://pve1:80067/api2/json/version", nil)
		_, ok := c.failover.suffix(req.URL)
		assert.False(t, ok)
		req, _ = http.NewRequest(http.MethodGet, "https://pve1:8006/api2/json/version?x=1", nil)
		s, ok := c.failover.suffix(req.URL)
		assert.True(t, ok)
		assert.Equal(t, "/api2/json/version?x=1", s)
	})
	t.Run("websocket helpers can find the real transport", func(t *testing.T) {
		tr := &http.Transport{}
		c := NewClient("https://pve1:8006", WithHTTPClient(&http.Client{Transport: tr}),
			WithFallbackEndpoints("https://pve2:8006"), WithRetry())
		assert.Same(t, tr, unwrapTransport(c.httpClient.Transport))
		assert.Nil(t, unwrapTransport(nil))
	})
}

func TestClassifyFailover(t *testing.T) {
	retry, bench, _ := classifyFailover(nil, errors.New("dial tcp: connection refused"))
	assert.True(t, retry)
	assert.True(t, bench)
	for code, want := range map[int][2]bool{
		200: {false, false}, 400: {false, false}, 401: {false, false}, 404: {false, false},
		429: {false, false}, 500: {false, false}, 501: {false, false},
		502: {true, true}, 503: {true, true}, 504: {true, true},
		595: {true, false}, 596: {true, false},
	} {
		retry, bench, _ := classifyFailover(&http.Response{StatusCode: code, Status: fmt.Sprint(code)}, nil)
		assert.Equal(t, want, [2]bool{retry, bench}, "status %d", code)
	}
}
