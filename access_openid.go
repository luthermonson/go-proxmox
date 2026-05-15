package proxmox

import (
	"context"
	"errors"
)

// OpenIDAuthURLResponse is what PVE returns from POST /access/openid/auth-url
// — a URL the caller redirects the browser to so the user can authenticate
// with the configured OIDC provider.
type OpenIDAuthURLResponse string

// OpenIDLoginResponse is the post-callback exchange result — same shape as
// the regular /access/ticket login response (ticket + CSRF token + user).
type OpenIDLoginResponse struct {
	Ticket              string `json:"ticket,omitempty"`
	CSRFPreventionToken string `json:"CSRFPreventionToken,omitempty"`
	Username            string `json:"username,omitempty"`
	Cap                 any    `json:"cap,omitempty"`
	ClusterName         string `json:"clustername,omitempty"`
}

// OpenIDAuthURL kicks off the OIDC flow. realm names the configured PVE OIDC
// realm; redirectURL is where the IdP will redirect after authentication
// (must match what's registered with the IdP). Returns the URL the caller
// should send the browser to.
func (c *Client) OpenIDAuthURL(ctx context.Context, realm, redirectURL string) (string, error) {
	if realm == "" || redirectURL == "" {
		return "", errors.New("realm and redirect-url are required")
	}
	body := map[string]string{
		"realm":        realm,
		"redirect-url": redirectURL,
	}
	var url string
	if err := c.Post(ctx, "/access/openid/auth-url", body, &url); err != nil {
		return "", err
	}
	return url, nil
}

// OpenIDLogin completes the OIDC dance after the IdP has redirected back to
// our redirectURL with code + state query params. PVE returns a normal
// session ticket on success.
func (c *Client) OpenIDLogin(ctx context.Context, code, state, redirectURL string) (*OpenIDLoginResponse, error) {
	if code == "" || state == "" || redirectURL == "" {
		return nil, errors.New("code, state, and redirect-url are required")
	}
	body := map[string]string{
		"code":         code,
		"state":        state,
		"redirect-url": redirectURL,
	}
	resp := &OpenIDLoginResponse{}
	if err := c.Post(ctx, "/access/openid/login", body, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
