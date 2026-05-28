package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Subnets lists the subnets configured under this VNet.
//
// GET /cluster/sdn/vnets/{vnet}/subnets
func (v *VNet) Subnets(ctx context.Context) (subnets []*VNetSubnet, err error) {
	if v.Name == "" {
		return nil, errors.New("vnet name is required")
	}
	if err = v.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", v.Name), &subnets); err != nil {
		return nil, err
	}
	for _, s := range subnets {
		s.client = v.client
		s.VNet = v.Name
	}
	return
}

// Subnet returns a handle for a single subnet under this VNet. No API call is
// made; use the returned handle's Read to populate it.
//
// GET /cluster/sdn/vnets/{vnet}/subnets/{subnet}
func (v *VNet) Subnet(name string) *VNetSubnet {
	return &VNetSubnet{client: v.client, VNet: v.Name, ID: name}
}

// NewSubnet creates a new subnet under this VNet. opts.Subnet (CIDR) is
// required; opts.Type defaults to "subnet" if empty.
//
// POST /cluster/sdn/vnets/{vnet}/subnets
func (v *VNet) NewSubnet(ctx context.Context, opts *SDNSubnetOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil || opts.Subnet == "" {
		return errors.New("subnet cidr is required")
	}
	if opts.Type == "" {
		opts.Type = "subnet"
	}
	if opts.VNet == "" {
		opts.VNet = v.Name
	}
	return v.client.Post(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", v.Name), opts, nil)
}

// Read populates the receiver with the subnet configuration.
//
// GET /cluster/sdn/vnets/{vnet}/subnets/{subnet}
func (s *VNetSubnet) Read(ctx context.Context) error {
	if s.VNet == "" || s.ID == "" {
		return errors.New("vnet and subnet id are required")
	}
	return s.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", s.VNet, s.ID), s)
}

// Update mutates the subnet configuration.
//
// PUT /cluster/sdn/vnets/{vnet}/subnets/{subnet}
func (s *VNetSubnet) Update(ctx context.Context, opts *SDNSubnetOptions) error {
	if s.VNet == "" || s.ID == "" {
		return errors.New("vnet and subnet id are required")
	}
	if opts == nil {
		opts = &SDNSubnetOptions{}
	}
	return s.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", s.VNet, s.ID), opts, nil)
}

// Delete removes the subnet.
//
// DELETE /cluster/sdn/vnets/{vnet}/subnets/{subnet}
func (s *VNetSubnet) Delete(ctx context.Context) error {
	if s.VNet == "" || s.ID == "" {
		return errors.New("vnet and subnet id are required")
	}
	return s.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", s.VNet, s.ID), nil)
}
