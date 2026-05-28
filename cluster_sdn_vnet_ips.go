package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// CreateIP creates a MAC/IP mapping in the IPAM for this VNet. opts.Zone and
// opts.IP are required.
//
// POST /cluster/sdn/vnets/{vnet}/ips
func (v *VNet) CreateIP(ctx context.Context, opts *SDNVNetIPOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil || opts.IP == "" {
		return errors.New("vnet ip is required")
	}
	if opts.Zone == "" {
		return errors.New("vnet zone is required")
	}
	return v.client.Post(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/ips", v.Name), opts, nil)
}

// UpdateIP updates a MAC/IP mapping in the IPAM (e.g. to associate a VMID).
//
// PUT /cluster/sdn/vnets/{vnet}/ips
func (v *VNet) UpdateIP(ctx context.Context, opts *SDNVNetIPOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil || opts.IP == "" {
		return errors.New("vnet ip is required")
	}
	if opts.Zone == "" {
		return errors.New("vnet zone is required")
	}
	return v.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/ips", v.Name), opts, nil)
}

// DeleteIP removes an IP mapping from the IPAM for this VNet.
//
// DELETE /cluster/sdn/vnets/{vnet}/ips
func (v *VNet) DeleteIP(ctx context.Context, opts *SDNVNetIPOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil || opts.IP == "" {
		return errors.New("vnet ip is required")
	}
	if opts.Zone == "" {
		return errors.New("vnet zone is required")
	}
	// The DELETE endpoint accepts query parameters rather than a body in this
	// client. Encode the identifying fields onto the URL.
	path := fmt.Sprintf("/cluster/sdn/vnets/%s/ips?zone=%s&ip=%s", v.Name, opts.Zone, opts.IP)
	if opts.MAC != "" {
		path += "&mac=" + opts.MAC
	}
	return v.client.Delete(ctx, path, nil)
}
