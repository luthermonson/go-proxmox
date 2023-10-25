package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPools(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pools, err := client.Pools(context.Background())
	assert.Nil(t, err)
	assert.Len(t, pools, 1)
}

func TestPoolGet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	pool, err := client.Pool(context.Background(), "test-pool")
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

	err := client.NewPool(context.Background(), "test-pool", "Test pool")
	assert.Nil(t, err)
}

func TestPoolUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	pool, err := client.Pool(ctx, "test-pool")

	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		err = pool.Update(ctx, &PoolUpdateOption{
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
	ctx := context.Background()

	pool, err := client.Pool(ctx, "test-pool")

	assert.Nil(t, err)
	assert.NotNil(t, pool)
	if pool != nil {
		err = pool.Delete(ctx)
		assert.Nil(t, err)
	}
}
