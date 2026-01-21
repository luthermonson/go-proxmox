package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClusterStorages(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storages, err := client.ClusterStorages(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storages)
	assert.Len(t, storages, 3)

	// Verify local storage
	assert.Equal(t, "local", storages[0].Storage)
	assert.Equal(t, "dir", storages[0].Type)
	assert.Equal(t, "vztmpl,iso,backup", storages[0].Content)
	assert.Equal(t, 0, storages[0].Shared)

	// Verify local-lvm storage
	assert.Equal(t, "local-lvm", storages[1].Storage)
	assert.Equal(t, "lvmthin", storages[1].Type)
	assert.Equal(t, "images,rootdir", storages[1].Content)
	assert.Equal(t, "data", storages[1].Thinpool)
	assert.Equal(t, "pve", storages[1].VgName)

	// Verify nfs storage
	assert.Equal(t, "nfs-storage", storages[2].Storage)
	assert.Equal(t, "nfs", storages[2].Type)
	assert.Equal(t, 1, storages[2].Shared)
	assert.Equal(t, "node1,node2", storages[2].Nodes)
}

func TestClusterStorage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storage, err := client.ClusterStorage(ctx, "local")
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "local", storage.Storage)
	assert.Equal(t, "dir", storage.Type)
	assert.Equal(t, "vztmpl,iso,backup", storage.Content)
	assert.Equal(t, "/var/lib/vz", storage.Path)
	assert.Equal(t, 0, storage.Shared)
}

func TestClusterStorage_LVM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storage, err := client.ClusterStorage(ctx, "local-lvm")
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "local-lvm", storage.Storage)
	assert.Equal(t, "lvmthin", storage.Type)
	assert.Equal(t, "data", storage.Thinpool)
	assert.Equal(t, "pve", storage.VgName)
}

func TestNewClusterStorage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task, err := client.NewClusterStorage(ctx,
		ClusterStorageOptions{Name: "storage", Value: "test-storage"},
		ClusterStorageOptions{Name: "type", Value: "dir"},
		ClusterStorageOptions{Name: "path", Value: "/mnt/test"},
		ClusterStorageOptions{Name: "content", Value: "iso,vztmpl"},
	)
	assert.Nil(t, err)
	assert.Nil(t, task) // Task is nil for successful operations with null data
}

func TestUpdateClusterStorage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task, err := client.UpdateClusterStorage(ctx, "local",
		ClusterStorageOptions{Name: "content", Value: "vztmpl,iso,backup,snippets"},
	)
	assert.Nil(t, err)
	assert.Nil(t, task)
}

func TestDeleteClusterStorage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task, err := client.DeleteClusterStorage(ctx, "test-storage")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "storage", task.Type)
}

func TestStorage_GetContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storage := &Storage{
		client: client,
		Node:   "node1",
		Name:   "local",
	}

	content, err := storage.GetContent(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, content)
	assert.Len(t, content, 3)

	// Verify ISO content
	assert.Equal(t, "local:iso/debian-12.0.0-amd64-netinst.iso", content[0].Volid)
	assert.Equal(t, "iso", content[0].Format)
	assert.Equal(t, uint64(654311424), content[0].Size)

	// Verify vztmpl content
	assert.Equal(t, "local:vztmpl/debian-12-standard_12.0-1_amd64.tar.zst", content[1].Volid)
	assert.Equal(t, "tar.zst", content[1].Format)
	assert.Equal(t, uint64(128974848), content[1].Size)

	// Verify backup content
	assert.Equal(t, "local:backup/vzdump-qemu-100-2023_08_28-12_00_00.vma.zst", content[2].Volid)
	assert.Equal(t, "vma.zst", content[2].Format)
	assert.Equal(t, uint64(2147483648), content[2].Size)
	assert.Equal(t, uint64(100), content[2].VMID)
}

func TestStorage_DeleteContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storage := &Storage{
		client: client,
		Node:   "node1",
		Name:   "local",
	}

	task, err := storage.DeleteContent(ctx, "local:iso/test.iso")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "imgdel", task.Type)
}

