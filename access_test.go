package proxmox

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSession_NilBeforeAuth(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	assert.Nil(t, client.Session())
}

func TestRefreshTicket_NoSession(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	assert.ErrorIs(t, client.RefreshTicket(context.Background()), ErrNoSession)
}

func TestRefreshTicket_Success(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient(WithCredentials(&Credentials{Username: "root@pam", Password: "1234"}))
	ctx := context.Background()

	_, err := client.Ticket(ctx, client.credentials)
	require.NoError(t, err)
	require.NotNil(t, client.Session())

	// Re-register the POST /access/ticket mock so a second call (the refresh) matches.
	gock.New(TestURI).
		Post("^/access/ticket$").
		Reply(200).
		JSON(`{"data": {"username": "root@pam", "ticket": "REFRESHED-TICKET", "CSRFPreventionToken": "REFRESHED-CSRF"}}`)

	require.NoError(t, client.RefreshTicket(ctx))

	s := client.Session()
	require.NotNil(t, s)
	assert.Equal(t, "REFRESHED-TICKET", s.Ticket)
	assert.Equal(t, "REFRESHED-CSRF", s.CSRFPreventionToken)
}

// TestCreateSession_PrefersRefreshOverFullReauth proves that when an existing
// session has expired, CreateSession sends the previous ticket as the password
// (renewal) rather than the originally-stored credential password (full reauth).
// Two distinct mocks are registered, each matching only one of the two possible
// password values, so the test fails if the wrong path is taken.
func TestCreateSession_PrefersRefreshOverFullReauth(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient(WithCredentials(&Credentials{Username: "root@pam", Password: "1234"}))
	ctx := context.Background()

	// Initial auth — consumes the default pve9x POST /access/ticket mock.
	_, err := client.Ticket(ctx, client.credentials)
	require.NoError(t, err)
	originalTicket := client.Session().Ticket
	require.NotEmpty(t, originalTicket)

	// Force the session to look expired so CreateSession does not no-op.
	client.sessionMux.Lock()
	client.sessionExpiresAt = time.Now().Add(-time.Hour)
	client.sessionMux.Unlock()

	// Renewal path: matches when the body carries password=<previous ticket>.
	gock.New(TestURI).
		Post("^/access/ticket$").
		BodyString(`"password":"` + regexp.QuoteMeta(originalTicket) + `"`).
		Reply(200).
		JSON(`{"data": {"username": "root@pam", "ticket": "RENEWED", "CSRFPreventionToken": "RENEWED-CSRF"}}`)

	// Full-reauth path: matches when the body carries the original credential password.
	// If this mock is hit, the test will catch it via the assertion below.
	gock.New(TestURI).
		Post("^/access/ticket$").
		BodyString(`"password":"1234"`).
		Reply(200).
		JSON(`{"data": {"username": "root@pam", "ticket": "FULL-REAUTH", "CSRFPreventionToken": "FULL-REAUTH-CSRF"}}`)

	require.NoError(t, client.CreateSession(ctx))
	assert.Equal(t, "RENEWED", client.Session().Ticket,
		"CreateSession should prefer ticket renewal when an existing session is present")
}

