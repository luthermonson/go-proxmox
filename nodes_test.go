package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_Nodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	nodes, err := client.Nodes(ctx)
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
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)
	assert.Equal(t, "node1", node.Name)
	assert.NotNil(t, node.client)

	v, err := node.Version(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)

	node, err = client.Node(ctx, "doesntexist")
	assert.NotNil(t, err)
}
