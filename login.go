package proxmox

import (
	"encoding/json"
)

func (c *Client) Login(credentials Credentials) (*Session, error) {
	var session *Session
	credJSON, err := json.Marshal(credentials)
	if err != nil {
		return session, err
	}

	if err := c.Post("/access/ticket", credJSON, &session); err != nil {
		return session, err
	}

	return session, nil
}
