package proxmox

import (
	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPoolList(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	_, err := client.Pools().List()
	assert.Nil(t, err)
}

func TestPoolGet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pools().Get("test-pool")
	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		assert.Equal(t, "test-pool", pool.PoolID)
		assert.Equal(t, "Test pool", pool.Comment)
		assert.Len(t, pool.Members, 2)
	}
}

func TestPoolCreate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	err := client.Pools().Create(&PoolCreateOption{
		PoolID:  "test-pool",
		Comment: "Test pool",
	})
	assert.Nil(t, err)
}

func TestPoolUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pools().Get("test-pool")

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

	pool, err := client.Pools().Get("test-pool")

	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		err = pool.Delete()
		assert.Nil(t, err)
	}
}
