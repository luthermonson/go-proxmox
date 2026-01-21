package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNetwork(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}

	network, err := node.Network(ctx, "vmbr0")
	assert.Nil(t, err)
	assert.Equal(t, network.Iface, "vmbr0")
}

func TestNode1Networks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}

	networks, err := node.Networks(ctx)
	assert.Nil(t, err)
	assert.Len(t, networks, 5)
}

func TestNode2Networks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node2",
	}

	networks, err := node.Networks(ctx)
	assert.Nil(t, err)
	assert.Len(t, networks, 2)
}

func TestNetworksPve8(t *testing.T) {
	mocks.ProxmoxVE8x(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}

	networks, err := node.Networks(ctx)
	assert.Nil(t, err)
	assert.Len(t, networks, 5)
}

func TestNetworksPve8NetworksOfType(t *testing.T) {
	mocks.ProxmoxVE8x(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}

	networks, err := node.Networks(ctx, "any_bridge")
	assert.Nil(t, err)
	assert.Len(t, networks, 1)

	_, err = node.Networks(ctx, "any_bridge", "second_argument")
	assert.NotNil(t, err)
}
