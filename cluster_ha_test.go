package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_HAGroups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	groups, err := cluster.HAGroups(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, groups)
	assert.Equal(t, "test-group", groups[0].Group)
}

func TestCluster_HAGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	group, err := cluster.HAGroup(ctx, "test-group")
	assert.Nil(t, err)
	assert.Equal(t, "test-group", group.Group)
	assert.Equal(t, "node1,node2", group.Nodes)
}

func TestCluster_NewHAGroup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewHAGroup(ctx, &HAGroupCreateOption{
		Group: "test-group",
		Nodes: "node1,node2",
	})
	assert.Nil(t, err)
}

func TestCluster_HAGroupUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.HAGroupUpdate(ctx, "test-group", &HAGroupUpdateOption{
		Comment: "updated",
	})
	assert.Nil(t, err)
}

func TestCluster_HAGroupDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.HAGroupDelete(ctx, "test-group"))
}

func TestCluster_HAResources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	resources, err := cluster.HAResources(ctx, "")
	assert.Nil(t, err)
	assert.NotEmpty(t, resources)
	assert.Equal(t, "vm:100", resources[0].SID)
}

func TestCluster_HAResource(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	r, err := cluster.HAResource(ctx, "vm:100")
	assert.Nil(t, err)
	assert.Equal(t, "vm:100", r.SID)
	assert.NotNil(t, r.State)
	assert.Equal(t, "started", *r.State)
	assert.NotNil(t, r.MaxRelocate)
	assert.Equal(t, 1, *r.MaxRelocate)
}

func TestCluster_NewHAResource(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewHAResource(ctx, &HAResourceCreateOption{
		SID:   "vm:100",
		Group: "test-group",
		State: Ptr("started"),
	})
	assert.Nil(t, err)
}

func TestCluster_HAResourceUpdateAndDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.HAResourceUpdate(ctx, "vm:100", &HAResourceUpdateOption{
		State: Ptr("disabled"),
	}))
	// purge=true exercises the query-param path
	assert.Nil(t, cluster.HAResourceDelete(ctx, "vm:100", true))
}

func TestCluster_HAResourceMigrateAndRelocate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.HAResourceMigrate(ctx, "vm:100", "node2"))
	assert.Nil(t, cluster.HAResourceRelocate(ctx, "vm:100", "node2"))
}

func TestCluster_HARules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	rules, err := cluster.HARules(ctx, "", "")
	assert.Nil(t, err)
	assert.NotEmpty(t, rules)
	assert.Equal(t, "rule-1", rules[0].Rule)
}

func TestCluster_HARule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	r, err := cluster.HARule(ctx, "rule-1")
	assert.Nil(t, err)
	assert.Equal(t, "rule-1", r.Rule)
	assert.Equal(t, "node-affinity", r.Type)
}

func TestCluster_NewHARule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewHARule(ctx, &HARuleCreateOption{
		Rule:      "rule-1",
		Type:      "node-affinity",
		Resources: "vm:100",
		Nodes:     "node1",
	})
	assert.Nil(t, err)
}

func TestCluster_HARuleUpdateAndDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.HARuleUpdate(ctx, "rule-1", &HARuleUpdateOption{Comment: "updated"}))
	assert.Nil(t, cluster.HARuleDelete(ctx, "rule-1"))
}

func TestCluster_HAStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	status, err := cluster.HAStatus(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, status)
	assert.Equal(t, "node1", status[0].ID)
}

func TestCluster_HAManagerStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	mgr, err := cluster.HAManagerStatus(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, mgr.NodeStatus)
	assert.Equal(t, "online", mgr.NodeStatus["node1"])
}

func TestCluster_HAArm(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	assert.Nil(t, cluster.HAArm(context.Background()))
}

func TestCluster_HADisarm(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	assert.Nil(t, cluster.HADisarm(context.Background(), "freeze"))
	assert.Error(t, cluster.HADisarm(context.Background(), ""))
}
