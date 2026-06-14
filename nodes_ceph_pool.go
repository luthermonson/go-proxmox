package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// /nodes/{node}/ceph/pool/* — Ceph pool wrappers.
//
// List + create stay on *Node. The per-pool getter returns a *CephPool handle;
// read / update / delete / status are methods on the handle.

// CephPools lists every Ceph pool visible to the node with the same settings
// exposed by the per-pool PUT endpoint plus a few read-only stats. Each entry
// is returned with `client` + `Node` populated so callers can chain methods.
func (n *Node) CephPools(ctx context.Context) (pools []*CephPool, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/pool", n.Name), &pools); err != nil {
		return nil, err
	}
	for _, p := range pools {
		p.client = n.client
		p.Node = n.Name
	}
	return pools, nil
}

// CreateCephPool creates a Ceph pool. opts.Name is required. For erasure-coded
// pools pass opts.ErasureCoding (k/m required, the rest optional); PVE will
// additionally create a replicated metadata pool. Returns a Task.
func (n *Node) CreateCephPool(ctx context.Context, opts *CephPoolOptions) (*Task, error) {
	if opts == nil || opts.Name == "" {
		return nil, errors.New("ceph pool name is required")
	}
	body, err := cephPoolBody(opts)
	if err != nil {
		return nil, err
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/pool", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// CephPool returns an operations handle for the pool with the given name. No
// API call is made — the handle holds Node + name and dispatches when methods
// are invoked.
func (n *Node) CephPool(name string) *CephPool {
	return &CephPool{client: n.client, Node: n.Name, PoolName: name}
}

// cephPoolBody round-trips opts through JSON so the json tags on
// CephPoolOptions drive (de)serialization, then splices in the EC config in
// the form PVE expects (a single comma-separated key=value string under
// "erasure-coding"). Returning a map lets callers further mutate the body
// (e.g. drop "name" on PUT) without reflecting back into the struct.
func cephPoolBody(opts *CephPoolOptions) (map[string]any, error) {
	raw, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}
	body := map[string]any{}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	if opts.ErasureCoding != nil {
		body["erasure-coding"] = opts.ErasureCoding.String()
	}
	return body, nil
}

// SubResources returns the per-pool directory index (currently just "status").
// To fetch the actual pool configuration / utilization use Status.
func (p *CephPool) SubResources(ctx context.Context) (subdirs []*CephPoolSubdir, err error) {
	if p.PoolName == "" {
		return nil, errors.New("ceph pool name is required")
	}
	return subdirs, p.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/pool/%s", p.Node, p.PoolName), &subdirs)
}

// Update changes pool settings. Name is taken from the handle (URL path);
// any Name set on opts is ignored. Returns a Task.
func (p *CephPool) Update(ctx context.Context, opts *CephPoolOptions) (*Task, error) {
	if p.PoolName == "" {
		return nil, errors.New("ceph pool name is required")
	}
	if opts == nil {
		return nil, errors.New("ceph pool options are required")
	}
	// PUT body must not include Name (the URL path supplies it). Re-marshal
	// through a map so we can drop it explicitly without mutating the caller's
	// struct, and so EC config (which has no JSON tag) is still honored.
	body, err := cephPoolBody(opts)
	if err != nil {
		return nil, err
	}
	delete(body, "name")
	var upid UPID
	if err := p.client.Put(ctx, fmt.Sprintf("/nodes/%s/ceph/pool/%s", p.Node, p.PoolName), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, p.client), nil
}

// Delete destroys the pool. Pass force to destroy a pool even if it is in use.
// removeStorages also strips any pveceph-managed storage.cfg entries pointing
// at the pool. removeECProfile drops the EC profile when applicable (PVE
// defaults this to true server-side; pass false to keep it).
func (p *CephPool) Delete(ctx context.Context, force, removeStorages, removeECProfile bool) (*Task, error) {
	if p.PoolName == "" {
		return nil, errors.New("ceph pool name is required")
	}
	q := url.Values{}
	if force {
		q.Set("force", "1")
	}
	if removeStorages {
		q.Set("remove_storages", "1")
	}
	// PVE defaults remove_ecprofile to 1; only pass it when caller wants false
	// to override that default.
	if !removeECProfile {
		q.Set("remove_ecprofile", "0")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/pool/%s", p.Node, p.PoolName)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := p.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, p.client), nil
}

// Status returns the current configuration and (optionally) statistics for the
// pool. Set verbose=true to include usage and IO statistics in the Statistics
// field.
func (p *CephPool) Status(ctx context.Context, verbose bool) (status *CephPoolStatus, err error) {
	if p.PoolName == "" {
		return nil, errors.New("ceph pool name is required")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/pool/%s/status", p.Node, p.PoolName)
	if verbose {
		path += "?verbose=1"
	}
	return status, p.client.Get(ctx, path, &status)
}
