package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Ceph daemon-registry endpoints under /nodes/{node}/ceph/{mon,mgr,mds}.
// All three daemon types share the same CRUD shape: list / create-by-id /
// delete-by-id. Lists + creates stay on *Node; per-daemon getters return a
// *CephMon / *CephMgr / *CephMDS handle whose .Delete() removes the daemon.

// --- MON (Monitor) ---------------------------------------------------------

// CephMons lists the Ceph monitor daemons PVE knows about on this node.
// Includes both running monitors visible to the cluster and configured
// daemons that are currently stopped/unknown. Each returned entry has
// `client` + `Node` populated so callers can chain `.Delete()` directly.
func (n *Node) CephMons(ctx context.Context) (mons []*CephMon, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mon", n.Name), &mons); err != nil {
		return nil, err
	}
	for _, m := range mons {
		m.client = n.client
		m.Node = n.Name
	}
	return mons, nil
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

// CephMon returns an operations handle for the monitor with the given monid.
// No API call is made.
func (n *Node) CephMon(monid string) *CephMon {
	return &CephMon{client: n.client, Node: n.Name, Name: monid}
}

// Delete destroys the monitor. PVE refuses to remove the last monitor of the
// cluster. Does not touch any manager running on the same node — use *CephMgr.Delete
// for that.
func (m *CephMon) Delete(ctx context.Context) (*Task, error) {
	if m.Name == "" {
		return nil, errors.New("monid is required")
	}
	var upid UPID
	if err := m.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mon/%s", m.Node, m.Name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, m.client), nil
}

// --- MGR (Manager) ---------------------------------------------------------

// CephMgrs lists the Ceph manager daemons PVE knows about on this node.
// Each returned entry has `client` + `Node` populated for chaining.
func (n *Node) CephMgrs(ctx context.Context) (mgrs []*CephMgr, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mgr", n.Name), &mgrs); err != nil {
		return nil, err
	}
	for _, m := range mgrs {
		m.client = n.client
		m.Node = n.Name
	}
	return mgrs, nil
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

// CephMgr returns an operations handle for the manager with the given id.
// No API call is made.
func (n *Node) CephMgr(id string) *CephMgr {
	return &CephMgr{client: n.client, Node: n.Name, Name: id}
}

// Delete destroys the manager.
func (m *CephMgr) Delete(ctx context.Context) (*Task, error) {
	if m.Name == "" {
		return nil, errors.New("mgr id is required")
	}
	var upid UPID
	if err := m.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mgr/%s", m.Node, m.Name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, m.client), nil
}

// --- MDS (Metadata Server) -------------------------------------------------

// CephMDSs lists the CephFS metadata-server daemons PVE knows about on this
// node. Includes both active/standby MDSes visible to the cluster and
// configured daemons that are currently stopped/unknown. Each returned entry
// has `client` + `Node` populated for chaining.
func (n *Node) CephMDSs(ctx context.Context) (mdss []*CephMDS, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/mds", n.Name), &mdss); err != nil {
		return nil, err
	}
	for _, m := range mdss {
		m.client = n.client
		m.Node = n.Name
	}
	return mdss, nil
}

// CreateCephMDS creates a CephFS metadata server on this node with the given
// name. Pass an empty name to default to the nodename. Set opts.HotStandby
// to spin the daemon up as a standby-replay MDS that polls and replays the
// log of an active MDS for faster failover.
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

// CephMDS returns an operations handle for the metadata server with the given
// name. No API call is made.
func (n *Node) CephMDS(name string) *CephMDS {
	return &CephMDS{client: n.client, Node: n.Name, Name: name}
}

// Delete destroys the metadata server.
func (m *CephMDS) Delete(ctx context.Context) (*Task, error) {
	if m.Name == "" {
		return nil, errors.New("mds name is required")
	}
	var upid UPID
	if err := m.client.Delete(ctx, fmt.Sprintf("/nodes/%s/ceph/mds/%s", m.Node, m.Name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, m.client), nil
}
