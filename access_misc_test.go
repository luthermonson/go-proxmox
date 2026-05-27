package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_AccessIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := mockClient().AccessIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "ticket")
	assert.Contains(t, subdirs, "openid")
	assert.Contains(t, subdirs, "tfa")
}

func TestClient_OpenIDIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := mockClient().OpenIDIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"auth-url", "login"}, subdirs)
}

func TestClient_GetTicket(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, mockClient().GetTicket(context.Background()))
}

func TestClient_VerifyVNCTicket(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := mockClient().VerifyVNCTicket(context.Background(), &VerifyVNCTicketOptions{
		AuthID:    "root@pam",
		Path:      "/nodes/node1/qemu/100",
		Privs:     "VM.Console",
		VNCTicket: "PVE:vncticket123",
		Port:      5901,
	})
	assert.Nil(t, err)

	assert.NotNil(t, mockClient().VerifyVNCTicket(context.Background(), nil))
	assert.NotNil(t, mockClient().VerifyVNCTicket(context.Background(), &VerifyVNCTicketOptions{AuthID: "u"}))
}
