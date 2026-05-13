package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_FWGroups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	groups, err := cluster.FWGroups(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, groups)
	assert.GreaterOrEqual(t, len(groups), 1)

	// Check first group
	assert.NotEmpty(t, groups[0].Group)
	assert.NotNil(t, groups[0].client)
}

func TestCluster_FWGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, "test-group", group.Group)
	assert.NotNil(t, group.client)
	assert.NotEmpty(t, group.Rules)
	assert.GreaterOrEqual(t, len(group.Rules), 1)
}

func TestCluster_NewFWGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	newGroup := &FirewallSecurityGroup{
		Group:   "new-group",
		Comment: "New test group",
	}

	err = cluster.NewFWGroup(ctx, newGroup)
	assert.Nil(t, err)
}

func TestFirewallSecurityGroup_GetRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)

	rules, err := group.GetRules(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, rules)
	assert.GreaterOrEqual(t, len(rules), 1)

	// Check first rule
	rule := rules[0]
	assert.NotEmpty(t, rule.Type)
	assert.NotEmpty(t, rule.Action)
}

func TestFirewallSecurityGroup_RuleCreate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)

	newRule := &FirewallRule{
		Type:    "in",
		Action:  "ACCEPT",
		Enable:  1,
		Proto:   "tcp",
		Dport:   "443",
		Comment: "Allow HTTPS",
	}

	err = group.RuleCreate(ctx, newRule)
	assert.Nil(t, err)
}

func TestFirewallSecurityGroup_RuleUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)

	updateRule := &FirewallRule{
		Pos:     0,
		Type:    "in",
		Action:  "DROP",
		Enable:  1,
		Proto:   "tcp",
		Dport:   "22",
		Comment: "Block SSH",
	}

	err = group.RuleUpdate(ctx, updateRule)
	assert.Nil(t, err)
}

func TestFirewallSecurityGroup_RuleDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)

	err = group.RuleDelete(ctx, 0)
	assert.Nil(t, err)
}

func TestFirewallSecurityGroup_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.FWGroup(ctx, "test-group")
	assert.Nil(t, err)

	err = group.Delete(ctx)
	assert.Nil(t, err)
}

// ---- cluster-level firewall (rules / aliases / ipset / options / macros / refs) ----

func TestCluster_FirewallRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	rules, err := cluster.FirewallRules(ctx)
	assert.Nil(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "in", rules[0].Type)
	assert.Equal(t, "ACCEPT", rules[0].Action)
	assert.Equal(t, 1, rules[0].Enable)

	rule, err := cluster.FirewallRule(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, "ACCEPT", rule.Action)

	assert.Nil(t, cluster.NewFirewallRule(ctx, &FirewallRule{Type: "in", Action: "ACCEPT", Proto: "tcp", Dport: "443"}))
	assert.Nil(t, cluster.FirewallRuleUpdate(ctx, &FirewallRule{Pos: 0, Action: "DROP"}))
	assert.Nil(t, cluster.FirewallRuleDelete(ctx, 0))
}

func TestCluster_FirewallAliases(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	aliases, err := cluster.FirewallAliases(ctx)
	assert.Nil(t, err)
	assert.Len(t, aliases, 1)
	assert.Equal(t, "test-alias", aliases[0].Name)
	assert.Equal(t, "10.0.0.0/24", aliases[0].Cidr)

	alias, err := cluster.FirewallAlias(ctx, "test-alias")
	assert.Nil(t, err)
	assert.Equal(t, "test-alias", alias.Name)

	assert.Nil(t, cluster.NewFirewallAlias(ctx, &FirewallAliasCreateOption{Name: "test-alias", CIDR: "10.0.0.0/24"}))
	assert.Nil(t, cluster.FirewallAliasUpdate(ctx, "test-alias", &FirewallAliasUpdateOption{CIDR: "10.0.0.0/16", Comment: "expanded"}))
	assert.Nil(t, cluster.FirewallAliasDelete(ctx, "test-alias"))
}

func TestCluster_FirewallIPSet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	ipsets, err := cluster.FirewallIPSets(ctx)
	assert.Nil(t, err)
	assert.Len(t, ipsets, 1)
	assert.Equal(t, "test-ipset", ipsets[0].Name)

	entries, err := cluster.FirewallIPSet(ctx, "test-ipset")
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.True(t, entries[1].NoMatch, "second mock entry advertises nomatch=true")

	assert.Nil(t, cluster.NewFirewallIPSet(ctx, &FirewallIPSetCreationOption{Name: "test-ipset", Comment: "test"}))
	assert.Nil(t, cluster.FirewallIPSetDelete(ctx, "test-ipset", false))
	assert.Nil(t, cluster.FirewallIPSetDelete(ctx, "test-ipset", true))

	entry, err := cluster.FirewallIPSetEntry(ctx, "test-ipset", "10.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, "10.0.0.1", entry.CIDR)

	assert.Nil(t, cluster.NewFirewallIPSetEntry(ctx, "test-ipset", &FirewallIPSetEntryCreationOption{CIDR: "10.0.0.1"}))
	assert.Nil(t, cluster.FirewallIPSetEntryUpdate(ctx, "test-ipset", "10.0.0.1", &FirewallIPSetEntryUpdateOption{Comment: "renamed"}))
	assert.Nil(t, cluster.FirewallIPSetEntryDelete(ctx, "test-ipset", "10.0.0.1"))
}

func TestCluster_FirewallOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	opts, err := cluster.FirewallOptions(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, opts.Ebtables, "mock returns ebtables=1, must round-trip as non-nil *IntOrBool")
	assert.Equal(t, IntOrBool(true), *opts.Ebtables)
	assert.Equal(t, "DROP", opts.PolicyIn)

	assert.Nil(t, cluster.FirewallOptionsUpdate(ctx, &FirewallClusterOptionUpdateOption{
		Enable:   1,
		PolicyIn: "DROP",
	}))
}

func TestCluster_FirewallMacrosAndRefs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	macros, err := cluster.FirewallMacros(ctx)
	assert.Nil(t, err)
	assert.Len(t, macros, 2)
	assert.Equal(t, "HTTP", macros[0].Macro)

	refs, err := cluster.FirewallRefs(ctx, "")
	assert.Nil(t, err)
	assert.Len(t, refs, 2)

	refs, err = cluster.FirewallRefs(ctx, "alias")
	assert.Nil(t, err)
	assert.NotEmpty(t, refs)
}
