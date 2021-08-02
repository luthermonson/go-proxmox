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

func WithCredentials(credentials Credentials) Option {
	return func(c *Client) {
		c.Login(credentials)
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
