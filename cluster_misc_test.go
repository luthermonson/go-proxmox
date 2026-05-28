package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_Log(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.Log(context.Background(), 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "node1", entries[0].Node)
}

func TestCluster_Log_WithMax(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.Log(context.Background(), 50)
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
}

func TestCluster_NotificationEndpointsSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subs, err := cluster.NotificationEndpointsSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subs, "sendmail")
	assert.Contains(t, subs, "gotify")
	assert.Contains(t, subs, "smtp")
	assert.Contains(t, subs, "webhook")
}

func TestFirewallSecurityGroup_GetRule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	group, err := cluster.FWGroup(context.Background(), "test-group")
	assert.Nil(t, err)
	assert.NotNil(t, group)

	rule, err := group.GetRule(context.Background(), 0)
	assert.Nil(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "in", rule.Type)
	assert.Equal(t, "ACCEPT", rule.Action)

	// blank-group guard
	empty := &FirewallSecurityGroup{}
	_, err = empty.GetRule(context.Background(), 0)
	assert.Error(t, err)
}
