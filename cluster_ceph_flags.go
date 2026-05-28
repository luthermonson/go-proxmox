package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Wrappers for /cluster/ceph/flags + /cluster/ceph/metadata — cluster-wide
// Ceph flag inspection / mutation and a per-service version snapshot.

// CephFlags returns the catalog of Ceph OSD-map flags with their current
// enabled state.
//
// GET /cluster/ceph/flags
func (cl *Cluster) CephFlags(ctx context.Context) (flags []*CephFlag, err error) {
	err = cl.client.Get(ctx, "/cluster/ceph/flags", &flags)
	return
}

// CephFlag returns the current state of a single flag. Useful as a quick poll
// helper without parsing the full catalog.
//
// GET /cluster/ceph/flags/{flag}
func (cl *Cluster) CephFlag(ctx context.Context, flag string) (value string, err error) {
	if flag == "" {
		err = errors.New("ceph flag: name is required")
		return
	}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/ceph/flags/%s", flag), &value)
	return
}

// CephMetadata returns the Ceph services version snapshot — versions and
// device-id mapping per OSD, MON, MGR, MDS. PVE's response is a wide
// version-dependent shape; we surface it as a typed root with the common
// service buckets and let callers reach into the per-service maps directly.
//
// GET /cluster/ceph/metadata
func (cl *Cluster) CephMetadata(ctx context.Context) (meta *CephMetadata, err error) {
	meta = &CephMetadata{}
	err = cl.client.Get(ctx, "/cluster/ceph/metadata", meta)
	return
}

// SetCephFlags toggles multiple ceph flags atomically and returns a *Task for
// the worker that applies them. Each pointer field in opts: true sets, false
// unsets, nil leaves the flag alone.
//
// PUT /cluster/ceph/flags
func (cl *Cluster) SetCephFlags(ctx context.Context, opts *CephFlagsUpdateOptions) (*Task, error) {
	if opts == nil {
		opts = &CephFlagsUpdateOptions{}
	}
	var upid UPID
	if err := cl.client.Put(ctx, "/cluster/ceph/flags", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// SetCephFlag sets or clears a single ceph flag synchronously. The PVE schema
// describes this as a sync wrapper around the bulk endpoint; no UPID is
// returned.
//
// PUT /cluster/ceph/flags/{flag}
func (cl *Cluster) SetCephFlag(ctx context.Context, flag string, value bool) error {
	if flag == "" {
		return errors.New("ceph flag: name is required")
	}
	body := map[string]bool{"value": value}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/ceph/flags/%s", flag), body, nil)
}
