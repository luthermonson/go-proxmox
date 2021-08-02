package proxmox

import (
	"encoding/json"
	"fmt"
)

func (c *Client) Login(credentials Credentials) (Session, error) {
	var session Session
	credJSON, err := json.Marshal(credentials)
	if err != nil {
		return session, err
	}

	if err := c.Post("/access/ticket", credJSON, &session); err != nil {
		return session, err
	}
	c.session = &session

	return session, nil
}

func (c *Client) APIToken(tokenID, secret string) {
	c.token = fmt.Sprintf("%s=%s", tokenID, secret)
}
