package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func sdnNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_SDNIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := sdnNode().SDNIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"vnets", "zones", "fabrics"}, subdirs)
}

func TestNode_SDNFabricIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := sdnNode().SDNFabricIndex(context.Background(), "fab1")
	assert.Nil(t, err)
	assert.Equal(t, []string{"routes", "neighbors", "interfaces"}, subdirs)

	_, err = sdnNode().SDNFabricIndex(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNFabricInterfaces(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ifaces, err := sdnNode().SDNFabricInterfaces(context.Background(), "fab1")
	assert.Nil(t, err)
	assert.Len(t, ifaces, 2)
	assert.Equal(t, "eth0", ifaces[0].Name)
	assert.Equal(t, "up", ifaces[0].State)

	_, err = sdnNode().SDNFabricInterfaces(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNFabricNeighbors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	neighbors, err := sdnNode().SDNFabricNeighbors(context.Background(), "fab1")
	assert.Nil(t, err)
	assert.Len(t, neighbors, 1)
	assert.Equal(t, "Established", neighbors[0].Status)
	assert.Equal(t, "8h24m12s", neighbors[0].Uptime)

	_, err = sdnNode().SDNFabricNeighbors(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNFabricRoutes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	routes, err := sdnNode().SDNFabricRoutes(context.Background(), "fab1")
	assert.Nil(t, err)
	assert.Len(t, routes, 2)
	assert.Equal(t, "10.0.0.0/24", routes[0].Route)
	assert.Len(t, routes[0].Via, 2)

	_, err = sdnNode().SDNFabricRoutes(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNVNetIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := sdnNode().SDNVNetIndex(context.Background(), "vnet1")
	assert.Nil(t, err)
	assert.Equal(t, []string{"mac-vrf"}, subdirs)

	_, err = sdnNode().SDNVNetIndex(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNVNetMACVRF(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := sdnNode().SDNVNetMACVRF(context.Background(), "vnet1")
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", entries[0].MAC)

	_, err = sdnNode().SDNVNetMACVRF(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNZones(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	zones, err := sdnNode().SDNZones(context.Background())
	assert.Nil(t, err)
	assert.Len(t, zones, 2)
	assert.Equal(t, "zone1", zones[0].Zone)
	assert.Equal(t, "available", zones[0].Status)
}

func TestNode_SDNZoneIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := sdnNode().SDNZoneIndex(context.Background(), "zone1")
	assert.Nil(t, err)
	assert.Equal(t, []string{"content", "bridges", "ip-vrf"}, subdirs)

	_, err = sdnNode().SDNZoneIndex(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNZoneBridges(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	bridges, err := sdnNode().SDNZoneBridges(context.Background(), "zone1")
	assert.Nil(t, err)
	assert.Len(t, bridges, 1)
	assert.Equal(t, "vnet1", bridges[0].Name)
	assert.Equal(t, "1", bridges[0].VLANFiltering)
	assert.Len(t, bridges[0].Ports, 2)
	assert.Equal(t, float64(100), bridges[0].Ports[0].VMID)
	assert.Equal(t, float64(100), bridges[0].Ports[0].PrimaryVLAN)

	_, err = sdnNode().SDNZoneBridges(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNZoneContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	content, err := sdnNode().SDNZoneContent(context.Background(), "zone1")
	assert.Nil(t, err)
	assert.Len(t, content, 2)
	assert.Equal(t, "vnet1", content[0].VNet)
	assert.Equal(t, "awaiting reload", content[1].StatusMsg)

	_, err = sdnNode().SDNZoneContent(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_SDNZoneIPVRF(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := sdnNode().SDNZoneIPVRF(context.Background(), "zone1")
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "10.0.0.0/24", entries[0].IP)
	assert.Equal(t, 20, entries[0].Metric)
	assert.Equal(t, "bgp", entries[0].Protocol)

	_, err = sdnNode().SDNZoneIPVRF(context.Background(), "")
	assert.NotNil(t, err)
}
