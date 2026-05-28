package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_JoinAPIVersion(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, err := cluster.JoinAPIVersion(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, v)
}

func TestCluster_JoinInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	info, err := cluster.JoinInfo(context.Background(), "")
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "node1", info.PreferredNode)
	assert.Equal(t, "abc123", info.ConfigDigest)
	assert.Len(t, info.NodeList, 1)
	assert.Equal(t, "node1", info.NodeList[0].Name)
	assert.NotNil(t, info.Totem)
}

func TestCluster_JoinCluster(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	status, err := cluster.JoinCluster(context.Background(), &ClusterJoinOptions{
		Hostname:    "10.0.0.1",
		Password:    "secret",
		Fingerprint: "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99",
	})
	assert.Nil(t, err)
	assert.Equal(t, "OK", status)

	_, err = cluster.JoinCluster(context.Background(), nil)
	assert.Error(t, err)
	_, err = cluster.JoinCluster(context.Background(), &ClusterJoinOptions{Hostname: "x"})
	assert.Error(t, err)
	_, err = cluster.JoinCluster(context.Background(), &ClusterJoinOptions{Hostname: "x", Password: "p"})
	assert.Error(t, err)
}

func TestCluster_ConfigNodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	nodes, err := cluster.ConfigNodes(context.Background())
	assert.Nil(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "node1", nodes[0].Node)
}

func TestCluster_QDevice(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	status, err := cluster.QDevice(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "running", status["state"])
}

func TestCluster_Totem(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	totem, err := cluster.Totem(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "knet", totem["transport"])
}

func TestCluster_CreateCluster(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	status, err := cluster.CreateCluster(context.Background(), &ClusterCreateOptions{ClusterName: "mycluster"})
	assert.Nil(t, err)
	assert.Equal(t, "OK", status)

	_, err = cluster.CreateCluster(context.Background(), nil)
	assert.Error(t, err)
	_, err = cluster.CreateCluster(context.Background(), &ClusterCreateOptions{})
	assert.Error(t, err)
}

func TestCluster_AddConfigNode(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	res, err := cluster.AddConfigNode(context.Background(), "node2", &ClusterAddNodeOptions{NewNodeIP: "10.0.0.2"})
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.CorosyncConf)
	assert.NotEmpty(t, res.CorosyncAuthkey)

	_, err = cluster.AddConfigNode(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteConfigNode(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	assert.Nil(t, cluster.DeleteConfigNode(context.Background(), "node2"))
	assert.Error(t, cluster.DeleteConfigNode(context.Background(), ""))
}
