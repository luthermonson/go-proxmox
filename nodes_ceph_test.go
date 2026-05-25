package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cephNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_CephIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	entries, err := cephNode().CephIndex(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Subdir)
	}
	assert.Contains(t, names, "osd")
	assert.Contains(t, names, "status")
	assert.Contains(t, names, "log")
}

func TestNode_InitCeph(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// nil opts — accept all PVE defaults
	task, err := cephNode().InitCeph(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "cephcreate", task.Type)

	// with opts — exercise the body-building branches
	task, err = cephNode().InitCeph(context.Background(), &CephInitOptions{
		Network:        "10.10.10.0/24",
		ClusterNetwork: "10.10.20.0/24",
		Size:           3,
		MinSize:        2,
		PGBits:         6,
		DisableCephx:   true,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "cephcreate", task.Type)
}

func TestNode_StartCeph(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	task, err := cephNode().StartCeph(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "srvstart", task.Type)

	// explicit service= form value
	task, err = cephNode().StartCeph(context.Background(), "osd.0")
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestNode_StopCeph(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	task, err := cephNode().StopCeph(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "srvstop", task.Type)
}

func TestNode_RestartCeph(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	task, err := cephNode().RestartCeph(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "srvrestart", task.Type)
}

func TestNode_CephStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	status, err := cephNode().CephStatus(context.Background())
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "d11c6ea1-7ab2-41fa-99c5-b85f4d7ffd49", status.Fsid)
	assert.Equal(t, "HEALTH_OK", status.Health.Status)
	assert.Contains(t, status.QuorumNames, "node1")
}

func TestNode_CephLog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	entries, err := cephNode().CephLog(context.Background(), 0, 0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, 1, entries[0].N)
	assert.Contains(t, entries[0].T, "osdmap")

	// non-zero start/limit — exercise the query-string branches
	entries, err = cephNode().CephLog(context.Background(), 100, 50)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
}

func TestNode_CephCrush(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	crush, err := cephNode().CephCrush(context.Background())
	require.NoError(t, err)
	assert.Contains(t, crush, "begin crush map")
	assert.Contains(t, crush, "replicated_rule")
}

func TestNode_CephRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	rules, err := cephNode().CephRules(context.Background())
	require.NoError(t, err)
	require.Len(t, rules, 2)
	assert.Equal(t, "replicated_rule", rules[0].Name)
	assert.Equal(t, "erasure-code", rules[1].Name)
}

func TestNode_CephCmdSafety(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	safety, err := cephNode().CephCmdSafety(context.Background(), "osd", "3", "stop")
	require.NoError(t, err)
	require.NotNil(t, safety)
	assert.True(t, safety.Safe)

	// validation — empty args must be rejected client-side
	_, err = cephNode().CephCmdSafety(context.Background(), "", "3", "stop")
	assert.Error(t, err)
	_, err = cephNode().CephCmdSafety(context.Background(), "osd", "", "stop")
	assert.Error(t, err)
	_, err = cephNode().CephCmdSafety(context.Background(), "osd", "3", "")
	assert.Error(t, err)
}
