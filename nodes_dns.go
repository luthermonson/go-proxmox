package proxmox

import (
	"context"
	"fmt"
)

// DNS returns the resolver configuration for this node — what gets written to
// /etc/resolv.conf. Any of search/dns1/dns2/dns3 may be empty if the node has
// no value set for that slot.
func (n *Node) DNS(ctx context.Context) (*NodeDNS, error) {
	var dns NodeDNS
	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/dns", n.Name), &dns); err != nil {
		return nil, err
	}
	return &dns, nil
}

// UpdateDNS rewrites the node's resolver configuration. The Search field is
// required by PVE; an empty Search will be rejected by the server. The three
// DNS slots are optional individually but supplied together — sending the
// struct replaces all of them.
func (n *Node) UpdateDNS(ctx context.Context, dns *NodeDNS) error {
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/dns", n.Name), dns, nil)
}
