package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func cephPoolNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestNode_CephPools(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pools, err := cephPoolNode().CephPools(context.Background())
	assert.Nil(t, err)
	assert.Len(t, pools, 2)
	assert.Equal(t, "rbd", pools[0].PoolName)
	assert.Equal(t, "replicated", pools[0].Type)
	assert.Equal(t, 3, pools[0].Size)
	assert.Equal(t, 2, pools[0].MinSize)
	assert.Equal(t, "on", pools[0].PgAutoscaleMode)
	assert.Equal(t, "replicated_rule", pools[0].CrushRuleName)
	assert.Equal(t, "cephfs_metadata", pools[1].PoolName)
}

func TestNode_CreateCephPool(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// happy path: minimal pool
	task, err := cephPoolNode().CreateCephPool(context.Background(), &CephPoolOptions{
		Name:            "rbd",
		Size:            intPtr(3),
		MinSize:         intPtr(2),
		Application:     "rbd",
		PgAutoscaleMode: "on",
	})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "cephcreatepool", task.Type)

	// happy path: erasure-coded pool — the EC config should be serialized.
	task, err = cephPoolNode().CreateCephPool(context.Background(), &CephPoolOptions{
		Name:        "ec-rbd",
		Application: "rbd",
		ErasureCoding: &CephPoolErasureCoding{
			K:             4,
			M:             2,
			FailureDomain: "host",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "cephcreatepool", task.Type)

	// nil opts must error before any HTTP call
	_, err = cephPoolNode().CreateCephPool(context.Background(), nil)
	assert.NotNil(t, err)

	// missing name must error
	_, err = cephPoolNode().CreateCephPool(context.Background(), &CephPoolOptions{})
	assert.NotNil(t, err)
}

func TestCephPool_SubResources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := cephPoolNode().CephPool("rbd").SubResources(context.Background())
	assert.Nil(t, err)
	assert.Len(t, subdirs, 1)
	assert.Equal(t, "status", subdirs[0].Subdir)

	_, err = cephPoolNode().CephPool("").SubResources(context.Background())
	assert.NotNil(t, err)
}

func TestCephPool_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephPoolNode().CephPool("rbd").Update(context.Background(), &CephPoolOptions{
		Size:            intPtr(4),
		PgAutoscaleMode: "warn",
	})
	assert.Nil(t, err)
	assert.Equal(t, "cephsetpool", task.Type)

	_, err = cephPoolNode().CephPool("").Update(context.Background(), &CephPoolOptions{})
	assert.NotNil(t, err)

	_, err = cephPoolNode().CephPool("rbd").Update(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestCephPool_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// default: removeECProfile=true matches PVE default so no override is sent
	task, err := cephPoolNode().CephPool("rbd").Delete(context.Background(), true, true, true)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroypool", task.Type)

	// force=false, removeStorages=false, removeECProfile=false should still hit
	task, err = cephPoolNode().CephPool("rbd").Delete(context.Background(), false, false, false)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroypool", task.Type)

	_, err = cephPoolNode().CephPool("").Delete(context.Background(), false, false, true)
	assert.NotNil(t, err)
}

func TestCephPool_Status(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	status, err := cephPoolNode().CephPool("rbd").Status(context.Background(), true)
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "rbd", status.Name)
	assert.Equal(t, 1, status.ID)
	assert.Equal(t, 3, status.Size)
	assert.Equal(t, 2, status.MinSize)
	assert.Equal(t, "on", status.PgAutoscaleMode)
	assert.Equal(t, "rbd", status.Application)
	assert.Contains(t, status.ApplicationList, "rbd")
	assert.True(t, status.HashPSPool)
	assert.True(t, status.UseGMTHitset)
	assert.NotNil(t, status.Statistics)

	// also verify the non-verbose path is callable
	_, err = cephPoolNode().CephPool("rbd").Status(context.Background(), false)
	assert.Nil(t, err)

	_, err = cephPoolNode().CephPool("").Status(context.Background(), false)
	assert.NotNil(t, err)
}

func TestCephPoolErasureCoding_String(t *testing.T) {
	// nil receiver returns empty string for safety
	var nilEC *CephPoolErasureCoding
	assert.Equal(t, "", nilEC.String())

	// k/m only
	ec := &CephPoolErasureCoding{K: 4, M: 2}
	assert.Equal(t, "k=4,m=2", ec.String())

	// full options in canonical order
	ec = &CephPoolErasureCoding{
		K:             4,
		M:             2,
		DeviceClass:   "ssd",
		FailureDomain: "host",
		Profile:       "ec42",
	}
	assert.Equal(t, "k=4,m=2,device-class=ssd,failure-domain=host,profile=ec42", ec.String())
}

func TestCephPoolOptions_OmitsUnsetPointerFields(t *testing.T) {
	// Regression test mirroring AGENTS.md: pointer fields must drop out of the
	// marshalled body when left nil so PVE server-side defaults survive.
	body, err := cephPoolBody(&CephPoolOptions{Name: "rbd"})
	assert.Nil(t, err)
	assert.Equal(t, "rbd", body["name"])
	_, hasSize := body["size"]
	assert.False(t, hasSize, "size should not be present when nil")
	_, hasMinSize := body["min_size"]
	assert.False(t, hasMinSize, "min_size should not be present when nil")
	_, hasPgNum := body["pg_num"]
	assert.False(t, hasPgNum, "pg_num should not be present when nil")
	_, hasAddStorages := body["add_storages"]
	assert.False(t, hasAddStorages, "add_storages should not be present when nil")

	// And present when explicitly set, even to a zero value.
	body, err = cephPoolBody(&CephPoolOptions{Name: "rbd", Size: intPtr(0), AddStorages: boolPtr(false)})
	assert.Nil(t, err)
	assert.EqualValues(t, 0, body["size"])
	assert.Equal(t, false, body["add_storages"])
}
