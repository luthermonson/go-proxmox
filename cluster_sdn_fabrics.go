package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// SDNFabricsIndex returns the directory entries under /cluster/sdn/fabrics
// (currently {"fabric", "node", "all"} subdirs).
//
// GET /cluster/sdn/fabrics
func (cl *Cluster) SDNFabricsIndex(ctx context.Context) (entries []map[string]any, err error) {
	err = cl.client.Get(ctx, "/cluster/sdn/fabrics", &entries)
	return
}

// SDNFabricsAll returns the combined view of all fabrics and their member
// nodes — useful for rendering the whole SDN underlay in one call.
//
// GET /cluster/sdn/fabrics/all
func (cl *Cluster) SDNFabricsAll(ctx context.Context) (all *SDNFabricsAll, err error) {
	all = &SDNFabricsAll{}
	err = cl.client.Get(ctx, "/cluster/sdn/fabrics/all", all)
	if err != nil {
		return nil, err
	}
	for _, f := range all.Fabrics {
		f.client = cl.client
	}
	for _, n := range all.Nodes {
		n.client = cl.client
	}
	return
}

// SDNFabrics lists configured fabrics. pending/running toggle the
// returned configuration (PVE distinguishes pending changes from running config).
//
// GET /cluster/sdn/fabrics/fabric
func (cl *Cluster) SDNFabrics(ctx context.Context, pending, running bool) (fabrics []*SDNFabric, err error) {
	path := "/cluster/sdn/fabrics/fabric"
	q := url.Values{}
	if pending {
		q.Set("pending", "1")
	}
	if running {
		q.Set("running", "1")
	}
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	if err = cl.client.Get(ctx, path, &fabrics); err != nil {
		return nil, err
	}
	for _, f := range fabrics {
		f.client = cl.client
	}
	return
}

// SDNFabric returns a handle for a single fabric. No API call is made.
//
// GET /cluster/sdn/fabrics/fabric/{id}
func (cl *Cluster) SDNFabric(id string) *SDNFabric {
	return &SDNFabric{client: cl.client, ID: id}
}

// NewSDNFabric creates a new fabric. opts.ID and opts.Protocol are required.
//
// POST /cluster/sdn/fabrics/fabric
func (cl *Cluster) NewSDNFabric(ctx context.Context, opts *SDNFabricOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("sdn fabric id is required")
	}
	if opts.Protocol == "" {
		return errors.New("sdn fabric protocol is required")
	}
	return cl.client.Post(ctx, "/cluster/sdn/fabrics/fabric", opts, nil)
}

// Read populates the receiver with the current fabric configuration.
//
// GET /cluster/sdn/fabrics/fabric/{id}
func (f *SDNFabric) Read(ctx context.Context) error {
	if f.ID == "" {
		return errors.New("sdn fabric id is required")
	}
	return f.client.Get(ctx, fmt.Sprintf("/cluster/sdn/fabrics/fabric/%s", f.ID), f)
}

// Update mutates a fabric configuration.
//
// PUT /cluster/sdn/fabrics/fabric/{id}
func (f *SDNFabric) Update(ctx context.Context, opts *SDNFabricOptions) error {
	if f.ID == "" {
		return errors.New("sdn fabric id is required")
	}
	if opts == nil {
		opts = &SDNFabricOptions{}
	}
	return f.client.Put(ctx, fmt.Sprintf("/cluster/sdn/fabrics/fabric/%s", f.ID), opts, nil)
}

// Delete removes the fabric.
//
// DELETE /cluster/sdn/fabrics/fabric/{id}
func (f *SDNFabric) Delete(ctx context.Context) error {
	if f.ID == "" {
		return errors.New("sdn fabric id is required")
	}
	return f.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/fabrics/fabric/%s", f.ID), nil)
}

// Nodes lists nodes participating in this fabric.
//
// GET /cluster/sdn/fabrics/node/{fabric_id}
func (f *SDNFabric) Nodes(ctx context.Context) (nodes []*SDNFabricNode, err error) {
	if f.ID == "" {
		return nil, errors.New("sdn fabric id is required")
	}
	if err = f.client.Get(ctx, fmt.Sprintf("/cluster/sdn/fabrics/node/%s", f.ID), &nodes); err != nil {
		return nil, err
	}
	for _, n := range nodes {
		n.client = f.client
		n.FabricID = f.ID
	}
	return
}

// Node returns a handle for a single node in this fabric. No API call is made.
//
// GET /cluster/sdn/fabrics/node/{fabric_id}/{node_id}
func (f *SDNFabric) Node(nodeID string) *SDNFabricNode {
	return &SDNFabricNode{client: f.client, FabricID: f.ID, NodeID: nodeID}
}

// AddNode adds a node to this fabric. opts.NodeID is required.
//
// POST /cluster/sdn/fabrics/node/{fabric_id}
func (f *SDNFabric) AddNode(ctx context.Context, opts *SDNFabricNodeOptions) error {
	if f.ID == "" {
		return errors.New("sdn fabric id is required")
	}
	if opts == nil || opts.NodeID == "" {
		return errors.New("sdn fabric node id is required")
	}
	return f.client.Post(ctx, fmt.Sprintf("/cluster/sdn/fabrics/node/%s", f.ID), opts, nil)
}

// SDNFabricNodes lists all SDN fabric/node pairs across every fabric — the
// flat alternative to per-fabric iteration via SDNFabric.Nodes.
//
// GET /cluster/sdn/fabrics/node
func (cl *Cluster) SDNFabricNodes(ctx context.Context) (nodes []*SDNFabricNode, err error) {
	if err = cl.client.Get(ctx, "/cluster/sdn/fabrics/node", &nodes); err != nil {
		return nil, err
	}
	for _, n := range nodes {
		n.client = cl.client
	}
	return
}

// Read populates the receiver with the current fabric-node configuration.
//
// GET /cluster/sdn/fabrics/node/{fabric_id}/{node_id}
func (n *SDNFabricNode) Read(ctx context.Context) error {
	if n.FabricID == "" || n.NodeID == "" {
		return errors.New("sdn fabric and node id are required")
	}
	return n.client.Get(ctx, fmt.Sprintf("/cluster/sdn/fabrics/node/%s/%s", n.FabricID, n.NodeID), n)
}

// Update mutates a fabric-node configuration.
//
// PUT /cluster/sdn/fabrics/node/{fabric_id}/{node_id}
func (n *SDNFabricNode) Update(ctx context.Context, opts *SDNFabricNodeOptions) error {
	if n.FabricID == "" || n.NodeID == "" {
		return errors.New("sdn fabric and node id are required")
	}
	if opts == nil {
		opts = &SDNFabricNodeOptions{}
	}
	return n.client.Put(ctx, fmt.Sprintf("/cluster/sdn/fabrics/node/%s/%s", n.FabricID, n.NodeID), opts, nil)
}

// Delete removes the node from the fabric.
//
// DELETE /cluster/sdn/fabrics/node/{fabric_id}/{node_id}
func (n *SDNFabricNode) Delete(ctx context.Context) error {
	if n.FabricID == "" || n.NodeID == "" {
		return errors.New("sdn fabric and node id are required")
	}
	return n.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/fabrics/node/%s/%s", n.FabricID, n.NodeID), nil)
}
