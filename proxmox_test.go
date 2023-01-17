package proxmox

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

const (
	TestURI = "http://test.localhost"
)

// proxmox api returns everything in a { data: {} } key and thie just abstracts that so the gock JSON calls are cleaner
func data(d map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": d,
	}
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

func TestClient_Version(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Get("/version").
		Reply(200).
		JSON(data(map[string]interface{}{
			"version": "7.3-4",
			"repoid":  "d69b70d4",
			"release": "7.3",
		}))

	v, err := NewClient(TestURI).Version()
	assert.Nil(t, err)
	assert.Equal(t, "7.3-4", v.Version)
	assert.Equal(t, "d69b70d4", v.RepoID)
	assert.Equal(t, "7.3", v.Release)
}

func TestClient_Get(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Get("/test").
		Reply(200).
		JSON(data(map[string]interface{}{
			"test": "data",
		}))

	var testData map[string]string
	client := NewClient(TestURI)
	err := client.Get("/test", &testData)
	assert.Nil(t, err)
	assert.Equal(t, "data", testData["test"])
}

func TestClient_Post(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Post("/test").
		MatchType("json").
		Reply(200).
		JSON(data(map[string]interface{}{
			"test": "data",
		}))

	client := NewClient(TestURI)
	var testData map[string]string
	err := client.Post("/test", map[string]string{"test": "data"}, &testData)
	assert.Nil(t, err)
	assert.Equal(t, "data", testData["test"])
}

func TestClient_Put(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Put("/test").
		Reply(200).
		JSON(data(map[string]interface{}{
			"test": "data",
		}))

	client := NewClient(TestURI)
	var testData map[string]string
	err := client.Put("/test", map[string]string{"test": "data"}, &testData)
	assert.Nil(t, err)
	assert.Equal(t, "data", testData["test"])
}

func TestClient_Delete(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Delete("/test").
		Reply(200).
		JSON(data(map[string]interface{}{
			"test": "data",
		}))

	client := NewClient(TestURI)
	var testData map[string]string
	err := client.Delete("/test", &testData)
	assert.Nil(t, err)
	assert.Equal(t, "data", testData["test"])
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
