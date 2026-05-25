package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Ceph daemon-registry endpoints under /nodes/{node}/ceph/{mon,mgr,mds}.
// All three daemon types share the same CRUD shape: a GET list, a
// POST-by-id create, and a DELETE-by-id destroy. The create/destroy
// endpoints return a UPID — wrapped with *Task like other long-running
// per-node operations.

// --- MON (Monitor) ---------------------------------------------------------

// CephMons lists the Ceph monitor daemons PVE knows about on this node.
// Includes both running monitors visible to the cluster and configured
// daemons that are currently stopped/unknown.
func (n *Node) CephMons(ctx context.Context) (mons []*CephMonDaemon, err error) {
	return mons, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mon", n.Name), &mons)
}

// CreateCephMon creates a Ceph monitor on this node with the given monid.
// Pass an empty monid to default to the nodename. Also auto-creates a
// manager for the first monitor in the cluster.
func (n *Node) CreateCephMon(ctx context.Context, monid string, opts *CephMonOptions) (*Task, error) {
	if monid == "" {
		monid = n.Name
	}
	body := map[string]any{}
	if opts != nil && opts.MonAddress != "" {
		body["mon-address"] = opts.MonAddress
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/mon/%s", n.Name, monid), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteCephMon destroys a Ceph monitor on this node. PVE refuses to remove
// the last monitor of the cluster. Does not touch any manager running on
// the same node — use DeleteCephMgr for that.
func (n *Node) DeleteCephMon(ctx context.Context, monid string) (*Task, error) {
	if monid == "" {
		return nil, errors.New("monid is required")
	}
	var upid UPID
	if err := n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mon/%s", n.Name, monid), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- MGR (Manager) ---------------------------------------------------------

// CephMgrs lists the Ceph manager daemons PVE knows about on this node.
func (n *Node) CephMgrs(ctx context.Context) (mgrs []*CephMgrDaemon, err error) {
	return mgrs, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mgr", n.Name), &mgrs)
}

// CreateCephMgr creates a Ceph manager on this node with the given id.
// Pass an empty id to default to the nodename.
func (n *Node) CreateCephMgr(ctx context.Context, id string) (*Task, error) {
	if id == "" {
		id = n.Name
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/mgr/%s", n.Name, id), map[string]any{}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteCephMgr destroys a Ceph manager on this node.
func (n *Node) DeleteCephMgr(ctx context.Context, id string) (*Task, error) {
	if id == "" {
		return nil, errors.New("mgr id is required")
	}
	var upid UPID
	if err := n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mgr/%s", n.Name, id), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- MDS (Metadata Server) -------------------------------------------------

// CephMDSs lists the CephFS metadata-server daemons PVE knows about on
// this node. Includes both active/standby MDSes visible to the cluster
// and configured daemons that are currently stopped/unknown.
func (n *Node) CephMDSs(ctx context.Context) (mdss []*CephMDSDaemon, err error) {
	return mdss, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mds", n.Name), &mdss)
}

// CreateCephMDS creates a CephFS metadata server on this node with the
// given name. Pass an empty name to default to the nodename. Set
// opts.HotStandby to spin the daemon up as a standby-replay MDS that
// polls and replays the log of an active MDS for faster failover.
func (n *Node) CreateCephMDS(ctx context.Context, name string, opts *CephMDSOptions) (*Task, error) {
	if name == "" {
		name = n.Name
	}
	body := map[string]any{}
	if opts != nil && bool(opts.HotStandby) {
		body["hotstandby"] = 1
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/mds/%s", n.Name, name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteCephMDS destroys a CephFS metadata server on this node.
func (n *Node) DeleteCephMDS(ctx context.Context, name string) (*Task, error) {
	if name == "" {
		return nil, errors.New("mds name is required")
	}
	var upid UPID
	if err := n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mds/%s", n.Name, name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}
