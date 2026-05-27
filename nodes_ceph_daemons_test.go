package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNode_CephMons(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	mons, err := cephNode().CephMons(context.Background())
	assert.Nil(t, err)
	assert.Len(t, mons, 2)
	assert.Equal(t, "node1", mons[0].Name)
	assert.Equal(t, "running", mons[0].State)
	assert.Equal(t, 0, mons[0].Rank)
}

func TestNode_CreateCephMon(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CreateCephMon(context.Background(), "node1", &CephMonOptions{MonAddress: "10.0.0.1"})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "cephcreatemon", task.Type)

	// empty monid defaults to nodename
	task, err = cephNode().CreateCephMon(context.Background(), "", nil)
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatemon", task.Type)
}

func TestNode_DeleteCephMon(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CephMon("node1").Delete(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroymon", task.Type)

	// empty monid → error
	_, err = cephNode().CephMon("").Delete(context.Background())
	assert.NotNil(t, err)
}

func TestNode_CephMgrs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	mgrs, err := cephNode().CephMgrs(context.Background())
	assert.Nil(t, err)
	assert.Len(t, mgrs, 1)
	assert.Equal(t, "node1", mgrs[0].Name)
	assert.Equal(t, "active", mgrs[0].State)
}

func TestNode_CreateCephMgr(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CreateCephMgr(context.Background(), "node1")
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatemgr", task.Type)

	// empty id defaults to nodename
	task, err = cephNode().CreateCephMgr(context.Background(), "")
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatemgr", task.Type)
}

func TestNode_DeleteCephMgr(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CephMgr("node1").Delete(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroymgr", task.Type)

	_, err = cephNode().CephMgr("").Delete(context.Background())
	assert.NotNil(t, err)
}

func TestNode_CephMDSs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	mdss, err := cephNode().CephMDSs(context.Background())
	assert.Nil(t, err)
	assert.Len(t, mdss, 2)
	assert.Equal(t, "node1", mdss[0].Name)
	assert.Equal(t, "up:active", mdss[0].State)
	assert.Equal(t, "cephfs", mdss[0].FSName)
}

func TestNode_CreateCephMDS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CreateCephMDS(context.Background(), "node1", &CephMDSOptions{HotStandby: true})
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatemds", task.Type)

	// empty name defaults to nodename
	task, err = cephNode().CreateCephMDS(context.Background(), "", nil)
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatemds", task.Type)
}

func TestNode_DeleteCephMDS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephNode().CephMDS("node1").Delete(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroymds", task.Type)

	_, err = cephNode().CephMDS("").Delete(context.Background())
	assert.NotNil(t, err)
}
