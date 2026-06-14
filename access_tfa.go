package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// This file wraps /access/tfa/* — the modern TFA endpoint family that
// supports listing, adding, updating, and removing individual TFA entries
// per user. The older /access/users/{userid}/tfa surface (which returns the
// summary `TFA` struct) is wrapped in access.go as User.GetTFA and is
// retained for backward compatibility.

// TFAUserEntry is one row in GET /access/tfa — a user that has at least one
// TFA entry configured.
type TFAUserEntry struct {
	UserID    string         `json:"userid,omitempty"`
	Entries   []TFAEntryInfo `json:"entries,omitempty"`
	TOTP      bool           `json:"totp,omitempty"`
	YubicoOTP bool           `json:"yubico,omitempty"`
	U2F       bool           `json:"u2f,omitempty"`
	Webauthn  bool           `json:"webauthn,omitempty"`
	Recovery  []int          `json:"recovery,omitempty"`
}

// TFAEntryInfo is the read shape of a single TFA entry.
type TFAEntryInfo struct {
	ID          string    `json:"id,omitempty"`
	Type        string    `json:"type,omitempty"` // totp | webauthn | u2f | yubico | recovery
	Description string    `json:"description,omitempty"`
	Created     int64     `json:"created,omitempty"`
	Enable      IntOrBool `json:"enable,omitempty"`
}

// TFAEntryOptions is the POST body for adding a TFA entry.
//   - Type is required ("totp" | "webauthn" | "u2f" | "yubico" | "recovery").
//   - For TOTP: set TOTP (the otpauth:// URI) and Value (the current OTP to prove enrollment).
//   - For Webauthn/U2F: set Challenge / Value with the client's signed assertion.
//   - Password is the requesting user's current password (required when changing another user's TFA).
type TFAEntryOptions struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	TOTP        string `json:"totp,omitempty"`
	Value       string `json:"value,omitempty"`
	Challenge   string `json:"challenge,omitempty"`
	Password    string `json:"password,omitempty"`
}

// TFAEntryUpdateOptions is the PUT body for updating an existing entry.
// Only Enable / Description are settable.
type TFAEntryUpdateOptions struct {
	Enable      *bool  `json:"enable,omitempty"`
	Description string `json:"description,omitempty"`
	Password    string `json:"password,omitempty"`
}

// TFAUsers lists every user with at least one TFA entry configured.
func (c *Client) TFAUsers(ctx context.Context) (users []*TFAUserEntry, err error) {
	err = c.Get(ctx, "/access/tfa", &users)
	return
}

// TFAEntries lists all TFA entries for a given user.
func (c *Client) TFAEntries(ctx context.Context, userid string) (entries []*TFAEntryInfo, err error) {
	if userid == "" {
		err = errors.New("userid is required")
		return
	}
	err = c.Get(ctx, fmt.Sprintf("/access/tfa/%s", userid), &entries)
	return
}

// TFAEntry reads a single TFA entry by id.
func (c *Client) TFAEntry(ctx context.Context, userid, id string) (entry *TFAEntryInfo, err error) {
	if userid == "" || id == "" {
		err = errors.New("userid and entry id are required")
		return
	}
	err = c.Get(ctx, fmt.Sprintf("/access/tfa/%s/%s", userid, id), &entry)
	return
}

// NewTFAEntry adds a TFA entry for a user. Returns the new entry id on
// success — PVE wraps it in a {"id": "..."} envelope inside data.
func (c *Client) NewTFAEntry(ctx context.Context, userid string, opts *TFAEntryOptions) (id string, err error) {
	if userid == "" {
		return "", errors.New("userid is required")
	}
	if opts == nil || opts.Type == "" {
		return "", errors.New("tfa entry type is required")
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err = c.Post(ctx, fmt.Sprintf("/access/tfa/%s", userid), opts, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

// UpdateTFAEntry mutates an existing entry (enable/disable, description).
func (c *Client) UpdateTFAEntry(ctx context.Context, userid, id string, opts *TFAEntryUpdateOptions) error {
	if userid == "" || id == "" {
		return errors.New("userid and entry id are required")
	}
	if opts == nil {
		opts = &TFAEntryUpdateOptions{}
	}
	return c.Put(ctx, fmt.Sprintf("/access/tfa/%s/%s", userid, id), opts, nil)
}

// DeleteTFAEntry removes a single TFA entry. password is the caller's current
// password when changing another user's TFA (PVE may require it server-side).
// Pass "" to omit.
func (c *Client) DeleteTFAEntry(ctx context.Context, userid, id, password string) error {
	if userid == "" || id == "" {
		return errors.New("userid and entry id are required")
	}
	path := fmt.Sprintf("/access/tfa/%s/%s", userid, id)
	if password != "" {
		// PVE accepts password via body on DELETE; we pass it as a body map.
		return c.Delete(ctx, path, map[string]string{"password": password})
	}
	return c.Delete(ctx, path, nil)
}

// UnlockUserTFA clears the TFA lockout flag set after too many failed attempts
// for the given user. Unlike User.UnlockTFA (which actually removes the user's
// TFA configuration via the legacy endpoint), this leaves the entries intact —
// it just resets the failure counter. PUT /access/users/{userid}/unlock-tfa.
func (c *Client) UnlockUserTFA(ctx context.Context, userid string) error {
	if userid == "" {
		return errors.New("userid is required")
	}
	return c.Put(ctx, fmt.Sprintf("/access/users/%s/unlock-tfa", userid), nil, nil)
}