func TestStorage_DownloadURL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	storage := &Storage{
		client: client,
		Node:   "node1",
		Name:   "local",
	}

	task, err := storage.DownloadURL(ctx, "iso", "debian-12.iso", "https://example.com/debian-12.iso")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "download", task.Type)
}

func TestStorage_UnmarshalJSON_LargeValues(t *testing.T) {
	// Test handling of large storage values (>1PB) that come back as floats in scientific notation
	tests := []struct {
		name     string
		json     string
		expected Storage
	}{
		{
			name: "Large total value in scientific notation",
			json: `{
				"storage": "large-storage",
				"enabled": 1,
				"active": 1,
				"total": 1.12589990684262e+15,
				"used": 5.5e+14,
				"avail": 5.7589990684262e+14,
				"type": "dir",
				"shared": 0
			}`,
			expected: Storage{
				Name:    "large-storage",
				Storage: "large-storage",
				Enabled: 1,
				Active:  1,
				Total:   uint64(1125899906842620),
				Used:    uint64(550000000000000),
				Avail:   uint64(575899906842620),
				Type:    "dir",
				Shared:  0,
			},
		},
		{
			name: "Normal integer values",
			json: `{
				"storage": "normal-storage",
				"enabled": 1,
				"active": 1,
				"total": 1000000000,
				"used": 500000000,
				"avail": 500000000,
				"type": "lvm",
				"shared": 1
			}`,
			expected: Storage{
				Name:    "normal-storage",
				Storage: "normal-storage",
				Enabled: 1,
				Active:  1,
				Total:   uint64(1000000000),
				Used:    uint64(500000000),
				Avail:   uint64(500000000),
				Type:    "lvm",
				Shared:  1,
			},
		},
		{
			name: "UsedFraction as float",
			json: `{
				"storage": "frac-storage",
				"enabled": 1,
				"active": 1,
				"total": 1000000,
				"used": 750000,
				"avail": 250000,
				"used_fraction": 0.75,
				"type": "zfs",
				"shared": 0
			}`,
			expected: Storage{
				Name:         "frac-storage",
				Storage:      "frac-storage",
				Enabled:      1,
				Active:       1,
				Total:        uint64(1000000),
				Used:         uint64(750000),
				Avail:        uint64(250000),
				UsedFraction: 0.75,
				Type:         "zfs",
				Shared:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storage Storage
			err := storage.UnmarshalJSON([]byte(tt.json))
			assert.Nil(t, err)
			assert.Equal(t, tt.expected.Name, storage.Name)
			assert.Equal(t, tt.expected.Storage, storage.Storage)
			assert.Equal(t, tt.expected.Enabled, storage.Enabled)
			assert.Equal(t, tt.expected.Active, storage.Active)
			assert.Equal(t, tt.expected.Total, storage.Total)
			assert.Equal(t, tt.expected.Used, storage.Used)
			assert.Equal(t, tt.expected.Avail, storage.Avail)
			assert.Equal(t, tt.expected.UsedFraction, storage.UsedFraction)
			assert.Equal(t, tt.expected.Type, storage.Type)
			assert.Equal(t, tt.expected.Shared, storage.Shared)
		})
	}
}

func TestStorages_UnmarshalJSON(t *testing.T) {
	// Test that Storages slice unmarshaling works correctly
	json := `[
		{
			"storage": "storage1",
			"enabled": 1,
			"active": 1,
			"total": 1.5e+15,
			"type": "dir"
		},
		{
			"storage": "storage2",
			"enabled": 1,
			"active": 1,
			"total": 2000000000,
			"type": "lvm"
		}
	]`

	var storages Storages
	err := storages.UnmarshalJSON([]byte(json))
	assert.Nil(t, err)
	assert.Len(t, storages, 2)
	assert.Equal(t, "storage1", storages[0].Storage)
	assert.Equal(t, uint64(1500000000000000), storages[0].Total)
	assert.Equal(t, "storage2", storages[1].Storage)
	assert.Equal(t, uint64(2000000000), storages[1].Total)
}
