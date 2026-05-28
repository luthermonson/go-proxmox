package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_BulkActionSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subs, err := cluster.BulkActionSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subs, "guest")
}

func TestCluster_BulkActionGuestSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subs, err := cluster.BulkActionGuestSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subs, "start")
	assert.Contains(t, subs, "shutdown")
	assert.Contains(t, subs, "migrate")
	assert.Contains(t, subs, "suspend")
}

func TestCluster_BulkStartGuests(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	task, err := cluster.BulkStartGuests(context.Background(), &BulkStartOptions{VMIDs: []int{100, 101}})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
}

func TestCluster_BulkShutdownGuests(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	task, err := cluster.BulkShutdownGuests(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestCluster_BulkSuspendGuests(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	task, err := cluster.BulkSuspendGuests(context.Background(), &BulkSuspendOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestCluster_BulkMigrateGuests(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	task, err := cluster.BulkMigrateGuests(context.Background(), &BulkMigrateOptions{Target: "node2"})
	assert.Nil(t, err)
	assert.NotNil(t, task)
}
