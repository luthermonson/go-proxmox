package proxmox

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/diskfs/go-diskfs/backend"
	"github.com/diskfs/go-diskfs/backend/file"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luthermonson/go-proxmox/tests/mocks"
)

func TestVirtualMachine_Ping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	assert.Nil(t, vm.Ping(ctx))
	assert.Equal(t, StringOrUint64(101), vm.VMID)
	assert.Equal(t, IntOrBool(true), vm.Spice)
}

func TestVirtualMachine_RRDData(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	rdddata, err := vm.RRDData(ctx, TimeframeHour)
	assert.Nil(t, err)
	assert.Len(t, rdddata, 70)
}

func TestVirtualMachineClone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vmTemplate := VirtualMachine{
		client:   client,
		Node:     "node1",
		Template: true,
		VMID:     101,
	}
	cloneOptions := VirtualMachineCloneOptions{
		NewID: 102,
	}
	newID, _, err := vmTemplate.Clone(ctx, &cloneOptions)
	assert.Nil(t, err)
	assert.Equal(t, cloneOptions.NewID, newID)
}

func TestVirtualMachineMonitor(t *testing.T) {
	mocks.On(mockConfig)
	client := mockClient()
	defer mocks.Off()
	ctx := context.Background()
	vmTemplate := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}
	out, err := vmTemplate.Monitor(ctx, "help")
	assert.Nil(t, err)
	assert.Equal(t, "help text", out)
}

func TestVirtualMachineCloneWithoutNewID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vmTemplate := VirtualMachine{
		client:   client,
		Node:     "node1",
		Template: true,
		VMID:     101,
	}
	cloneOptions := VirtualMachineCloneOptions{}
	newID, _, err := vmTemplate.Clone(ctx, &cloneOptions)
	assert.Nil(t, err)
	assert.Equal(t, 100, newID)
}

func TestVirtualMachineState(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	runningVM := VirtualMachine{
		Status:    "running",
		QMPStatus: "running",
	}
	assert.False(t, runningVM.IsStopped())
	assert.False(t, runningVM.IsPaused())
	assert.False(t, runningVM.IsHibernated())
	assert.True(t, runningVM.IsRunning())
	stoppedVM := VirtualMachine{
		Status:    "stopped",
		QMPStatus: "stopped",
	}
	assert.True(t, stoppedVM.IsStopped())
	assert.False(t, stoppedVM.IsPaused())
	assert.False(t, stoppedVM.IsHibernated())
	assert.False(t, stoppedVM.IsRunning())
	pausedVM := VirtualMachine{
		Status:    "running",
		QMPStatus: "paused",
	}
	assert.False(t, pausedVM.IsStopped())
	assert.True(t, pausedVM.IsPaused())
	assert.False(t, pausedVM.IsHibernated())
	assert.False(t, pausedVM.IsRunning())
	hibernatedVM := VirtualMachine{
		Status:    "stopped",
		QMPStatus: "stopped",
		Lock:      "suspended",
	}
	assert.False(t, hibernatedVM.IsStopped())
	assert.False(t, hibernatedVM.IsPaused())
	assert.True(t, hibernatedVM.IsHibernated())
	assert.False(t, hibernatedVM.IsRunning())
}

func TestVirtualMachineStateWithoutQMPStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	runningVM := VirtualMachine{
		Status: "running",
	}
	assert.False(t, runningVM.IsStopped())
	assert.False(t, runningVM.IsPaused())
	assert.False(t, runningVM.IsHibernated())
	assert.True(t, runningVM.IsRunning())
	stoppedVM := VirtualMachine{
		Status: "stopped",
	}
	assert.True(t, stoppedVM.IsStopped())
	assert.False(t, stoppedVM.IsPaused())
	assert.False(t, stoppedVM.IsHibernated())
	assert.False(t, stoppedVM.IsRunning())
	hibernatedVM := VirtualMachine{
		Status: "stopped",
		Lock:   "suspended",
	}
	assert.False(t, hibernatedVM.IsStopped())
	assert.False(t, hibernatedVM.IsPaused())
	assert.True(t, hibernatedVM.IsHibernated())
	assert.False(t, hibernatedVM.IsRunning())
}


