package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Disks wraps the /nodes/{node}/disks/* family — physical disk discovery,
// SMART, GPT init/wipe, and CRUD for directory / LVM / LVM-thin / ZFS
// storages backed by raw block devices. All write ops return a Task since
// PVE runs them via the task queue (mkfs, zpool create, etc.).

// --- discovery / SMART / GPT / wipe ----------------------------------------

// Disks lists every block device PVE sees on the node. Pass includePartitions
// to also surface partition entries, skipSMART to avoid the per-disk SMART
// probe, and diskType ("unused"|"journal_disks"|"") to filter.
func (n *Node) Disks(ctx context.Context, includePartitions, skipSMART bool, diskType string) (disks []*Disk, err error) {
	q := url.Values{}
	if includePartitions {
		q.Set("include-partitions", "1")
	}
	if skipSMART {
		q.Set("skipsmart", "1")
	}
	if diskType != "" {
		q.Set("type", diskType)
	}
	path := fmt.Sprintf("/nodes/%s/disks/list", n.Name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	return disks, n.client.Get(ctx, path, &disks)
}

// DiskSMART returns SMART data for one disk. Pass healthOnly=true for just
// the "PASS"/"FAIL" string (faster — skips the full attribute dump).
func (n *Node) DiskSMART(ctx context.Context, disk string, healthOnly bool) (smart *DiskSMART, err error) {
	if disk == "" {
		return nil, errors.New("disk path is required")
	}
	q := url.Values{}
	q.Set("disk", disk)
	if healthOnly {
		q.Set("healthonly", "1")
	}
	return smart, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/smart?%s", n.Name, q.Encode()), &smart)
}

// DiskInitGPT writes a fresh GPT partition table to disk. uuid is optional;
// pass "" to let the kernel assign one. Returns a Task.
func (n *Node) DiskInitGPT(ctx context.Context, disk, uuid string) (*Task, error) {
	if disk == "" {
		return nil, errors.New("disk path is required")
	}
	body := map[string]any{"disk": disk}
	if uuid != "" {
		body["uuid"] = uuid
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/disks/initgpt", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DiskWipe zeroes the first and last megabytes of disk and clears the
// partition table — destructive. Returns a Task.
func (n *Node) DiskWipe(ctx context.Context, disk string) (*Task, error) {
	if disk == "" {
		return nil, errors.New("disk path is required")
	}
	body := map[string]any{"disk": disk}
	var upid UPID
	if err := n.client.Put(ctx, fmt.Sprintf("/nodes/%s/disks/wipedisk", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- directory-mount storages ----------------------------------------------

// Directories lists per-node directory-mount storages backed by raw devices.
func (n *Node) Directories(ctx context.Context) (dirs []*NodeDirectory, err error) {
	return dirs, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/directory", n.Name), &dirs)
}

// NewDirectory formats a block device and mounts it as a per-node directory
// storage. opts.Name and opts.Device are required. opts.Filesystem defaults
// to ext4 server-side.
func (n *Node) NewDirectory(ctx context.Context, opts *NodeDirectoryOptions) (*Task, error) {
	if opts == nil || opts.Name == "" || opts.Device == "" {
		return nil, errors.New("directory name and device are required")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/disks/directory", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteDirectory tears down a directory-mount storage. cleanupConfig also
// removes the storage.cfg entry; cleanupDisks wipes the underlying disk so
// it can be repurposed.
func (n *Node) DeleteDirectory(ctx context.Context, name string, cleanupConfig, cleanupDisks bool) (*Task, error) {
	if name == "" {
		return nil, errors.New("directory name is required")
	}
	q := url.Values{}
	if cleanupConfig {
		q.Set("cleanup-config", "1")
	}
	if cleanupDisks {
		q.Set("cleanup-disks", "1")
	}
	path := fmt.Sprintf("/nodes/%s/disks/directory/%s", n.Name, name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- LVM volume groups -----------------------------------------------------

// LVMs returns the nested LVM tree (volume groups → physical volumes).
func (n *Node) LVMs(ctx context.Context) (lvm *NodeLVMTree, err error) {
	return lvm, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/lvm", n.Name), &lvm)
}

// NewLVM creates an LVM volume group on a block device. opts.Name and
// opts.Device are required.
func (n *Node) NewLVM(ctx context.Context, opts *NodeLVMOptions) (*Task, error) {
	if opts == nil || opts.Name == "" || opts.Device == "" {
		return nil, errors.New("lvm name and device are required")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/disks/lvm", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteLVM removes an LVM volume group. See DeleteDirectory for cleanup
// semantics.
func (n *Node) DeleteLVM(ctx context.Context, name string, cleanupConfig, cleanupDisks bool) (*Task, error) {
	if name == "" {
		return nil, errors.New("lvm name is required")
	}
	q := url.Values{}
	if cleanupConfig {
		q.Set("cleanup-config", "1")
	}
	if cleanupDisks {
		q.Set("cleanup-disks", "1")
	}
	path := fmt.Sprintf("/nodes/%s/disks/lvm/%s", n.Name, name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- LVM-thin pools --------------------------------------------------------

// LVMThins lists LVM-thin pools.
func (n *Node) LVMThins(ctx context.Context) (thins []*NodeLVMThin, err error) {
	return thins, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/lvmthin", n.Name), &thins)
}

// NewLVMThin creates an LVM-thin pool on a block device.
func (n *Node) NewLVMThin(ctx context.Context, opts *NodeLVMThinOptions) (*Task, error) {
	if opts == nil || opts.Name == "" || opts.Device == "" {
		return nil, errors.New("lvmthin name and device are required")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/disks/lvmthin", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteLVMThin removes an LVM-thin pool. Unlike directory / LVM / ZFS,
// this endpoint *requires* the parent volume-group name to disambiguate
// when a node has multiple VGs containing thin pools by the same name.
func (n *Node) DeleteLVMThin(ctx context.Context, name, volumeGroup string, cleanupConfig, cleanupDisks bool) (*Task, error) {
	if name == "" || volumeGroup == "" {
		return nil, errors.New("lvmthin name and volume-group are required")
	}
	q := url.Values{}
	q.Set("volume-group", volumeGroup)
	if cleanupConfig {
		q.Set("cleanup-config", "1")
	}
	if cleanupDisks {
		q.Set("cleanup-disks", "1")
	}
	path := fmt.Sprintf("/nodes/%s/disks/lvmthin/%s?%s", n.Name, name, q.Encode())
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// --- ZFS pools -------------------------------------------------------------

// ZFSPools lists ZFS pools visible to PVE on the node.
func (n *Node) ZFSPools(ctx context.Context) (pools []*NodeZFSPoolSummary, err error) {
	return pools, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/zfs", n.Name), &pools)
}

// ZFSPool returns the detailed status of a single pool, including the
// underlying vdev/device tree and any failure-action recommendation.
func (n *Node) ZFSPool(ctx context.Context, name string) (pool *NodeZFSPool, err error) {
	if name == "" {
		return nil, errors.New("zfs pool name is required")
	}
	return pool, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/disks/zfs/%s", n.Name, name), &pool)
}

// NewZFSPool creates a ZFS pool. opts.Name, opts.Devices (space-separated),
// and opts.RaidLevel are required.
func (n *Node) NewZFSPool(ctx context.Context, opts *NodeZFSPoolOptions) (*Task, error) {
	if opts == nil || opts.Name == "" || opts.Devices == "" || opts.RaidLevel == "" {
		return nil, errors.New("zfs pool name, devices, and raidlevel are required")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/disks/zfs", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// DeleteZFSPool destroys a ZFS pool. cleanupConfig/cleanupDisks semantics
// match DeleteDirectory.
func (n *Node) DeleteZFSPool(ctx context.Context, name string, cleanupConfig, cleanupDisks bool) (*Task, error) {
	if name == "" {
		return nil, errors.New("zfs pool name is required")
	}
	q := url.Values{}
	if cleanupConfig {
		q.Set("cleanup-config", "1")
	}
	if cleanupDisks {
		q.Set("cleanup-disks", "1")
	}
	path := fmt.Sprintf("/nodes/%s/disks/zfs/%s", n.Name, name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	var upid UPID
	if err := n.client.Delete(ctx, path, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}
