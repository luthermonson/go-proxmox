package proxmox

import (
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
