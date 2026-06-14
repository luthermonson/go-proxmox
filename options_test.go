package proxmox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithClient(t *testing.T) {
	httpClient := http.Client{Timeout: time.Second * 10}
	client := NewClient("", WithClient(&httpClient))
	assert.Equal(t, client.httpClient, &http.Client{Timeout: time.Second * 10})
}

func TestWithLogins(t *testing.T) {
	client := NewClient("", WithLogins("root@pam", "1234"))
	assert.Equal(t, client.credentials, &Credentials{Username: "root@pam", Password: "1234"})
}

func TestWithCredentials(t *testing.T) {
	client := NewClient("", WithCredentials(&Credentials{
		Username: "root@pam",
		Password: "1234",
	}))
	assert.Equal(t, client.credentials, &Credentials{Username: "root@pam", Password: "1234"})
}

func TestWithAPIToken(t *testing.T) {
	client := NewClient("", WithAPIToken("root@pam!test", "1234"))
	assert.Equal(t, client.token, "root@pam!test=1234")
}

func TestWithSession(t *testing.T) {
	client := NewClient("", WithSession("ticket", "csrf"))
	assert.Equal(t, client.session, &Session{Ticket: "ticket", CSRFPreventionToken: "csrf"})
}

func TestWithUserAgent(t *testing.T) {
	client := NewClient("", WithUserAgent("test-ua"))
	assert.Equal(t, client.userAgent, "test-ua")
}

func TestWithLogger(t *testing.T) {
	client := NewClient("", WithLogger(&LeveledLogger{Level: 1}))
	assert.Equal(t, client.log, &LeveledLogger{Level: 1})
}

// --- transport / TLS options ---------------------------------------------

func TestWithTimeout(t *testing.T) {
	client := NewClient("", WithTimeout(15*time.Second))
	assert.Equal(t, 15*time.Second, client.httpClient.Timeout)
	// must not have polluted the package-global default
	assert.Equal(t, time.Duration(0), http.DefaultClient.Timeout)
}

func TestWithTimeout_AppliesToCallerHTTPClient(t *testing.T) {
	custom := &http.Client{}
	client := NewClient("", WithHTTPClient(custom), WithTimeout(7*time.Second))
	assert.Equal(t, 7*time.Second, custom.Timeout)
	assert.Same(t, custom, client.httpClient)
}

func TestWithInsecureSkipVerify(t *testing.T) {
	client := NewClient("", WithInsecureSkipVerify())
	tr := client.httpClient.Transport.(*http.Transport)
	assert.NotNil(t, tr.TLSClientConfig)
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	// must not have polluted the global default
	defaultTr := http.DefaultTransport.(*http.Transport)
	assert.False(t, defaultTr.TLSClientConfig != nil && defaultTr.TLSClientConfig.InsecureSkipVerify,
		"WithInsecureSkipVerify must not mutate http.DefaultTransport")
}

func TestWithRootCAs(t *testing.T) {
	pool := x509.NewCertPool()
	client := NewClient("", WithRootCAs(pool))
	tr := client.httpClient.Transport.(*http.Transport)
	assert.Same(t, pool, tr.TLSClientConfig.RootCAs)
}

func TestWithRootCAFile_AppendsCerts(t *testing.T) {
	// Self-signed PEM generated once for this test. Any valid PEM works;
	// we don't need the corresponding private key because we only test that
	// AppendCertsFromPEM accepts it.
	const certPEM = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "ca.pem")
	assert.NoError(t, os.WriteFile(path, []byte(certPEM), 0o600))

	client := NewClient("", WithRootCAFile(path))
	tr := client.httpClient.Transport.(*http.Transport)
	assert.NotNil(t, tr.TLSClientConfig)
	assert.NotNil(t, tr.TLSClientConfig.RootCAs)
}

func TestWithRootCAFile_MissingFileLogsAndNoOps(t *testing.T) {
	// Missing file should not panic, should not crash, should not set RootCAs.
	client := NewClient("", WithRootCAFile("/does/not/exist.pem"))
	if client.httpClient.Transport != nil {
		if tr, ok := client.httpClient.Transport.(*http.Transport); ok && tr.TLSClientConfig != nil {
			assert.Nil(t, tr.TLSClientConfig.RootCAs)
		}
	}
}

