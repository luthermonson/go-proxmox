package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_Mappings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	entries, err := cluster.Mappings(ctx)
	assert.Nil(t, err)
	assert.Len(t, entries, 3)
}

// --- dir --------------------------------------------------------------------

func TestCluster_DirMappings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	mappings, err := cluster.DirMappings(ctx, "")
	assert.Nil(t, err)
	assert.Len(t, mappings, 1)
	assert.Equal(t, "shared-iso", mappings[0].ID)
	assert.Len(t, mappings[0].Map, 2)
}

func TestCluster_DirMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	m, err := cluster.DirMapping(ctx, "shared-iso")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "shared-iso", m.ID)
	assert.Equal(t, "d1", m.Digest)

	_, err = cluster.DirMapping(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewDirMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewDirMapping(ctx, &ClusterDirMappingOptions{
		ID:  "shared-iso",
		Map: []string{"node=node1,path=/srv/iso"},
	})
	assert.Nil(t, err)

	err = cluster.NewDirMapping(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateDirMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateDirMapping(ctx, "shared-iso", &ClusterDirMappingOptions{
		Description: "updated",
	})
	assert.Nil(t, err)
	err = cluster.UpdateDirMapping(ctx, "shared-iso", nil)
	assert.Nil(t, err)
	err = cluster.UpdateDirMapping(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteDirMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteDirMapping(ctx, "shared-iso")
	assert.Nil(t, err)
	err = cluster.DeleteDirMapping(ctx, "")
	assert.Error(t, err)
}

// --- pci --------------------------------------------------------------------

func TestCluster_PCIMappings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	mappings, err := cluster.PCIMappings(ctx, "")
	assert.Nil(t, err)
	assert.Len(t, mappings, 1)
	assert.Equal(t, "gpu0", mappings[0].ID)
}

func TestCluster_PCIMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	m, err := cluster.PCIMapping(ctx, "gpu0")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "gpu0", m.ID)
	assert.Equal(t, "p1", m.Digest)
	_, err = cluster.PCIMapping(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewPCIMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewPCIMapping(ctx, &ClusterPCIMappingOptions{
		ID:  "gpu0",
		Map: []string{"node=node1,path=0000:01:00.0,id=10de:1eb8"},
	})
	assert.Nil(t, err)
	err = cluster.NewPCIMapping(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdatePCIMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdatePCIMapping(ctx, "gpu0", &ClusterPCIMappingOptions{MDev: true})
	assert.Nil(t, err)
	err = cluster.UpdatePCIMapping(ctx, "gpu0", nil)
	assert.Nil(t, err)
	err = cluster.UpdatePCIMapping(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeletePCIMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeletePCIMapping(ctx, "gpu0")
	assert.Nil(t, err)
	err = cluster.DeletePCIMapping(ctx, "")
	assert.Error(t, err)
}

// --- usb --------------------------------------------------------------------

func TestCluster_USBMappings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	mappings, err := cluster.USBMappings(ctx, "")
	assert.Nil(t, err)
	assert.Len(t, mappings, 1)
	assert.Equal(t, "yubikey", mappings[0].ID)
}

func TestCluster_USBMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	m, err := cluster.USBMapping(ctx, "yubikey")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "yubikey", m.ID)
	assert.Equal(t, "u1", m.Digest)
	_, err = cluster.USBMapping(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewUSBMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewUSBMapping(ctx, &ClusterUSBMappingOptions{
		ID:  "yubikey",
		Map: []string{"node=node1,id=1050:0407"},
	})
	assert.Nil(t, err)
	err = cluster.NewUSBMapping(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateUSBMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateUSBMapping(ctx, "yubikey", &ClusterUSBMappingOptions{Description: "x"})
	assert.Nil(t, err)
	err = cluster.UpdateUSBMapping(ctx, "yubikey", nil)
	assert.Nil(t, err)
	err = cluster.UpdateUSBMapping(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_Mappings_CheckNode(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// Each listing helper appends a check-node query string when given a node
	// name. The shared pve9x mocks ignore extra params, so the request still
	// matches and we cover the URL-building branch.
	_, err = cluster.DirMappings(ctx, "node1")
	assert.Nil(t, err)
	_, err = cluster.PCIMappings(ctx, "node1")
	assert.Nil(t, err)
	_, err = cluster.USBMappings(ctx, "node1")
	assert.Nil(t, err)
}

func TestCluster_DeleteUSBMapping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteUSBMapping(ctx, "yubikey")
	assert.Nil(t, err)
	err = cluster.DeleteUSBMapping(ctx, "")
	assert.Error(t, err)
}
