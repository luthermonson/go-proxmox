package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// SDNDNSList lists configured SDN DNS plugins. typ filters by plugin type
// (currently only "powerdns" is supported by PVE); pass "" for all.
//
// GET /cluster/sdn/dns
func (cl *Cluster) SDNDNSList(ctx context.Context, typ string) (dns []*SDNDNS, err error) {
	path := "/cluster/sdn/dns"
	if typ != "" {
		q := url.Values{}
		q.Set("type", typ)
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &dns); err != nil {
		return nil, err
	}
	for _, d := range dns {
		d.client = cl.client
	}
	return
}

// SDNDNS returns a handle for a single SDN DNS plugin. No API call is made.
//
// GET /cluster/sdn/dns/{dns}
func (cl *Cluster) SDNDNS(name string) *SDNDNS {
	return &SDNDNS{client: cl.client, DNS: name}
}

// NewSDNDNS creates a new SDN DNS plugin. opts.DNS, opts.Type, opts.URL and
// opts.Key are required.
//
// POST /cluster/sdn/dns
func (cl *Cluster) NewSDNDNS(ctx context.Context, opts *SDNDNSOptions) error {
	if opts == nil || opts.DNS == "" {
		return errors.New("sdn dns name is required")
	}
	if opts.Type == "" {
		return errors.New("sdn dns type is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/dns", opts, nil)
}

// Read populates the receiver with the current configuration.
//
// GET /cluster/sdn/dns/{dns}
func (d *SDNDNS) Read(ctx context.Context) error {
	if d.DNS == "" {
		return errors.New("sdn dns name is required")
	}
	return d.client.Get(ctx, fmt.Sprintf("/cluster/sdn/dns/%s", d.DNS), d)
}

// Update mutates an existing SDN DNS plugin configuration.
//
// PUT /cluster/sdn/dns/{dns}
func (d *SDNDNS) Update(ctx context.Context, opts *SDNDNSOptions) error {
	if d.DNS == "" {
		return errors.New("sdn dns name is required")
	}
	if opts == nil {
		opts = &SDNDNSOptions{}
	}
	return d.client.Put(ctx, fmt.Sprintf("/cluster/sdn/dns/%s", d.DNS), opts, nil)
}

// Delete removes the SDN DNS plugin.
//
// DELETE /cluster/sdn/dns/{dns}
func (d *SDNDNS) Delete(ctx context.Context) error {
	if d.DNS == "" {
		return errors.New("sdn dns name is required")
	}
	return d.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/dns/%s", d.DNS), nil)
}
