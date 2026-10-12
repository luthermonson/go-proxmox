package proxmox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errorResponse builds a response the way net/http would hand it to
// handleResponse: PVE puts the first line of its die message in the reason
// phrase, so Status carries text gock cannot express.
func errorResponse(code int, status, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Status:     status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request: &http.Request{
			Method: http.MethodGet,
			URL:    &url.URL{Scheme: "https", Host: "pve:8006", Path: "/api2/json/nodes/pve/qemu/100/config", RawQuery: "current=1"},
		},
	}
}

// handleError runs a synthetic error response through handleResponse.
func handleError(c *Client, code int, status, body string) error {
	res := errorResponse(code, status, body)
	defer func() { _ = res.Body.Close() }()
	return c.handleResponse(res, nil)
}

func TestAPIError_ParsesBody(t *testing.T) {
	cases := []struct {
		name        string
		code        int
		status      string
		body        string
		wantMessage string
		wantErrors  map[string]string
		wantError   string
	}{
		{
			name:        "json message with multi-line detail",
			code:        500,
			status:      "500 unable to create VM 100 - first line",
			body:        `{"data":null,"message":"unable to create VM 100 - first line\nsecond line\n"}`,
			wantMessage: "unable to create VM 100 - first line\nsecond line",
			wantError:   "500 unable to create VM 100 - first line: second line",
		},
		{
			name:        "json message identical to status line",
			code:        500,
			status:      "500 Configuration file 'nodes/pve/qemu-server/100.conf' does not exist",
			body:        `{"data":null,"message":"Configuration file 'nodes/pve/qemu-server/100.conf' does not exist\n"}`,
			wantMessage: "Configuration file 'nodes/pve/qemu-server/100.conf' does not exist",
			wantError:   "500 Configuration file 'nodes/pve/qemu-server/100.conf' does not exist",
		},
		{
			name:        "parameter verification with errors map",
			code:        400,
			status:      "400 Parameter verification failed.",
			body:        `{"data":null,"errors":{"vmid":"invalid format - value does not look like a valid VM ID\n","cores":"value must have a minimum value of 1"},"message":"Parameter verification failed.\n"}`,
			wantMessage: "Parameter verification failed.",
			wantErrors: map[string]string{
				"vmid":  "invalid format - value does not look like a valid VM ID",
				"cores": "value must have a minimum value of 1",
			},
			wantError: `bad request: 400 Parameter verification failed. - {"vmid":"invalid format - value does not look like a valid VM ID\n","cores":"value must have a minimum value of 1"}`,
		},
		{
			name:        "empty body",
			code:        500,
			status:      "500 Internal Server Error",
			body:        "",
			wantMessage: "Internal Server Error",
			wantError:   "500 Internal Server Error",
		},
		{
			name:        "non-json body from a reverse proxy",
			code:        502,
			status:      "502 Bad Gateway",
			body:        "<html><body>bad gateway</body></html>",
			wantMessage: "Bad Gateway",
			wantError:   "502 Bad Gateway",
		},
		{
			name:        "non-string errors values are kept as raw json",
			code:        400,
			status:      "400 Parameter verification failed.",
			body:        `{"data":null,"errors":{"net0":{"bridge":"missing"}}}`,
			wantMessage: "Parameter verification failed.",
			wantErrors:  map[string]string{"net0": `{"bridge":"missing"}`},
			wantError:   `bad request: 400 Parameter verification failed. - {"net0":{"bridge":"missing"}}`,
		},
		{
			name:        "node unreachable pseudo-status",
			code:        595,
			status:      "595 Connection refused",
			body:        "",
			wantMessage: "Connection refused",
			wantError:   "595 Connection refused",
		},
	}

	client := NewClient(TestURI)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := handleError(client, tc.code, tc.status, tc.body)
			require.Error(t, err)

			var apiErr *APIError
			require.True(t, errors.As(err, &apiErr), "want *APIError, got %T", err)
			assert.Equal(t, tc.code, apiErr.StatusCode)
			assert.Equal(t, tc.status, apiErr.Status)
			assert.Equal(t, tc.wantMessage, apiErr.Message)
			assert.Equal(t, tc.wantErrors, apiErr.Errors)
			assert.Equal(t, http.MethodGet, apiErr.Method)
			assert.Equal(t, "/api2/json/nodes/pve/qemu/100/config", apiErr.Path, "path excludes the query string")
			assert.Equal(t, tc.wantError, err.Error())
		})
	}
}

