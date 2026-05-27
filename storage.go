package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// validContent enumerates the values Proxmox's
// /nodes/{node}/storage/{storage}/upload endpoint accepts. The Proxmox API
// rejects everything else (including "snippets" — those have to be placed
// on the storage path directly, e.g. via SCP/SFTP, since there is no REST
// upload path for them as of PVE 9.x).
var validContent = map[string]struct{}{
	"iso":    {},
	"vztmpl": {},
	"import": {},
}

func (c *Client) ClusterStorages(ctx context.Context) (storages ClusterStorages, err error) {
	err = c.Get(ctx, "/storage", &storages)
	if err != nil {
		return
	}

	for _, s := range storages {
		s.client = c
	}
	return
}

func (c *Client) ClusterStorage(ctx context.Context, name string) (storage *ClusterStorage, err error) {
	err = c.Get(ctx, fmt.Sprintf("/storage/%s", name), &storage)
	if err != nil {
		return
	}

	storage.client = c
	return
}

func (c *Client) DeleteClusterStorage(ctx context.Context, name string) (*Task, error) {
	var upid UPID
	err := c.Delete(ctx, fmt.Sprintf("/storage/%s", name), &upid)
	if err != nil {
		return nil, err
	}
	return NewTask(upid, c), nil
}

func (c *Client) NewClusterStorage(ctx context.Context, options ...ClusterStorageOptions) (*Task, error) {
	var upid UPID

	data := make(map[string]interface{})
	for _, option := range options {
		data[option.Name] = option.Value
	}
	err := c.Post(ctx, "/storage", data, &upid)

	if err != nil {
		return nil, err
	}
	return NewTask(upid, c), nil
}

func (c *Client) UpdateClusterStorage(ctx context.Context, name string, options ...ClusterStorageOptions) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	for _, option := range options {
		data[option.Name] = option.Value
	}
	err := c.Put(ctx, fmt.Sprintf("/storage/%s", name), data, &upid)
	if err != nil {
		return nil, err
	}
	return NewTask(upid, c), nil
}

func (s *Storage) Upload(content, file string) (*Task, error) {
	return s.upload(content, file, nil)
}

func (s *Storage) UploadWithName(content, file string, storageFilename string) (*Task, error) {
	return s.upload(content, file, &map[string]string{"filename": storageFilename})
}

func (s *Storage) UploadWithHash(content, file string, storageFilename *string, checksum, checksumAlgorithm string) (*Task, error) {
	extraArgs := map[string]string{
		"checksum":           checksum,
		"checksum-algorithm": checksumAlgorithm,
	}
	if storageFilename != nil {
		extraArgs["filename"] = *storageFilename
	}
	return s.upload(content, file, &extraArgs)
}

func (s *Storage) upload(content, file string, extraArgs *map[string]string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, validContentError()
	}

	stat, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("file is a directory %s", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// The upload's filename goes in the file part's Content-Disposition.
	// It must NOT also be sent as a form field — Proxmox treats both as
	// the same parameter and rejects the request when they collide.
	filename := filepath.Base(file)
	data := map[string]string{"content": content}
	if extraArgs != nil {
		for k, v := range *extraArgs {
			if k == "filename" {
				filename = v
				continue
			}
			data[k] = v
		}
	}

	var upid UPID
	if err := s.client.UploadReader(
		fmt.Sprintf("/nodes/%s/storage/%s/upload", s.Node, s.Name),
		data, filename, f, stat.Size(), &upid,
	); err != nil {
		return nil, err
	}

	return NewTask(upid, s.client), nil
}

// UploadString uploads contents directly as a file with the given storage
// filename without writing to a temporary file. Useful when the payload is
// already in memory. content must be one of the values accepted by the
// Proxmox upload endpoint (iso, vztmpl, import).
func (s *Storage) UploadString(content, storageFilename, contents string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, validContentError()
	}

	body := strings.NewReader(contents)
	// storageFilename is communicated via the file part's Content-Disposition
	// (UploadReader's filename arg) — it must not also be a form field.
	data := map[string]string{"content": content}

	var upid UPID
	if err := s.client.UploadReader(
		fmt.Sprintf("/nodes/%s/storage/%s/upload", s.Node, s.Name),
		data, storageFilename, body, int64(body.Len()), &upid,
	); err != nil {
		return nil, err
	}

	return NewTask(upid, s.client), nil
}

