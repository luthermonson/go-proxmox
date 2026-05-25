package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// --- ACME directory / metadata (read-only discovery) -----------------------

// ACMEDirectories returns the list of ACME CA directories PVE knows about
// (Let's Encrypt prod, staging, etc.).
func (cl *Cluster) ACMEDirectories(ctx context.Context) (dirs []*ACMEDirectory, err error) {
	err = cl.client.Get(ctx, "/cluster/acme/directories", &dirs)
	return
}

// ACMEChallengeSchema returns the catalog of supported challenge plugin
// schemas (DNS providers etc.) PVE can configure.
func (cl *Cluster) ACMEChallengeSchema(ctx context.Context) (schemas []*ACMEChallengeSchema, err error) {
	err = cl.client.Get(ctx, "/cluster/acme/challenge-schema", &schemas)
	return
}

// ACMETermsOfService returns the URL of the CA's ToS document for a given
// directory URL. directory is optional — PVE defaults to Let's Encrypt prod.
func (cl *Cluster) ACMETermsOfService(ctx context.Context, directory string) (tosURL string, err error) {
	path := "/cluster/acme/tos"
	if directory != "" {
		q := url.Values{}
		q.Set("directory", directory)
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &tosURL)
	return
}

// ACMEMeta returns the metadata document of an ACME CA directory (caa
// identities, EAB requirement, ToS URL, website). directory is optional.
func (cl *Cluster) ACMEMeta(ctx context.Context, directory string) (meta *ACMEMeta, err error) {
	path := "/cluster/acme/meta"
	if directory != "" {
		q := url.Values{}
		q.Set("directory", directory)
		path = path + "?" + q.Encode()
	}
	meta = &ACMEMeta{}
	err = cl.client.Get(ctx, path, meta)
	return
}

// --- ACME accounts ---------------------------------------------------------

// ACMEAccounts lists configured ACME accounts on the cluster.
func (cl *Cluster) ACMEAccounts(ctx context.Context) (accounts []*ACMEAccountIndex, err error) {
	err = cl.client.Get(ctx, "/cluster/acme/account", &accounts)
	return
}

// ACMEAccount returns the full account record (contact, directory URL, EAB
// settings, account JSON from the CA). Pass "" to read the "default" account.
func (cl *Cluster) ACMEAccount(ctx context.Context, name string) (account *ACMEAccount, err error) {
	if name == "" {
		name = "default"
	}
	account = &ACMEAccount{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/acme/account/%s", name), account)
	return
}

// NewACMEAccount registers a new ACME account with the CA. PVE runs this as
// a task because it does a real HTTP round-trip to the CA. opts.Contact is
// required; everything else (including Name) defaults sensibly.
func (cl *Cluster) NewACMEAccount(ctx context.Context, opts *ACMEAccountOptions) (*Task, error) {
	if opts == nil || opts.Contact == "" {
		return nil, errors.New("acme account contact is required")
	}
	var upid UPID
	if err := cl.client.Post(ctx, "/cluster/acme/account", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// UpdateACMEAccount mutates an existing account. Only `contact` is mutable
// per the PVE schema; the rest is set at creation. Pass "" for name to
// target the "default" account.
func (cl *Cluster) UpdateACMEAccount(ctx context.Context, name, contact string) (*Task, error) {
	if name == "" {
		name = "default"
	}
	body := map[string]any{}
	if contact != "" {
		body["contact"] = contact
	}
	var upid UPID
	if err := cl.client.Put(ctx, fmt.Sprintf("/cluster/acme/account/%s", name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// DeleteACMEAccount deactivates an ACME account with the CA and removes it
// from PVE. Async because PVE calls the CA's deactivate endpoint.
func (cl *Cluster) DeleteACMEAccount(ctx context.Context, name string) (*Task, error) {
	if name == "" {
		name = "default"
	}
	var upid UPID
	if err := cl.client.Delete(ctx, fmt.Sprintf("/cluster/acme/account/%s", name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// --- ACME plugins (DNS challenge providers) --------------------------------

// ACMEPlugins lists configured ACME challenge plugins (DNS providers,
// standalone, etc.). Pass a non-empty pluginType to filter by challenge
// type (e.g. "dns", "standalone").
func (cl *Cluster) ACMEPlugins(ctx context.Context, pluginType string) (plugins []*ACMEPlugin, err error) {
	path := "/cluster/acme/plugins"
	if pluginType != "" {
		q := url.Values{}
		q.Set("type", pluginType)
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &plugins)
	return
}

// ACMEPlugin reads a single plugin's configuration.
func (cl *Cluster) ACMEPlugin(ctx context.Context, id string) (plugin *ACMEPlugin, err error) {
	if id == "" {
		err = errors.New("acme plugin id can not be empty")
		return
	}
	plugin = &ACMEPlugin{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/acme/plugins/%s", id), plugin)
	return
}

// NewACMEPlugin creates a new ACME challenge plugin. opts.ID and opts.Type
// ("dns" | "standalone") are required by PVE.
func (cl *Cluster) NewACMEPlugin(ctx context.Context, opts *ACMEPluginOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("acme plugin id can not be empty")
	}
	if opts.Type == "" {
		return errors.New("acme plugin type is required")
	}
	return cl.client.Post(ctx, "/cluster/acme/plugins", opts, nil)
}

// UpdateACMEPlugin mutates an existing plugin. Pass opts.Delete to reset a
// comma-separated list of keys back to their PVE defaults.
func (cl *Cluster) UpdateACMEPlugin(ctx context.Context, id string, opts *ACMEPluginOptions) error {
	if id == "" {
		return errors.New("acme plugin id can not be empty")
	}
	if opts == nil {
		opts = &ACMEPluginOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/acme/plugins/%s", id), opts, nil)
}

// DeleteACMEPlugin removes an ACME challenge plugin.
func (cl *Cluster) DeleteACMEPlugin(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("acme plugin id can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/acme/plugins/%s", id), nil)
}
