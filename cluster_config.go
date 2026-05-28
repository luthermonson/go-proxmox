package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Wrappers for /cluster/config/* — corosync cluster bootstrap, join,
// node membership, and read-only quorum / totem inspection. The POST and
// DELETE endpoints under this tree are protected and intended to be driven
// by `pvecm` on the affected node; we expose them for completeness so
// callers building tooling on top of the API can drive them.

// JoinAPIVersion returns the version of the cluster join API on this node.
//
// GET /cluster/config/apiversion
func (cl *Cluster) JoinAPIVersion(ctx context.Context) (version int, err error) {
	err = cl.client.Get(ctx, "/cluster/config/apiversion", &version)
	return
}

// JoinInfo returns the parameters a prospective member needs to join this
// cluster: nodelist, totem config, preferred-node, config digest, and per-node
// SSH/cert fingerprints. node is optional — empty defaults to the connected
// node.
//
// GET /cluster/config/join
func (cl *Cluster) JoinInfo(ctx context.Context, node string) (info *ClusterJoinInfo, err error) {
	path := "/cluster/config/join"
	if node != "" {
		path = path + "?node=" + node
	}
	info = &ClusterJoinInfo{}
	err = cl.client.Get(ctx, path, info)
	return
}

// JoinCluster joins this node to an existing cluster. PVE returns a status
// string on success (not a UPID). opts.Hostname, opts.Password, and
// opts.Fingerprint are required.
//
// POST /cluster/config/join
func (cl *Cluster) JoinCluster(ctx context.Context, opts *ClusterJoinOptions) (status string, err error) {
	if opts == nil || opts.Hostname == "" {
		err = errors.New("cluster join: hostname is required")
		return
	}
	if opts.Password == "" {
		err = errors.New("cluster join: password is required")
		return
	}
	if opts.Fingerprint == "" {
		err = errors.New("cluster join: fingerprint is required")
		return
	}
	err = cl.client.Post(ctx, "/cluster/config/join", opts, &status)
	return
}

// ConfigNodes lists the corosync node list (just names — use
// ClusterStatus/NodeStatuses for richer per-node state).
//
// GET /cluster/config/nodes
func (cl *Cluster) ConfigNodes(ctx context.Context) (nodes []*ClusterConfigNodeEntry, err error) {
	err = cl.client.Get(ctx, "/cluster/config/nodes", &nodes)
	return
}

// QDevice returns the QDevice (corosync external arbitrator) status. Returns
// an open-shape map because PVE's response shape depends on whether qdevice
// is configured and which net algorithm is in use.
//
// GET /cluster/config/qdevice
func (cl *Cluster) QDevice(ctx context.Context) (status map[string]any, err error) {
	status = map[string]any{}
	err = cl.client.Get(ctx, "/cluster/config/qdevice", &status)
	return
}

// Totem returns the corosync totem protocol settings (token timeouts, link
// modes, secauth, etc.). Returns an open-shape map because the totem config
// has many version-dependent knobs.
//
// GET /cluster/config/totem
func (cl *Cluster) Totem(ctx context.Context) (totem map[string]any, err error) {
	totem = map[string]any{}
	err = cl.client.Get(ctx, "/cluster/config/totem", &totem)
	return
}

// CreateCluster generates a new cluster configuration on this node. PVE
// returns a status string on success.
//
// POST /cluster/config
func (cl *Cluster) CreateCluster(ctx context.Context, opts *ClusterCreateOptions) (status string, err error) {
	if opts == nil || opts.ClusterName == "" {
		err = errors.New("cluster create: clustername is required")
		return
	}
	err = cl.client.Post(ctx, "/cluster/config", opts, &status)
	return
}

// AddConfigNode adds a node to the cluster configuration. PVE documents this
// call as "for internal use" — it returns corosync.conf bytes + the authkey
// the new node needs, which `pvecm add` normally consumes locally. Exposed
// here for tooling.
//
// POST /cluster/config/nodes/{node}
func (cl *Cluster) AddConfigNode(ctx context.Context, node string, opts *ClusterAddNodeOptions) (result *ClusterAddNodeResult, err error) {
	if node == "" {
		err = errors.New("cluster add node: node is required")
		return
	}
	if opts == nil {
		opts = &ClusterAddNodeOptions{}
	}
	result = &ClusterAddNodeResult{}
	err = cl.client.Post(ctx, fmt.Sprintf("/cluster/config/nodes/%s", node), opts, result)
	return
}

// DeleteConfigNode removes a node from the cluster configuration.
//
// DELETE /cluster/config/nodes/{node}
func (cl *Cluster) DeleteConfigNode(ctx context.Context, node string) error {
	if node == "" {
		return errors.New("cluster delete node: node is required")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/config/nodes/%s", node), nil)
}
