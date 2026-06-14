package proxmox

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// retryTestURI is intentionally distinct from TestURI so gock matchers in this
// file don't collide with any persistent fixtures registered by other tests.
const retryTestURI = "http://retry.test.localhost"

// fastNoopSleep makes retry tests deterministic by collapsing all backoffs to
// near-zero while still respecting context cancellation. Individual tests that
// need to assert real sleep duration (e.g. Retry-After) install their own.
func fastNoopSleep(ctx context.Context, _ time.Duration) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// newRetryClient builds a *Client configured with WithRetry. It also installs
// the fastNoopSleep into the resulting policy so tests don't actually wait the
// backoff window.
func newRetryClient(t *testing.T, retryOpts []RetryOption) *Client {
	t.Helper()
	c := NewClient(retryTestURI, WithRetry(retryOpts...))
	rt, ok := c.httpClient.Transport.(*retryRoundTripper)
	require.True(t, ok, "expected retryRoundTripper, got %T", c.httpClient.Transport)
	rt.policy.sleep = fastNoopSleep
	return c
}

func TestWithRetry_SuccessAfterTransient503(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusServiceUnavailable).
		BodyString("first")
	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusServiceUnavailable).
		BodyString("second")
	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusOK).
		JSON(`{"data": {"release": "9.0", "version": "9.0.0"}}`)

	client := newRetryClient(t, nil)

	ver, err := client.Version(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ver)
	assert.Equal(t, "9.0.0", ver.Version)
	assert.True(t, gock.IsDone(), "expected all three mocks to be consumed")
}

func TestWithRetry_StopsAfterMaxAttempts(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Times(2).
		Reply(http.StatusServiceUnavailable)

	client := newRetryClient(t, []RetryOption{WithRetryMax(2)})

	_, err := client.Version(context.Background())
	// 503 surfaces from handleResponse as a "503 Service Unavailable"-shaped
	// error (StatusInternalServerError/NotImplemented are special-cased, but
	// 503 falls through to JSON parsing of an empty body).
	require.Error(t, err)
	assert.True(t, gock.IsDone(), "expected exactly two attempts")
}

func TestWithRetry_RespectsRetryAfterOn429(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusTooManyRequests).
		SetHeader("Retry-After", "1")
	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusOK).
		JSON(`{"data": {"release": "9.0", "version": "9.0.0"}}`)

	// Custom sleep that records the actual durations the policy asked for.
	var observed atomic.Int64
	client := NewClient(retryTestURI, WithRetry())
	rt := client.httpClient.Transport.(*retryRoundTripper)
	rt.policy.sleep = func(ctx context.Context, d time.Duration) error {
		observed.Store(int64(d))
		// Still sleep a tiny bit so the timing of the call is observable,
		// but never the full 1s the header asked for.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
			return nil
		}
	}

	_, err := client.Version(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, observed.Load(), int64(time.Second),
		"Retry-After: 1 should request at least 1s of backoff")
	assert.True(t, gock.IsDone())
}

func TestWithRetry_DoesNotRetry4xxOtherThan429(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusUnauthorized).
		JSON(`{}`)
	// Register a fallback success that the client should never consume.
	mockTwo := gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusOK).
		JSON(`{"data": {"release": "9.0", "version": "9.0.0"}}`)

	client := newRetryClient(t, nil)

	_, err := client.Version(context.Background())
	// 401 surfaces as ErrNotAuthorized via Req's auth handling.
	assert.ErrorIs(t, err, ErrNotAuthorized)
	assert.True(t, mockTwo.Mock.Request().Counter > 0,
		"second mock should still be registered (unconsumed)")

	// Tear down the unconsumed mock manually so the suite is clean.
	gock.Off()
}

func TestWithRetry_CustomConditionRetries418(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusTeapot)
	gock.New(retryTestURI).
		Get("^/version$").
		Reply(http.StatusOK).
		JSON(`{"data": {"release": "9.0", "version": "9.0.0"}}`)

	cond := func(res *http.Response, err error) bool {
		if err != nil {
			return true
		}
		return res != nil && res.StatusCode == http.StatusTeapot
	}
	client := newRetryClient(t, []RetryOption{WithRetryCondition(cond)})

	ver, err := client.Version(context.Background())
	require.NoError(t, err)
	require.NotNil(t, ver)
	assert.Equal(t, "9.0.0", ver.Version)
	assert.True(t, gock.IsDone(), "expected both mocks consumed")
}

func TestWithRetry_ContextCancellationAborts(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Persist().
		Reply(http.StatusServiceUnavailable)

	client := NewClient(retryTestURI, WithRetry(WithRetryMax(5)))
	rt := client.httpClient.Transport.(*retryRoundTripper)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel on the first sleep call — that simulates a caller pulling the
	// plug between attempts. Use the inner ctx (the one the request carries)
	// so the policy actually sees Done.
	rt.policy.sleep = func(c context.Context, _ time.Duration) error {
		cancel()
		return c.Err()
	}

	_, err := client.Version(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled),
		"expected error to wrap context.Canceled, got %v", err)
}

