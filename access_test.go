package proxmox

import (
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, IntOrBool(true), perms["/"]["Datastore.Allocate"])

	// test path option
	perms, err = client.Permissions(&PermissionsOptions{
		Path: "path",
	})
	assert.Nil(t, err)
	assert.Equal(t, IntOrBool(true), perms["path"]["permission"])

	// test userid
	perms, err = client.Permissions(&PermissionsOptions{
		UserID: "userid",
	})
	assert.Nil(t, err)
	assert.Equal(t, IntOrBool(true), perms["path"]["permission"])

	// test both path and userid
	perms, err = client.Permissions(&PermissionsOptions{
		UserID: "userid",
		Path:   "path",
	})
	assert.Nil(t, err)
	assert.Equal(t, IntOrBool(true), perms["path"]["permission"])
}

func TestPassword(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	assert.Nil(t, client.Password("userid", "password"))
}

func TestDomains(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	ds, err := client.Domains()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ds))
}

func TestDomain(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	d, err := client.Domain("test")
	assert.Nil(t, err)
	assert.Equal(t, d.Realm, "test")
	assert.False(t, bool(d.AutoCreate))
}

func TestGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	g, err := client.Group("test")
	assert.Nil(t, err)
	assert.Equal(t, g.GroupID, "test")
	assert.Len(t, g.Members, 2)
}

func TestGroups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	gs, err := client.Groups()
	assert.Nil(t, err)
	assert.Len(t, gs, 2)
	for _, g := range gs {
		assert.Len(t, g.Members, 0) // empty from lister
		assert.NotEmpty(t, g.Users)
	}
}

func TestUser(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	u, err := client.User("root@pam")
	assert.Nil(t, err)
	assert.Equal(t, u.UserID, "root@pam")
}

func TestUsers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	users, err := client.Users()
	assert.Nil(t, err)
	assert.Len(t, users, 4)
}

func TestRole(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	u, err := client.Role("Administrator")
	assert.Nil(t, err)
	assert.Contains(t, u, "SDN.Allocate")
	assert.Len(t, u, 38)

	u, err = client.Role("NoAccess")
	assert.Nil(t, err)
	assert.Len(t, u, 0)
}

func TestRoles(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	roles, err := client.Roles()
	assert.Nil(t, err)
	assert.Len(t, roles, 16)
}

func TestACL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	acls, err := client.ACL()
	assert.Nil(t, err)
	assert.Len(t, acls, 1)
}

func TestNewDomain(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	assert.Nil(t, client.NewDomain("test", "t"))
}

func TestDomain_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Update())
	domain.Realm = "test"
	assert.Nil(t, domain.Update())
}

func TestDomain_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Delete())
	domain.Realm = "test"
	assert.Nil(t, domain.Delete())
}

func TestDomain_Sync(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Sync(DomainSyncOptions{}))
	domain.Realm = "test"
	assert.Nil(t, domain.Sync(DomainSyncOptions{}))
}
