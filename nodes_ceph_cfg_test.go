package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_CephCfg(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	entries, err := cephNode().CephCfg(context.Background())
	require.NoError(t, err)
	require.Len(t, entries, 3)
	names := []interface{}{entries[0]["name"], entries[1]["name"], entries[2]["name"]}
	assert.Contains(t, names, "db")
	assert.Contains(t, names, "raw")
	assert.Contains(t, names, "value")
}

func TestNode_CephCfgDB(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	entries, err := cephNode().CephCfgDB(context.Background())
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "global", entries[0].Section)
	assert.Equal(t, "auth-cluster-required", entries[0].Name)
	assert.Equal(t, "cephx", entries[0].Value)
	assert.Equal(t, "basic", entries[0].Level)
	assert.False(t, bool(entries[0].CanUpdateAtRuntime))

	assert.Equal(t, "osd", entries[1].Section)
	assert.Equal(t, "osd-pool-default-size", entries[1].Name)
	assert.Equal(t, "3", entries[1].Value)
	assert.True(t, bool(entries[1].CanUpdateAtRuntime))
}

func TestNode_CephCfgRaw(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	raw, err := cephNode().CephCfgRaw(context.Background())
	require.NoError(t, err)
	assert.Contains(t, raw, "[global]")
	assert.Contains(t, raw, "auth_cluster_required = cephx")
}

func TestNode_CephCfgValue(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	values, err := cephNode().CephCfgValue(context.Background(), "global:auth_cluster_required;osd:osd_pool_default_size")
	require.NoError(t, err)
	require.NotNil(t, values)
	assert.Equal(t, "cephx", values["global"]["auth-cluster-required"])
	assert.Equal(t, "3", values["osd"]["osd-pool-default-size"])

	_, err = cephNode().CephCfgValue(context.Background(), "")
	assert.Error(t, err)
}
