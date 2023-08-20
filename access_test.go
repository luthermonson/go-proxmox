package proxmox

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luthermonson/go-proxmox/tests/mocks"
)

func TestTicket(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// todo current mocks are hardcoded with test data, make configurable via mock config
	client := mockClient(WithCredentials(
		&Credentials{
			Username: "root@pam",
			Password: "1234",
		}))

	session, err := client.Ticket(client.credentials)
	assert.Nil(t, err)
	assert.Equal(t, "root@pam", session.Username)
	assert.Equal(t, "pve-cluster", session.ClusterName)
}

func TestPermissions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	perms, err := client.Permissions(nil)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(perms))
	assert.Equal(t, 1, perms["/"]["Datastore.Allocate"])

	// test path option
	perms, err = client.Permissions(&PermissionsOptions{
		Path: "path",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, perms["path"]["permission"])

	// test userid
	perms, err = client.Permissions(&PermissionsOptions{
		UserID: "userid",
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, perms["path"]["permission"])

	// test both path and userid
	perms, err = client.Permissions(&PermissionsOptions{
		UserID: "userid",
		Path:   "path",
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, perms["path"]["permission"])
}
