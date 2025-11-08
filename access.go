package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Deprecated: Use WithCredentials Option
func (c *Client) Login(ctx context.Context, username, password string) error {
	c.sessionMux.Lock()
	defer c.sessionMux.Unlock()

	_, err := c.Ticket(ctx, &Credentials{
		Username: username,
		Password: password,
	})

	return err
}

// Deprecated: Use the WithAPIToken Option
func (c *Client) APIToken(tokenID, secret string) {
	c.token = fmt.Sprintf("%s=%s", tokenID, secret)
}

func (c *Client) CreateSession(ctx context.Context) error {
	c.sessionMux.Lock()
	defer c.sessionMux.Unlock()

	if c.session != nil {
		return ErrSessionExists
	}

	if _, err := c.Ticket(ctx, c.credentials); err != nil {
		return err
	}

	return nil
}

func (c *Client) Ticket(ctx context.Context, credentials *Credentials) (*Session, error) {
	return c.session, c.Post(ctx, "/access/ticket", credentials, &c.session)
}

func (c *Client) ACL(ctx context.Context) (acl ACLs, err error) {
	return acl, c.Get(ctx, "/access/acl", &acl)
}

func (c *Client) UpdateACL(ctx context.Context, aclOptions ACLOptions) error {
	return c.Put(ctx, "/access/acl", &aclOptions, nil)
}

// Permissions get permissions for the current user for the client which passes no params, use Permission
func (c *Client) Permissions(ctx context.Context, o *PermissionsOptions) (permissions Permissions, err error) {
	u := url.URL{Path: "/access/permissions"}

	if o != nil { // params are optional
		params := url.Values{}
		if o.UserID != "" {
			params.Add("userid", o.UserID)
		}
		if o.Path != "" {
			params.Add("path", o.Path)
		}
		u.RawQuery = params.Encode()
	}

	return permissions, c.Get(ctx, u.String(), &permissions)
}

func (c *Client) Password(ctx context.Context, userid, password string) error {
	return c.Put(ctx, "/access/password", map[string]string{
		"userid":   userid,
		"password": password,
	}, nil)
}

// NewDomain create a new domain with the required two parameters pull it and use domain.Update to configure
func (c *Client) NewDomain(ctx context.Context, realm string, domainType DomainType) error {
	return c.Post(ctx, "/access/domains", map[string]string{
		"realm": realm,
		"type":  string(domainType),
	}, nil)
}

func (c *Client) Domain(ctx context.Context, realm string) (domain *Domain, err error) {
	err = c.Get(ctx, fmt.Sprintf("/access/domains/%s", realm), &domain)
	if nil == err {
		domain.Realm = realm
		domain.client = c
	}
	return
}

func (c *Client) Domains(ctx context.Context) (domains Domains, err error) {
	err = c.Get(ctx, "/access/domains", &domains)
	if nil == err {
		for _, d := range domains {
			d.client = c
		}
	}
	return
}

func (d *Domain) Update(ctx context.Context) error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Put(ctx, fmt.Sprintf("/access/domains/%s", d.Realm), d, nil)
}

func (d *Domain) Delete(ctx context.Context) error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Delete(ctx, fmt.Sprintf("/access/domains/%s", d.Realm), nil)
}

func (d *Domain) Sync(ctx context.Context, options DomainSyncOptions) error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Post(ctx, fmt.Sprintf("/access/domains/%s", d.Realm), options, nil)
}

// NewGroup makes a new group, comment is option and can be left empty
func (c *Client) NewGroup(ctx context.Context, groupid, comment string) error {
	return c.Post(ctx, "/access/groups", map[string]string{
		"groupid": groupid,
		"comment": comment,
	}, nil)
}

func (c *Client) Group(ctx context.Context, groupid string) (group *Group, err error) {
	err = c.Get(ctx, fmt.Sprintf("/access/groups/%s", groupid), &group)
	if nil == err {
		group.GroupID = groupid
		group.client = c
	}
	return
}

