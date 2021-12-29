package proxmox

import (
	"fmt"
	"path/filepath"
)

func (s *Storage) Upload(content, filename string) error {
	if content != "iso" || content != "vztmpl" {
		return fmt.Errorf("only iso and vztmpl allowed")
	}

	s.client.Post()

	return nil
}

func (s *Storage) ISO(name string) (iso *ISO, err error) {
	err = s.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s/%s", s.Node, s.Name, s.Name, "iso", name), &iso)
	if err != nil {
		return nil, err
	}

	iso.client = s.client
	iso.Node = s.Node
	iso.Storage = s.Name
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

func (v *VzTmpl) Delete() error {
	return deleteVolume(v.client, v.Node, v.Storage, v.VolID, v.Path, "vztmpl")
}

func (b *Backup) Delete() error {
	return deleteVolume(b.client, b.Node, b.Storage, b.VolID, b.Path, "backup")
}

func (i *ISO) Delete() error {
	return deleteVolume(i.client, i.Node, i.Storage, i.VolID, i.Path, "iso")
}

func deleteVolume(c *Client, n, s, v, p, t string) error {
	var res string
	if v == "" && p == "" {
		return fmt.Errorf("volid or path required for a delete")
	}

	if v == "" {
		// volid not returned in the volume endpoints, need to generate
		v = fmt.Sprintf("%s:%s/%s", s, t, filepath.Base(p))
	}

	return c.Delete(fmt.Sprintf("/nodes/%s/storage/%s/content/%s?delay=5", n, s, v), &res)
}
