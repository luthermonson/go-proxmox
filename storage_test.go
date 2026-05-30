package proxmox

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/luthermonson/go-proxmox/tests/mocks/capture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	jsonStr := `[
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
	err := storages.UnmarshalJSON([]byte(jsonStr))
	assert.Nil(t, err)
	assert.Len(t, storages, 2)
	assert.Equal(t, "storage1", storages[0].Storage)
	assert.Equal(t, uint64(1500000000000000), storages[0].Total)
	assert.Equal(t, "storage2", storages[1].Storage)
	assert.Equal(t, uint64(2000000000), storages[1].Total)
}

func TestStorage_MarshalUnmarshalRoundTrip(t *testing.T) {
	// Test that marshalling and unmarshalling preserves the Name field
	// This tests the bug reported in issue #241 where the Storage field
	// (which marshals to "Storage" in JSON) would overwrite the Name field
	// (which marshals to "storage" in JSON) during round-trip.
	original := Storage{
		Name:    "iso",
		Content: "iso",
		Enabled: 1,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(original)
	assert.Nil(t, err)

	// Unmarshal back to Storage
	var unmarshalled Storage
	err = json.Unmarshal(jsonBytes, &unmarshalled)
	assert.Nil(t, err)

	// The Name field should be preserved after round-trip
	assert.Equal(t, original.Name, unmarshalled.Name, "Name field should be preserved after marshal/unmarshal round-trip")
	assert.Equal(t, original.Content, unmarshalled.Content)
	assert.Equal(t, original.Enabled, unmarshalled.Enabled)
}

func TestStorage_Upload_InvalidContent(t *testing.T) {
	storage := &Storage{Node: "node1", Name: "local"}
	_, err := storage.Upload("not-a-real-content-type", "some-file")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "iso")
}

func TestStorage_UploadString_InvalidContent(t *testing.T) {
	storage := &Storage{Node: "node1", Name: "local"}
	_, err := storage.UploadString("not-a-real-content-type", "user-data", "#cloud-config\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "iso")
}

func TestStorage_UploadString(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	const want = "fake iso payload"
	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}

	task, err := storage.UploadString("iso", "tiny.iso", want)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "imgcopy", task.Type)

	require.NotNil(t, capture.LastUpload, "upload matcher did not run")
	assert.Equal(t, "iso", capture.LastUpload.Fields["content"])
	assert.Equal(t, "tiny.iso", capture.LastUpload.Filename)
	assert.Equal(t, want, capture.LastUpload.Body)
	// Proxmox treats "filename" as a single parameter; sending it both as a
	// form field and as the file part name causes empty-body 4xx responses.
	_, ok := capture.LastUpload.Fields["filename"]
	assert.False(t, ok, "filename must not appear as a form field; it lives in the file part's Content-Disposition")
}

func TestStorage_UploadWithName_FilenameNotDuplicated(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	tmp, err := os.CreateTemp("", "upload-test-*.iso")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, err = tmp.WriteString("fake iso data")
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	_, err = storage.UploadWithName("iso", tmp.Name(), "renamed.iso")
	require.NoError(t, err)

	require.NotNil(t, capture.LastUpload)
	assert.Equal(t, "renamed.iso", capture.LastUpload.Filename)
	_, ok := capture.LastUpload.Fields["filename"]
	assert.False(t, ok, "filename must not appear as a form field")
}

func TestClient_UploadReader(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	const payload = "the body"
	body := strings.NewReader(payload)
	err := mockClient().UploadReader(
		"/nodes/node1/storage/local/upload",
		map[string]string{"content": "iso"},
		"hello.txt", body, int64(body.Len()), nil,
	)
	require.NoError(t, err)

	require.NotNil(t, capture.LastUpload)
	assert.Equal(t, "iso", capture.LastUpload.Fields["content"])
	assert.Equal(t, "hello.txt", capture.LastUpload.Filename)
	assert.Equal(t, payload, capture.LastUpload.Body)
}

func TestStorage_PreviewPruneBackups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}

	items, err := storage.PreviewPruneBackups(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, items, 4)

	marks := make(map[string]*PruneBackupItem, len(items))
	for _, it := range items {
		marks[it.Mark] = it
	}
	require.Contains(t, marks, "keep")
	require.Contains(t, marks, "remove")
	require.Contains(t, marks, "protected")
	require.Contains(t, marks, "renamed")

	assert.Equal(t, "local:backup/vzdump-qemu-100-2024_01_08-03_00_00.vma.zst", marks["remove"].Volid)
	assert.Equal(t, "qemu", marks["remove"].Type)
	assert.Equal(t, uint64(100), marks["remove"].VMID)
	assert.Equal(t, StringOrUint64(1704682800), marks["remove"].Ctime)
}

func TestStorage_PreviewPruneBackups_WithFilters(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}

	items, err := storage.PreviewPruneBackups(context.Background(), &StoragePruneBackupsOptions{
		PruneBackups: "keep-last=3,keep-monthly=4",
		Type:         "qemu",
		VMID:         100,
	})
	require.NoError(t, err)
	require.NotEmpty(t, items)
}

func TestStorage_PruneBackups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}

	task, err := storage.PruneBackups(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "prunebackups", task.Type)
}

func TestStorage_PruneBackups_WithOpts(t *testing.T) {
	// Hit the opts-not-nil / query-string-appended branch.
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	task, err := storage.PruneBackups(context.Background(), &StoragePruneBackupsOptions{
		PruneBackups: "keep-last=1", Type: "qemu", VMID: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestStorage_ISO_RegeneratesVolID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	iso, err := storage.ISO(context.Background(), "no-volid.iso")
	require.NoError(t, err)
	assert.Equal(t, "local:iso/no-volid.iso", iso.VolID)
}

func TestStorage_VzTmpl_RegeneratesVolID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	vz, err := storage.VzTmpl(context.Background(), "no-volid.tar.zst")
	require.NoError(t, err)
	assert.Equal(t, "local:vztmpl/no-volid.tar.zst", vz.VolID)
}

func TestStorage_Import_RegeneratesVolID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	storage := &Storage{client: mockClient(), Node: "node1", Name: "esxi"}
	imp, err := storage.Import(context.Background(), "no-volid.vmx")
	require.NoError(t, err)
	assert.Equal(t, "esxi:import/no-volid.vmx", imp.VolID)
}

func TestStoragePruneBackupsOptions_QueryString(t *testing.T) {
	cases := []struct {
		name string
		opts *StoragePruneBackupsOptions
		want string
	}{
		{"nil", nil, ""},
		{"empty", &StoragePruneBackupsOptions{}, ""},
		{
			"all fields",
			&StoragePruneBackupsOptions{PruneBackups: "keep-last=3", Type: "qemu", VMID: 100},
			"prune-backups=keep-last%3D3&type=qemu&vmid=100",
		},
		{"vmid only", &StoragePruneBackupsOptions{VMID: 100}, "vmid=100"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.opts.queryString())
		})
	}
}

func TestStorage_UploadWithHash_InvalidContent(t *testing.T) {
	storage := &Storage{Node: "node1", Name: "local"}
	rename := "renamed.iso"
	_, err := storage.UploadWithHash("not-a-real-content", "file", &rename, "deadbeef", "sha256")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "iso")
}

func TestStorage_UploadWithHash(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	tmp, err := os.CreateTemp("", "upload-hash-*.iso")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, err = tmp.WriteString("fake iso data")
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	rename := "hashed.iso"
	task, err := storage.UploadWithHash("iso", tmp.Name(), &rename, "deadbeef", "sha256")
	require.NoError(t, err)
	require.NotNil(t, task)

	require.NotNil(t, capture.LastUpload)
	assert.Equal(t, "hashed.iso", capture.LastUpload.Filename)
	assert.Equal(t, "deadbeef", capture.LastUpload.Fields["checksum"])
	assert.Equal(t, "sha256", capture.LastUpload.Fields["checksum-algorithm"])

	// nil rename branch: filename derives from file basename.
	task, err = storage.UploadWithHash("iso", tmp.Name(), nil, "cafebabe", "md5")
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestStorage_DownloadURLWithHash(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	task, err := storage.DownloadURLWithHash(
		context.Background(), "iso", "debian.iso",
		"https://example.com/debian.iso", "deadbeef", "sha256",
	)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "download", task.Type)
}

func TestStorage_DownloadURL_InvalidContent(t *testing.T) {
	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	_, err := storage.DownloadURL(context.Background(), "bogus", "x", "https://e/x")
	require.Error(t, err)
}

func TestStorage_ISO(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	iso, err := storage.ISO(context.Background(), "debian-12.iso")
	require.NoError(t, err)
	require.NotNil(t, iso)
	assert.Equal(t, "local:iso/debian-12.iso", iso.VolID)
	assert.Equal(t, "node1", iso.Node)
	assert.Equal(t, "local", iso.Storage)

	task, err := iso.Delete(context.Background())
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "imgdel", task.Type)
}

func TestStorage_VzTmpl_Detail(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	vz, err := storage.VzTmpl(context.Background(), "debian-12-standard.tar.zst")
	require.NoError(t, err)
	require.NotNil(t, vz)
	assert.Equal(t, "local:vztmpl/debian-12-standard.tar.zst", vz.VolID)
	assert.Equal(t, "local", vz.Storage)

	task, err := vz.Delete(context.Background())
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "imgdel", task.Type)
}

func TestStorage_Import(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "esxi"}
	imp, err := storage.Import(context.Background(), "MyVM.vmx")
	require.NoError(t, err)
	require.NotNil(t, imp)
	assert.Equal(t, "esxi:import/MyVM.vmx", imp.VolID)
	assert.Equal(t, "esxi", imp.Storage)
}

func TestStorage_Backup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	b, err := storage.Backup(context.Background(), "vzdump-qemu-100.vma.zst")
	require.NoError(t, err)
	require.NotNil(t, b)
	assert.Equal(t, "local:backup/vzdump-qemu-100.vma.zst", b.VolID)

	task, err := b.Delete(context.Background())
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestStorage_DeleteVolume_FromPath(t *testing.T) {
	// deleteVolume's path-only branch rebuilds volid from filepath.Base(path)
	// for callers that don't carry a VolID. Construct a Backup directly so the
	// reconstruction path is exercised.
	mocks.On(mockConfig)
	defer mocks.Off()

	b := &Backup{Content: Content{
		client:  mockClient(),
		Node:    "node1",
		Storage: "local",
		Path:    "/var/lib/vz/dump/from-path.vma.zst",
	}}
	task, err := b.Delete(context.Background())
	require.NoError(t, err)
	require.NotNil(t, task)
}

func TestStorage_DeleteVolume_MissingFields(t *testing.T) {
	// Both VolID and Path empty must surface an error without making a request.
	b := &Backup{Content: Content{client: mockClient(), Node: "node1", Storage: "local"}}
	_, err := b.Delete(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "volid or path required")
}

func TestStorage_ImportMetadata(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	storage := &Storage{client: mockClient(), Node: "node1", Name: "esxi"}

	meta, err := storage.ImportMetadata(context.Background(), "esxi:ha-datacenter/MyVM/MyVM.vmx")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, "vm", meta.Type)
	assert.Equal(t, "esxi", meta.Source)
	assert.Equal(t, "imported-vm", meta.CreateArgs["name"])
	assert.Equal(t, float64(4096), meta.CreateArgs["memory"])

	require.Len(t, meta.Disks, 2)
	assert.Equal(t, "esxi:ha-datacenter/MyVM/MyVM.vmdk", meta.Disks["scsi0"])

	require.Len(t, meta.Net, 1)
	net0, ok := meta.Net["net0"].(map[string]interface{})
	require.True(t, ok, "net0 should unmarshal as a map")
	assert.Equal(t, "vmxnet3", net0["model"])

	require.Len(t, meta.Warnings, 1)
	assert.Equal(t, "guest-is-running", meta.Warnings[0].Type)
	assert.Equal(t, "power", meta.Warnings[0].Key)
	assert.Equal(t, "poweredOn", meta.Warnings[0].Value)
}
