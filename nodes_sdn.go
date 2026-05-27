package proxmox

import (
	"context"
	"fmt"
)

// /nodes/{node}/sdn — runtime visibility into SDN on this node. All endpoints
// are read-only; mutations live on Cluster (see cluster_sdn.go).

// SDNIndex enumerates the children of /nodes/{node}/sdn (typically "vnets",
// "zones", "fabrics"). The PVE schema documents it as a directory index, so
// we collapse the {"subdir": ...} link objects into a flat []string.
func (n *Node) SDNIndex(ctx context.Context) ([]string, error) {
	return n.sdnDiridx(ctx, fmt.Sprintf("/nodes/%s/sdn", n.Name))
}

// SDNFabricIndex enumerates the children of /nodes/{node}/sdn/fabrics/{fabric}
// (typically "routes", "neighbors", "interfaces").
func (n *Node) SDNFabricIndex(ctx context.Context, fabric string) ([]string, error) {
	if fabric == "" {
		return nil, fmt.Errorf("fabric is required")
	}
	return n.sdnDiridx(ctx, fmt.Sprintf("/nodes/%s/sdn/fabrics/%s", n.Name, fabric))
}

// SDNFabricInterfaces returns the interfaces participating in the named fabric.
func (n *Node) SDNFabricInterfaces(ctx context.Context, fabric string) (ifaces []*SDNFabricInterface, err error) {
	if fabric == "" {
		return nil, fmt.Errorf("fabric is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/fabrics/%s/interfaces", n.Name, fabric), &ifaces)
	return
}

// SDNFabricNeighbors returns the FRR neighbor table for the named fabric.
func (n *Node) SDNFabricNeighbors(ctx context.Context, fabric string) (neighbors []*SDNFabricNeighbor, err error) {
	if fabric == "" {
		return nil, fmt.Errorf("fabric is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/fabrics/%s/neighbors", n.Name, fabric), &neighbors)
	return
}

// SDNFabricRoutes returns routes learned/configured for the named fabric.
func (n *Node) SDNFabricRoutes(ctx context.Context, fabric string) (routes []*SDNFabricRoute, err error) {
	if fabric == "" {
		return nil, fmt.Errorf("fabric is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/fabrics/%s/routes", n.Name, fabric), &routes)
	return
}

// SDNVNetIndex enumerates the children of /nodes/{node}/sdn/vnets/{vnet}
// (currently just "mac-vrf" on EVPN zones).
func (n *Node) SDNVNetIndex(ctx context.Context, vnet string) ([]string, error) {
	if vnet == "" {
		return nil, fmt.Errorf("vnet is required")
	}
	return n.sdnDiridx(ctx, fmt.Sprintf("/nodes/%s/sdn/vnets/%s", n.Name, vnet))
}

// SDNVNetMACVRF returns the MAC VRF for a VNet in an EVPN zone — entries
// either self-originated by this node or learned via BGP.
func (n *Node) SDNVNetMACVRF(ctx context.Context, vnet string) (entries []*SDNMACVRFEntry, err error) {
	if vnet == "" {
		return nil, fmt.Errorf("vnet is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/vnets/%s/mac-vrf", n.Name, vnet), &entries)
	return
}

// SDNZones returns the runtime status of every SDN zone visible to the node.
// Distinct from Cluster.SDNZones, which returns config — this returns
// per-node deployment state ("available", "pending", "error").
func (n *Node) SDNZones(ctx context.Context) (zones []*SDNZoneStatus, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/zones", n.Name), &zones)
	return
}

// SDNZoneIndex enumerates the children of /nodes/{node}/sdn/zones/{zone}
// (typically "content", "bridges", "ip-vrf").
func (n *Node) SDNZoneIndex(ctx context.Context, zone string) ([]string, error) {
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}
	return n.sdnDiridx(ctx, fmt.Sprintf("/nodes/%s/sdn/zones/%s", n.Name, zone))
}

// SDNZoneBridges returns the bridges (vnets) deployed for the zone, with
// their member ports — useful for correlating guest NICs to VLANs.
func (n *Node) SDNZoneBridges(ctx context.Context, zone string) (bridges []*SDNZoneBridge, err error) {
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/zones/%s/bridges", n.Name, zone), &bridges)
	return
}

// SDNZoneContent lists VNets in the zone with their per-node deployment status.
func (n *Node) SDNZoneContent(ctx context.Context, zone string) (content []*SDNZoneContent, err error) {
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/zones/%s/content", n.Name, zone), &content)
	return
}

// SDNZoneIPVRF returns the IP VRF of an EVPN zone (entries from BGP and the
// kernel routing table, excluding the /32s for guests on this host — those
// go through the vnet bridge directly).
func (n *Node) SDNZoneIPVRF(ctx context.Context, zone string) (entries []*SDNIPVRFEntry, err error) {
	if zone == "" {
		return nil, fmt.Errorf("zone is required")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/sdn/zones/%s/ip-vrf", n.Name, zone), &entries)
	return
}

// sdnDiridx is a shared helper for the four SDN directory-index endpoints.
// PVE returns [{"subdir":"..."}] — we collapse to []string.
func (n *Node) sdnDiridx(ctx context.Context, path string) ([]string, error) {
	var items []struct {
		Subdir string `json:"subdir"`
	}
	if err := n.client.Get(ctx, path, &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Subdir)
	}
	return out, nil
}
