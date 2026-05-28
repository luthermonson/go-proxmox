package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// SDNControllers lists configured SDN controllers. typ filters by plugin type
// (e.g. "bgp", "evpn"); pass "" for all.
//
// GET /cluster/sdn/controllers
func (cl *Cluster) SDNControllers(ctx context.Context, typ string) (controllers []*SDNController, err error) {
	path := "/cluster/sdn/controllers"
	if typ != "" {
		q := url.Values{}
		q.Set("type", typ)
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &controllers); err != nil {
		return nil, err
	}
	for _, c := range controllers {
		c.client = cl.client
	}
	return
}

// SDNController returns a handle for a single SDN controller. No API call is
// made; use the returned handle's Read to populate it.
//
// GET /cluster/sdn/controllers/{controller}
func (cl *Cluster) SDNController(name string) *SDNController {
	return &SDNController{client: cl.client, Controller: name}
}

// NewSDNController creates a new SDN controller object. opts.Controller and
// opts.Type are required.
//
// POST /cluster/sdn/controllers
func (cl *Cluster) NewSDNController(ctx context.Context, opts *SDNControllerOptions) error {
	if opts == nil || opts.Controller == "" {
		return errors.New("sdn controller name is required")
	}
	if opts.Type == "" {
		return errors.New("sdn controller type is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/controllers", opts, nil)
}

// Read populates the receiver with the current configuration of the controller.
//
// GET /cluster/sdn/controllers/{controller}
func (c *SDNController) Read(ctx context.Context) error {
	if c.Controller == "" {
		return errors.New("sdn controller name is required")
	}
	return c.client.Get(ctx, fmt.Sprintf("/cluster/sdn/controllers/%s", c.Controller), c)
}

// Update mutates an existing controller. opts.Delete may be a comma-separated
// list of keys to reset to PVE defaults.
//
// PUT /cluster/sdn/controllers/{controller}
func (c *SDNController) Update(ctx context.Context, opts *SDNControllerOptions) error {
	if c.Controller == "" {
		return errors.New("sdn controller name is required")
	}
	if opts == nil {
		opts = &SDNControllerOptions{}
	}
	return c.client.Put(ctx, fmt.Sprintf("/cluster/sdn/controllers/%s", c.Controller), opts, nil)
}

// Delete removes the SDN controller.
//
// DELETE /cluster/sdn/controllers/{controller}
func (c *SDNController) Delete(ctx context.Context) error {
	if c.Controller == "" {
		return errors.New("sdn controller name is required")
	}
	return c.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/controllers/%s", c.Controller), nil)
}
