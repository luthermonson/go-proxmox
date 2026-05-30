package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNDNSList(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	dns, err := cluster.SDNDNSList(context.Background(), "")
	assert.Nil(t, err)
	assert.Len(t, dns, 1)
	assert.Equal(t, "pdns1", dns[0].DNS)
}

func TestSDNDNS_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	d := cluster.SDNDNS("pdns1")
	assert.Nil(t, d.Read(context.Background()))
	assert.Equal(t, "powerdns", d.Type)
	assert.Equal(t, 3600, d.TTL)

	assert.Nil(t, d.Update(context.Background(), &SDNDNSOptions{TTL: 7200}))
	assert.Nil(t, d.Delete(context.Background()))
}

func TestCluster_NewSDNDNS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNDNS(context.Background(), &SDNDNSOptions{DNS: "pdns2", Type: "powerdns", URL: "https://x", Key: "k"})
	assert.Nil(t, err)

	assert.NotNil(t, cluster.NewSDNDNS(context.Background(), nil))
	assert.NotNil(t, cluster.NewSDNDNS(context.Background(), &SDNDNSOptions{DNS: "x"}))
}

func TestCluster_SDNDNSListFiltered(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	dns, err := cluster.SDNDNSList(context.Background(), "powerdns")
	assert.Nil(t, err)
	assert.Len(t, dns, 1)
}

func TestSDNDNS_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	d := cluster.SDNDNS("pdns1")
	assert.Nil(t, d.Update(context.Background(), nil))
}

func TestSDNDNS_EmptyName_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	d := &SDNDNS{client: mockClient()}
	assert.Error(t, d.Read(context.Background()))
	assert.Error(t, d.Update(context.Background(), &SDNDNSOptions{}))
	assert.Error(t, d.Delete(context.Background()))
}

func TestSDNDNS_Read_NotFound(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	d := cluster.SDNDNS("missing")
	assert.Error(t, d.Read(context.Background()))
}
