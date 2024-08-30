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
	assert.Equal(t, 20, len(rs))

	// type param test
	rs, err = cluster.Resources(ctx, "node")
	assert.Equal(t, 1, len(rs))
}
