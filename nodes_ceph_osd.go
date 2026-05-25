package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Methods on *Node that wrap the /nodes/{node}/ceph/osd/* family — Ceph OSD
// (Object Storage Daemon) lifecycle (create/destroy), state toggles
// (in/out/scrub), and introspection (list/index/metadata/lv-info).
//
// createosd and destroyosd return UPIDs (wrapped as *Task); in/out/scrub
// return null and are fire-and-forget; the GETs return data structures.

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

// CephOSD is the per-OSD index page (GET /nodes/{node}/ceph/osd/{osdid}).
// PVE returns a free-form list of child resource descriptors; surfaced as
// raw maps because the schema declares no concrete fields.
func (n *Node) CephOSD(ctx context.Context, osdID int) (items []map[string]interface{}, err error) {
	return items, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d", n.Name, osdID), &items)
}

// DeleteCephOSD destroys an OSD. With cleanup=true PVE also zaps the
// underlying logical volumes (via `ceph-volume lvm zap --destroy`), removes
// the VG's PV, and wipes any leftover journal/block.db/block.wal partitions.
// Returns a Task.
func (n *Node) DeleteCephOSD(ctx context.Context, osdID int, cleanup bool) (*Task, error) {
	q := url.Values{}
	if cleanup {
		q.Set("cleanup", "1")
	}
	path := fmt.Sprintf("/nodes/%s/ceph/osd/%d", n.Name, osdID)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// CephOSDIn marks an OSD as `in` (eligible to hold data). Returns no value
// per the PVE schema — succeed = nil err.
func (n *Node) CephOSDIn(ctx context.Context, osdID int) error {
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/in", n.Name, osdID), nil, nil)
}

// CephOSDOut marks an OSD as `out` (data will migrate off it).
func (n *Node) CephOSDOut(ctx context.Context, osdID int) error {
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/out", n.Name, osdID), nil, nil)
}

// CephOSDScrub instructs an OSD to scrub. Pass deep=true for a deep scrub
// (checksum-verifies every object) instead of the default metadata-only scrub.
func (n *Node) CephOSDScrub(ctx context.Context, osdID int, deep bool) error {
	body := map[string]interface{}{}
	if deep {
		body["deep"] = 1
	}
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/scrub", n.Name, osdID), body, nil)
}

// CephOSDLVInfo returns LVM details for one of the OSD's logical volumes.
// devType is "block" (default), "db", or "wal" — pass "" for block.
func (n *Node) CephOSDLVInfo(ctx context.Context, osdID int, devType string) (info *CephOSDLVInfo, err error) {
	path := fmt.Sprintf("/nodes/%s/ceph/osd/%d/lv-info", n.Name, osdID)
	if devType != "" {
		q := url.Values{}
		q.Set("type", devType)
		path = path + "?" + q.Encode()
	}
	return info, n.client.Get(ctx, path, &info)
}

// CephOSDMetadata returns daemon-level info plus the list of backing devices
// for one OSD.
func (n *Node) CephOSDMetadata(ctx context.Context, osdID int) (details *CephOSDDetails, err error) {
	return details, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/osd/%d/metadata", n.Name, osdID), &details)
}
