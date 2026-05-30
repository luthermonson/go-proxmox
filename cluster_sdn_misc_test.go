package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNIndex(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
}

func TestCluster_SDNLock(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	token, err := cluster.SDNLock(context.Background(), false)
	assert.Nil(t, err)
	assert.Equal(t, SDNLockToken("tok-abc123"), token)
}

func TestCluster_SDNReleaseLock(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.SDNReleaseLock(context.Background(), SDNLockToken("tok-abc123"), false)
	assert.Nil(t, err)

	err = cluster.SDNReleaseLock(context.Background(), "", true)
	assert.Nil(t, err)
}

func TestCluster_SDNRollback(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	assert.Nil(t, cluster.SDNRollback(context.Background(), SDNLockToken("tok-abc123"), true))
	assert.Nil(t, cluster.SDNRollback(context.Background(), "", false))
}

func TestCluster_SDNDryRun(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	diff, err := cluster.SDNDryRun(context.Background(), "node1")
	assert.Nil(t, err)
	assert.Contains(t, diff.FRRDiff, "bgp 65000")
	assert.Contains(t, diff.InterfacesDiff, "vnet1")

	_, err = cluster.SDNDryRun(context.Background(), "")
	assert.NotNil(t, err)
}

func TestCluster_SDNLock_AllowPending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	token, err := cluster.SDNLock(context.Background(), true)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)
}