func TestWithClientCertificate(t *testing.T) {
	cert := tls.Certificate{Certificate: [][]byte{{0x01, 0x02, 0x03}}}
	client := NewClient("", WithClientCertificate(cert))
	tr := client.httpClient.Transport.(*http.Transport)
	assert.Len(t, tr.TLSClientConfig.Certificates, 1)
}

func TestTLSOptions_Compose_OrderIndependent(t *testing.T) {
	pool := x509.NewCertPool()
	cert := tls.Certificate{Certificate: [][]byte{{0x01}}}

	// Order A: insecure → root → cert
	a := NewClient("", WithInsecureSkipVerify(), WithRootCAs(pool), WithClientCertificate(cert))
	atc := a.httpClient.Transport.(*http.Transport).TLSClientConfig
	assert.True(t, atc.InsecureSkipVerify)
	assert.Same(t, pool, atc.RootCAs)
	assert.Len(t, atc.Certificates, 1)

	// Order B: reversed — all three must still land
	b := NewClient("", WithClientCertificate(cert), WithRootCAs(pool), WithInsecureSkipVerify())
	btc := b.httpClient.Transport.(*http.Transport).TLSClientConfig
	assert.True(t, btc.InsecureSkipVerify)
	assert.Same(t, pool, btc.RootCAs)
	assert.Len(t, btc.Certificates, 1)
}

func TestTLSOptions_ComposeWithHTTPClient_BothOrders(t *testing.T) {
	custom := &http.Client{Transport: &http.Transport{}}
	// TLS option before WithHTTPClient
	NewClient("", WithInsecureSkipVerify(), WithHTTPClient(custom))
	tr := custom.Transport.(*http.Transport)
	assert.NotNil(t, tr.TLSClientConfig, "TLS option before WithHTTPClient should mutate the caller's transport")
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)

	custom2 := &http.Client{Transport: &http.Transport{}}
	NewClient("", WithHTTPClient(custom2), WithInsecureSkipVerify())
	tr2 := custom2.Transport.(*http.Transport)
	assert.NotNil(t, tr2.TLSClientConfig)
	assert.True(t, tr2.TLSClientConfig.InsecureSkipVerify)
}

// --- auth ergonomics ------------------------------------------------------

func TestWithOTP_StashesOnClient(t *testing.T) {
	client := NewClient("",
		WithCredentials(&Credentials{Username: "root@pam", Password: "pw"}),
		WithOTP("123456"),
	)
	assert.Equal(t, "123456", client.otp)
}

func TestWithOTP_ThreadedIntoTicketAndConsumed(t *testing.T) {
	mockConfig := mockConfig
	defer gock.Off()

	// Match the ticket POST and assert the body contains otp=123456.
	gock.New(mockConfig.URI).
		Post("/access/ticket").
		MatchType("json").
		AddMatcher(func(r *http.Request, _ *gock.Request) (bool, error) {
			body := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(body)
			return assert.ObjectsAreEqual(true, true) && containsByte(body, []byte(`"otp":"123456"`)), nil
		}).
		Reply(200).
		JSON(`{"data":{"ticket":"PVE:root@pam:0000:hex","CSRFPreventionToken":"csrf","username":"root@pam"}}`)

	client := NewClient(mockConfig.URI,
		WithCredentials(&Credentials{Username: "root@pam", Password: "pw"}),
		WithOTP("123456"),
	)
	assert.NoError(t, client.CreateSession(context.Background()))

	// OTP must have been consumed (zeroed out) so a subsequent re-auth
	// doesn't resend the same single-use code.
	assert.Equal(t, "", client.otp)
}

func TestWithDefaultRealm_AppendedWhenMissing(t *testing.T) {
	client := NewClient("", WithDefaultRealm("pam"))
	merged := mergeCredsForTest(client, &Credentials{Username: "root", Password: "x"})
	assert.Equal(t, "pam", merged.Realm)
	assert.Equal(t, "root", merged.Username, "username should not be rewritten when only Realm is filled in")
}

func TestWithDefaultRealm_NotAppendedWhenUsernameHasAtSuffix(t *testing.T) {
	client := NewClient("", WithDefaultRealm("pam"))
	merged := mergeCredsForTest(client, &Credentials{Username: "root@pve", Password: "x"})
	assert.Equal(t, "", merged.Realm, "WithDefaultRealm must not override an explicit @realm")
}

