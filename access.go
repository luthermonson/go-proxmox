package proxmox

import (
	"encoding/json"
	"fmt"
)

func (c *Client) Login(username, password string) error {
	_, err := c.Ticket(&Credentials{
		Username: username,
		Password: password,
	})

	return err
}

func (c *Client) APIToken(tokenID, secret string) {
	c.token = fmt.Sprintf("%s=%s", tokenID, secret)
}

func (c *Client) Ticket(credentials *Credentials) (*Session, error) {
	credJSON, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	if err := c.Post("/access/ticket", credJSON, &c.session); err != nil {
		return nil, err
	}

	return c.session, nil
}
