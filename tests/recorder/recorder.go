// Package recorder wraps go-vcr to provide a project-shaped helper for
// recording and replaying HTTP interactions with a Proxmox VE API. It is the
// successor to the gock-based fixtures under tests/mocks.
//
// Modes are selected by env var:
//
//   - PROXMOX_URL set    → ModeRecordOnce. Missing cassettes are recorded
//                          against the live server; existing cassettes are
//                          replay-only. To fully re-record an existing
//                          cassette, delete it first.
//   - PROXMOX_URL unset  → ModeReplayOnly. Hitting the network or making an
//                          unmatched request is a hard error. CI default.
//
// Cassettes live under tests/recorder/testdata/<TestName>.yaml and are
// expected to be scrubbed of credentials and host-identifying information by
// the BeforeSaveHook chain configured here; see scrubBody for the rules.
//
// Recording is normally driven by `mage record:pveN` against a freshly
// installed PVE in a nested VM. See SEED.md for the deterministic state the
// nested host must produce so re-records are byte-stable.
package recorder

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	vcrrecorder "gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// FakeHost is the synthetic host:port written into every recorded URL so the
// real Proxmox host name never lands in version control.
const FakeHost = "pve.example.test:8006"

// FakeURL is the corresponding base URL callers can pass to proxmox.NewClient
// in replay mode. The recorder's matcher is method+URL, so the client's
// configured URL must agree with what was written into the cassette.
const FakeURL = "https://" + FakeHost + "/api2/json"

// New returns a configured *vcrrecorder.Recorder for the calling test. The
// cassette path is derived from t.Name(); mode is selected from the env. The
// recorder is registered with t.Cleanup so callers don't have to remember a
// defer.
func New(t *testing.T) *vcrrecorder.Recorder {
	t.Helper()

	realURL := os.Getenv("PROXMOX_URL")
	mode := vcrrecorder.ModeReplayOnly
	if realURL != "" {
		mode = vcrrecorder.ModeRecordOnce
	}

	cassettePath := filepath.Join("testdata", t.Name())

	opts := []vcrrecorder.Option{
		vcrrecorder.WithMode(mode),
		vcrrecorder.WithSkipRequestLatency(true),
		vcrrecorder.WithMatcher(matchMethodURLBody),
		// Proxmox uses self-signed TLS by default; the recorder needs a
		// transport that tolerates it during recording. No effect in replay.
		vcrrecorder.WithRealTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		}),
	}

	realHost := hostOnly(realURL)
	for _, h := range scrubHooks(realHost) {
		opts = append(opts, vcrrecorder.WithHook(h, vcrrecorder.BeforeSaveHook))
	}

	r, err := vcrrecorder.New(cassettePath, opts...)
	if err != nil {
		t.Fatalf("recorder.New(%s): %v", cassettePath, err)
	}
	t.Cleanup(func() {
		if err := r.Stop(); err != nil {
			t.Errorf("recorder.Stop: %v", err)
		}
	})
	return r
}

// matchMethodURLBody is the project's cassette matcher. It compares method,
// URL, and (for non-GET) request body — the three things that actually
// determine what Proxmox returns. Headers are intentionally ignored: clients
// vary in User-Agent / Accept / Auth shape and those have no effect on the
// response payload we care about.
func matchMethodURLBody(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}
	if r.URL.String() != i.URL {
		return false
	}
	if r.Body == nil {
		return i.Body == ""
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(buf))
	return string(buf) == i.Body
}

// hostOnly extracts host:port from a URL like https://host:port/api2/json.
// Returns "" if the URL is empty.
func hostOnly(u string) string {
	if u == "" {
		return ""
	}
	s := strings.TrimPrefix(u, "https://")
	s = strings.TrimPrefix(s, "http://")
	if i := strings.Index(s, "/"); i > 0 {
		s = s[:i]
	}
	return s
}

