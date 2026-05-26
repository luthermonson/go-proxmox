package proxmox

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
	"github.com/stretchr/testify/assert"
)

const (
	TestURI = "http://test.localhost"
)

var mockConfig = config.Config{
	URI: TestURI,
}

func mockClient(options ...Option) *Client {
	return NewClient(mockConfig.URI, options...)
}

func TestMakeTag(t *testing.T) {
	assert.Equal(t, "go-proxmox+tagname", MakeTag("tagname"))
}

func TestEncodeSSHKeys(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{
			name: "empty",
			in:   nil,
			want: "",
		},
		{
			name: "single key encodes spaces as %20 not +",
			in:   []string{"ssh-rsa AAAAB3NzaC1yc2EAAAA user@host"},
			want: "ssh-rsa%20AAAAB3NzaC1yc2EAAAA%20user%40host",
		},
		{
			name: "literal plus in key is preserved as %2B, not turned into %20",
			in:   []string{"ssh-rsa A+B C"},
			want: "ssh-rsa%20A%2BB%20C",
		},
		{
			name: "multiple keys are joined with newline before encoding",
			in:   []string{"ssh-rsa AAA u@h1", "ssh-ed25519 BBB u@h2"},
			want: "ssh-rsa%20AAA%20u%40h1%0Assh-ed25519%20BBB%20u%40h2",
		},
		{
			name: "no plus signs leak through",
			in:   []string{"a b c d"},
			want: "a%20b%20c%20d",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeSSHKeys(tc.in...)
			assert.Equal(t, tc.want, got)
			assert.NotContains(t, got, "+",
				"EncodeSSHKeys output must never contain '+'; PVE rejects it")
		})
	}
}

func TestCSV_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		body string
		want CSV
	}{
		{
			name: "comma separated string",
			body: `"node1,node2, node3"`,
			want: CSV{"node1", "node2", "node3"},
		},
		{
			name: "array compatibility",
			body: `["node1","node2"]`,
			want: CSV{"node1", "node2"},
		},
		{
			name: "empty string",
			body: `""`,
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got CSV
			assert.NoError(t, json.Unmarshal([]byte(tc.body), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCSV_MarshalJSON(t *testing.T) {
	body, err := json.Marshal(CSV{"node1", "node2"})
	assert.NoError(t, err)
	assert.Equal(t, `"node1,node2"`, string(body))
}

// options tested in options_test.go
func TestNewClient(t *testing.T) {
	v := NewClient(TestURI)
	assert.Equal(t, http.DefaultClient, v.httpClient)
	assert.Equal(t, v.baseURL, TestURI)
	assert.Equal(t, v.userAgent, DefaultUserAgent)
}

func TestClient_authHeaders(t *testing.T) {
	cases := []struct {
		input  http.Header
		expect http.Header
		client *Client
	}{
		{
			input: http.Header{},
			expect: http.Header{
				"Accept":        []string{"application/json"},
				"Authorization": []string{"PVEAPIToken=root@pam!test=1234"},
				"User-Agent":    []string{"go-proxmox/dev"},
			},
			client: NewClient("", WithAPIToken("root@pam!test", "1234")),
		},
		{
			input: http.Header{},
			expect: http.Header{
				"Accept":              []string{"application/json"},
				"Cookie":              []string{"PVEAuthCookie=ticket"},
				"Csrfpreventiontoken": []string{"csrftoken"},
				"User-Agent":          []string{"go-proxmox/dev"},
			},
			client: NewClient("", WithSession("ticket", "csrftoken")),
		},
	}

	for _, test := range cases {
		test.client.authHeaders(&test.input)
		assert.Equal(t, test.expect, test.input)
	}
}

func TestClient_TermWebSocket_APITokenUnsupported(t *testing.T) {
	c := NewClient(TestURI, WithAPIToken("root@pam!test", "secret"))
	send, recv, errs, closer, err := c.TermWebSocket("/nodes/n/lxc/100/vncwebsocket?port=1&vncticket=t", &Term{})
	assert.Nil(t, send)
	assert.Nil(t, recv)
	assert.Nil(t, errs)
	assert.Nil(t, closer)
	assert.ErrorIs(t, err, ErrAPITokenWebSocketUnsupported)
	assert.True(t, IsAPITokenWebSocketUnsupported(err))
}

func TestClient_VNCWebSocket_APITokenUnsupported(t *testing.T) {
	c := NewClient(TestURI, WithAPIToken("root@pam!test", "secret"))
	send, recv, errs, closer, err := c.VNCWebSocket("/nodes/n/qemu/100/vncwebsocket?port=1&vncticket=t", &VNC{})
	assert.Nil(t, send)
	assert.Nil(t, recv)
	assert.Nil(t, errs)
	assert.Nil(t, closer)
	assert.ErrorIs(t, err, ErrAPITokenWebSocketUnsupported)
}

func TestClient_Version7(t *testing.T) {
	mocks.ProxmoxVE7x(mockConfig)
	defer mocks.Off()

	v, err := mockClient().Version(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "7.7-7", v.Version)
	assert.Equal(t, "777777", v.RepoID)
	assert.Equal(t, "7.7", v.Release)
}

func TestClient_Version6(t *testing.T) {
	mocks.ProxmoxVE6x(mockConfig)
	defer mocks.Off()

	v, err := mockClient().Version(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "6.6-6", v.Version)
	assert.Equal(t, "666666", v.RepoID)
	assert.Equal(t, "6.6", v.Release)
}

func TestClient_Version9(t *testing.T) {
	mocks.ProxmoxVE9x(mockConfig)
	defer mocks.Off()

	v, err := mockClient().Version(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "9.1-1", v.Version)
	assert.Equal(t, "9a1b2c3d", v.RepoID)
	assert.Equal(t, "9.1", v.Release)
}

func TestClientMethods(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	var err error

	var v Version
	err = client.Get(ctx, "/version", &v)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)

	err = client.Post(ctx, "/version", struct{}{}, &v)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)

	err = client.Put(ctx, "/version", struct{}{}, &v)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)

	err = client.Delete(ctx, "/version", &v)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)
}

func TestClient_handleResponse(t *testing.T) {
	// todo test if logs exclude /access/ticket requests
	// todo test data key vs no data key

	client := NewClient(TestURI)

	// bad json
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader("{\"data\":{\"test\": \"data\"}")),
	}
	testData := map[string]string{}
	err := client.handleResponse(resp, &testData)
	assert.NotNil(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Len(t, testData, 0)

	// good json
	resp = &http.Response{
		Body: io.NopCloser(strings.NewReader("{\"data\":{\"test\": \"data\"}}")),
	}
	testData = map[string]string{}
	assert.Nil(t, client.handleResponse(resp, &testData))
	assert.Equal(t, "data", testData["test"])

	// bad requests with error key
	resp = &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("{\"errors\":\"error content\"}")),
	}
	testData = map[string]string{}
	err = client.handleResponse(resp, &testData)
	assert.NotNil(t, err)
	assert.Equal(t, "bad request:  - \"error content\"", err.Error())

	// bad requests with no errors key
	resp = &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("{\"test\":\"data\"}")),
	}
	testData = map[string]string{}
	err = client.handleResponse(resp, &testData)
	assert.NotNil(t, err)
	assert.Equal(t, "bad request:  - {\"test\":\"data\"}", err.Error())
}
