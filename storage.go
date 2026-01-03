package proxmox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

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

	if storageFilename != nil {
		return s.upload(content, file, &map[string]string{"filename": *storageFilename})
	}

	return s.upload(content, file, nil)
}

func (s *Storage) upload(content, file string, extraArgs *map[string]string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, fmt.Errorf("only iso, vztmpl and import allowed")
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

	var upid UPID
	data := map[string]string{"content": content}
	if extraArgs != nil {
		for k, v := range *extraArgs {
			data[k] = v
		}
	}

	if err := s.client.Upload(fmt.Sprintf("/nodes/%s/storage/%s/upload", s.Node, s.Name),
		data, f, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, s.client), nil
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
		return nil, fmt.Errorf("only iso, vztmpl and import allowed")
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