// scrubHooks is the ordered chain of BeforeSaveHooks. Each scrub stage has one
// job; order matters only for the body pass, where host-name replacement runs
// before the regex sweep so the synthetic FakeHost survives subsequent passes.
func scrubHooks(realHost string) []vcrrecorder.HookFunc {
	return []vcrrecorder.HookFunc{
		stripAuthHeaders,
		rewriteHost(realHost),
		scrubBodies(realHost),
	}
}

func stripAuthHeaders(i *cassette.Interaction) error {
	for _, h := range []string{"Authorization", "Cookie", "Csrfpreventiontoken"} {
		i.Request.Headers.Del(h)
	}
	for _, h := range []string{"Set-Cookie", "Server"} {
		i.Response.Headers.Del(h)
	}
	return nil
}

func rewriteHost(realHost string) vcrrecorder.HookFunc {
	if realHost == "" {
		return func(*cassette.Interaction) error { return nil }
	}
	return func(i *cassette.Interaction) error {
		i.Request.URL = strings.ReplaceAll(i.Request.URL, realHost, FakeHost)
		i.Request.Host = FakeHost
		return nil
	}
}

var (
	sshKeyRe = regexp.MustCompile(`(?m)ssh-(?:rsa|ed25519|ecdsa-sha2-[a-z0-9-]+)\s+[A-Za-z0-9+/=]+(?:\s+\S+)?`)
	pemRe    = regexp.MustCompile(`-----BEGIN [A-Z ]+-----[\s\S]+?-----END [A-Z ]+-----`)
	pveTktRe = regexp.MustCompile(`PVE:[A-Za-z0-9+/=:_-]+`)
	ipv4Re   = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	macRe    = regexp.MustCompile(`\b(?:[0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}\b`)
)

func scrubBodies(realHost string) vcrrecorder.HookFunc {
	return func(i *cassette.Interaction) error {
		i.Request.Body = scrubBody(i.Request.Body, realHost)
		i.Response.Body = scrubBody(i.Response.Body, realHost)
		return nil
	}
}

// scrubBody applies the redaction passes to a request or response body.
// Order: host rewrite first (so the synthetic FakeHost survives the regex
// sweep), then key/PEM/ticket redactions, then numeric address normalization.
// IPv4 and MAC redaction run last so they do not eat into the encoded body
// of an SSH key or a PEM block.
func scrubBody(body, realHost string) string {
	if body == "" {
		return body
	}
	if realHost != "" {
		body = strings.ReplaceAll(body, realHost, FakeHost)
	}
	body = sshKeyRe.ReplaceAllString(body, "ssh-ed25519 REDACTED-SSH-KEY recorder@example.test")
	body = pemRe.ReplaceAllString(body, "-----BEGIN REDACTED-----\nREDACTED\n-----END REDACTED-----")
	body = pveTktRe.ReplaceAllString(body, "PVE:REDACTED-TICKET")
	body = ipv4Re.ReplaceAllStringFunc(body, redactIPv4)
	body = macRe.ReplaceAllStringFunc(body, redactMAC)
	return body
}

// redactIPv4 maps an IPv4 address to a deterministic address in TEST-NET-1
// (192.0.2.0/24, RFC 5737). Deterministic so re-records produce stable diffs.
func redactIPv4(s string) string {
	var sum uint32
	for i := 0; i < len(s); i++ {
		sum = sum*31 + uint32(s[i])
	}
	return fmt.Sprintf("192.0.2.%d", (sum%254)+1)
}

// redactMAC maps a MAC to a deterministic address in BC:24:11:xx:xx:xx, which
// is the OUI Proxmox uses for auto-generated VM MACs. Deterministic for stable
// diffs across re-records.
func redactMAC(s string) string {
	var sum uint32
	for i := 0; i < len(s); i++ {
		sum = sum*31 + uint32(s[i])
	}
	return fmt.Sprintf("BC:24:11:%02X:%02X:%02X", (sum>>16)&0xFF, (sum>>8)&0xFF, sum&0xFF)
}
