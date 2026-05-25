package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// This file closes the per-VM firewall coverage gaps under
// /nodes/{node}/qemu/{vmid}/firewall/*. Existing rules/ipset/options helpers
// remain in virtual_machine.go.

// Firewall fetches the directory-rooted firewall record (aliases + ipset +
// rules + options) for the VM in one call. Cheaper than four round-trips
// when you just need a snapshot.
func (v *VirtualMachine) Firewall(ctx context.Context) (firewall *Firewall, err error) {
	return firewall, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall", v.Node, v.VMID), &firewall)
}

// FirewallLog returns the per-VM firewall log entries. start/limit page the
// result; since/until filter by UNIX epoch — pass 0 to omit.
func (v *VirtualMachine) FirewallLog(ctx context.Context, start, limit, since, until int) (entries []*FirewallLogEntry, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/qemu/%d/firewall/log", v.Node, v.VMID)}
	q := url.Values{}
	if start > 0 {
		q.Set("start", strconv.Itoa(start))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if since > 0 {
		q.Set("since", strconv.Itoa(since))
	}
	if until > 0 {
		q.Set("until", strconv.Itoa(until))
	}
	u.RawQuery = q.Encode()
	return entries, v.client.Get(ctx, u.String(), &entries)
}

// FirewallRefs lists IPSets and aliases reachable from this VM's scope —
// useful when authoring rules that reference cluster/node-level objects
// alongside VM-local ones. Pass refType ("alias"|"ipset"|"") to filter.
func (v *VirtualMachine) FirewallRefs(ctx context.Context, refType string) (refs []*FirewallRef, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/qemu/%d/firewall/refs", v.Node, v.VMID)}
	if refType != "" {
		q := url.Values{}
		q.Set("type", refType)
		u.RawQuery = q.Encode()
	}
	return refs, v.client.Get(ctx, u.String(), &refs)
}

// GetFirewallAliases lists per-VM firewall aliases.
func (v *VirtualMachine) GetFirewallAliases(ctx context.Context) (aliases []*FirewallAlias, err error) {
	return aliases, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/aliases", v.Node, v.VMID), &aliases)
}

// NewFirewallAlias creates a per-VM firewall alias.
func (v *VirtualMachine) NewFirewallAlias(ctx context.Context, alias *FirewallAlias) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/aliases", v.Node, v.VMID), alias, nil)
}

// GetFirewallAlias reads a single per-VM firewall alias.
func (v *VirtualMachine) GetFirewallAlias(ctx context.Context, name string) (alias *FirewallAlias, err error) {
	return alias, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/aliases/%s", v.Node, v.VMID, name), &alias)
}

// UpdateFirewallAlias mutates an existing per-VM alias.
func (v *VirtualMachine) UpdateFirewallAlias(ctx context.Context, name string, alias *FirewallAlias) error {
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/aliases/%s", v.Node, v.VMID, name), alias, nil)
}

// DeleteFirewallAlias removes a per-VM firewall alias.
func (v *VirtualMachine) DeleteFirewallAlias(ctx context.Context, name string) error {
	return v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/aliases/%s", v.Node, v.VMID, name), nil)
}
