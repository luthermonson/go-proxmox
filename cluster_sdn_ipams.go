package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// SDNIPAMs lists configured IPAM backends. typ filters by plugin type
// ("netbox", "phpipam", "pve"); pass "" for all.
//
// GET /cluster/sdn/ipams
func (cl *Cluster) SDNIPAMs(ctx context.Context, typ string) (ipams []*SDNIPAM, err error) {
	path := "/cluster/sdn/ipams"
	if typ != "" {
		q := url.Values{}
		q.Set("type", typ)
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &ipams); err != nil {
		return nil, err
	}
	for _, i := range ipams {
		i.client = cl.client
	}
	return
}

// SDNIPAM returns a handle for a single IPAM backend. No API call is made.
//
// GET /cluster/sdn/ipams/{ipam}
func (cl *Cluster) SDNIPAM(name string) *SDNIPAM {
	return &SDNIPAM{client: cl.client, IPAM: name}
}

// NewSDNIPAM creates a new IPAM backend. opts.IPAM and opts.Type are required.
//
// POST /cluster/sdn/ipams
func (cl *Cluster) NewSDNIPAM(ctx context.Context, opts *SDNIPAMOptions) error {
	if opts == nil || opts.IPAM == "" {
		return errors.New("sdn ipam name is required")
	}
	if opts.Type == "" {
		return errors.New("sdn ipam type is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/ipams", opts, nil)
}

// Read populates the receiver with the current configuration.
//
// GET /cluster/sdn/ipams/{ipam}
func (i *SDNIPAM) Read(ctx context.Context) error {
	if i.IPAM == "" {
		return errors.New("sdn ipam name is required")
	}
	return i.client.Get(ctx, fmt.Sprintf("/cluster/sdn/ipams/%s", i.IPAM), i)
}

// Update mutates an existing IPAM backend configuration.
//
// PUT /cluster/sdn/ipams/{ipam}
func (i *SDNIPAM) Update(ctx context.Context, opts *SDNIPAMOptions) error {
	if i.IPAM == "" {
		return errors.New("sdn ipam name is required")
	}
	if opts == nil {
		opts = &SDNIPAMOptions{}
	}
	return i.client.Put(ctx, fmt.Sprintf("/cluster/sdn/ipams/%s", i.IPAM), opts, nil)
}

// Delete removes the IPAM backend.
//
// DELETE /cluster/sdn/ipams/{ipam}
func (i *SDNIPAM) Delete(ctx context.Context) error {
	if i.IPAM == "" {
		return errors.New("sdn ipam name is required")
	}
	return i.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/ipams/%s", i.IPAM), nil)
}

// Status returns the list of IP entries from this IPAM. The shape of each
// entry is plugin-specific so it's returned as a free-form slice of maps; use
// the typed IPAM struct (cluster/sdn types) for known fields.
//
// GET /cluster/sdn/ipams/{ipam}/status
func (i *SDNIPAM) Status(ctx context.Context) (entries []map[string]any, err error) {
	if i.IPAM == "" {
		return nil, errors.New("sdn ipam name is required")
	}
	err = i.client.Get(ctx, fmt.Sprintf("/cluster/sdn/ipams/%s/status", i.IPAM), &entries)
	return
}
