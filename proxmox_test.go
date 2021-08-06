package proxmox

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	username           string
	password           string
	tokenID            string
	secret             string
	otp                string
	nodeName           string
	nodeStorage        string
	applianceName      string
	isoURL             string
	insecureHTTPClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func init() {
	username = os.Getenv("PROXMOX_USERNAME")
	password = os.Getenv("PROXMOX_PASSWORD")
	otp = os.Getenv("PROXMOX_OTP")
	tokenID = os.Getenv("PROXMOX_TOKENID")
	secret = os.Getenv("PROXMOX_SECRET")
	nodeName = os.Getenv("PROXMOX_NODE_NAME")
	nodeStorage = os.Getenv("PROXMOX_NODE_STORAGE")
	applianceName = os.Getenv("PROXMOX_CONTAINER_TEMPLATE") // alpine-3.14-default_20210623_amd64.tar.xz
	isoURL = os.Getenv("PROXMOX_ISO_URL")                   // https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/x86_64/alpine-virt-3.14.1-x86_64.iso
}

func ClientFromEnv() *Client {
	return NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
	)
}

func ClientFromLogins() *Client {
	client := NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithLogins(username, password),
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
	client := ClientFromLogins()
	version, err := client.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, version.Version)
}
