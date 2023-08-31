package proxmox

import (
	"errors"
	"fmt"
	"net/url"
)

// Deprecated: Use WithCredentials Option
func (c *Client) Login(username, password string) error {
	_, err := c.Ticket(&Credentials{
		Username: username,
		Password: password,
	})

	return err
}

// Deprecated: Use the WithAPIToken Option
func (c *Client) APIToken(tokenID, secret string) {
	c.token = fmt.Sprintf("%s=%s", tokenID, secret)
}

func (c *Client) Ticket(credentials *Credentials) (*Session, error) {
	return c.session, c.Post("/access/ticket", credentials, &c.session)
}

func (c *Client) ACL() (acl ACLs, err error) {
	return acl, c.Get("/access/acl", &acl)
}

func (c *Client) UpdateACL(acl ACL) error {
	return c.Put("/access/acl", &acl, nil)
}

// Permissions get permissions for the current user for the client which passes no params, use Permission
func (c *Client) Permissions(o *PermissionsOptions) (permissions Permissions, err error) {
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

	return permissions, c.Get(u.String(), &permissions)
}

func (c *Client) Password(userid, password string) error {
	return c.Post("/access/password", map[string]string{
		"userid":   userid,
		"password": password,
	}, nil)
}

// NewDomain create a new domain with the required two parameters pull it and use domain.Update to configure
func (c *Client) NewDomain(realm string, domainType DomainType) error {
	return c.Post("/access/domains", map[string]string{
		"realm": realm,
		"type":  string(domainType),
	}, nil)
}

func (c *Client) Domain(realm string) (domain *Domain, err error) {
	err = c.Get(fmt.Sprintf("/access/domains/%s", realm), &domain)
	if nil == err {
		domain.Realm = realm
		domain.client = c
	}
	return
}

func (c *Client) Domains() (domains Domains, err error) {
	err = c.Get("/access/domains", &domains)
	if nil == err {
		for _, d := range domains {
			d.client = c
		}
	}
	return
}

func (d *Domain) Update() error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Put(fmt.Sprintf("/access/domains/%s", d.Realm), d, nil)
}

func (d *Domain) Delete() error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Delete(fmt.Sprintf("/access/domains/%s", d.Realm), nil)
}

func (d *Domain) Sync(options DomainSyncOptions) error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	return d.client.Post(fmt.Sprintf("/access/domains/%s", d.Realm), options, nil)
}

// NewGroup makes a new group, comment is option and can be left empty
func (c *Client) NewGroup(groupid, comment string) error {
	return c.Post("/access/groups", map[string]string{
		"groupid": groupid,
		"comment": comment,
	}, nil)
}

func (c *Client) Group(groupid string) (group *Group, err error) {
	err = c.Get(fmt.Sprintf("/access/groups/%s", groupid), &group)
	if nil == err {
		group.GroupID = groupid
		group.client = c
	}
	return
}

func (c *Client) Groups() (groups Groups, err error) {
	err = c.Get("/access/groups", &groups)
	if nil == err {
		for _, g := range groups {
			g.client = c
		}
	}
	return
}

func (g *Group) Update() error {
	return g.client.Put(fmt.Sprintf("/access/groups/%s", g.GroupID), g, nil)
}

func (g *Group) Delete() error {
	return g.client.Delete(fmt.Sprintf("/access/groups/%s", g.GroupID), nil)
}

func (c *Client) User(userid string) (user *User, err error) {
	err = c.Get(fmt.Sprintf("/access/users/%s", userid), &user)
	if nil == err {
		user.UserID = userid
		user.client = c
	}
	return
}

func (c *Client) Users() (users Users, err error) {
	err = c.Get("/access/users", &users)
	if nil == err {
		for _, g := range users {
			g.client = c
		}
	}
	return
}

func (u *User) Update() error {
	return u.client.Put(fmt.Sprintf("/access/users/%s", u.UserID), u, nil)
}

func (u *User) Delete() error {
	return u.client.Delete(fmt.Sprintf("/access/users/%s", u.UserID), nil)
}

func (c *Client) Role(roleid string) (role Permission, err error) {
	err = c.Get(fmt.Sprintf("/access/roles/%s", roleid), &role)
	return
}

func (c *Client) Roles() (roles Roles, err error) {
	err = c.Get("/access/roles", &roles)
	if nil == err {
		for _, g := range roles {
			g.client = c
		}
	}
	return
}

func (r *Role) Update() error {
	return r.client.Put(fmt.Sprintf("/access/roles/%s", r.RoleID), r, nil)
}

func (r *Role) Delete() error {
	return r.client.Delete(fmt.Sprintf("/access/roles/%s", r.RoleID), nil)
}