func TestWithDefaultRealm_StripsLeadingAt(t *testing.T) {
	client := NewClient("", WithDefaultRealm("@pam"))
	assert.Equal(t, "pam", client.defaultRealm, "leading '@' should be stripped")
}

func TestWithEagerAuth_SetsFlag(t *testing.T) {
	client := NewClient("", WithEagerAuth())
	assert.True(t, client.eagerAuth)
}

func TestWithEagerAuth_RunsCreateSessionAtConstruction(t *testing.T) {
	mockConfig := mockConfig
	defer gock.Off()

	called := false
	gock.New(mockConfig.URI).
		Post("/access/ticket").
		AddMatcher(func(r *http.Request, _ *gock.Request) (bool, error) {
			called = true
			return true, nil
		}).
		Reply(200).
		JSON(`{"data":{"ticket":"PVE:root@pam:0000:hex","CSRFPreventionToken":"csrf","username":"root@pam"}}`)

	NewClient(mockConfig.URI,
		WithCredentials(&Credentials{Username: "root@pam", Password: "pw"}),
		WithEagerAuth(),
	)
	assert.True(t, called, "WithEagerAuth should POST /access/ticket synchronously in NewClient")
}

func TestWithEagerAuth_NoOpWithToken(t *testing.T) {
	defer gock.Off()
	// Even with WithEagerAuth, token auth should not trigger a ticket POST.
	gock.New("http://nope").
		Post("/access/ticket").
		Reply(500)

	NewClient("http://nope",
		WithAPIToken("root@pam!t", "secret"),
		WithEagerAuth(),
	)
	// If the eager auth had fired, gock would have served the 500 and our
	// transport would still be intact; the real assertion is just that
	// NewClient didn't panic and we got here.
	assert.True(t, gock.IsPending(), "ticket POST should not have been consumed for token auth")
}

// --- test helpers ---------------------------------------------------------

func mergeCredsForTest(c *Client, base *Credentials) *Credentials {
	c.credentials = base
	return c.sessionCredentials()
}

func containsByte(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// --- retry option ---------------------------------------------------------

func TestWithRetry_InstallsWrapper(t *testing.T) {
	client := NewClient("", WithRetry())
	_, ok := client.httpClient.Transport.(*retryRoundTripper)
	assert.True(t, ok, "WithRetry() should install a *retryRoundTripper on the transport")
	// Sanity: the default policy is non-zero.
	rt := client.httpClient.Transport.(*retryRoundTripper)
	assert.Equal(t, 3, rt.policy.maxAttempts)
	assert.Equal(t, 200*time.Millisecond, rt.policy.initialBackoff)
	assert.Equal(t, 5*time.Second, rt.policy.maxBackoff)
}

// --- proxy options --------------------------------------------------------

// transportFromClient extracts the underlying *http.Transport from a Client,
// failing the test if the transport is missing or has been replaced with a
// non-*http.Transport RoundTripper.
func transportFromClient(t *testing.T, c *Client) *http.Transport {
	t.Helper()
	require.NotNil(t, c.httpClient, "client.httpClient should be set after NewClient")
	require.NotNil(t, c.httpClient.Transport, "client.httpClient.Transport should be promoted by transport options")
	tr, ok := c.httpClient.Transport.(*http.Transport)
	require.True(t, ok, "client.httpClient.Transport should be *http.Transport")
	return tr
}

func TestWithProxy(t *testing.T) {
	proxyURL, err := url.Parse("http://proxy.example.com:3128")
	require.NoError(t, err)

	t.Run("default client gets proxy", func(t *testing.T) {
		c := NewClient("", WithProxy(proxyURL))
		tr := transportFromClient(t, c)
		require.NotNil(t, tr.Proxy, "Proxy should be set")
		got, err := tr.Proxy(&http.Request{})
		require.NoError(t, err)
		assert.Equal(t, proxyURL, got)
		// Ensure we did not mutate http.DefaultTransport.
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			assert.NotSame(t, dt, tr, "should not mutate http.DefaultTransport")
		}
	})

	t.Run("WithHTTPClient then WithProxy mutates custom client's transport", func(t *testing.T) {
		custom := &http.Client{Timeout: 7 * time.Second}
		c := NewClient("", WithHTTPClient(custom), WithProxy(proxyURL))
		assert.Same(t, custom, c.httpClient, "WithHTTPClient's client should be preserved")
		tr := transportFromClient(t, c)
		require.NotNil(t, tr.Proxy)
		got, err := tr.Proxy(&http.Request{})
		require.NoError(t, err)
		assert.Equal(t, proxyURL, got)
	})

	t.Run("WithProxy then WithHTTPClient still applies proxy to custom client", func(t *testing.T) {
		custom := &http.Client{Timeout: 7 * time.Second}
		c := NewClient("", WithProxy(proxyURL), WithHTTPClient(custom))
		assert.Same(t, custom, c.httpClient, "WithHTTPClient's client should be preserved")
		tr := transportFromClient(t, c)
		require.NotNil(t, tr.Proxy)
		got, err := tr.Proxy(&http.Request{})
		require.NoError(t, err)
		assert.Equal(t, proxyURL, got)
	})

	t.Run("custom client with pre-set transport is reused, not replaced", func(t *testing.T) {
		preTransport := &http.Transport{}
		custom := &http.Client{Transport: preTransport}
		c := NewClient("", WithHTTPClient(custom), WithProxy(proxyURL))
		assert.Same(t, preTransport, c.httpClient.Transport, "existing *http.Transport should be reused")
		require.NotNil(t, preTransport.Proxy)
		got, err := preTransport.Proxy(&http.Request{})
		require.NoError(t, err)
		assert.Equal(t, proxyURL, got)
	})

	t.Run("non-http.Transport RoundTripper logs and no-ops", func(t *testing.T) {
		rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return nil, nil
		})
		custom := &http.Client{Transport: rt}
		c := NewClient("", WithHTTPClient(custom), WithProxy(proxyURL))
		// Transport untouched; still the original RoundTripperFunc.
		assert.Equal(t, reflect.ValueOf(rt).Pointer(),
			reflect.ValueOf(c.httpClient.Transport).Pointer(),
			"custom RoundTripper should not be replaced")
	})
}

