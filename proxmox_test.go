package proxmox

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, test.input, test.expect)
	}
}