func validContentError() error {
	keys := make([]string, 0, len(validContent))
	for k := range validContent {
		keys = append(keys, k)
	}
	return fmt.Errorf("invalid content type, allowed: %s", strings.Join(keys, ", "))
}

func (s *Storage) DownloadURL(ctx context.Context, content, filename, url string) (*Task, error) {
	return s.downloadURL(ctx, content, filename, url, nil)
}

func (s *Storage) DownloadURLWithHash(ctx context.Context, content, filename, url string, checksum, checksumAlgorithm string) (*Task, error) {
	return s.downloadURL(ctx, content, filename, url, &map[string]string{
		"checksum":           checksum,
		"checksum-algorithm": checksumAlgorithm,
	})
}

func (s *Storage) downloadURL(ctx context.Context, content, filename, url string, extraArgs *map[string]string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, validContentError()
	}

	var upid UPID
	data := map[string]string{
		"content":  content,
		"filename": filename,
		"url":      url,
	}

	if extraArgs != nil {
		for k, v := range *extraArgs {
			data[k] = v
		}
	}
	err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/download-url", s.Node, s.Name), data, &upid)
	if err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

func (s *Storage) GetContent(ctx context.Context) (content []*StorageContent, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", s.Node, s.Name), &content)
	return content, err
}

func (s *Storage) DeleteContent(ctx context.Context, content string) (*Task, error) {
	var upid UPID
	err := s.client.Delete(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s", s.Node, s.Name, content), &upid)
	if err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

func (s *Storage) ISO(ctx context.Context, name string) (iso *ISO, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "iso", name), &iso)
	if err != nil {
		return nil, err
	}

	iso.client = s.client
	iso.Node = s.Node
	iso.Storage = s.Name
	if iso.VolID == "" {
		iso.VolID = fmt.Sprintf("%s:iso/%s", iso.Storage, name)
	}
	return
}

func (s *Storage) VzTmpl(ctx context.Context, name string) (vztmpl *VzTmpl, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "vztmpl", name), &vztmpl)
	if err != nil {
		return nil, err
	}

	vztmpl.client = s.client
	vztmpl.Node = s.Node
	vztmpl.Storage = s.Name
	if vztmpl.VolID == "" {
		vztmpl.VolID = fmt.Sprintf("%s:vztmpl/%s", vztmpl.Storage, name)
	}
	return
}

func (s *Storage) Import(ctx context.Context, name string) (imp *Import, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "import", name), &imp)
	if err != nil {
		return nil, err
	}

	imp.client = s.client
	imp.Node = s.Node
	imp.Storage = s.Name
	if imp.VolID == "" {
		imp.VolID = fmt.Sprintf("%s:import/%s", imp.Storage, name)
	}
	return
}

func (s *Storage) Backup(ctx context.Context, name string) (backup *Backup, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "backup", name), &backup)
	if err != nil {
		return nil, err
	}

	backup.client = s.client
	backup.Node = s.Node
	backup.Storage = s.Name
	return
}

func (v *VzTmpl) Delete(ctx context.Context) (*Task, error) {
	return deleteVolume(ctx, v.client, v.Node, v.Storage, v.VolID, v.Path, "vztmpl")
}

func (b *Backup) Delete(ctx context.Context) (*Task, error) {
	return deleteVolume(ctx, b.client, b.Node, b.Storage, b.VolID, b.Path, "backup")
}

func (i *ISO) Delete(ctx context.Context) (*Task, error) {
	return deleteVolume(ctx, i.client, i.Node, i.Storage, i.VolID, i.Path, "iso")
}

