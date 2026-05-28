package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_BackupInfoSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subs, err := cluster.BackupInfoSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subs, "not-backed-up")
}

func TestCluster_GuestsNotInBackup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	guests, err := cluster.GuestsNotInBackup(context.Background())
	assert.Nil(t, err)
	assert.Len(t, guests, 1)
	assert.Equal(t, 100, guests[0].VMID)
	assert.Equal(t, "qemu", guests[0].Type)
}

func TestClusterBackup_IncludedVolumes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	backup, err := cluster.Backup(context.Background(), "backup-1")
	assert.Nil(t, err)
	assert.NotNil(t, backup)

	root, err := backup.IncludedVolumes(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, root)
	assert.Len(t, root.Children, 1)
	assert.Equal(t, 100, root.Children[0].ID)
	assert.Len(t, root.Children[0].Children, 1)

	// blank id guard
	bare := &ClusterBackup{}
	_, err = bare.IncludedVolumes(context.Background())
	assert.Error(t, err)
}
