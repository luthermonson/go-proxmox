package proxmox

import (
	"fmt"
	"net/http"
)

type Option func(*Client)

// Deprecated: Use WithHTTPClient
func WithClient(client *http.Client) Option {
	return WithHTTPClient(client)
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Deprecated: Use WithCredential
func WithLogins(username, password string) Option {
	return WithCredentials(&Credentials{
		Username: username,
		Password: password,
	})
}

func WithCredentials(credentials *Credentials) Option {
	return func(c *Client) {
		c.credentials = credentials
	}
}

func WithAPIToken(tokenID, secret string) Option {
	return func(c *Client) {
		c.token = fmt.Sprintf("%s=%s", tokenID, secret)
	}
}

// WithSession experimental
func WithSession(ticket, CSRFPreventionToken string) Option {
	return func(c *Client) {
		c.session = &Session{
			Ticket:              ticket,
			CSRFPreventionToken: CSRFPreventionToken,
		}
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
