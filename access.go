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
	var res string
	return c.Post("/access/password", map[string]string{"userid": userid, "password": password}, &res)
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

func (c *Client) Domain(realm string) (domain *Domain, err error) {
	err = c.Get(fmt.Sprintf("/access/domains/%s", realm), &domain)
	if nil == err {
		domain.Realm = realm
		domain.client = c
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
	var ret string
	return d.client.Delete(fmt.Sprintf("/access/domains/%s", d.Realm), &ret)
}

func (d *Domain) Sync(options DomainSyncOptions) error {
	if d.Realm == "" {
		return errors.New("realm can not be empty")
	}
	var ret string
	return d.client.Post(fmt.Sprintf("/access/domains/%s", d.Realm), options, &ret)
}
