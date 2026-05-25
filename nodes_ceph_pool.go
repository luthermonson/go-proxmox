package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// This file wraps the /nodes/{node}/ceph/pool/* family — Ceph pool list /
// create / read / update / delete plus the pool-status endpoint. Mutating
// endpoints return *Task because PVE runs them via the task queue (the API
// answers with a UPID string).

// CephPools lists every Ceph pool visible to the node, with the same
// settings exposed by the per-pool PUT endpoint plus a few read-only stats.
func (n *Node) CephPools(ctx context.Context) (pools []*CephPool, err error) {
	return pools, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/pool", n.Name), &pools)
}

// CreateCephPool creates a Ceph pool. opts.Name is required. For erasure-
// coded pools pass opts.ErasureCoding (k/m are required, the rest optional);
// PVE will additionally create a replicated metadata pool. Returns a Task.
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

// CephPool returns the sub-resource directory index for a pool (currently
// just "status"). To fetch the actual pool configuration / utilization use
// CephPoolStatus.
func (n *Node) CephPool(ctx context.Context, name string) (subdirs []*CephPoolSubdir, err error) {
	if name == "" {
		return nil, errors.New("ceph pool name is required")
	}
	return subdirs, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/pool/%s", n.Name, name), &subdirs)
}

// UpdateCephPool changes pool settings. Name is taken from the URL path;
// any Name set on opts is ignored. Returns a Task.
func (n *Node) UpdateCephPool(ctx context.Context, name string, opts *CephPoolOptions) (*Task, error) {
	if name == "" {
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
	if err := n.client.Put(ctx, fmt.Sprintf("/nodes/%s/ceph/pool/%s", n.Name, name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteCephPool destroys a pool. Pass force to destroy a pool even if it is
// in use. removeStorages also strips any pveceph-managed storage.cfg entries
// pointing at the pool. removeECProfile drops the EC profile when applicable
// (PVE defaults this to true server-side; pass false to keep it).
func (n *Node) DeleteCephPool(ctx context.Context, name string, force, removeStorages, removeECProfile bool) (*Task, error) {
	if name == "" {
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
	path := fmt.Sprintf("/nodes/%s/ceph/pool/%s", n.Name, name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// CephPoolStatus returns the current configuration and (optionally)
// statistics for a single pool. Set verbose=true to include usage and IO
// statistics in the Statistics field.
func (n *Node) CephPoolStatus(ctx context.Context, name string, verbose bool) (status *CephPoolStatus, err error) {
	if name == "" {
		return nil, errors.New("ceph pool name is required")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/pool/%s/status", n.Name, name)
	if verbose {
		path = path + "?verbose=1"
	}
	return status, n.client.Get(ctx, path, &status)
}
