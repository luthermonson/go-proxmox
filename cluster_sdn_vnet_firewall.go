package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// FirewallIndex returns the per-VNet firewall directory entries
// ({"options", "rules"}).
//
// GET /cluster/sdn/vnets/{vnet}/firewall
func (v *VNet) FirewallIndex(ctx context.Context) (entries []map[string]any, err error) {
	if v.Name == "" {
		return nil, errors.New("vnet name is required")
	}
	err = v.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall", v.Name), &entries)
	return
}

// FirewallOptions reads the per-VNet firewall toggle/policy/log-level.
//
// GET /cluster/sdn/vnets/{vnet}/firewall/options
func (v *VNet) FirewallOptions(ctx context.Context) (*SDNVNetFirewallOptions, error) {
	if v.Name == "" {
		return nil, errors.New("vnet name is required")
	}
	out := &SDNVNetFirewallOptions{}
	if err := v.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/options", v.Name), out); err != nil {
		return nil, err
	}
	return out, nil
}

// FirewallOptionsUpdate mutates the per-VNet firewall options.
//
// PUT /cluster/sdn/vnets/{vnet}/firewall/options
func (v *VNet) FirewallOptionsUpdate(ctx context.Context, opts *SDNVNetFirewallOptionsUpdate) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil {
		opts = &SDNVNetFirewallOptionsUpdate{}
	}
	return v.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/options", v.Name), opts, nil)
}

// FirewallRules lists the firewall rules on this VNet.
//
// GET /cluster/sdn/vnets/{vnet}/firewall/rules
func (v *VNet) FirewallRules(ctx context.Context) (rules []*SDNVNetFirewallRule, err error) {
	if v.Name == "" {
		return nil, errors.New("vnet name is required")
	}
	err = v.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/rules", v.Name), &rules)
	return
}

// FirewallRule reads a single firewall rule on this VNet by position.
//
// GET /cluster/sdn/vnets/{vnet}/firewall/rules/{pos}
func (v *VNet) FirewallRule(ctx context.Context, pos int) (*SDNVNetFirewallRule, error) {
	if v.Name == "" {
		return nil, errors.New("vnet name is required")
	}
	rule := &SDNVNetFirewallRule{}
	if err := v.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/rules/%d", v.Name, pos), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// NewFirewallRule creates a new firewall rule on this VNet. opts.Type and
// opts.Action are required.
//
// POST /cluster/sdn/vnets/{vnet}/firewall/rules
func (v *VNet) NewFirewallRule(ctx context.Context, opts *SDNVNetFirewallRuleOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil || opts.Type == "" || opts.Action == "" {
		return errors.New("vnet firewall rule type and action are required")
	}
	return v.client.Post(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/rules", v.Name), opts, nil)
}

// FirewallRuleUpdate mutates an existing firewall rule by position.
//
// PUT /cluster/sdn/vnets/{vnet}/firewall/rules/{pos}
func (v *VNet) FirewallRuleUpdate(ctx context.Context, pos int, opts *SDNVNetFirewallRuleOptions) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	if opts == nil {
		opts = &SDNVNetFirewallRuleOptions{}
	}
	return v.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/rules/%d", v.Name, pos), opts, nil)
}

// FirewallRuleDelete removes a firewall rule by position.
//
// DELETE /cluster/sdn/vnets/{vnet}/firewall/rules/{pos}
func (v *VNet) FirewallRuleDelete(ctx context.Context, pos int) error {
	if v.Name == "" {
		return errors.New("vnet name is required")
	}
	return v.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/firewall/rules/%d", v.Name, pos), nil)
}
