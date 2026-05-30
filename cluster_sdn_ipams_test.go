package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNIPAMs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	ipams, err := cluster.SDNIPAMs(context.Background(), "")
	assert.Nil(t, err)
	assert.Len(t, ipams, 2)
	assert.Equal(t, "pve", ipams[0].IPAM)
}

func TestSDNIPAM_ReadUpdateDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	i := cluster.SDNIPAM("pve")
	assert.Nil(t, i.Read(context.Background()))
	assert.Equal(t, "pve", i.Type)

	assert.Nil(t, i.Update(context.Background(), &SDNIPAMOptions{}))
	assert.Nil(t, i.Delete(context.Background()))
}

func TestSDNIPAM_Status(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	i := cluster.SDNIPAM("pve")
	entries, err := i.Status(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "vm100", entries[0]["hostname"])
}

func TestCluster_NewSDNIPAM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNIPAM(context.Background(), &SDNIPAMOptions{IPAM: "netbox2", Type: "netbox", URL: "https://x"})
	assert.Nil(t, err)

	assert.NotNil(t, cluster.NewSDNIPAM(context.Background(), nil))
	assert.NotNil(t, cluster.NewSDNIPAM(context.Background(), &SDNIPAMOptions{IPAM: "x"}))
}

func TestCluster_SDNIPAMsFiltered(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	ipams, err := cluster.SDNIPAMs(context.Background(), "pve")
	assert.Nil(t, err)
	assert.Len(t, ipams, 1)
}

func TestSDNIPAM_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	i := cluster.SDNIPAM("pve")
	assert.Nil(t, i.Update(context.Background(), nil))
}

func TestSDNIPAM_EmptyName_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	i := &SDNIPAM{client: mockClient()}
	assert.Error(t, i.Read(context.Background()))
	assert.Error(t, i.Update(context.Background(), &SDNIPAMOptions{}))
	assert.Error(t, i.Delete(context.Background()))
	_, err := i.Status(context.Background())
	assert.Error(t, err)
}

func TestSDNIPAM_Read_NotFound(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	i := cluster.SDNIPAM("missing")
	assert.Error(t, i.Read(context.Background()))
}
