package proxmox

import (
	"fmt"
	"net/http"
)

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithLogins(username, password string) Option {
	return func(c *Client) {
		c.credentials = &Credentials{
			Username: username,
			Password: password,
		}
	}
}

func WithAPIToken(tokenID, secret string) Option {
	return func(c *Client) {
		c.token = fmt.Sprintf("%s=%s", tokenID, secret)
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithLogger(logger LeveledLoggerInterface) Option {
	return func(c *Client) {
		c.log = logger
	}
}
