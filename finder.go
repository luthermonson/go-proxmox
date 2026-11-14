package proxmox

import (
	"context"
	"fmt"
)

// Resource types as reported in ClusterResource.Type.
const (
	resourceTypeQemu = "qemu"
	resourceTypeLXC  = "lxc"
)

// FindVirtualMachine resolves the node a QEMU VM lives on and returns a
// request-free *VirtualMachine handle already bound to that node, in a single
// GET /cluster/resources. It collapses the usual "query the cluster, then
// build a node ref, then act" chain into one call, for the common case where
// you know the VMID but not the node.
//
// The lookup is keyed on VMID, which is unique cluster-wide. A VMID with no
// matching QEMU guest returns an error that satisfies IsNotFound.
//
// Caveat: /cluster/resources is served from pvestatd's cached view and can lag
// reality by a few seconds, so a VM created, migrated, or destroyed moments
// ago may be resolved to a stale (or missing) node. Use this for resolution
// and routing; the returned handle's own methods hit the authoritative node
// for live truth.
func (c *Client) FindVirtualMachine(ctx context.Context, vmid int) (*VirtualMachine, error) {
	node, err := c.findGuestNode(ctx, vmid, resourceTypeQemu)
	if err != nil {
		return nil, err
	}
	return c.NodeRef(node).VirtualMachineRef(vmid), nil
}

// FindContainer is the LXC counterpart of FindVirtualMachine: it resolves the
// owning node for a container VMID in a single GET /cluster/resources and
// returns a request-free *Container handle bound to that node. The same VMID
// keying and eventual-consistency caveat apply.
func (c *Client) FindContainer(ctx context.Context, vmid int) (*Container, error) {
	node, err := c.findGuestNode(ctx, vmid, resourceTypeLXC)
	if err != nil {
		return nil, err
	}
	return c.NodeRef(node).ContainerRef(vmid), nil
}

// findGuestNode issues one GET /cluster/resources?type=vm (which covers both
// QEMU and LXC guests) and returns the node hosting the guest with the given
// VMID and resource type ("qemu" or "lxc"). The Cluster handle is built
// request-free so the only call made is the resources GET.
func (c *Client) findGuestNode(ctx context.Context, vmid int, kind string) (string, error) {
	cluster := &Cluster{client: c}
	resources, err := cluster.Resources(ctx, "vm")
	if err != nil {
		return "", err
	}

	for _, r := range resources {
		if r.Type == kind && r.VMID == uint64(vmid) {
			if r.Node == "" {
				return "", fmt.Errorf("cluster resource for vmid %d reported no node: %w", vmid, ErrNotFound)
			}
			return r.Node, nil
		}
	}

	noun := "virtual machine"
	if kind == resourceTypeLXC {
		noun = "container"
	}
	return "", fmt.Errorf("no %s with vmid %d in cluster resources: %w", noun, vmid, ErrNotFound)
}