// Error() must keep the pre-APIError text so callers matching on it, like
// WaitForAgent and CheckID in this package, keep working.
func TestAPIError_ErrorTextBackwardCompatible(t *testing.T) {
	client := NewClient(TestURI)

	// Was errors.New(res.Status) for 500 and 501.
	err := handleError(client, 500, "500 QEMU guest agent is not running", `{"data":null,"message":"QEMU guest agent is not running\n"}`)
	assert.Equal(t, "500 QEMU guest agent is not running", err.Error())
	assert.True(t, strings.Contains(err.Error(), "500 QEMU guest agent is not running"))

	err = handleError(client, 501, "501 Method 'GET /nodes/pve/nope' not implemented", `{"data":null,"message":"Method 'GET /nodes/pve/nope' not implemented"}`)
	assert.Equal(t, "501 Method 'GET /nodes/pve/nope' not implemented", err.Error())

	err = handleError(client, 400, "400 VM 100 already exists", `{"data":null,"message":"VM 100 already exists\n"}`)
	assert.Contains(t, err.Error(), "VM 100 already exists")
	assert.True(t, strings.HasPrefix(err.Error(), "bad request: 400 VM 100 already exists - "))
}

func TestAPIError_Sentinels(t *testing.T) {
	cases := []struct {
		name   string
		code   int
		status string
		body   string
		is     []error
		isNot  []error
	}{
		{"404", 404, "404 Not Found", "", []error{ErrNotFound}, []error{ErrBadRequest, ErrNotAuthorized}},
		{"401", 401, "401 No ticket", "", []error{ErrNotAuthorized}, []error{ErrForbidden, ErrNotFound}},
		{"403", 403, "403 Permission check failed (/vms/100, VM.Audit)", "", []error{ErrNotAuthorized, ErrForbidden}, []error{ErrNotFound}},
		{"400", 400, "400 Parameter verification failed.", `{"errors":{"cores":"value must have a minimum value of 1"}}`, []error{ErrBadRequest}, []error{ErrNotFound}},
		{"400 no such storage", 400, "400 Parameter verification failed.", `{"errors":{"storage":"No such storage."}}`, []error{ErrBadRequest, ErrNotFound}, nil},
		{"429", 429, "429 Too Many Requests", "", []error{ErrRateLimited}, []error{ErrNotFound}},
		{"595", 595, "595 Connection refused", "", []error{ErrNodeUnreachable}, []error{ErrNotFound}},
		{"596", 596, "596 Connection timed out", "", []error{ErrNodeUnreachable}, nil},
		{"597 is not node unreachable", 597, "597", "", nil, []error{ErrNodeUnreachable}},
		{"500 generic", 500, "500 unable to create VM 100 - lock timeout", "", nil, []error{ErrNotFound, ErrBadRequest}},
		{"501 is not not-found", 501, "501 Method 'GET /x' not implemented", "", nil, []error{ErrNotFound}},
		{"503", 503, "503 Service Unavailable", "", nil, []error{ErrNotFound, ErrRateLimited}},
	}

	client := NewClient(TestURI)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := handleError(client, tc.code, tc.status, tc.body)
			require.Error(t, err)
			for _, s := range tc.is {
				assert.ErrorIs(t, err, s)
			}
			for _, s := range tc.isNot {
				assert.NotErrorIs(t, err, s)
			}
			// Wrapping must not break classification.
			wrapped := fmt.Errorf("listing vms: %w", err)
			for _, s := range tc.is {
				assert.ErrorIs(t, wrapped, s)
			}
		})
	}
}

