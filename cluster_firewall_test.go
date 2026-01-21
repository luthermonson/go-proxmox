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
