package proxmox

import (
	"context"
	"errors"
)

// Trailing /access/* gaps — directory indexes and the VNC-ticket verifier.
// The big surfaces (Ticket, OpenIDLogin, TFA*) already live in access.go,
// access_openid.go, and access_tfa.go.

// AccessIndex enumerates the children of /access ("ticket", "openid",
// "tfa", "permissions", "password", "domains", ...).
func (c *Client) AccessIndex(ctx context.Context) ([]string, error) {
	var items []struct {
		Subdir string `json:"subdir"`
	}
	if err := c.Get(ctx, "/access", &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Subdir)
	}
	return out, nil
}

// OpenIDIndex enumerates the children of /access/openid ("auth-url",
// "login").
func (c *Client) OpenIDIndex(ctx context.Context) ([]string, error) {
	var items []struct {
		Subdir string `json:"subdir"`
	}
	if err := c.Get(ctx, "/access/openid", &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Subdir)
	}
	return out, nil
}

// GetTicket is a no-op placeholder that PVE exposes for HTML formatters
// that want a "login page" URL (GET /access/ticket). The endpoint returns
// null; the wrapper exists for surface coverage and can double as a
// world-permission liveness probe (no auth required).
func (c *Client) GetTicket(ctx context.Context) error {
	return c.Get(ctx, "/access/ticket", nil)
}

// VerifyVNCTicketOptions is the POST body for /access/vncticket. AuthID,
// Path, and VNCTicket are required; Privs lists privileges to check on
// Path; Port (optional) binds the verification to a specific VNC port.
type VerifyVNCTicketOptions struct {
	AuthID    string `json:"authid"`
	Path      string `json:"path"`
	Privs     string `json:"privs"`
	VNCTicket string `json:"vncticket"`
	Port      int    `json:"port,omitempty"`
}

// VerifyVNCTicket verifies a VNC ticket previously issued by a vncshell /
// vncproxy call. PVE returns null on success and 401 on failure. Useful
// for spice/vnc gateway services that re-authenticate clients.
func (c *Client) VerifyVNCTicket(ctx context.Context, opts *VerifyVNCTicketOptions) error {
	if opts == nil || opts.AuthID == "" || opts.Path == "" || opts.VNCTicket == "" {
		return errors.New("authid, path, and vncticket are required")
	}
	return c.Post(ctx, "/access/vncticket", opts, nil)
}