func (c *Client) Groups(ctx context.Context) (groups Groups, err error) {
	err = c.Get(ctx, "/access/groups", &groups)
	if nil == err {
		for _, g := range groups {
			g.client = c
		}
	}
	return
}

func (g *Group) Update(ctx context.Context) error {
	return g.client.Put(ctx, fmt.Sprintf("/access/groups/%s", g.GroupID), g, nil)
}

func (g *Group) Delete(ctx context.Context) error {
	return g.client.Delete(ctx, fmt.Sprintf("/access/groups/%s", g.GroupID), nil)
}

func (c *Client) User(ctx context.Context, userid string) (user *User, err error) {
	err = c.Get(ctx, fmt.Sprintf("/access/users/%s", userid), &user)
	if nil == err {
		user.UserID = userid
		user.client = c
	}
	return
}

func (c *Client) Users(ctx context.Context) (users Users, err error) {
	err = c.Get(ctx, "/access/users", &users)
	if nil == err {
		for _, g := range users {
			g.client = c
		}
	}
	return
}

func (c *Client) NewUser(ctx context.Context, user *NewUser) (err error) {
	return c.Post(ctx, "/access/users", user, nil)
}

func (u *User) Update(ctx context.Context, options UserOptions) error {
	return u.client.Put(ctx, fmt.Sprintf("/access/users/%s", u.UserID), &options, nil)
}

func (u *User) Delete(ctx context.Context) error {
	return u.client.Delete(ctx, fmt.Sprintf("/access/users/%s", u.UserID), nil)
}

func (u *User) GetAPITokens(ctx context.Context) (tokens Tokens, err error) {
	return tokens, u.client.Get(ctx, fmt.Sprintf("/access/users/%s/token", u.UserID), &tokens)
}

func (u *User) APIToken(ctx context.Context, tokenid string) (token Token, err error) {
	return token, u.client.Get(ctx, fmt.Sprintf("/access/users/%s/token/%s", u.UserID, tokenid), &token)
}

func (u *User) NewAPIToken(ctx context.Context, token Token) (newtoken NewAPIToken, err error) {
	return newtoken, u.client.Post(ctx, fmt.Sprintf("/access/users/%s/token/%s", u.UserID, token.TokenID), token, &newtoken)
}

func (u *User) UpdateAPIToken(ctx context.Context, tokenid string) (token Token, err error) {
	return token, u.client.Put(ctx, fmt.Sprintf("/access/users/%s/token/%s", u.UserID, tokenid), token, nil)
}

func (u *User) DeleteAPIToken(ctx context.Context, tokenid string) error {
	return u.client.Delete(ctx, fmt.Sprintf("/access/users/%s/token/%s", u.UserID, tokenid), nil)
}

func (u *User) GetTFA(ctx context.Context) (tfa TFA, err error) {
	return tfa, u.client.Get(ctx, fmt.Sprintf("/access/users/%s/tfa", u.UserID), &tfa)
}

func (u *User) UnlockTFA(ctx context.Context) error {
	return u.client.Delete(ctx, fmt.Sprintf("/access/users/%s/tfa", u.UserID), nil)
}

func (c *Client) Role(ctx context.Context, roleid string) (role Permission, err error) {
	err = c.Get(ctx, fmt.Sprintf("/access/roles/%s", roleid), &role)
	return
}

func (c *Client) NewRole(ctx context.Context, roleID string, privs string) (err error) {
	return c.Post(ctx, "/access/roles", map[string]string{
		"roleid": roleID,
		"privs":  privs,
	}, nil)
}

func (c *Client) Roles(ctx context.Context) (roles Roles, err error) {
	err = c.Get(ctx, "/access/roles", &roles)
	if nil == err {
		for _, g := range roles {
			g.client = c
		}
	}
	return
}

func (r *Role) Update(ctx context.Context) error {
	return r.client.Put(ctx, fmt.Sprintf("/access/roles/%s", r.RoleID), r, nil)
}

func (r *Role) Delete(ctx context.Context) error {
	return r.client.Delete(ctx, fmt.Sprintf("/access/roles/%s", r.RoleID), nil)
}
