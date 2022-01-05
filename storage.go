package proxmox

import (
	"fmt"
	"os"
	"path/filepath"
)

var validContent = map[string]struct{}{
	"iso":    struct{}{},
	"vztmpl": struct{}{},
}

func (s *Storage) Upload(content, file string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, fmt.Errorf("only iso and vztmpl allowed")
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
	defer f.Close()

	var upid UPID
	if err := s.client.Upload(fmt.Sprintf("/nodes/%s/storage/%s/upload", s.Node, s.Name),
		map[string]string{"content": content}, f, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, s.client), nil
}

func (s *Storage) DownloadURL(content, filename, url string) (*Task, error) {
	if _, ok := validContent[content]; !ok {
		return nil, fmt.Errorf("only iso and vztmpl allowed")
	}

	var upid UPID
	s.client.Post(fmt.Sprintf("/nodes/%s/storage/%s/download-url", s.Node, s.Name), map[string]string{
		"content":  content,
		"filename": filename,
		"url":      url,
	}, &upid)
	return NewTask(upid, s.client), nil
}

func (s *Storage) ISO(name string) (iso *ISO, err error) {
	err = s.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "iso", name), &iso)
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

func (s *Storage) VzTmpl(name string) (vztmpl *VzTmpl, err error) {
	err = s.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "vztmpl", name), &vztmpl)
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

func (s *Storage) Backup(name string) (backup *Backup, err error) {
	err = s.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "backup", name), &backup)
	if err != nil {
		return nil, err
	}

	backup.client = s.client
	backup.Node = s.Node
	backup.Storage = s.Name
	return
}

func (v *VzTmpl) Delete() (*Task, error) {
	return deleteVolume(v.client, v.Node, v.Storage, v.VolID, v.Path, "vztmpl")
}

func (b *Backup) Delete() (*Task, error) {
	return deleteVolume(b.client, b.Node, b.Storage, b.VolID, b.Path, "backup")
}

func (i *ISO) Delete() (*Task, error) {
	return deleteVolume(i.client, i.Node, i.Storage, i.VolID, i.Path, "iso")
}

func deleteVolume(c *Client, n, s, v, p, t string) (*Task, error) {
	var upid UPID
	if v == "" && p == "" {
		return nil, fmt.Errorf("volid or path required for a delete")
	}

	if v == "" {
		// volid not returned in the volume endpoints, need to generate
		v = fmt.Sprintf("%s:%s/%s", s, t, filepath.Base(p))
	}

	err := c.Delete(fmt.Sprintf("/nodes/%s/storage/%s/content/%s?delay=5", n, s, v), &upid)
	return NewTask(upid, c), err
}
