package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, cluster.Version)
	assert.Equal(t, "cluster", cluster.ID)
	for _, n := range cluster.Nodes {
		assert.Contains(t, n.ID, "node/node")
		assert.Equal(t, "node", n.Type)
	}
}

func TestNextID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	nextid, err := cluster.NextID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 100, nextid)
}

func TestCheckID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	checkIDFree, err := cluster.CheckID(ctx, 100)
	assert.Nil(t, err)
	assert.Equal(t, true, checkIDFree)
	checkIDTaken, err := cluster.CheckID(ctx, 200)
	assert.Nil(t, err)
	assert.Equal(t, false, checkIDTaken)
}

func TestCluster_Resources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	rs, err := cluster.Resources(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 20, len(rs))

	// type param test
	rs, err = cluster.Resources(ctx, "node")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
}

func TestCluster_SDNZones(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	zones, err := cluster.SDNZones(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(zones))

	// type param test
	zones, err = cluster.SDNZones(ctx, "vxlan")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(zones))
	assert.Equal(t, "vxlan", zones[0].Type)
	assert.Equal(t, "test1", zones[0].Name)
	assert.Equal(t, "pve", zones[0].IPAM)
}

func TestCluster_SDNVNets(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	vnets, err := cluster.SDNVNets(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(vnets))

	// vnet name test
	vnet, err := cluster.SDNVNet(ctx, "user1")
	assert.Nil(t, err)
	assert.Equal(t, "user1", vnet.Name)
	assert.Equal(t, "vnet", vnet.Type)
	assert.Equal(t, "test1", vnet.Zone)
	assert.Equal(t, 1, vnet.VlanAware)
	assert.Equal(t, uint16(10), vnet.Tag)
}
