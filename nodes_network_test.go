package proxmox

import (
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNetwork(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	node := Node{
		client: client,
		Name:   "node1",
	}

	network, err := node.Network("vmbr0")
	assert.Nil(t, err)
	assert.Equal(t, network.Iface, "vmbr0")
}

func TestNetworks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	node := Node{
		client: client,
		Name:   "node1",
	}

	networks, err := node.Networks()
	assert.Nil(t, err)
	assert.Len(t, networks, 2)
}
