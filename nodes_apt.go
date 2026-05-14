package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// APT returns the directory index for /nodes/{node}/apt — a small list of
// child handles (changelog, repositories, update, versions). Mostly useful as
// a probe; the real data lives on the children.
func (n *Node) APT(ctx context.Context) (entries []*APTIndexEntry, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/apt", n.Name), &entries)
	return
}

// APTUpdates lists the packages with available upgrades on the node, as
// produced by the last `apt-get update`. Empty list means the index is clean
// or has never been refreshed; call APTUpdate to resync first.
func (n *Node) APTUpdates(ctx context.Context) (updates []*APTUpdate, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/apt/update", n.Name), &updates)
	return
}

// APTUpdate resynchronizes the apt package index (apt-get update) on the
// node. notify=true asks PVE to send its configured "new packages available"
// notification; quiet=true suppresses progress output in the task log.
// Returns the worker task.
func (n *Node) APTUpdate(ctx context.Context, notify, quiet bool) (*Task, error) {
	body := map[string]interface{}{}
	if notify {
		body["notify"] = 1
	}
	if quiet {
		body["quiet"] = 1
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/apt/update", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// APTChangelog returns the changelog text for a single package. version is
// optional — when empty PVE picks the candidate version. The endpoint returns
// a single string, not structured data.
func (n *Node) APTChangelog(ctx context.Context, name, version string) (changelog string, err error) {
	q := url.Values{}
	q.Set("name", name)
	if version != "" {
		q.Set("version", version)
	}
	path := fmt.Sprintf("/nodes/%s/apt/changelog?%s", n.Name, q.Encode())
	err = n.client.Get(ctx, path, &changelog)
	return
}

// APTRepositories returns the parsed contents of /etc/apt/sources.list(.d) on
// the node, plus a digest used for optimistic-concurrency on writes and a
// catalog of standard repositories PVE knows how to add.
func (n *Node) APTRepositories(ctx context.Context) (repos *APTRepositories, err error) {
	repos = &APTRepositories{}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/apt/repositories", n.Name), repos)
	return
}

// APTChangeRepository enables or disables an existing repository entry,
// identified by the containing file path and its index within that file.
// digest is optional; pass the value from APTRepositories to detect concurrent
// edits.
func (n *Node) APTChangeRepository(ctx context.Context, path string, index int, enabled bool, digest string) error {
	body := map[string]interface{}{
		"path":    path,
		"index":   index,
		"enabled": 0,
	}
	if enabled {
		body["enabled"] = 1
	}
	if digest != "" {
		body["digest"] = digest
	}
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/apt/repositories", n.Name), body, nil)
}

// APTAddRepository adds one of PVE's standard repositories (identified by
// handle, e.g. "no-subscription", "enterprise") to the node's apt
// configuration. digest is optional for optimistic concurrency.
func (n *Node) APTAddRepository(ctx context.Context, handle, digest string) error {
	body := map[string]interface{}{"handle": handle}
	if digest != "" {
		body["digest"] = digest
	}
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/apt/repositories", n.Name), body, nil)
}

// APTVersions returns the installed versions of the Proxmox-relevant packages
// (pve-manager, kernel, qemu-server, etc.) — the same data shown on the GUI
// "Updates → Package Versions" panel.
func (n *Node) APTVersions(ctx context.Context) (versions []*APTPackageVersion, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/apt/versions", n.Name), &versions)
	return
}
