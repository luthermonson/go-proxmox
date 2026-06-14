package proxmox

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Option func(*Client)

// Deprecated: Use WithHTTPClient
func WithClient(client *http.Client) Option {
	return WithHTTPClient(client)
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Deprecated: Use WithCredential
func WithLogins(username, password string) Option {
	return WithCredentials(&Credentials{
		Username: username,
		Password: password,
	})
}

func WithCredentials(credentials *Credentials) Option {
	return func(c *Client) {
		c.credentials = credentials
	}
}

func WithAPIToken(tokenID, secret string) Option {
	return func(c *Client) {
		c.token = fmt.Sprintf("%s=%s", tokenID, secret)
	}
}

// WithSession experimental
func WithSession(ticket, csrfPreventionToken string) Option {
	return func(c *Client) {
		c.session = &Session{
			Ticket:              ticket,
			CSRFPreventionToken: csrfPreventionToken,
		}
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithLogger(logger LeveledLoggerInterface) Option {
	return func(c *Client) {
		c.log = logger
	}
}

// --- transport / TLS options ---------------------------------------------

// WithTimeout sets the *http.Client.Timeout used for every request. Composes
// with WithHTTPClient regardless of option order — if the caller passed their
// own *http.Client, the timeout is applied to it.
//
// The default http.DefaultClient has no timeout. Without this option,
// a hung PVE node means a hung caller forever; setting at least a generous
// upper bound is recommended for any non-interactive use.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithInsecureSkipVerify disables TLS certificate verification. For lab use
// only — production clusters should pin the cluster's CA via WithRootCAs or
// WithRootCAFile instead.
//
// Composes with WithHTTPClient, WithRootCAs, WithRootCAFile, and
// WithClientCertificate; option order doesn't matter.
func WithInsecureSkipVerify() Option {
	return func(c *Client) {
		c.tlsMods = append(c.tlsMods, func(tc *tls.Config) {
			tc.InsecureSkipVerify = true
		})
	}
}

// WithRootCAs sets the *x509.CertPool used to verify the cluster's TLS
// certificate. Use this when the cluster presents a cert chain signed by your
// org's CA. Composes with the other TLS options.
func WithRootCAs(pool *x509.CertPool) Option {
	return func(c *Client) {
		c.tlsMods = append(c.tlsMods, func(tc *tls.Config) {
			tc.RootCAs = pool
		})
	}
}

// WithRootCAFile loads a PEM-encoded CA bundle from path and appends every
// certificate it contains to the TLS root pool. Convenience wrapper around
// WithRootCAs for the common single-file case.
//
// The file is read at NewClient time (when this option is evaluated). If the
// file can't be read or contains no valid certificates, the error is logged
// via the client logger and the option is a no-op — option funcs can't
// return errors. Callers who need the file-IO error surfaced explicitly
// should read the file themselves and pass the resulting pool to
// WithRootCAs.
func WithRootCAFile(path string) Option {
	return func(c *Client) {
		data, err := os.ReadFile(path)
		if err != nil {
			c.log.Errorf("WithRootCAFile: read %q: %v", path, err)
			return
		}
		c.tlsMods = append(c.tlsMods, func(tc *tls.Config) {
			if tc.RootCAs == nil {
				tc.RootCAs = x509.NewCertPool()
			}
			if !tc.RootCAs.AppendCertsFromPEM(data) {
				c.log.Errorf("WithRootCAFile: %q contained no valid PEM certificates", path)
			}
		})
	}
}

// WithClientCertificate adds a client certificate for mutual TLS. Appends to
// tls.Config.Certificates so multiple calls compose. Composes with the other
// TLS options.
func WithClientCertificate(cert tls.Certificate) Option {
	return func(c *Client) {
		c.tlsMods = append(c.tlsMods, func(tc *tls.Config) {
			tc.Certificates = append(tc.Certificates, cert)
		})
	}
}

// --- auth ergonomics ------------------------------------------------------

// WithOTP supplies a one-time password (TOTP, Yubico OTP, etc.) for the
// initial /access/ticket call when the user has two-factor auth enabled.
// The code is consumed exactly once on the first CreateSession call;
// subsequent RefreshTicket calls renew the session via the ticket itself
// and do not need a new OTP.
//
// Requires WithCredentials. No effect when using token auth — tokens bypass
// 2FA by design.
//
// If the session is fully lost later (PVE restart, ticket past the renewal
// window) and the client tries to re-authenticate, that call will fail
// because the OTP is single-use. Callers in that scenario must construct a
// fresh client with a fresh OTP — the library cannot keep a TOTP around.
func WithOTP(otp string) Option {
	return func(c *Client) {
		c.otp = otp
	}
}

// WithDefaultRealm auto-appends "@<realm>" to Credentials.Username if the
// supplied username has no @realm suffix and Credentials.Realm is empty.
// Saves the most common credential-auth typo ("root" vs "root@pam").
//
// No effect when token auth is used or when the username already carries
// a realm.
func WithDefaultRealm(realm string) Option {
	return func(c *Client) {
		c.defaultRealm = strings.TrimPrefix(realm, "@")
	}
}

// WithProxy routes all client traffic through the given proxy URL. The proxy
// function is applied to the underlying *http.Transport so every request goes
// through u (use http://, https://, or socks5:// schemes).
//
// Composes with WithHTTPClient regardless of option order: the proxy is
// applied to whatever client the constructor settles on after all options
// have run, via the shared finalizeOptions step. If the resulting transport
// is not an *http.Transport (custom RoundTripper), the option logs a debug
// warning and no-ops — set .Proxy yourself on a transport you control before
// passing it to WithHTTPClient.
func WithProxy(u *url.URL) Option {
	return func(c *Client) {
		c.proxyFunc = func(*http.Request) (*url.URL, error) {
			return u, nil
		}
	}
}

// WithProxyFromEnvironment uses Go's standard http.ProxyFromEnvironment
// lookup (HTTP_PROXY, HTTPS_PROXY, NO_PROXY env vars). Env vars are read
// per-request by http.ProxyFromEnvironment, not at option-eval time, so
// later changes to the environment are picked up on the next request.
//
// Composes with WithHTTPClient the same way as WithProxy.
func WithProxyFromEnvironment() Option {
	return func(c *Client) {
		c.proxyFunc = http.ProxyFromEnvironment
	}
}

// WithEagerAuth makes NewClient call CreateSession synchronously so the
// first user-facing request doesn't trigger PVE pveproxy's hardcoded
// 3-second 401 delay on unauthenticated requests. Pveproxy enforces this
// delay on every failed-or-missing-auth response — see
// PVE::APIServer::AnyEvent's `# always delay unauthorized calls by 3 seconds`
// block. With credential auth the client's first request is always
// unauthenticated (the cookie+CSRF aren't set until /access/ticket
// succeeds), so it eats the full 3s before the library retries with the
// ticket. WithEagerAuth pays that cost once at construction instead.
//
// Has no effect with token auth — tokens attach Authorization on every
// request and never trigger the 401 path. Has no effect when neither
// credentials nor token are set.
//
// Errors from the eager CreateSession are logged at debug level and
// otherwise swallowed; the next user request will retry via the existing
// lazy-auth path. To surface auth errors at startup explicitly, call
// (*Client).CreateSession yourself instead of using this option.
func WithEagerAuth() Option {
	return func(c *Client) {
		c.eagerAuth = true
	}
}

// --- shared transport plumbing -------------------------------------------
//
// ensureOwnHTTPClient, ensureTransport, ensureTLSConfig are private helpers
// shared by every transport-touching option (TLS, timeout — and by the
// Tier 3 options WithProxy and WithRetry, which land in sibling PRs).
// Their job is to give those options a writable *http.Transport without
// mutating http.DefaultClient or http.DefaultTransport when the caller
// didn't provide their own client.

// ensureOwnHTTPClient guarantees c.httpClient is something we're allowed to
// mutate. If it's nil or still pointing at http.DefaultClient, allocate a
// fresh *http.Client (Transport defaulted, Timeout zero) so subsequent
// option writes don't bleed into the global default.
func (c *Client) ensureOwnHTTPClient() {
	if c.httpClient == nil || c.httpClient == http.DefaultClient {
		c.httpClient = &http.Client{}
	}
}

// ensureTransport returns the *http.Transport on c.httpClient, promoting
// http.DefaultTransport to a clone if necessary. Returns nil if the
// client's RoundTripper is a custom non-Transport implementation — in that
// case the caller should debug-log and skip whatever option needed it.
func (c *Client) ensureTransport() *http.Transport {
	c.ensureOwnHTTPClient()
	if c.httpClient.Transport == nil {
		c.httpClient.Transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	if t, ok := c.httpClient.Transport.(*http.Transport); ok {
		return t
	}
	return nil
}

// ensureTLSConfig returns the *tls.Config on the client's transport,
// allocating a fresh one if absent. Returns nil if the transport isn't an
// *http.Transport — see ensureTransport.
func (c *Client) ensureTLSConfig() *tls.Config {
	t := c.ensureTransport()
	if t == nil {
		return nil
	}
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return t.TLSClientConfig
}

// WithRequestInterceptor registers a function called on every outgoing HTTP
// request after the client's auth headers are added and before the request is
// sent. Use cases: tracing (OpenTelemetry span injection), custom audit
// headers, correlation IDs, request logging.
//
// The interceptor fires from Client.Req, Client.Upload, and
// Client.UploadReader. Websocket upgrades (Client.TermWebSocket and
// Client.VNCWebSocket) are exempt — the dialer does not surface a request
// object the chain could mutate.
//
// Multiple WithRequestInterceptor options compose — each call appends to the
// interceptor chain. Interceptors run in registration order. The first
// non-nil error short-circuits the request, is wrapped with a
// "request interceptor:" prefix (so callers can errors.Is against their own
// sentinel errors), and is returned to the caller; the HTTP request is never
// sent.
//
// A nil fn is silently skipped at registration; nil entries in the chain are
// also skipped at request time.
func WithRequestInterceptor(fn func(*http.Request) error) Option {
	return func(c *Client) {
		if fn == nil {
			return
		}
		c.interceptors = append(c.interceptors, fn)
	}
}