func TestAPIError_Helpers(t *testing.T) {
	mk := func(code int, status string) error {
		return handleError(NewClient(TestURI), code, status, "")
	}
	assert.True(t, IsNotFound(mk(404, "404 Not Found")))
	assert.True(t, IsNotAuthorized(mk(401, "401 No ticket")))
	assert.True(t, IsForbidden(mk(403, "403 Forbidden")))
	assert.True(t, IsBadRequest(mk(400, "400 Parameter verification failed.")))
	assert.True(t, IsRateLimited(mk(429, "429 Too Many Requests")))
	assert.True(t, IsNodeUnreachable(mk(595, "595 Connection refused")))

	other := errors.New("other")
	for _, fn := range []func(error) bool{IsForbidden, IsBadRequest, IsRateLimited, IsNodeUnreachable} {
		assert.False(t, fn(other))
		assert.False(t, fn(nil))
	}

	apiErr, ok := AsAPIError(fmt.Errorf("wrapped: %w", mk(500, "500 boom")))
	require.True(t, ok)
	assert.Equal(t, 500, apiErr.StatusCode)
	_, ok = AsAPIError(other)
	assert.False(t, ok)
}

func TestAPIError_RetryAfter(t *testing.T) {
	res := errorResponse(429, "429 Too Many Requests", "")
	res.Header.Set("Retry-After", "7")
	defer func() { _ = res.Body.Close() }()
	err := NewClient(TestURI).handleResponse(res, nil)
	apiErr, ok := AsAPIError(err)
	require.True(t, ok)
	assert.Equal(t, 7*time.Second, apiErr.RetryAfter)
}

func TestMatchesNotFound(t *testing.T) {
	// Real PVE messages; see the notFoundPatterns table for sources.
	notFound := []string{
		"Configuration file 'nodes/pve/qemu-server/100.conf' does not exist",
		"storage 'nfs-backup' does not exist",
		"pool 'dev' does not exist",
		"role 'Auditor2' does not exist",
		"sdn vnet 'vnet9' does not exist",
		"snapshot 'pre-upgrade' does not exist",
		"OSD '3' does not exists on host 'pve'",
		"no such user ('bob@pve')",
		"no such cluster node 'pve9'",
		"no such job 'local-zfs-100-0'",
		"No such job 'backup-1234'",
		"no such resource 'vm:100'",
		"no such ha group 'ha1'",
		"No such storage.",
		"no such task",
		"mapping 'gpu0' not found",
		"Custom cputype 'mycpu' not found",
		"unable to find configuration file for VM 100 on node 'pve'",
	}
	for _, msg := range notFound {
		assert.True(t, matchesNotFound(msg), msg)
	}

	notNotFound := []string{
		"",
		"QEMU guest agent is not running",
		"VM 100 already exists",
		"unable to create VM 100 - lock timeout",
		"no such option 'foo'",
		"no such standard option 'pve-vmid'",
		"No such timezone",
		"invalid bootorder: device 'scsi9' does not exist'",
		"executable not found '/usr/bin/proxmox-backup-client'! Proxmox backup client not installed?",
		"Method 'GET /nodes/pve/nope' not implemented",
		"Parameter verification failed.",
	}
	for _, msg := range notNotFound {
		assert.False(t, matchesNotFound(msg), msg)
	}
}