func TestWithProxyFromEnvironment(t *testing.T) {
	c := NewClient("", WithProxyFromEnvironment())
	tr := transportFromClient(t, c)
	require.NotNil(t, tr.Proxy, "Proxy should be set")

	// Assert by function-pointer identity rather than by exercising the
	// returned func with a t.Setenv-driven HTTP_PROXY. http.ProxyFromEnvironment
	// populates its env lookup behind a process-wide sync.Once, so any prior
	// test that touches the default transport (every gock-using test does) primes
	// the cache to whatever the environment looked like at that moment, and a
	// later t.Setenv has no effect on the cached result. The pointer-identity
	// check still proves the option installed the expected proxy resolver.
	assert.Equal(t,
		reflect.ValueOf(http.ProxyFromEnvironment).Pointer(),
		reflect.ValueOf(tr.Proxy).Pointer(),
		"WithProxyFromEnvironment should install http.ProxyFromEnvironment as the transport proxy func")
}

// roundTripperFunc is a test helper that adapts a function into a
// http.RoundTripper for verifying the non-*http.Transport branch of
// ensureTransport.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// --- request interceptor --------------------------------------------------

// TestWithRequestInterceptor_RegistrationAppends verifies the option appends
// to the interceptor chain rather than replacing earlier entries, that nil
// callbacks are skipped at registration time, and that no-options returns
// nil.
func TestWithRequestInterceptor_RegistrationAppends(t *testing.T) {
	noop := func(*http.Request) error { return nil }

	c := NewClient("")
	assert.Nil(t, c.interceptors)

	c = NewClient("", WithRequestInterceptor(noop), WithRequestInterceptor(noop))
	assert.Len(t, c.interceptors, 2)

	// nil fn should be a silent no-op at registration time.
	c = NewClient("", WithRequestInterceptor(nil), WithRequestInterceptor(noop))
	assert.Len(t, c.interceptors, 1)
}

