package proxmox

import (
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_Nodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	nodes, err := client.Nodes()
	assert.Nil(t, err)
	for _, n := range nodes {
		assert.Contains(t, n.Node, "node")
		assert.Equal(t, n.Type, "node")
	}
	//assert.Equal(t, 6, len(testData))
}

func TestClient_Node(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	node, err := client.Node("node1")
	assert.Nil(t, err)
	assert.Equal(t, "node1", node.Name)
	assert.NotNil(t, node.client)

	v, err := node.Version()
	assert.Nil(t, err)
	assert.Equal(t, "7.4", v.Release)

	node, err = client.Node("doesntexist")
	assert.NotNil(t, err)
	assert.Nil(t, node)
}
