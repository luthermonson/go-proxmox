package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_DNS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	dns, err := node.DNS(context.Background())
	require.NoError(t, err)
	require.NotNil(t, dns)
	assert.Equal(t, "example.com", dns.Search)
	assert.Equal(t, "1.1.1.1", dns.DNS1)
	assert.Equal(t, "8.8.8.8", dns.DNS2)
	assert.Equal(t, "9.9.9.9", dns.DNS3)
}

func TestNode_UpdateDNS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	err := node.UpdateDNS(context.Background(), &NodeDNS{
		Search: "example.com",
		DNS1:   "1.1.1.1",
	})
	assert.NoError(t, err)
}