// TestWithRequestInterceptor_SingleSeesAuthHeader fires a real request
// through gock with API-token auth and asserts the interceptor was called
// exactly once and saw the Authorization header populated by authHeaders.
// This pins the "fires after authHeaders" contract.
func TestWithRequestInterceptor_SingleSeesAuthHeader(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Shared pve9x mock set already covers GET ^/version$.

	var (
		calls   int
		sawAuth string
		sawPath string
	)
	intercept := func(req *http.Request) error {
		calls++
		sawAuth = req.Header.Get("Authorization")
		if req.URL != nil {
			sawPath = req.URL.Path
		}
		return nil
	}

	c := NewClient(
		mockConfig.URI,
		WithAPIToken("root@pam!test", "secret"),
		WithRequestInterceptor(intercept),
	)

	_, err := c.Version(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, calls, "interceptor should fire exactly once per request")
	assert.Equal(t, "PVEAPIToken=root@pam!test=secret", sawAuth,
		"interceptor must run AFTER authHeaders so it can observe the wire-bound Authorization header")
	// The interceptor must see the URL.Path that's about to be sent, so
	// tracing exporters can attach span attributes.
	assert.Equal(t, "/version", sawPath)
}

// TestWithRequestInterceptor_OrderPreserved registers two interceptors and
// asserts they ran in registration order.
func TestWithRequestInterceptor_OrderPreserved(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Shared pve9x mock set already covers GET ^/version$.

	var order []string
	first := func(*http.Request) error { order = append(order, "first"); return nil }
	second := func(*http.Request) error { order = append(order, "second"); return nil }

	c := NewClient(
		mockConfig.URI,
		WithAPIToken("root@pam!test", "secret"),
		WithRequestInterceptor(first),
		WithRequestInterceptor(second),
	)

	_, err := c.Version(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"first", "second"}, order)
}

// errInterceptorSentinel is a package-level sentinel so the test can assert
// errors.Is unwraps through the "request interceptor:" wrapper.
var errInterceptorSentinel = errors.New("sentinel: interceptor refused")

// TestWithRequestInterceptor_ErrorAborts verifies that returning an error
// from an interceptor short-circuits the request (gock never matches), the
// returned error wraps the sentinel via fmt.Errorf("%w") so callers can
// errors.Is against their own sentinels, and a later interceptor in the
// chain never runs.
func TestWithRequestInterceptor_ErrorAborts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// Register a strict route that, if hit, would respond 200. We assert
	// after the call that this route is *still* pending (never matched).
	pending := gock.New(config.C.URI).
		Get("^/version$").
		Reply(200).
		JSON(`{"data":{"version":"9.0"}}`)

	bad := func(*http.Request) error { return errInterceptorSentinel }
	laterCalled := false
	later := func(*http.Request) error { laterCalled = true; return nil }

	c := NewClient(
		mockConfig.URI,
		WithAPIToken("root@pam!test", "secret"),
		WithRequestInterceptor(bad),
		WithRequestInterceptor(later),
	)

	_, err := c.Version(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, errInterceptorSentinel),
		"returned error must wrap the sentinel so callers can errors.Is")
	assert.True(t, strings.Contains(err.Error(), "request interceptor:"),
		"error message should be prefixed with 'request interceptor:' for diagnostics, got: %s", err.Error())
	assert.False(t, laterCalled, "first non-nil error must short-circuit the chain")

	// gock should never have matched — the request was aborted before
	// httpClient.Do was reached.
	assert.False(t, pending.Done(), "interceptor error must abort before httpClient.Do; gock saw a request when it should not have")
}

// TestWithRequestInterceptor_FiresForUpload verifies that the interceptor
// chain also runs for the Upload/UploadReader path, so tracing and auditing
// is uniform across every outgoing request the client makes.
func TestWithRequestInterceptor_FiresForUpload(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	var (
		calls   int
		sawAuth string
		sawPath string
	)
	intercept := func(req *http.Request) error {
		calls++
		sawAuth = req.Header.Get("Authorization")
		if req.URL != nil {
			sawPath = req.URL.Path
		}
		return nil
	}

	c := NewClient(
		mockConfig.URI,
		WithAPIToken("root@pam!test", "secret"),
		WithRequestInterceptor(intercept),
	)

	body := strings.NewReader("payload")
	err := c.UploadReader(
		"/nodes/node1/storage/local/upload",
		map[string]string{"content": "iso"},
		"hello.txt", body, int64(body.Len()), nil,
	)
	require.NoError(t, err)

	assert.Equal(t, 1, calls, "interceptor should fire once for an Upload/UploadReader request")
	assert.Equal(t, "PVEAPIToken=root@pam!test=secret", sawAuth,
		"interceptor must run after authHeaders on the upload path too")
	assert.Equal(t, "/nodes/node1/storage/local/upload", sawPath)
}