// TestCreateSession_FallsBackToFullReauth proves that when ticket renewal fails
// (e.g., the previous ticket is past its renewable window), CreateSession falls
// back to full credentials reauth instead of bubbling the renewal error.
func TestCreateSession_FallsBackToFullReauth(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient(WithCredentials(&Credentials{Username: "root@pam", Password: "1234"}))
	ctx := context.Background()

	_, err := client.Ticket(ctx, client.credentials)
	require.NoError(t, err)
	originalTicket := client.Session().Ticket
	require.NotEmpty(t, originalTicket)

	client.sessionMux.Lock()
	client.sessionExpiresAt = time.Now().Add(-time.Hour)
	client.sessionMux.Unlock()

	// Renewal attempt is rejected by PVE.
	gock.New(TestURI).
		Post("^/access/ticket$").
		BodyString(`"password":"` + regexp.QuoteMeta(originalTicket) + `"`).
		Reply(401).
		JSON(`{"data": null, "errors": {"password": "ticket expired"}}`)

	// Full-reauth attempt with the original credential password succeeds.
	gock.New(TestURI).
		Post("^/access/ticket$").
		BodyString(`"password":"1234"`).
		Reply(200).
		JSON(`{"data": {"username": "root@pam", "ticket": "FRESH-AFTER-FALLBACK", "CSRFPreventionToken": "FRESH-CSRF"}}`)

	require.NoError(t, client.CreateSession(ctx))
	assert.Equal(t, "FRESH-AFTER-FALLBACK", client.Session().Ticket)
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

func TestNewUser(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	user := NewUser{
		UserID: "test",
	}
	assert.Nil(t, client.NewUser(ctx, &user))

}

func TestAPITokens(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	user := &User{
		client: client,
		UserID: "test",
	}

	apitokens, err := user.GetAPITokens(ctx)
	assert.Nil(t, err)
	assert.Len(t, apitokens, 2)
}

func TestAPIToken(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	User := &User{
		client: client,
		UserID: "root@pam",
	}
	token, err := User.APIToken(ctx, "test")
	assert.Nil(t, err)
	assert.NotNil(t, token)

}

func TestUpdateAPIToken(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	User := &User{
		client: client,
		UserID: "userid",
	}
	token, err := User.UpdateAPIToken(ctx, "tokenid")
	assert.Nil(t, err)
	assert.NotNil(t, token)
}

// func TestNewAPIToken(t *testing.T) {
// mocks.On(mockConfig)
// defer mocks.Off()
// client := mockClient()
// ctx := context.Background()

// // Users
// user := &User{
// client: client,
// UserID: "userid",
// }

// token := &Token{
// TokenID: "test",
// Comment: "test",
// Expire:  0,
// }

// newToken, err := user.NewAPIToken(ctx, *token)
// assert.Nil(t, err)
// // Check if newToken.Value is not empty
// assert.NotEmpty(t, newToken.Value)
// // Check if fullTokenid = userid!tokenid
// assert.Equal(t, "userid!test", newToken.FullTokenID)
// }

func TestDeleteAPIToken(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	User := &User{
		client: client,
		UserID: "root@pam",
	}
	assert.Nil(t, User.DeleteAPIToken(ctx, "test"))
}

func TestNewRole(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	assert.Nil(t, client.NewRole(ctx, "test", "test"))
}

func TestGetTFA(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	user := &User{
		client: client,
		UserID: "userid",
	}

	tfa, err := user.GetTFA(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "userid", tfa.User)
}

func TestUnlockTFA(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	user := &User{
		client: client,
		UserID: "userid",
	}
	assert.Nil(t, user.UnlockTFA(ctx))
}

func TestClient_Login_Deprecated(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// Hits POST /access/ticket (the default mock returns a valid session).
	assert.Nil(t, client.Login(ctx, "root@pam", "1234"))
	assert.NotNil(t, client.Session())
}

func TestClient_APIToken_Deprecated(t *testing.T) {
	client := mockClient()
	client.APIToken("root@pam!t", "secret")
	assert.Equal(t, "root@pam!t=secret", client.token)
}

func TestClient_UpdateACL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	err := client.UpdateACL(ctx, ACLOptions{
		Path:      "/",
		Roles:     "Administrator",
		Users:     "alice@pve",
		Propagate: IntOrBool(true),
	})
	assert.Nil(t, err)
}

func TestUser_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	user := &User{client: client, UserID: "userid"}
	assert.Nil(t, user.Update(ctx, UserOptions{Comment: "x"}))
}

func TestUser_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	user := &User{client: client, UserID: "userid"}
	assert.Nil(t, user.Delete(ctx))
}

func TestUser_NewAPIToken(t *testing.T) {
	// Don't use the shared mockConfig pve9x fixtures: the generic
	// Post("^/access/users") mock there matches before any per-token mock,
	// shadowing this endpoint's response. Run with a minimal gock setup
	// scoped to this test.
	defer gock.Off()

	gock.New(TestURI).
		Post("^/access/users/test-user/token/newtok$").
		Reply(200).
		JSON(`{"data":{"full-tokenid":"test-user!newtok","value":"sekret","info":{"privsep":0}}}`)

	client := mockClient()
	ctx := context.Background()

	user := &User{client: client, UserID: "test-user"}
	newToken, err := user.NewAPIToken(ctx, Token{TokenID: "newtok", Comment: "x"})
	assert.Nil(t, err)
	assert.NotEmpty(t, newToken.Value)
	assert.Equal(t, "test-user!newtok", newToken.FullTokenID)
}

func TestRole_Update_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	role := &Role{client: client, RoleID: "test-role"}
	assert.Nil(t, role.Update(ctx))
	assert.Nil(t, role.Delete(ctx))
}
