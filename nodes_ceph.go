package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// Ceph wraps /nodes/{node}/ceph/* — the per-node entry point to the cluster's
// Ceph services. Everything Ceph is cluster-wide under the hood; the per-node
// prefix mostly exists so PVE can run the underlying ceph CLI on a node that
// actually has the cluster's keyring + admin socket. Methods here cover the
// root directory index, the systemd-style service lifecycle (init / start /
// stop / restart), and the read-only observability surface (status, log,
// crush map, CRUSH rules, cmd-safety).

// --- root index ------------------------------------------------------------

// CephIndex returns the /nodes/{node}/ceph directory index — a flat list of
// child handles (osd, mon, mgr, pool, fs, status, log, …). Mostly a probe;
// the actual resources are wrapped by sibling methods on *Node.
func (n *Node) CephIndex(ctx context.Context) (entries []*CephIndexEntry, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph", n.Name), &entries)
	return
}

// --- service lifecycle -----------------------------------------------------

// InitCeph performs the one-time Ceph bootstrap on the node — writes
// /etc/ceph/ceph.conf with the cluster fsid, default pool sizing, and
// auth/network settings. Idempotent: re-calling with an existing [global]
// section preserves the original values and silently ignores most params.
// Pass a nil *opts to accept all PVE defaults.
func (n *Node) InitCeph(ctx context.Context, opts *CephInitOptions) (*Task, error) {
	body := map[string]any{}
	if opts != nil {
		if opts.Network != "" {
			body["network"] = opts.Network
		}
		if opts.ClusterNetwork != "" {
			body["cluster-network"] = opts.ClusterNetwork
		}
		if opts.Size != 0 {
			body["size"] = opts.Size
		}
		if opts.MinSize != 0 {
			body["min_size"] = opts.MinSize
		}
		if opts.PGBits != 0 {
			body["pg_bits"] = opts.PGBits
		}
		if opts.DisableCephx {
			body["disable_cephx"] = 1
		}
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/init", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// StartCeph starts Ceph services on the node. service is optional (defaults
// to "ceph.target", i.e. all roles); pass e.g. "osd.3" or "mon" to target a
// single daemon or role.
func (n *Node) StartCeph(ctx context.Context, service string) (*Task, error) {
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/start", n.Name), cephServiceBody(service), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// StopCeph stops Ceph services on the node. See StartCeph for service syntax.
func (n *Node) StopCeph(ctx context.Context, service string) (*Task, error) {
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/stop", n.Name), cephServiceBody(service), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// RestartCeph restarts Ceph services on the node. See StartCeph for service
// syntax.
func (n *Node) RestartCeph(ctx context.Context, service string) (*Task, error) {
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/restart", n.Name), cephServiceBody(service), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// cephServiceBody builds the shared request body for the start/stop/restart
// trio — they all accept the same optional `service` form field.
func cephServiceBody(service string) map[string]any {
	body := map[string]any{}
	if service != "" {
		body["service"] = service
	}
	return body
}

// --- observability ---------------------------------------------------------

// CephStatus returns the raw `ceph status` output (mon/mgr/osd/pg maps,
// health checks, quorum). Cluster-wide — identical payload to
// Cluster.Ceph().Status; this node-level alias just runs the ceph CLI on the
// requested node.
func (n *Node) CephStatus(ctx context.Context) (status *ClusterCephStatus, err error) {
	status = &ClusterCephStatus{}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/status", n.Name), status)
	return
}

// CephLog reads the ceph cluster log. start is the 0-based offset of the
// first line to return; pass 0 for the head of the log. limit caps the
// number of lines (0 = PVE's default, typically 50).
func (n *Node) CephLog(ctx context.Context, start, limit int) (entries []*CephLogEntry, err error) {
	q := url.Values{}
	if start > 0 {
		q.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := fmt.Sprintf("/nodes/%s/ceph/log", n.Name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	err = n.client.Get(ctx, path, &entries)
	return
}

// CephCrush returns the OSD CRUSH map as a textual dump (the format produced
// by `ceph osd crush dump`). PVE returns it as a single string blob.
func (n *Node) CephCrush(ctx context.Context) (crush string, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/crush", n.Name), &crush)
	return
}

// CephRules lists the configured CRUSH rules (one entry per rule, name only —
// the body of each rule lives in the CRUSH map dumped by CephCrush).
func (n *Node) CephRules(ctx context.Context) (rules []*CephRule, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/rules", n.Name), &rules)
	return
}

// CephCmdSafety asks Ceph whether a planned mutation is safe right now —
// e.g. stopping an OSD without losing data redundancy. service must be one
// of "osd"|"mon"|"mds", action must be "stop"|"destroy", id is the
// service-specific identifier (numeric for osd, name for mon/mds).
func (n *Node) CephCmdSafety(ctx context.Context, service, id, action string) (safety *CephCmdSafety, err error) {
	if service == "" || id == "" || action == "" {
		return nil, errors.New("ceph cmd-safety requires service, id, and action")
	}
	q := url.Values{}
	q.Set("service", service)
	q.Set("id", id)
	q.Set("action", action)
	safety = &CephCmdSafety{}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/cmd-safety?%s", n.Name, q.Encode()), safety)
	return
}
