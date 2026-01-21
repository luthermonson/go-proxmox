package proxmox

import (
	"context"
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