func TestVirtualMachine_Config(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	// Config() with options updates the config and returns a task
	task, err := vm.Config(ctx, VirtualMachineOption{Name: "tags", Value: "test;demo"})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmconfig", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Start(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Start(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmstart", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Stop(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Stop(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmstop", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Shutdown(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Shutdown(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmshutdown", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Reboot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Reboot(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmreboot", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Reset(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Reset(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmreset", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Pause(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Pause(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmsuspend", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Resume(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.Resume(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmresume", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   999,
		Node:   "node1",
	}

	task, err := vm.Delete(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmdestroy", task.Type)
	assert.Equal(t, "999", task.ID)
}

func TestVirtualMachine_AgentPing(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	assert.Nil(t, vm.AgentPing(context.Background()))
}

func TestVirtualMachine_AgentGetTime(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	ts, err := vm.AgentGetTime(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, AgentTime(1715600000000000000), ts)
}

func TestVirtualMachine_AgentGetTimezone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	tz, err := vm.AgentGetTimezone(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "UTC", tz)
}

func TestVirtualMachine_AgentGetUsers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	users, err := vm.AgentGetUsers(context.Background())
	assert.Nil(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "root", users[0].User)
	assert.Equal(t, "WORKGROUP", users[1].Domain)
}

func TestVirtualMachine_AgentGetVCPUs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	cpus, err := vm.AgentGetVCPUs(context.Background())
	assert.Nil(t, err)
	assert.Len(t, cpus, 2)
	assert.Equal(t, 1, cpus[1].LogicalID)
	assert.True(t, cpus[1].CanOffline)
}

func TestVirtualMachine_AgentGetFsInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	fs, err := vm.AgentGetFsInfo(context.Background())
	assert.Nil(t, err)
	assert.Len(t, fs, 1)
	assert.Equal(t, "/", fs[0].Mountpoint)
	assert.Equal(t, "ext4", fs[0].Type)
	assert.Len(t, fs[0].Disk, 1)
	assert.Equal(t, "/dev/sda1", fs[0].Disk[0].Dev)
}

func TestVirtualMachine_AgentGetMemoryBlocks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	blocks, err := vm.AgentGetMemoryBlocks(context.Background())
	assert.Nil(t, err)
	assert.Len(t, blocks, 2)
	assert.True(t, blocks[1].CanOffline)
}

func TestVirtualMachine_AgentGetInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	info, err := vm.AgentGetInfo(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "7.2.0", info.Version)
	assert.Len(t, info.SupportedCommands, 2)
}

func TestVirtualMachine_AgentFsfreezeFreezeThawStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	ctx := context.Background()

	n, err := vm.AgentFsfreezeFreeze(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)

	status, err := vm.AgentFsfreezeStatus(ctx)
	assert.Nil(t, err)
	assert.Equal(t, AgentFsfreezeStatus("thawed"), status)

	n, err = vm.AgentFsfreezeThaw(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
}

func TestVirtualMachine_AgentFstrim(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	out, err := vm.AgentFstrim(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, out, "/")
}

func TestVirtualMachine_AgentShutdown(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	assert.Nil(t, vm.AgentShutdown(context.Background()))
}

func TestVirtualMachine_AgentSuspend(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	ctx := context.Background()
	assert.Nil(t, vm.AgentSuspendDisk(ctx))
	assert.Nil(t, vm.AgentSuspendHybrid(ctx))
	assert.Nil(t, vm.AgentSuspendRAM(ctx))
}

func TestVirtualMachine_AgentFileRead(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	out, err := vm.AgentFileRead(context.Background(), "/etc/hostname")
	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, "hello world\n", out.Content)
	assert.Equal(t, IntOrBool(false), out.Truncated)
}

func TestVirtualMachine_AgentFileWrite(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	assert.Nil(t, vm.AgentFileWrite(context.Background(), "/tmp/foo", []byte("hello")))
}

func TestVirtualMachine_Snapshots(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	snapshots, err := vm.Snapshots(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, snapshots)
	assert.Len(t, snapshots, 3)
	assert.Equal(t, "current", snapshots[0].Name)
	assert.Equal(t, "snap1", snapshots[1].Name)
	assert.Equal(t, "Before upgrade", snapshots[1].Description)
	assert.Equal(t, "snap2", snapshots[2].Name)
}

func TestVirtualMachine_NewSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.NewSnapshot(ctx, "test-snapshot")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmsnapshot", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_DeleteSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		Node:   "node1",
		VMID:   100,
	}
	task, err := vm.DeleteSnapshot(ctx, "snap2")
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestVirtualMachine_SnapshotRollback(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	task, err := vm.SnapshotRollback(ctx, "snap1")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmrollback", task.Type)
	assert.Equal(t, "100", task.ID)
}

func cleanupISO(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Logf("removing test iso %s: %v", path, err)
	}
}

func closeBackend(t *testing.T, bk backend.Storage) {
	t.Helper()
	if err := bk.Close(); err != nil {
		t.Logf("closing iso backend: %v", err)
	}
}

func TestWithCloudInitStorage(t *testing.T) {
	var cfg cloudInitConfig
	WithCloudInitStorage("templates")(&cfg)
	assert.Equal(t, "templates", cfg.storage)
}

func TestResolveCloudInitStorage_FallbackWhenUnset(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	require.NoError(t, err)

	storage, err := resolveCloudInitStorage(ctx, node, &cloudInitConfig{})
	require.NoError(t, err)
	require.NotNil(t, storage)
	assert.Contains(t, storage.Content, "iso",
		"fallback should resolve to an iso-capable storage")
}

func TestResolveCloudInitStorage_NamedISOCapable(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	require.NoError(t, err)

	storage, err := resolveCloudInitStorage(ctx, node, &cloudInitConfig{storage: "local"})
	require.NoError(t, err)
	require.NotNil(t, storage)
	assert.Equal(t, "local", storage.Name)
	assert.Contains(t, storage.Content, "iso")
}

