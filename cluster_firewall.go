package proxmox

import (
	"fmt"
)

func (cl *Cluster) FWGroups() (groups []*FirewallSecurityGroup, err error) {
	err = cl.client.Get("/cluster/firewall/groups", &groups)

	if nil == err {
		for _, g := range groups {
			g.client = cl.client
		}
	}
	return
}

func (cl *Cluster) FWGroup(name string) (group *FirewallSecurityGroup, err error) {
	group = &FirewallSecurityGroup{}
	err = cl.client.Get(fmt.Sprintf("/cluster/firewall/groups/%s", name), &group.Rules)
	if nil == err {
		group.Group = name
		group.client = cl.client
	}
	return
}

func (cl *Cluster) NewFWGroup(group *FirewallSecurityGroup) error {
	return cl.client.Post(fmt.Sprintf("/cluster/firewall/groups"), group, &group)
}

func (g *FirewallSecurityGroup) GetRules() ([]*FirewallRule, error) {
	return g.Rules, g.client.Get(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), &g.Rules)
}

func (g *FirewallSecurityGroup) Delete() error {
	return g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), nil)
}

func (g *FirewallSecurityGroup) RuleCreate(rule *FirewallRule) error {
	return g.client.Post(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), rule, nil)
}

func (g *FirewallSecurityGroup) RuleUpdate(rule *FirewallRule) error {
	return g.client.Put(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rule.Pos), rule, nil)
}

func (g *FirewallSecurityGroup) RuleDelete(rulePos int) error {
	return g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rulePos), nil)
}
