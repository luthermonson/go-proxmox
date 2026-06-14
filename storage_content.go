package proxmox

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
)

// This file fills the remaining /nodes/{node}/storage/{storage}/* gaps —
// content allocation/update/copy, OCI registry pull, file-restore listing,
// and RRD data. Upload, GetContent, DeleteContent, prunebackups, and
// import-metadata are wrapped in storage.go.

// StorageContentAllocOptions is the body for POST /content — allocate a new
// disk image on the storage. Filename and Size are required; PVE picks
// Format based on the storage type when unset.
type StorageContentAllocOptions struct {
	Filename string `json:"filename"`
	Size     string `json:"size"` // e.g. "1024" (KB) or "4G"
	VMID     uint64 `json:"vmid"`
	Format   string `json:"format,omitempty"`
}

// AllocContent creates a new image volume on the storage. Returns the new
// volid (e.g. "local-lvm:vm-100-disk-1"). Synchronous — most LVM/ZFS-backed
// allocations finish quickly enough that PVE returns the volid directly.
func (s *Storage) AllocContent(ctx context.Context, opts *StorageContentAllocOptions) (volid string, err error) {
	if opts == nil || opts.Filename == "" || opts.Size == "" || opts.VMID == 0 {
		return "", errors.New("filename, size, and vmid are required")
	}
	err = s.client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", s.Node, s.Name), opts, &volid)
	return
}

// StorageContentUpdateOptions is the PUT body for /content/{volume}.
// Protected currently only applies to backup volumes.
type StorageContentUpdateOptions struct {
	Notes     string `json:"notes,omitempty"`
	Protected *bool  `json:"protected,omitempty"`
}

// UpdateContent mutates a volume's metadata (currently: notes + protected
// flag on backups). Pass volume as the full PVE volid (e.g.
// "local:backup/vzdump-qemu-100-2026_01_01-12_00_00.vma.zst").
func (s *Storage) UpdateContent(ctx context.Context, volume string, opts *StorageContentUpdateOptions) error {
	if volume == "" {
		return errors.New("volume is required")
	}
	if opts == nil {
		opts = &StorageContentUpdateOptions{}
	}
	return s.client.Put(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s", s.Node, s.Name, volume), opts, nil)
}

// CopyContent clones a source volume to a target volid, optionally on a
// different node. Returns a Task — copying multi-GB images takes a while.
func (s *Storage) CopyContent(ctx context.Context, sourceVolume, targetVolume, targetNode string) (*Task, error) {
	if sourceVolume == "" || targetVolume == "" {
		return nil, errors.New("source and target volumes are required")
	}
	body := map[string]string{"target": targetVolume}
	if targetNode != "" {
		body["target_node"] = targetNode
	}
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s", s.Node, s.Name, sourceVolume), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// OCIRegistryPull downloads an OCI image from a registry into the storage.
// reference is the OCI image ref (e.g. "docker.io/library/alpine:latest").
// filename is optional — PVE will derive one from the reference when unset.
// Returns a Task.
func (s *Storage) OCIRegistryPull(ctx context.Context, reference, filename string) (*Task, error) {
	if reference == "" {
		return nil, errors.New("oci reference is required")
	}
	body := map[string]string{"reference": reference}
	if filename != "" {
		body["filename"] = filename
	}
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/oci-registry-pull", s.Node, s.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// StorageFileRestoreEntry is one row from GET /file-restore/list.
type StorageFileRestoreEntry struct {
	Filepath string    `json:"filepath,omitempty"`
	Type     string    `json:"type,omitempty"` // "f" (file), "d" (directory), "l" (link)
	Text     string    `json:"text,omitempty"`
	Size     uint64    `json:"size,omitempty"`
	Mtime    int64     `json:"mtime,omitempty"`
	Leaf     IntOrBool `json:"leaf,omitempty"`
}

// FileRestoreList lists entries inside a PBS-backed backup volume at the
// given filesystem path. Pass filepath="/" for the root. PVE only supports
// this on PBS storages.
func (s *Storage) FileRestoreList(ctx context.Context, volume, filepath string) (entries []*StorageFileRestoreEntry, err error) {
	if volume == "" || filepath == "" {
		return nil, errors.New("volume and filepath are required")
	}
	// PVE requires filepath base64-encoded — applies to both directories
	// (which list children) and individual files (which download content).
	q := url.Values{}
	q.Set("volume", volume)
	q.Set("filepath", base64.StdEncoding.EncodeToString([]byte(filepath)))
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/file-restore/list?%s", s.Node, s.Name, q.Encode()), &entries)
	return
}

// RRDData returns the storage's historical IO/usage timeseries. timeframe
// is one of hour/day/week/month/year; cf is the consolidation function
// (AVERAGE | MAX) — empty defaults to AVERAGE server-side.
func (s *Storage) RRDData(ctx context.Context, timeframe Timeframe, cf ConsolidationFunction) (data []*RRDData, err error) {
	q := url.Values{}
	q.Set("timeframe", string(timeframe))
	if cf != "" {
		q.Set("cf", string(cf))
	}
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/rrddata?%s", s.Node, s.Name, q.Encode()), &data)
	return
}
