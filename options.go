package proxmox

import "net/http"

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}
