package proxmox

import (
	"context"
	"fmt"
)

func (cl *Cluster) FWGroups(ctx context.Context) (groups []*FirewallSecurityGroup, err error) {
	err = cl.client.Get(ctx, "/cluster/firewall/groups", &groups)

	if nil == err {
		for _, g := range groups {
			g.client = cl.client
		}
	}
	return
}

func (cl *Cluster) FWGroup(ctx context.Context, name string) (group *FirewallSecurityGroup, err error) {
	group = &FirewallSecurityGroup{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/firewall/groups/%s", name), &group.Rules)
	if nil == err {
		group.Group = name
		group.client = cl.client
	}
	return
}

func (cl *Cluster) NewFWGroup(ctx context.Context, group *FirewallSecurityGroup) error {
	return cl.client.Post(ctx, "/cluster/firewall/groups", group, &group)
}

func (g *FirewallSecurityGroup) GetRules(ctx context.Context) ([]*FirewallRule, error) {
	return g.Rules, g.client.Get(ctx, fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), &g.Rules)
}

func (g *FirewallSecurityGroup) Delete(ctx context.Context) error {
	return g.client.Delete(ctx, fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), nil)
}

func (g *FirewallSecurityGroup) RuleCreate(ctx context.Context, rule *FirewallRule) error {
	return g.client.Post(ctx, fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), rule, nil)
}

func (g *FirewallSecurityGroup) RuleUpdate(ctx context.Context, rule *FirewallRule) error {
	return g.client.Put(ctx, fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rule.Pos), rule, nil)
}

func (g *FirewallSecurityGroup) RuleDelete(ctx context.Context, rulePos int) error {
	return g.client.Delete(ctx, fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rulePos), nil)
}

// ---- /cluster/firewall/rules --------------------------------------------------

func (cl *Cluster) FirewallRules(ctx context.Context) (rules []*FirewallRule, err error) {
	err = cl.client.Get(ctx, "/cluster/firewall/rules", &rules)
	return
}

func (cl *Cluster) FirewallRule(ctx context.Context, pos int) (rule *FirewallRule, err error) {
	rule = &FirewallRule{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/firewall/rules/%d", pos), rule)
	return
}

func (cl *Cluster) NewFirewallRule(ctx context.Context, rule *FirewallRule) error {
	return cl.client.Post(ctx, "/cluster/firewall/rules", rule, nil)
}

func (cl *Cluster) FirewallRuleUpdate(ctx context.Context, rule *FirewallRule) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/firewall/rules/%d", rule.Pos), rule, nil)
}

func (cl *Cluster) FirewallRuleDelete(ctx context.Context, pos int) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/firewall/rules/%d", pos), nil)
}

// ---- /cluster/firewall/aliases ------------------------------------------------

func (cl *Cluster) FirewallAliases(ctx context.Context) (aliases []*FirewallAlias, err error) {
	err = cl.client.Get(ctx, "/cluster/firewall/aliases", &aliases)
	return
}

func (cl *Cluster) FirewallAlias(ctx context.Context, name string) (alias *FirewallAlias, err error) {
	alias = &FirewallAlias{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/firewall/aliases/%s", name), alias)
	return
}

func (cl *Cluster) NewFirewallAlias(ctx context.Context, alias *FirewallAliasCreateOption) error {
	return cl.client.Post(ctx, "/cluster/firewall/aliases", alias, nil)
}

func (cl *Cluster) FirewallAliasUpdate(ctx context.Context, name string, update *FirewallAliasUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/firewall/aliases/%s", name), update, nil)
}

func (cl *Cluster) FirewallAliasDelete(ctx context.Context, name string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/firewall/aliases/%s", name), nil)
}

// ---- /cluster/firewall/ipset --------------------------------------------------

func (cl *Cluster) FirewallIPSets(ctx context.Context) (ipsets []*FirewallIPSet, err error) {
	err = cl.client.Get(ctx, "/cluster/firewall/ipset", &ipsets)
	return
}

// FirewallIPSet returns the entries (CIDRs) of a single ipset. The GET on the
// collection-style URL returns the members, not metadata about the ipset itself
// — that matches the PVE wire format.
func (cl *Cluster) FirewallIPSet(ctx context.Context, name string) (entries []*FirewallIPSetEntry, err error) {
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/firewall/ipset/%s", name), &entries)
	return
}

func (cl *Cluster) NewFirewallIPSet(ctx context.Context, ipset *FirewallIPSetCreationOption) error {
	return cl.client.Post(ctx, "/cluster/firewall/ipset", ipset, nil)
}

// FirewallIPSetDelete removes the ipset. PVE rejects deletion of a non-empty
// ipset unless force=1 is set; we pass it as a query param so the JSON body
// stays empty (Delete in this client doesn't take a body).
func (cl *Cluster) FirewallIPSetDelete(ctx context.Context, name string, force bool) error {
	path := fmt.Sprintf("/cluster/firewall/ipset/%s", name)
	if force {
		path += "?force=1"
	}
	return cl.client.Delete(ctx, path, nil)
}

// ---- /cluster/firewall/ipset/{name}/{cidr} -----------------------------------

func (cl *Cluster) FirewallIPSetEntry(ctx context.Context, name, cidr string) (entry *FirewallIPSetEntry, err error) {
	entry = &FirewallIPSetEntry{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/firewall/ipset/%s/%s", name, cidr), entry)
	return
}

func (cl *Cluster) NewFirewallIPSetEntry(ctx context.Context, name string, entry *FirewallIPSetEntryCreationOption) error {
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/firewall/ipset/%s", name), entry, nil)
}

func (cl *Cluster) FirewallIPSetEntryUpdate(ctx context.Context, name, cidr string, entry *FirewallIPSetEntryUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/firewall/ipset/%s/%s", name, cidr), entry, nil)
}

func (cl *Cluster) FirewallIPSetEntryDelete(ctx context.Context, name, cidr string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/firewall/ipset/%s/%s", name, cidr), nil)
}

// ---- /cluster/firewall/options -----------------------------------------------

func (cl *Cluster) FirewallOptions(ctx context.Context) (opts *FirewallClusterOption, err error) {
	opts = &FirewallClusterOption{}
	err = cl.client.Get(ctx, "/cluster/firewall/options", opts)
	return
}

func (cl *Cluster) FirewallOptionsUpdate(ctx context.Context, opts *FirewallClusterOptionUpdateOption) error {
	return cl.client.Put(ctx, "/cluster/firewall/options", opts, nil)
}

// ---- /cluster/firewall/macros + /cluster/firewall/refs -----------------------

func (cl *Cluster) FirewallMacros(ctx context.Context) (macros []*FirewallMacro, err error) {
	err = cl.client.Get(ctx, "/cluster/firewall/macros", &macros)
	return
}

// FirewallRefs lists alias/ipset references usable in rule source/dest. typ
// is optional — pass "alias" or "ipset" to filter, or "" for all.
func (cl *Cluster) FirewallRefs(ctx context.Context, typ string) (refs []*FirewallRef, err error) {
	path := "/cluster/firewall/refs"
	if typ != "" {
		path += "?type=" + typ
	}
	err = cl.client.Get(ctx, path, &refs)
	return
}
