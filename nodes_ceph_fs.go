package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// /nodes/{node}/ceph/fs/* — CephFS filesystem wrappers.
//
// List + create stay on *Node. The per-fs getter returns a *CephFS handle
// whose .Delete() removes the filesystem.

// CephFSs lists the CephFS filesystems known to the cluster, as seen by this
// node. Each entry carries the metadata pool and one or more data pools, with
// `client` + `Node` populated for chaining.
func (n *Node) CephFSs(ctx context.Context) (fss []*CephFS, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/fs", n.Name), &fss); err != nil {
		return nil, err
	}
	for _, f := range fss {
		f.client = n.client
		f.Node = n.Name
	}
	return fss, nil
}

// CreateCephFS creates a new CephFS filesystem with the given name. opts is
// optional; PVE defaults PgNum to 128 and AddStorage to false. The endpoint
// runs as a worker task — returns the Task so the caller can wait on it.
func (n *Node) CreateCephFS(ctx context.Context, name string, opts *CephFSOptions) (*Task, error) {
	if name == "" {
		return nil, errors.New("cephfs name is required")
	}
	if opts == nil {
		opts = &CephFSOptions{}
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/fs/%s", n.Name, name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// CephFS returns an operations handle for the CephFS filesystem with the given
// name. No API call is made.
func (n *Node) CephFS(name string) *CephFS {
	return &CephFS{client: n.client, Node: n.Name, Name: name}
}

// Delete destroys the CephFS filesystem. removePools also drops the backing
// metadata + data pools; removeStorages also removes the pveceph-managed
// storage.cfg entries. PVE refuses the call if any non-disabled cephfs storage
// entry still references the filesystem.
func (f *CephFS) Delete(ctx context.Context, removePools, removeStorages bool) (*Task, error) {
	if f.Name == "" {
		return nil, errors.New("cephfs name is required")
	}
	q := url.Values{}
	if removePools {
		q.Set("remove-pools", "1")
	}
	if removeStorages {
		q.Set("remove-storages", "1")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/fs/%s", f.Node, f.Name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := f.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, f.client), nil
}
