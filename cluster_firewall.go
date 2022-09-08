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

func (cl *Cluster) NewFWGroup(group *FirewallSecurityGroup) (err error) {
	err = cl.client.Post(fmt.Sprintf("/cluster/firewall/groups"), group, &group)
	return
}

func (g *FirewallSecurityGroup) GetRules() (rules []*FirewallRule, err error) {
	err = g.client.Get(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), &g.Rules)
	rules = g.Rules
	return
}
func (g *FirewallSecurityGroup) Delete() (err error) {
	err = g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), nil)
	return
}
func (g *FirewallSecurityGroup) RuleCreate(rule *FirewallRule) (err error) {
	err = g.client.Post(fmt.Sprintf("/cluster/firewall/groups/%s", g.Group), rule, nil)
	return
}
func (g *FirewallSecurityGroup) RuleUpdate(rule *FirewallRule) (err error) {
	err = g.client.Put(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rule.Pos), rule, nil)
	return
}
func (g *FirewallSecurityGroup) RuleDelete(rulePos int) (err error) {
	err = g.client.Delete(fmt.Sprintf("/cluster/firewall/groups/%s/%d", g.Group, rulePos), nil)
	return
}
