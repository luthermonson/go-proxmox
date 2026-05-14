package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Mappings lists the resource-type directory under /cluster/mapping (e.g. dir,
// pci, usb). See https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/mapping
func (cl *Cluster) Mappings(ctx context.Context) (entries ClusterMappings, err error) {
	err = cl.client.Get(ctx, "/cluster/mapping", &entries)
	return
}

// --- directory mappings -----------------------------------------------------

// DirMappings lists directory mappings. Pass a non-empty checkNode to ask PVE
// to validate each entry against that node (populates each entry's Checks).
func (cl *Cluster) DirMappings(ctx context.Context, checkNode string) (mappings ClusterDirMappings, err error) {
	path := "/cluster/mapping/dir"
	if checkNode != "" {
		q := url.Values{}
		q.Set("check-node", checkNode)
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &mappings)
	return
}

// DirMapping reads a single directory mapping by id.
func (cl *Cluster) DirMapping(ctx context.Context, id string) (m *ClusterDirMapping, err error) {
	if id == "" {
		err = errors.New("dir mapping id can not be empty")
		return
	}
	m = &ClusterDirMapping{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/mapping/dir/%s", id), m); err != nil {
		return
	}
	if m.ID == "" {
		m.ID = id
	}
	return
}

// NewDirMapping creates a directory mapping. opts.ID and opts.Map are required.
func (cl *Cluster) NewDirMapping(ctx context.Context, opts *ClusterDirMappingOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("dir mapping id can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/mapping/dir", opts, nil)
}

// UpdateDirMapping mutates an existing directory mapping.
func (cl *Cluster) UpdateDirMapping(ctx context.Context, id string, opts *ClusterDirMappingOptions) error {
	if id == "" {
		return errors.New("dir mapping id can not be empty")
	}
	if opts == nil {
		opts = &ClusterDirMappingOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/mapping/dir/%s", id), opts, nil)
}

// DeleteDirMapping removes a directory mapping.
func (cl *Cluster) DeleteDirMapping(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("dir mapping id can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/mapping/dir/%s", id), nil)
}

// --- PCI mappings -----------------------------------------------------------

// PCIMappings lists PCI hardware mappings. See DirMappings for checkNode.
func (cl *Cluster) PCIMappings(ctx context.Context, checkNode string) (mappings ClusterPCIMappings, err error) {
	path := "/cluster/mapping/pci"
	if checkNode != "" {
		q := url.Values{}
		q.Set("check-node", checkNode)
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &mappings)
	return
}

// PCIMapping reads a single PCI mapping by id.
func (cl *Cluster) PCIMapping(ctx context.Context, id string) (m *ClusterPCIMapping, err error) {
	if id == "" {
		err = errors.New("pci mapping id can not be empty")
		return
	}
	m = &ClusterPCIMapping{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/mapping/pci/%s", id), m); err != nil {
		return
	}
	if m.ID == "" {
		m.ID = id
	}
	return
}

// NewPCIMapping creates a PCI hardware mapping.
func (cl *Cluster) NewPCIMapping(ctx context.Context, opts *ClusterPCIMappingOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("pci mapping id can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/mapping/pci", opts, nil)
}

// UpdatePCIMapping mutates an existing PCI mapping.
func (cl *Cluster) UpdatePCIMapping(ctx context.Context, id string, opts *ClusterPCIMappingOptions) error {
	if id == "" {
		return errors.New("pci mapping id can not be empty")
	}
	if opts == nil {
		opts = &ClusterPCIMappingOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/mapping/pci/%s", id), opts, nil)
}

// DeletePCIMapping removes a PCI mapping.
func (cl *Cluster) DeletePCIMapping(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("pci mapping id can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/mapping/pci/%s", id), nil)
}

// --- USB mappings -----------------------------------------------------------

// USBMappings lists USB hardware mappings. See DirMappings for checkNode.
func (cl *Cluster) USBMappings(ctx context.Context, checkNode string) (mappings ClusterUSBMappings, err error) {
	path := "/cluster/mapping/usb"
	if checkNode != "" {
		q := url.Values{}
		q.Set("check-node", checkNode)
		path = path + "?" + q.Encode()
	}
	err = cl.client.Get(ctx, path, &mappings)
	return
}

// USBMapping reads a single USB mapping by id.
func (cl *Cluster) USBMapping(ctx context.Context, id string) (m *ClusterUSBMapping, err error) {
	if id == "" {
		err = errors.New("usb mapping id can not be empty")
		return
	}
	m = &ClusterUSBMapping{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/mapping/usb/%s", id), m); err != nil {
		return
	}
	if m.ID == "" {
		m.ID = id
	}
	return
}

// NewUSBMapping creates a USB hardware mapping.
func (cl *Cluster) NewUSBMapping(ctx context.Context, opts *ClusterUSBMappingOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("usb mapping id can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/mapping/usb", opts, nil)
}

// UpdateUSBMapping mutates an existing USB mapping.
func (cl *Cluster) UpdateUSBMapping(ctx context.Context, id string, opts *ClusterUSBMappingOptions) error {
	if id == "" {
		return errors.New("usb mapping id can not be empty")
	}
	if opts == nil {
		opts = &ClusterUSBMappingOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/mapping/usb/%s", id), opts, nil)
}

// DeleteUSBMapping removes a USB mapping.
func (cl *Cluster) DeleteUSBMapping(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("usb mapping id can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/mapping/usb/%s", id), nil)
}
