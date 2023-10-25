package proxmox

import (
	"context"
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
	ctx := context.Background()

	session, err := client.Ticket(ctx, client.credentials)
	assert.Nil(t, err)
	assert.Equal(t, "root@pam", session.Username)
	assert.Equal(t, "pve-cluster", session.ClusterName)
}

func TestPermissions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	perms, err := client.Permissions(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(perms))
	assert.Equal(t, IntOrBool(true), perms["/"]["Datastore.Allocate"])

	// test path option
	perms, err = client.Permissions(ctx, &PermissionsOptions{
		Path: "path",
	})
	assert.Nil(t, err)
	assert.Equal(t, IntOrBool(true), perms["path"]["permission"])

	// test userid
	perms, err = client.Permissions(ctx, &PermissionsOptions{
		UserID: "userid",
	})
	assert.Nil(t, err)
	assert.Equal(t, IntOrBool(true), perms["path"]["permission"])

	// test both path and userid
	perms, err = client.Permissions(ctx, &PermissionsOptions{
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
	ctx := context.Background()

	assert.Nil(t, client.Password(ctx, "userid", "password"))
}

func TestDomains(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	ds, err := client.Domains(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(ds))
}

func TestDomain(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	d, err := client.Domain(ctx, "test")
	assert.Nil(t, err)
	assert.Equal(t, d.Realm, "test")
	assert.False(t, bool(d.AutoCreate))
}

func TestNewGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	assert.Nil(t, client.NewGroup(ctx, "groupid", "comment"))
}

func TestGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	g, err := client.Group(ctx, "test")
	assert.Nil(t, err)
	assert.Equal(t, g.GroupID, "test")
	assert.Len(t, g.Members, 2)
}

func TestGroups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	gs, err := client.Groups(ctx)
	assert.Nil(t, err)
	assert.Len(t, gs, 2)
	for _, g := range gs {
		assert.Len(t, g.Members, 0) // empty from lister
		assert.NotEmpty(t, g.Users)
	}
}

func TestGroup_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	group := Group{
		client: client,
	}
	assert.Error(t, group.Update(ctx)) // no groupid
	group.GroupID = "groupid"
	assert.Nil(t, group.Update(ctx))
}

func TestGroup_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	group := Group{
		client: client,
	}

	assert.Error(t, group.Delete(ctx))
	group.GroupID = "groupid"
	assert.Nil(t, group.Delete(ctx))
}

func TestUser(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	u, err := client.User(ctx, "root@pam")
	assert.Nil(t, err)
	assert.Equal(t, u.UserID, "root@pam")
}

func TestUsers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	users, err := client.Users(ctx)
	assert.Nil(t, err)
	assert.Len(t, users, 4)
}

func TestRole(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	u, err := client.Role(ctx, "Administrator")
	assert.Nil(t, err)
	assert.Contains(t, u, "SDN.Allocate")
	assert.Len(t, u, 38)

	u, err = client.Role(ctx, "NoAccess")
	assert.Nil(t, err)
	assert.Len(t, u, 0)
}

func TestRoles(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	roles, err := client.Roles(ctx)
	assert.Nil(t, err)
	assert.Len(t, roles, 16)
}

func TestACL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	acls, err := client.ACL(ctx)
	assert.Nil(t, err)
	assert.Len(t, acls, 1)
}

func TestNewDomain(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	assert.Nil(t, client.NewDomain(ctx, "test", "t"))
}

func TestDomain_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Update(ctx))
	domain.Realm = "test"
	assert.Nil(t, domain.Update(ctx))
}

func TestDomain_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Delete(ctx))
	domain.Realm = "test"
	assert.Nil(t, domain.Delete(ctx))
}

func TestDomain_Sync(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// no realm name
	domain := Domain{
		client: client,
	}

	assert.Error(t, domain.Sync(ctx, DomainSyncOptions{}))
	domain.Realm = "test"
	assert.Nil(t, domain.Sync(ctx, DomainSyncOptions{}))
}
