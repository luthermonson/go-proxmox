package proxmox

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	creds              Credentials
	tokenID            string
	secret             string
	insecureHTTPClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func init() {
	creds.Username = os.Getenv("PROXMOX_USERNAME")
	creds.Password = os.Getenv("PROXMOX_PASSWORD")
	creds.Otp = os.Getenv("PROXMOX_OTP")
	tokenID = os.Getenv("PROXMOX_TOKENID")
	secret = os.Getenv("PROXMOX_SECRET")
}

func ClientFromEnv() *Client {
	return NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
	)
}

func ClientFromCredentials() *Client {
	client := NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithCredentials(creds),
	)

	return client
}

func ClientFromToken() *Client {
	return NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithAPIToken(tokenID, secret),
	)
}

func TestVersion(t *testing.T) {
	client := ClientFromCredentials()
	version, err := client.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, version.Version)
}
