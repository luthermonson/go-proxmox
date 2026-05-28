package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_CephFlags(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	flags, err := cluster.CephFlags(context.Background())
	assert.Nil(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "noout", flags[0].Name)
	assert.True(t, flags[0].Value)
}

func TestCluster_CephFlag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, err := cluster.CephFlag(context.Background(), "noout")
	assert.Nil(t, err)
	assert.NotEmpty(t, v)

	_, err = cluster.CephFlag(context.Background(), "")
	assert.Error(t, err)
}

func TestCluster_CephMetadata(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	meta, err := cluster.CephMetadata(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, meta)
	assert.NotNil(t, meta.Version)
	assert.Equal(t, "18.2.0", meta.Version.Version)
	assert.Len(t, meta.OSD, 1)
}

func TestCluster_SetCephFlags(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	noout := true
	task, err := cluster.SetCephFlags(context.Background(), &CephFlagsUpdateOptions{NoOut: &noout})
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestCluster_SetCephFlag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	assert.Nil(t, cluster.SetCephFlag(context.Background(), "noout", true))
	assert.Error(t, cluster.SetCephFlag(context.Background(), "", true))
}
