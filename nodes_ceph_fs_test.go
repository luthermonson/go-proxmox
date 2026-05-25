package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_CephFSs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	fss, err := cephNode().CephFSs(context.Background())
	require.NoError(t, err)
	require.Len(t, fss, 1)
	assert.Equal(t, "cephfs", fss[0].Name)
	assert.Equal(t, "cephfs_metadata", fss[0].MetadataPool)
	assert.Equal(t, 7, fss[0].MetadataPoolID)
	assert.Equal(t, "cephfs_data", fss[0].DataPool)
	assert.Equal(t, []string{"cephfs_data", "cephfs_data_ec"}, fss[0].DataPools)
	assert.Equal(t, []int{8, 9}, fss[0].DataPoolIDs)
}

func TestNode_CreateCephFS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	task, err := cephNode().CreateCephFS(context.Background(), "cephfs", &CephFSOptions{PgNum: 128, AddStorage: true})
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "cephfs-create", task.Type)

	// nil opts is fine — PVE applies its own defaults.
	task, err = cephNode().CreateCephFS(context.Background(), "cephfs", nil)
	require.NoError(t, err)
	assert.NotNil(t, task)

	_, err = cephNode().CreateCephFS(context.Background(), "", nil)
	assert.Error(t, err)
}

func TestNode_DeleteCephFS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	task, err := cephNode().CephFS("cephfs").Delete(context.Background(), true, true)
	require.NoError(t, err)
	assert.Equal(t, "cephfs-destroy", task.Type)

	// no cleanup flags
	task, err = cephNode().CephFS("cephfs").Delete(context.Background(), false, false)
	require.NoError(t, err)
	assert.NotNil(t, task)

	_, err = cephNode().CephFS("").Delete(context.Background(), false, false)
	assert.Error(t, err)
}
