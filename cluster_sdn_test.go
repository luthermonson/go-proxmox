package proxmox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNSubnets(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subnets, err := cluster.SDNSubnets(context.Background(), "user1")
	assert.Nil(t, err)
	assert.Len(t, subnets, 1)
	assert.Equal(t, "10.0.0.0/24", subnets[0].CIDR)
}

func TestCluster_SDNApply(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	task, err := cluster.SDNApply(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
}

func TestCluster_SDNVNets_ClientThreaded(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	vnets, err := cluster.SDNVNets(context.Background())
	assert.Nil(t, err)
	assert.Len(t, vnets, 5)
	assert.Equal(t, "user1", vnets[0].Name)
	// client should be threaded onto each entry so chained mutations work
	assert.NotNil(t, vnets[0].client)
}

func TestCluster_SDNVNet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, err := cluster.SDNVNet(context.Background(), "user1")
	assert.Nil(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "user1", v.Name)
	assert.NotNil(t, v.client)
}

func TestCluster_NewSDNVNet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNVNet(context.Background(), &VNetOptions{Name: "user99", Zone: "test1", Type: "vnet"})
	assert.Nil(t, err)
}

func TestCluster_UpdateSDNVNet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.UpdateSDNVNet(context.Background(), &VNet{Name: "user1", Zone: "test1", Alias: "renamed"})
	assert.Nil(t, err)
}

func TestCluster_DeleteSDNVNet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.DeleteSDNVNet(context.Background(), "user1")
	assert.Nil(t, err)
}

func TestCluster_SDNZonesFiltered(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	// no filter
	zones, err := cluster.SDNZones(context.Background())
	assert.Nil(t, err)
	assert.Len(t, zones, 2)

	// with filter — variadic, multiple strings get joined and spaces stripped
	zones, err = cluster.SDNZones(context.Background(), "vx", "lan ")
	assert.Nil(t, err)
	assert.Len(t, zones, 1)
}

func TestCluster_SDNZone_Basic(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	zone, err := cluster.SDNZone(context.Background(), "test1")
	assert.Nil(t, err)
	assert.Equal(t, "test1", zone.Name)
	assert.Equal(t, "vxlan", zone.Type)
	// Pending is an object of not-yet-applied changes. The old `bool` typing
	// made the whole GET fail to decode whenever pending data was present.
	assert.Equal(t, "changed", zone.State)
	assert.Equal(t, "host1,host2,host3", zone.Pending["nodes"])
}

func TestSDNZoneOptions_ExitNodeTags(t *testing.T) {
	// Write path: the SDN zone schema spells the exit-node keys without
	// hyphens (exitnodes / exitnodes-local-routing / exitnodes-primary); the
	// old `exit-nodes*` tags were silently ignored by PVE.
	opts := &SDNZoneOptions{
		Name:                  "evpn1",
		Type:                  "evpn",
		ExitNodes:             "node1,node2",
		ExitNodesLocalRouting: IntOrBool(true),
		ExitNodesPrimary:      "node1",
	}
	b, err := json.Marshal(opts)
	assert.Nil(t, err)
	s := string(b)
	assert.Contains(t, s, `"exitnodes":"node1,node2"`)
	assert.Contains(t, s, `"exitnodes-local-routing":1`)
	assert.Contains(t, s, `"exitnodes-primary":"node1"`)
	assert.NotContains(t, s, `"exit-nodes"`)
}

func TestCluster_NewSDNZone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNZone(context.Background(), &SDNZoneOptions{Name: "newzone", Type: "simple"})
	assert.Nil(t, err)
}

func TestCluster_UpdateSDNZone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.UpdateSDNZone(context.Background(), &SDNZoneOptions{Name: "test1", Type: "vxlan", MTU: 1450})
	assert.Nil(t, err)
}

func TestCluster_DeleteSDNZone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.DeleteSDNZone(context.Background(), "test1")
	assert.Nil(t, err)
}
