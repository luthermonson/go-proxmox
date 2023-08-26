package proxmox

import (
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPools(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pools, err := client.Pools()
	assert.Nil(t, err)
	assert.Len(t, pools, 1)
}

func TestPoolGet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pool("test-pool")
	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		assert.Equal(t, "test-pool", pool.PoolID)
		assert.Equal(t, "Test pool", pool.Comment)
		assert.Len(t, pool.Members, 3)
	}
}

func TestPoolCreate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	err := client.NewPool("test-pool", "Test pool")
	assert.Nil(t, err)
}

func TestPoolUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pool("test-pool")

	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		err = pool.Update(&PoolUpdateOption{
			Comment:         "Test pool updated",
			Delete:          true,
			Storage:         "local-zfs",
			VirtualMachines: "100",
		})
		assert.Nil(t, err)
	}
}

func TestPoolDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pool("test-pool")

	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		err = pool.Delete()
		assert.Nil(t, err)
	}
}
