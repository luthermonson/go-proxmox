package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVirtualMachine_GetSnapshotConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cfg, err := vm100().GetSnapshotConfig(context.Background(), "snap1")
	assert.Nil(t, err)
	assert.Equal(t, "Before upgrade", cfg["description"])
	assert.Equal(t, "snap0", cfg["parent"])
}

func TestVirtualMachine_UpdateSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := vm100().UpdateSnapshot(context.Background(), "snap1", &VirtualMachineSnapshotUpdateOptions{Description: "renamed"})
	assert.Nil(t, err)
	// nil options path
	assert.Nil(t, vm100().UpdateSnapshot(context.Background(), "snap1", nil))
}
