// +build nodes

package proxmox

import (
	"fmt"
	"strings"
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
	node, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, nodeName)
}

func TestNode_Storage(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, nodeName)

	//storages, err := n.Storages()
	//assert.Nil(t, err)
	//fmt.Println(storages)
}

func TestContainers(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, nodeName)

	t.Run("get Containers for node "+node.Name, func(t *testing.T) {
		_, err := node.Containers()
		assert.Nil(t, err)
	})
}

func TestNode_Appliances(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, nodeName)

	t.Run("get Containers for node "+node.Name, func(t *testing.T) {
		aplinfos, err := node.Appliances()
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(aplinfos), 1)
	})
}

func TestNode_DownloadAppliance(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, nodeName)

	var aplinfos Appliances
	t.Run("get Containers for node "+node.Name, func(t *testing.T) {
		aplinfos, err = node.Appliances()
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(aplinfos), 1)
	})

	t.Run("download appliance "+applianceName, func(t *testing.T) {
		_, err := node.DownloadAppliance("doesnt-exist", nodeStorage)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such template"))
	})

	t.Run("download appliance "+applianceName, func(t *testing.T) {
		for _, a := range aplinfos {
			if a.Template == applianceName {
				ret, err := node.DownloadAppliance(a.Template, nodeStorage)
				assert.Nil(t, err)
				fmt.Println(ret)
				assert.True(t, strings.HasPrefix("UPID:"+node.Name, ret))
			}
		}
	})
}