// End to end through Client.Get with gock.
func TestAPIError_ThroughClient(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Get("^/nodes/node1/qemu/999/status/current$").
		Reply(500).
		JSON(`{"data":null,"message":"Configuration file 'nodes/node1/qemu-server/999.conf' does not exist\n"}`)
	gock.New(TestURI).
		Get("^/nodes/node2/status$").
		Reply(595).
		BodyString("")
	gock.New(TestURI).
		Post("^/nodes/node1/qemu$").
		Reply(400).
		JSON(`{"data":null,"errors":{"vmid":"VM 101 already exists"},"message":"Parameter verification failed.\n"}`)
	gock.New(TestURI).
		Get("^/cluster/ha/resources/vm:100$").
		Reply(404).
		JSON(`{"data":null,"message":"no such resource 'vm:100'\n"}`)

	client := mockClient()
	ctx := context.Background()

	err := client.Get(ctx, "/nodes/node1/qemu/999/status/current", &VirtualMachine{})
	assert.True(t, IsNotFound(err), "%v", err)
	apiErr, ok := AsAPIError(err)
	require.True(t, ok)
	assert.Equal(t, "/nodes/node1/qemu/999/status/current", apiErr.Path)
	assert.Equal(t, http.MethodGet, apiErr.Method)

	err = client.Get(ctx, "/nodes/node2/status", nil)
	assert.True(t, IsNodeUnreachable(err), "%v", err)

	err = client.Post(ctx, "/nodes/node1/qemu", map[string]int{"vmid": 101}, nil)
	assert.True(t, IsBadRequest(err), "%v", err)
	apiErr, ok = AsAPIError(err)
	require.True(t, ok)
	assert.Equal(t, map[string]string{"vmid": "VM 101 already exists"}, apiErr.Errors)
	assert.Equal(t, "Parameter verification failed.", apiErr.Message)

	// A 404 used to be silently swallowed when v was nil.
	err = client.Get(ctx, "/cluster/ha/resources/vm:100", nil)
	assert.True(t, IsNotFound(err), "%v", err)

	assert.True(t, gock.IsDone())
}

func TestAPIError_NotAuthorizedThroughReq(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Get("^/version$").
		Reply(403).
		JSON(`{"data":null,"message":"Permission check failed (/, Sys.Audit)\n"}`)

	client := NewClient(TestURI, WithAPIToken("root@pam!t", "secret"))
	_, err := client.Version(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotAuthorized)
	assert.ErrorIs(t, err, ErrForbidden)
	// Old text was exactly ErrNotAuthorized.Error(); it is still the prefix.
	assert.True(t, strings.HasPrefix(err.Error(), ErrNotAuthorized.Error()), err.Error())
	assert.Contains(t, err.Error(), "Permission check failed (/, Sys.Audit)")
	apiErr, ok := AsAPIError(err)
	require.True(t, ok)
	assert.Equal(t, 403, apiErr.StatusCode)
}

func TestAPIError_RetryExhaustedSurfacesStatus(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Times(3).
		Reply(http.StatusServiceUnavailable).
		JSON(`{"data":null,"message":"cluster not ready - no quorum?\n"}`)

	client := newRetryClient(t, nil)
	_, err := client.Version(context.Background())
	require.Error(t, err)
	apiErr, ok := AsAPIError(err)
	require.True(t, ok, "%T %v", err, err)
	assert.Equal(t, http.StatusServiceUnavailable, apiErr.StatusCode)
	assert.Equal(t, "cluster not ready - no quorum?", apiErr.Message)
	assert.True(t, gock.IsDone(), "all three attempts should have been made")
}

func TestAPIError_RetryRateLimited(t *testing.T) {
	defer gock.Off()

	gock.New(retryTestURI).
		Get("^/version$").
		Times(2).
		Reply(http.StatusTooManyRequests).
		SetHeader("Retry-After", "1")

	client := newRetryClient(t, []RetryOption{WithRetryMax(2)})
	_, err := client.Version(context.Background())
	assert.True(t, IsRateLimited(err), "%v", err)
	apiErr, ok := AsAPIError(err)
	require.True(t, ok)
	assert.Equal(t, time.Second, apiErr.RetryAfter)
	assert.True(t, gock.IsDone())
}
