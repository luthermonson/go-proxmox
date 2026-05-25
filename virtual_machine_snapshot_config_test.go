package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVirtualMachineSnapshot_Config(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cfg, err := vm100().Snapshot("snap1").Config(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "Before upgrade", cfg["description"])
	assert.Equal(t, "snap0", cfg["parent"])
}

func TestVirtualMachineSnapshot_UpdateConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := vm100().Snapshot("snap1").UpdateConfig(context.Background(), &VirtualMachineSnapshotUpdateOptions{Description: "renamed"})
	assert.Nil(t, err)
	// nil options path
	assert.Nil(t, vm100().Snapshot("snap1").UpdateConfig(context.Background(), nil))
}
