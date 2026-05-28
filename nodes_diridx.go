package proxmox

import (
	"context"
	"fmt"
)

// Directory-index ("diridx") wrappers for /nodes/{node}/* endpoints. PVE
// publishes these endpoints as the canonical permission/capability probes:
// they're ACL-filtered, so the returned list tells the caller which
// subresources they can read, and they exist on every supported PVE version,
// giving a stable surface for capability discovery across upgrades. The web
// UI itself uses these to decide what menu items to show.
//
// Each helper here issues a GET against the directory-index path and
// collapses the `[{"subdir":"..."}]` link-objects into a flat []string. The
// helpers are named `<area>Diridx` so the endpoints scanner (see
// mage/endpoints) treats the call site as a GET against the helper's path
// argument.

// Subdirs enumerates the children of /nodes/{node} (typically "qemu",
// "lxc", "storage", "network", "tasks", "scan", "services", "subscription",
// etc.). Useful as a permission probe: PVE filters the list by the caller's
// ACLs, so the result tells you which sub-APIs the credential can read.
func (n *Node) Subdirs(ctx context.Context) ([]string, error) {
	return n.nodeDiridx(ctx, fmt.Sprintf("/nodes/%s", n.Name))
}

// FirewallSubdirs enumerates the children of /nodes/{node}/firewall
// ("rules", "options", "log").
func (n *Node) FirewallSubdirs(ctx context.Context) ([]string, error) {
	return n.firewallDiridx(ctx, fmt.Sprintf("/nodes/%s/firewall", n.Name))
}

// DisksSubdirs enumerates the children of /nodes/{node}/disks ("list",
// "smart", "initgpt", "wipedisk", "directory", "lvm", "lvmthin", "zfs").
func (n *Node) DisksSubdirs(ctx context.Context) ([]string, error) {
	return n.disksDiridx(ctx, fmt.Sprintf("/nodes/%s/disks", n.Name))
}

// Subdirs enumerates the children of /nodes/{node}/replication/{id}
// ("status", "log", "schedule_now"). /nodes/{node}/replication (without
// {id}) is a true list endpoint, not a diridx; use (*Node).Replications for
// that.
func (r *NodeReplicationJob) Subdirs(ctx context.Context) ([]string, error) {
	if r.ID == "" {
		return nil, fmt.Errorf("replication id is required")
	}
	return r.replicationDiridx(ctx, fmt.Sprintf("/nodes/%s/replication/%s", r.Node, r.ID))
}

// Subdirs enumerates the children of /nodes/{node}/services/{service}
// ("state", "start", "stop", "restart", "reload"). /nodes/{node}/services
// (without {service}) is a true list endpoint — use (*Node).Services.
func (s *NodeService) Subdirs(ctx context.Context) ([]string, error) {
	if s.Name == "" {
		return nil, fmt.Errorf("service name is required")
	}
	return s.serviceDiridx(ctx, fmt.Sprintf("/nodes/%s/services/%s", s.Node, s.Name))
}

// Subdirs enumerates the children of /nodes/{node}/tasks/{upid}
// ("log", "status"). Use as a permission probe before calling Log/Ping.
func (t *Task) Subdirs(ctx context.Context) ([]string, error) {
	if t.UPID == "" {
		return nil, fmt.Errorf("task upid is required")
	}
	return t.taskDiridx(ctx, fmt.Sprintf("/nodes/%s/tasks/%s", t.Node, t.UPID))
}

// StorageStatus is the response of GET /nodes/{node}/storage/{storage}.
// Despite its path being a diridx, PVE returns the storage's capability /
// status object directly: type + plugin metadata (Active, Enabled, Shared),
// declared Content types, and (when active) capacity counters. Numeric
// fields are int/uint64 to match what the JSON envelope unmarshals into.
type StorageStatus struct {
	// Type is the storage plugin ("dir", "lvm", "lvmthin", "zfs", "nfs",
	// "cifs", "pbs", "rbd", "cephfs", …).
	Type string `json:"type,omitempty"`
	// Content is the comma-joined list of content types the storage holds
	// ("images,rootdir,iso,vztmpl,backup,snippets,import").
	Content string `json:"content,omitempty"`
	// Active is 1 when the storage is currently mounted/reachable on this
	// node, 0 otherwise (e.g. unreachable NFS, disabled storage). Capacity
	// fields are only populated when Active=1.
	Active int `json:"active,omitempty"`
	// Enabled is 1 when the storage is administratively enabled on this
	// node (it may still be Active=0 if the underlying transport is down).
	Enabled int `json:"enabled,omitempty"`
	// Shared is 1 when PVE treats this storage as shared across the
	// cluster (NFS, CIFS, Ceph, …) and 0 for node-local plugins.
	Shared int `json:"shared,omitempty"`
	// Total is the storage's total capacity in bytes; 0 when inactive.
	Total uint64 `json:"total,omitempty"`
	// Used is bytes currently used; 0 when inactive.
	Used uint64 `json:"used,omitempty"`
	// Avail is bytes available; 0 when inactive.
	Avail uint64 `json:"avail,omitempty"`
	// UsedFraction is Used/Total as a float in [0,1]; 0 when inactive.
	UsedFraction float64 `json:"used_fraction,omitempty"`
}

// Status returns the storage's status / capability object from
// GET /nodes/{node}/storage/{storage}. Despite living at the directory-index
// path, PVE publishes the storage's status payload here (Active, Content
// types, Type, Enabled, capacity counters), not the usual `[{"subdir":...}]`
// envelope — see StorageStatus.
func (s *Storage) Status(ctx context.Context) (status *StorageStatus, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s", s.Node, s.Name), &status)
	return
}

// --- diridx helpers ---------------------------------------------------------
//
// Each helper does the same thing: GET path, decode [{"subdir":"x"}, ...] to
// []string. They're per-receiver so the scanner's
// `<recv>.<area>Diridx(ctx, path)` regex picks up every call site and treats
// it as a GET against the path argument.

func (n *Node) nodeDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, n.client, path)
}

func (n *Node) firewallDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, n.client, path)
}

func (n *Node) disksDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, n.client, path)
}

func (r *NodeReplicationJob) replicationDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, r.client, path)
}

func (s *NodeService) serviceDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, s.client, path)
}

func (t *Task) taskDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, t.client, path)
}

// decodeSubdirList is the underlying [{"subdir":...}] decoder. Centralised
// so the per-receiver `<area>Diridx` helpers stay trivial one-liners.
func decodeSubdirList(ctx context.Context, c *Client, path string) ([]string, error) {
	var items []struct {
		Subdir string `json:"subdir"`
	}
	if err := c.Get(ctx, path, &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Subdir)
	}
	return out, nil
}
