package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// /nodes/{node}/ceph/osd/* — Ceph OSD (Object Storage Daemon) wrappers.
//
// List + create stay on *Node. The per-OSD getter returns a *CephOSD handle;
// state toggles (in/out/scrub), destroy, and per-OSD introspection (lv-info,
// metadata) are methods on the handle.

// CephOSDs returns the cluster CRUSH tree plus any cluster-wide OSD flags
// from GET /nodes/{node}/ceph/osd. The CRUSH bucket hierarchy is recursive
// and per-bucket properties are not statically typed.
func (n *Node) CephOSDs(ctx context.Context) (tree *CephOSDTree, err error) {
	return tree, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/osd", n.Name), &tree)
}

// CreateCephOSD provisions a new OSD on a block device. opts.Dev is required.
// Returns a Task — PVE runs the ceph-volume / mkfs work via the task queue.
func (n *Node) CreateCephOSD(ctx context.Context, opts *CephOSDCreateOptions) (*Task, error) {
	if opts == nil || opts.Dev == "" {
		return nil, errors.New("ceph osd dev is required")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// CephOSD returns an operations handle for the OSD with the given id. No API
// call is made — the handle holds Node + id and dispatches when methods are
// invoked.
func (n *Node) CephOSD(osdID int) *CephOSD {
	return &CephOSD{client: n.client, Node: n.Name, ID: osdID}
}

// SubResources returns the per-OSD index page (GET /nodes/{node}/ceph/osd/{id}).
// PVE returns a free-form list of child resource descriptors; surfaced as raw
// maps because the schema declares no concrete fields.
func (o *CephOSD) SubResources(ctx context.Context) (items []map[string]interface{}, err error) {
	return items, o.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d", o.Node, o.ID), &items)
}

// Delete destroys the OSD. With cleanup=true PVE also zaps the underlying
// logical volumes (via `ceph-volume lvm zap --destroy`), removes the VG's PV,
// and wipes any leftover journal/block.db/block.wal partitions. Returns a Task.
func (o *CephOSD) Delete(ctx context.Context, cleanup bool) (*Task, error) {
	q := url.Values{}
	if cleanup {
		q.Set("cleanup", "1")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/osd/%d", o.Node, o.ID)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := o.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, o.client), nil
}

// In marks the OSD as `in` (eligible to hold data). Returns no value per the
// PVE schema — success = nil err.
func (o *CephOSD) In(ctx context.Context) error {
	return o.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/in", o.Node, o.ID), nil, nil)
}

// Out marks the OSD as `out` (data will migrate off it).
func (o *CephOSD) Out(ctx context.Context) error {
	return o.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/out", o.Node, o.ID), nil, nil)
}

// Scrub instructs the OSD to scrub. Pass deep=true for a deep scrub
// (checksum-verifies every object) instead of the default metadata-only scrub.
func (o *CephOSD) Scrub(ctx context.Context, deep bool) error {
	body := map[string]interface{}{}
	if deep {
		body["deep"] = 1
	}
	return o.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/scrub", o.Node, o.ID), body, nil)
}

// LVInfo returns LVM details for one of the OSD's logical volumes. devType is
// "block" (default), "db", or "wal" — pass "" for block.
func (o *CephOSD) LVInfo(ctx context.Context, devType string) (info *CephOSDLVInfo, err error) {
	path := fmt.Sprintf("/nodes/%s/ceph/osd/%d/lv-info", o.Node, o.ID)
	if devType != "" {
		q := url.Values{}
		q.Set("type", devType)
		path = path + "?" + q.Encode()
	}
	return info, o.client.Get(ctx, path, &info)
}

// Metadata returns daemon-level info plus the list of backing devices.
func (o *CephOSD) Metadata(ctx context.Context) (details *CephOSDDetails, err error) {
	return details, o.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/metadata", o.Node, o.ID), &details)
}
