package proxmox

import (
	"context"
	"testing"

	"github.com/h2non/gock"
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

// PVE only touches the description when the field is present, so clearing it
// must send an empty description rather than an empty body.
func TestVirtualMachineSnapshot_UpdateConfig_ClearsDescription(t *testing.T) {
	defer gock.Off()
	gock.New(TestURI).
		Put("^/nodes/node1/qemu/100/snapshot/snap1/config$").
		JSON(map[string]string{"description": ""}).
		Times(2).
		Reply(200).
		JSON(`{"data": null}`)

	ctx := context.Background()
	assert.Nil(t, vm100().Snapshot("snap1").UpdateConfig(ctx, nil))
	assert.Nil(t, vm100().Snapshot("snap1").UpdateConfig(ctx, &VirtualMachineSnapshotUpdateOptions{}))
	assert.True(t, gock.IsDone())
}