// PreviewPruneBackups returns the list of backup volumes the prune call would
// keep, remove, retain by protection flag, or skip due to non-standard
// naming. Pass nil opts to use the storage's configured retention spec across
// every guest. This is a dryrun — nothing on disk changes.
func (s *Storage) PreviewPruneBackups(ctx context.Context, opts *StoragePruneBackupsOptions) ([]*PruneBackupItem, error) {
	p := fmt.Sprintf("/nodes/%s/storage/%s/prunebackups", s.Node, s.Name)
	if q := opts.queryString(); q != "" {
		p = p + "?" + q
	}
	var items []*PruneBackupItem
	err := s.client.Get(ctx, p, &items)
	return items, err
}

// PruneBackups deletes the backup volumes a PreviewPruneBackups call with the
// same opts would mark "remove". Returns the task so callers can Wait on it.
// Note that backups added/removed between preview and prune may shift which
// volumes get deleted; the preview is informational, not a transaction.
func (s *Storage) PruneBackups(ctx context.Context, opts *StoragePruneBackupsOptions) (*Task, error) {
	p := fmt.Sprintf("/nodes/%s/storage/%s/prunebackups", s.Node, s.Name)
	if q := opts.queryString(); q != "" {
		p = p + "?" + q
	}
	var upid UPID
	if err := s.client.Delete(ctx, p, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

func (o *StoragePruneBackupsOptions) queryString() string {
	if o == nil {
		return ""
	}
	v := url.Values{}
	if o.PruneBackups != "" {
		v.Set("prune-backups", o.PruneBackups)
	}
	if o.Type != "" {
		v.Set("type", o.Type)
	}
	if o.VMID != 0 {
		v.Set("vmid", strconv.FormatUint(o.VMID, 10))
	}
	return v.Encode()
}

// ImportMetadata fetches the metadata Proxmox extracts from an importable
// guest volume — currently ESXi-sourced VM disks on an "import"-capable
// storage. Use this as a pre-flight before constructing a VM with the
// "import-from=" disk option to see the disks/network mapping PVE detected
// and any warnings about unsupported fields.
//
// volume is the standard PVE volume identifier, e.g.
// "esxi-store:ha-datacenter/MyVM/MyVM.vmx".
func (s *Storage) ImportMetadata(ctx context.Context, volume string) (*ImportMetadata, error) {
	p := fmt.Sprintf("/nodes/%s/storage/%s/import-metadata?%s",
		s.Node, s.Name, url.Values{"volume": {volume}}.Encode())
	var meta ImportMetadata
	if err := s.client.Get(ctx, p, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// Identity returns the storage's stable id + plugin type. PBS storages
// surface a content-addressed datastore id here for namespace tracking;
// other plugins typically return the storage name as the id.
func (s *Storage) Identity(ctx context.Context) (id *StorageIdentity, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/identity", s.Node, s.Name), &id)
	return
}

// RRD asks PVE to render a storage-utilization PNG and returns its on-disk
// filename. ds is a comma-separated list of datasources ("total,used");
// timeframe is hour/day/week/month/year (no decade for storage).
func (s *Storage) RRD(ctx context.Context, ds string, timeframe Timeframe, cf ConsolidationFunction) (rrd *NodeRRDImage, err error) {
	if ds == "" {
		return nil, fmt.Errorf("ds is required")
	}
	q := url.Values{}
	q.Set("ds", ds)
	q.Set("timeframe", string(timeframe))
	if cf != "" {
		q.Set("cf", string(cf))
	}
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/rrd?%s", s.Node, s.Name, q.Encode()), &rrd)
	return
}

func deleteVolume(ctx context.Context, c *Client, n, s, v, p, t string) (*Task, error) {
	var upid UPID
	if v == "" && p == "" {
		return nil, fmt.Errorf("volid or path required for a delete")
	}

	if v == "" {
		// volid not returned in the volume endpoints, need to generate
		v = fmt.Sprintf("%s:%s/%s", s, t, filepath.Base(p))
	}

	err := c.Delete(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s", n, s, v), &upid)
	return NewTask(upid, c), err
}