func TestWithRetry_POSTBodyResentIdentically(t *testing.T) {
	defer gock.Off()

	const body = `{"username":"root@pam","password":"hunter2"}`

	var firstBody, secondBody atomic.Value
	firstBody.Store("")
	secondBody.Store("")

	gock.New(retryTestURI).
		Post("^/access/ticket$").
		AddMatcher(func(req *http.Request, _ *gock.Request) (bool, error) {
			b, _ := io.ReadAll(req.Body)
			firstBody.Store(string(b))
			return true, nil
		}).
		Reply(http.StatusServiceUnavailable)

	gock.New(retryTestURI).
		Post("^/access/ticket$").
		AddMatcher(func(req *http.Request, _ *gock.Request) (bool, error) {
			b, _ := io.ReadAll(req.Body)
			secondBody.Store(string(b))
			return true, nil
		}).
		Reply(http.StatusOK).
		JSON(`{"data": {"ticket": "t", "CSRFPreventionToken": "c", "username": "root@pam"}}`)

	client := newRetryClient(t, nil)

	err := client.Req(context.Background(), http.MethodPost, "/access/ticket", []byte(body), nil)
	require.NoError(t, err, "expected the retry to succeed on the second attempt")

	assert.Equal(t, body, firstBody.Load().(string), "first attempt body")
	assert.Equal(t, body, secondBody.Load().(string), "retried body must match")
	assert.True(t, gock.IsDone())
}

// --- helper / unit-level tests --------------------------------------------

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		in   string
		want time.Duration
		ok   bool
	}{
		{"empty", "", 0, false},
		{"seconds", "5", 5 * time.Second, true},
		{"zero", "0", 0, true},
		{"negative seconds rejected", "-3", 0, false},
		{"http date in future", now.Add(3 * time.Second).UTC().Format(http.TimeFormat), 3 * time.Second, true},
		{"http date in past clamps to zero", now.Add(-1 * time.Hour).UTC().Format(http.TimeFormat), 0, true},
		{"garbage", "lol", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseRetryAfter(tc.in, now)
			assert.Equal(t, tc.ok, ok)
			if ok {
				// HTTP-date parsing rounds to whole seconds — tolerate ±1s.
				diff := got - tc.want
				if diff < 0 {
					diff = -diff
				}
				assert.LessOrEqual(t, diff, time.Second)
			}
		})
	}
}

func TestSnapshotBody_RewindsBytes(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://x", strings.NewReader("hello"))
	require.NoError(t, err)

	rewind, err := snapshotBody(req)
	require.NoError(t, err)
	require.NotNil(t, rewind)

	for i := 0; i < 3; i++ {
		rc := rewind()
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(b))
		_ = rc.Close()
	}
}

func TestSnapshotBody_NoBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://x", nil)
	require.NoError(t, err)
	rewind, err := snapshotBody(req)
	require.NoError(t, err)
	assert.Nil(t, rewind)
}

func TestIsIdempotent(t *testing.T) {
	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace} {
		assert.True(t, isIdempotent(m), "expected %s to be idempotent", m)
	}
	for _, m := range []string{http.MethodPost, http.MethodPatch, http.MethodConnect} {
		assert.False(t, isIdempotent(m), "expected %s to not be idempotent", m)
	}
}

func TestWithRetryMax_ClampsToOne(t *testing.T) {
	p := defaultRetryPolicy()
	WithRetryMax(-5)(p)
	assert.Equal(t, 1, p.maxAttempts)
}

func TestWithRetryBackoff_IgnoresNonPositive(t *testing.T) {
	p := defaultRetryPolicy()
	WithRetryBackoff(0, 0)(p)
	assert.Equal(t, defaultRetryInitialBackoff, p.initialBackoff)
	assert.Equal(t, defaultRetryMaxBackoff, p.maxBackoff)

	WithRetryBackoff(time.Second, 10*time.Second)(p)
	assert.Equal(t, time.Second, p.initialBackoff)
	assert.Equal(t, 10*time.Second, p.maxBackoff)
}

func TestWithRetryCondition_NilKeepsDefault(t *testing.T) {
	p := defaultRetryPolicy()
	WithRetryCondition(nil)(p)
	assert.True(t, p.condition(&http.Response{StatusCode: http.StatusServiceUnavailable}, nil))
}

// TestInstallRetryWrapper_PromotesDefaultClient verifies that wrapping the
// retry tripper onto a fresh client promotes c.httpClient off the shared
// http.DefaultClient singleton (via the shared ensureOwnHTTPClient helper),
// so installing the wrapper doesn't poke at package globals.
func TestInstallRetryWrapper_PromotesDefaultClient(t *testing.T) {
	c := &Client{httpClient: http.DefaultClient, retryPolicy: defaultRetryPolicy()}
	c.installRetryWrapper()
	assert.NotSame(t, http.DefaultClient, c.httpClient,
		"installRetryWrapper must not leave the shared DefaultClient in place")
	_, ok := c.httpClient.Transport.(*retryRoundTripper)
	assert.True(t, ok, "transport should be wrapped in *retryRoundTripper")
}
