package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_OpenIDAuthURL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	url, err := mockClient().OpenIDAuthURL(context.Background(), "oidc1", "https://pve.example.com/callback")
	assert.Nil(t, err)
	assert.Contains(t, url, "https://idp.example.com")

	_, err = mockClient().OpenIDAuthURL(context.Background(), "", "https://pve.example.com/callback")
	assert.NotNil(t, err)

	_, err = mockClient().OpenIDAuthURL(context.Background(), "oidc1", "")
	assert.NotNil(t, err)
}

func TestClient_OpenIDLogin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	resp, err := mockClient().OpenIDLogin(context.Background(), "code123", "state456", "https://pve.example.com/callback")
	assert.Nil(t, err)
	assert.Equal(t, "alice@pve", resp.Username)
	assert.NotEmpty(t, resp.Ticket)
	assert.NotEmpty(t, resp.CSRFPreventionToken)

	_, err = mockClient().OpenIDLogin(context.Background(), "", "state", "url")
	assert.NotNil(t, err)
}
