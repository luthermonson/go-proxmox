package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// CephFSs lists the CephFS filesystems known to the cluster, as seen by this
// node. Each entry carries the metadata pool and one or more data pools.
func (n *Node) CephFSs(ctx context.Context) (fss []*CephFS, err error) {
	return fss, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/fs", n.Name), &fss)
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

// DeleteCephFS destroys a CephFS filesystem. removePools also drops the
// backing metadata + data pools; removeStorages also removes the
// pveceph-managed storage.cfg entries. PVE refuses the call if any
// non-disabled cephfs storage entry still references the filesystem.
func (n *Node) DeleteCephFS(ctx context.Context, name string, removePools, removeStorages bool) (*Task, error) {
	if name == "" {
		return nil, errors.New("cephfs name is required")
	}
	q := url.Values{}
	if removePools {
		q.Set("remove-pools", "1")
	}
	if removeStorages {
		q.Set("remove-storages", "1")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/fs/%s", n.Name, name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}
