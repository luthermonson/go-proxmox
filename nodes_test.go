// +build nodes

package proxmox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodes(t *testing.T) {
	client := ClientFromLogins()
	nodes, err := client.Nodes()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)
	for _, n := range nodes {
		assert.NotEmpty(t, n.Node)
		var node *Node
		t.Run("get status for node "+n.Node, func(t *testing.T) {
			var err error
			node, err = client.Node(n.Node)
			assert.Nil(t, err)
			assert.Equal(t, n.MaxMem, node.Memory.Total)
			assert.Equal(t, n.Disk, node.RootFS.Used)
		})

		t.Run("get VMs for node "+n.Node, func(t *testing.T) {
			_, err := node.VirtualMachines()
			assert.Nil(t, err)
		})

		break // only pull status from one node
	}

	_, err = client.Node("doesnt-exist")
	assert.Contains(t, err.Error(), "500 hostname lookup 'doesnt-exist' failed - failed to get address info for: doesnt-exist:")
}

func TestNode(t *testing.T) {
	client := ClientFromLogins()
	n, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, n.Name, nodeName)
}

func TestContainers(t *testing.T) {
	client := ClientFromLogins()
	nodes, err := client.Nodes()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)
	for _, n := range nodes {
		assert.NotEmpty(t, n.Node)
		var node *Node
		t.Run("get status for node "+n.Node, func(t *testing.T) {
			var err error
			node, err = client.Node(n.Node)
			assert.Nil(t, err)
		})

		t.Run("get Containers for node "+n.Node, func(t *testing.T) {
			_, err := node.Containers()
			assert.Nil(t, err)
		})

		break // only pull status from one node
	}

	_, err = client.Node("doesnt-exist")
	assert.Contains(t, err.Error(), "500 hostname lookup 'doesnt-exist' failed - failed to get address info for: doesnt-exist:")
}

func TestNode_AvailableContainerTemplates(t *testing.T) {
	client := ClientFromLogins()
	nodes, err := client.Nodes()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)
	for _, n := range nodes {
		assert.NotEmpty(t, n.Node)
		var node *Node
		t.Run("get status for node "+n.Node, func(t *testing.T) {
			var err error
			node, err = client.Node(n.Node)
			assert.Nil(t, err)
		})

		t.Run("get Containers for node "+n.Node, func(t *testing.T) {
			aplinfos, err := node.AvailableContainerTemplates()
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(aplinfos), 1)
		})

		break // only pull status from one node
	}
}

func TestNode_ContainerTemplates(t *testing.T) {
	client := ClientFromLogins()
	nodes, err := client.Nodes()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)
	for _, n := range nodes {
		assert.NotEmpty(t, n.Node)
		var node *Node
		t.Run("get status for node "+n.Node, func(t *testing.T) {
			var err error
			node, err = client.Node(n.Node)
			assert.Nil(t, err)
		})

		t.Run("get Containers for node "+n.Node, func(t *testing.T) {
			aplinfos, err := node.AvailableContainerTemplates()
			assert.Nil(t, err)
			assert.GreaterOrEqual(t, len(aplinfos), 1)
		})

		break // only pull status from one node
	}
}
