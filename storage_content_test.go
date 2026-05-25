package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func storageRef(name string) *Storage {
	return &Storage{client: mockClient(), Node: "node1", Name: name}
}

func TestStorage_AllocContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	volid, err := storageRef("local-lvm").AllocContent(context.Background(), &StorageContentAllocOptions{
		Filename: "vm-100-disk-1",
		Size:     "4G",
		VMID:     100,
	})
	assert.Nil(t, err)
	assert.Equal(t, "local-lvm:vm-100-disk-1", volid)

	_, err = storageRef("local-lvm").AllocContent(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestStorage_UpdateContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	protected := true
	err := storageRef("local").UpdateContent(context.Background(),
		"local:backup/vzdump-qemu-100-2026_01_01-12_00_00.vma.zst",
		&StorageContentUpdateOptions{Notes: "keep — golden image", Protected: &protected})
	assert.Nil(t, err)

	err = storageRef("local").UpdateContent(context.Background(), "", nil)
	assert.NotNil(t, err)
}

func TestStorage_CopyContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := storageRef("local-lvm").CopyContent(context.Background(),
		"local-lvm:vm-100-disk-0", "local-lvm:vm-200-disk-0", "")
	assert.Nil(t, err)
	assert.Equal(t, "imgcopy", task.Type)

	_, err = storageRef("local-lvm").CopyContent(context.Background(), "", "x", "")
	assert.NotNil(t, err)
}

func TestStorage_OCIRegistryPull(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := storageRef("local").OCIRegistryPull(context.Background(), "docker.io/library/alpine:latest", "")
	assert.Nil(t, err)
	assert.Equal(t, "ocipull", task.Type)

	_, err = storageRef("local").OCIRegistryPull(context.Background(), "", "")
	assert.NotNil(t, err)
}

func TestStorage_FileRestoreList(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := storageRef("pbs").FileRestoreList(context.Background(), "pbs:backup/vm/100/2026-05-13T22:00:00Z", "/")
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "f", entries[0].Type)

	_, err = storageRef("pbs").FileRestoreList(context.Background(), "", "/")
	assert.NotNil(t, err)
}

func TestStorage_RRDData(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	data, err := storageRef("local").RRDData(context.Background(), TimeframeHour, "")
	assert.Nil(t, err)
	assert.Len(t, data, 2)
}
