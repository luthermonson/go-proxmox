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

func TestNode_NewNetwork(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	n := &Node{client: mockClient(), Name: "node1"}
	nw := &NodeNetwork{Iface: "vmbr99", Type: "bridge"}

	task, err := n.NewNetwork(context.Background(), nw)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	// Wiring side-effects from NewNetwork are applied to the input struct.
	assert.Equal(t, "node1", nw.Node)
	assert.NotNil(t, nw.client)
	assert.Equal(t, n, nw.NodeAPI)
}

func TestNode_NetworkReload(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	n := &Node{client: mockClient(), Name: "node1"}
	task, err := n.NetworkReload(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestNodeNetwork_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// No-op when Iface is empty — must not issue an HTTP call.
	empty := &NodeNetwork{client: mockClient(), Node: "node1"}
	assert.Nil(t, empty.Update(context.Background()))

	nw := &NodeNetwork{
		client: mockClient(),
		Node:   "node1",
		Iface:  "vmbr99",
		Type:   "bridge",
	}
	assert.Nil(t, nw.Update(context.Background()))
}

func TestNodeNetwork_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	n := &Node{client: mockClient(), Name: "node1"}

	// Empty Iface returns zero values without a request.
	empty := &NodeNetwork{client: mockClient(), Node: "node1", NodeAPI: n}
	task, err := empty.Delete(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, task)

	nw := &NodeNetwork{
		client:  mockClient(),
		Node:    "node1",
		NodeAPI: n,
		Iface:   "vmbr99",
	}
	task, err = nw.Delete(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, task) // task comes from NetworkReload post-delete.
}

func TestNode_IPAM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	n := &Node{client: mockClient(), Name: "node1"}
	ipam, err := n.IPAM(context.Background())
	assert.Nil(t, err)
	assert.Len(t, ipam, 2)
	assert.Equal(t, "vm100", ipam[0].Hostname)
	assert.Equal(t, "10.0.0.10", ipam[0].IP)
}
