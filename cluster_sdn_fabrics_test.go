package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNFabricsIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNFabricsIndex(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 3)
}

func TestCluster_SDNFabricsAll(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	all, err := cluster.SDNFabricsAll(context.Background())
	assert.Nil(t, err)
	assert.Len(t, all.Fabrics, 1)
	assert.Len(t, all.Nodes, 1)
	assert.Equal(t, "fab1", all.Fabrics[0].ID)
}

func TestCluster_SDNFabrics(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	fabrics, err := cluster.SDNFabrics(context.Background(), false, false)
	assert.Nil(t, err)
	assert.Len(t, fabrics, 2)
}

func TestSDNFabric_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	f := cluster.SDNFabric("fab1")
	assert.Nil(t, f.Read(context.Background()))
	assert.Equal(t, "openfabric", f.Protocol)

	assert.Nil(t, f.Update(context.Background(), &SDNFabricOptions{HelloInterval: 5}))
	assert.Nil(t, f.Delete(context.Background()))
}

func TestCluster_NewSDNFabric(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNFabric(context.Background(), &SDNFabricOptions{ID: "fab9", Protocol: "ospf"})
	assert.Nil(t, err)

	assert.NotNil(t, cluster.NewSDNFabric(context.Background(), nil))
	assert.NotNil(t, cluster.NewSDNFabric(context.Background(), &SDNFabricOptions{ID: "x"}))
}

func TestSDNFabric_Nodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	f := cluster.SDNFabric("fab1")
	nodes, err := f.Nodes(context.Background())
	assert.Nil(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "node1", nodes[0].NodeID)
	assert.Equal(t, "fab1", nodes[0].FabricID)
}

func TestCluster_SDNFabricNodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	nodes, err := cluster.SDNFabricNodes(context.Background())
	assert.Nil(t, err)
	assert.Len(t, nodes, 2)
}

func TestSDNFabricNode_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	f := cluster.SDNFabric("fab1")
	n := f.Node("node1")
	assert.Nil(t, n.Read(context.Background()))
	assert.Equal(t, "10.0.0.1", n.IP)

	assert.Nil(t, n.Update(context.Background(), &SDNFabricNodeOptions{}))
	assert.Nil(t, n.Delete(context.Background()))
}

func TestSDNFabric_AddNode(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	f := cluster.SDNFabric("fab1")
	err := f.AddNode(context.Background(), &SDNFabricNodeOptions{NodeID: "node3"})
	assert.Nil(t, err)

	assert.NotNil(t, f.AddNode(context.Background(), nil))
}