func TestResolveCloudInitStorage_NamedRejectsNonISO(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	require.NoError(t, err)

	_, err = resolveCloudInitStorage(ctx, node, &cloudInitConfig{storage: "local-lvm"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local-lvm")
	assert.Contains(t, err.Error(), "iso content")
}

func TestMakeCloudInitISO(t *testing.T) {
	userdata := "#cloud-config\npassword: test\n"
	metadata := "instance-id: test-vm\nlocal-hostname: test\n"

	isoPath, err := makeCloudInitISO("test-cloudinit.iso", userdata, metadata, "", "")
	require.NoError(t, err)
	defer cleanupISO(t, isoPath)

	assert.FileExists(t, isoPath)

	bk, err := file.OpenFromPath(isoPath, true)
	require.NoError(t, err)
	defer closeBackend(t, bk)

	fs, err := iso9660.Read(bk, 0, 0, blockSize)
	require.NoError(t, err)

	for filename, want := range map[string]string{
		"/user-data": userdata,
		"/meta-data": metadata,
	} {
		f, err := fs.OpenFile(filename, os.O_RDONLY)
		require.NoError(t, err, "opening %s", filename)
		got, err := io.ReadAll(f)
		require.NoError(t, err, "reading %s", filename)
		assert.Equal(t, want, string(got))
	}
}

func TestMakeCloudInitISO_AllFiles(t *testing.T) {
	userdata := "#cloud-config\n"
	metadata := "instance-id: vm-100\n"
	vendordata := "vendor: test\n"
	networkconfig := "network:\n  version: 2\n"

	isoPath, err := makeCloudInitISO("test-allfiles.iso", userdata, metadata, vendordata, networkconfig)
	require.NoError(t, err)
	defer cleanupISO(t, isoPath)

	bk, err := file.OpenFromPath(isoPath, true)
	require.NoError(t, err)
	defer closeBackend(t, bk)

	fs, err := iso9660.Read(bk, 0, 0, blockSize)
	require.NoError(t, err)

	expected := map[string]string{
		"/user-data":      userdata,
		"/meta-data":      metadata,
		"/vendor-data":    vendordata,
		"/network-config": networkconfig,
	}
	for filename, want := range expected {
		f, err := fs.OpenFile(filename, os.O_RDONLY)
		require.NoError(t, err, "opening %s", filename)
		got, err := io.ReadAll(f)
		require.NoError(t, err, "reading %s", filename)
		assert.Equal(t, want, string(got))
	}
}

func TestMakeCloudInitISO_JolietSVD(t *testing.T) {
	isoPath, err := makeCloudInitISO("test-joliet.iso", "userdata", "metadata", "", "")
	require.NoError(t, err)
	defer cleanupISO(t, isoPath)

	isoBytes, err := os.ReadFile(isoPath)
	require.NoError(t, err)

	// Scan volume descriptors starting at sector 16 for a Joliet SVD.
	// Type 0x02 + "CD001" signature + Joliet escape sequence at bytes 88-90.
	jolietEscapes := [][]byte{
		{0x25, 0x2F, 0x40}, // UCS-2 Level 1
		{0x25, 0x2F, 0x43}, // UCS-2 Level 2
		{0x25, 0x2F, 0x45}, // UCS-2 Level 3
	}

	var foundJoliet bool
	for i := 0; ; i++ {
		offset := int64(16+i) * blockSize
		if offset+blockSize > int64(len(isoBytes)) {
			break
		}
		vd := isoBytes[offset : offset+blockSize]
		if vd[0] == 0xFF {
			break
		}
		if vd[0] == 0x02 && string(vd[1:6]) == "CD001" {
			esc := vd[88:91]
			for _, valid := range jolietEscapes {
				if bytes.Equal(esc, valid) {
					foundJoliet = true
					break
				}
			}
		}
	}

	assert.True(t, foundJoliet, "Joliet Supplementary Volume Descriptor not found in ISO")
}

func TestMakeCloudInitISO_VolumeIdentifier(t *testing.T) {
	isoPath, err := makeCloudInitISO("test-volid.iso", "userdata", "metadata", "", "")
	require.NoError(t, err)
	defer cleanupISO(t, isoPath)

	isoBytes, err := os.ReadFile(isoPath)
	require.NoError(t, err)

	// PVD is at sector 16, volume identifier is at bytes 40-72 (32 bytes, space-padded).
	pvdOffset := int64(16) * blockSize
	require.Greater(t, int64(len(isoBytes)), pvdOffset+blockSize)

	pvd := isoBytes[pvdOffset : pvdOffset+blockSize]
	assert.Equal(t, byte(0x01), pvd[0], "expected PVD type")
	assert.Equal(t, "CD001", string(pvd[1:6]))

	volID := strings.TrimRight(string(pvd[40:72]), " \x00")
	assert.Equal(t, volumeIdentifier, volID)
}
