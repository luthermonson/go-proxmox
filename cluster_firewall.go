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
