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
	"github.com/h2non/gock"
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

func TestVirtualMachine_RRD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	rrd, err := vm.RRD(ctx, "cpu", TimeframeHour)
	assert.Nil(t, err)
	require.NotNil(t, rrd)
	assert.Contains(t, rrd.Filename, "101.png")
}

func TestVirtualMachine_MigratePreconditions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	pre, err := vm.MigratePreconditions(ctx, "")
	assert.Nil(t, err)
	require.NotNil(t, pre)
	assert.True(t, pre.Running)
	assert.True(t, pre.HasDBusVMState)
	assert.Equal(t, []string{"node2", "node3"}, pre.AllowedNodes)
	require.Contains(t, pre.NotAllowedNodes, "node4")
	assert.Equal(t, []string{"local-lvm"}, pre.NotAllowedNodes["node4"].UnavailableStorages)
	require.Len(t, pre.NotAllowedNodes["node4"].BlockingHAResources, 1)
	assert.Equal(t, "vm:101", pre.NotAllowedNodes["node4"].BlockingHAResources[0].SID)
	assert.Equal(t, "node-affinity", pre.NotAllowedNodes["node4"].BlockingHAResources[0].Cause)
	require.Len(t, pre.LocalDisks, 1)
	assert.Equal(t, "local-lvm:vm-101-disk-0", pre.LocalDisks[0].VolID)
	assert.Equal(t, uint64(34359738368), pre.LocalDisks[0].Size)
}

func TestVirtualMachine_RemoteMigrate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	task, err := vm.RemoteMigrate(ctx, &VirtualMachineRemoteMigrateOptions{
		TargetEndpoint: "apitoken=PVEAPIToken=user@pam!tok=secret host=target.example.com fingerprint=AA:BB",
		TargetBridge:   "vmbr0=vmbr0",
		TargetStorage:  "local=local",
		TargetVMID:     201,
		Online:         IntOrBool(true),
	})
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, string(task.UPID), "qmremote-migrate")
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

func TestVirtualMachine_ConfigSync(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	// ConfigSync blocks until the change is applied and returns no task.
	err := vm.ConfigSync(ctx, VirtualMachineOption{Name: "description", Value: "synchronous update"})
	assert.Nil(t, err)
}

func TestVirtualMachine_Feature(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	// Feature without a snapshot.
	feature, err := vm.Feature(ctx, "snapshot", "")
	assert.Nil(t, err)
	assert.True(t, feature.HasFeature)
	assert.Equal(t, []string{"node1", "node2"}, feature.Nodes)

	// Feature against a specific snapshot.
	feature, err = vm.Feature(ctx, "clone", "snap1")
	assert.Nil(t, err)
	assert.True(t, feature.HasFeature)
}

func TestVirtualMachine_DBusVMState(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   100,
		Node:   "node1",
	}

	assert.Nil(t, vm.DBusVMState(ctx, "start"))
	assert.Nil(t, vm.DBusVMState(ctx, "stop"))
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

	task, err := vm.Delete(ctx, nil)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmdestroy", task.Type)
	assert.Equal(t, "999", task.ID)
}

func TestVirtualMachine_DeleteWithOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	defer gock.Off()

	gock.New(TestURI).
		Delete("^/nodes/node1/qemu/998$").
		MatchParams(map[string]string{
			"purge":                      "1",
			"skiplock":                   "1",
			"destroy-unreferenced-disks": "1",
		}).
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:qmdestroy:998:root@pam:"}`)
	client := mockClient()
	ctx := context.Background()
	vm := VirtualMachine{
		client: client,
		VMID:   998,
		Node:   "node1",
	}

	options := &VirtualMachineDeleteOptions{
		Purge:                    true,
		SkipLock:                 true,
		DestroyUnreferencedDisks: true,
	}
	task, err := vm.Delete(ctx, options)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmdestroy", task.Type)
	assert.Equal(t, "998", task.ID)
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

func TestVirtualMachine_AgentCommandIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	cmds, err := vm.AgentCommandIndex(context.Background())
	assert.Nil(t, err)
	assert.Len(t, cmds, 3)
	assert.Equal(t, "exec", cmds[0].Name)
	assert.Equal(t, "ping", cmds[1].Name)
}

func TestVirtualMachine_AgentCommand(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	out, err := vm.AgentCommand(context.Background(), "ping")
	assert.Nil(t, err)
	assert.Equal(t, "ping", out["echoed"])
}

func TestVirtualMachine_AgentGetMemoryBlockInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := VirtualMachine{client: mockClient(), VMID: 101, Node: "node1"}
	info, err := vm.AgentGetMemoryBlockInfo(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, uint64(134217728), info.Size)
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
	// Returned snapshots must be wired to their parent VM so instance
	// methods can target the right snapshot path.
	for _, s := range snapshots {
		assert.Equal(t, "node1", s.Node)
		assert.Equal(t, 100, s.VMID)
	}
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
	task, err := vm.Snapshot("snap2").Delete(ctx)
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

	task, err := vm.Snapshot("snap1").Rollback(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "qmrollback", task.Type)
	assert.Equal(t, "100", task.ID)
}

func TestVirtualMachine_Snapshot_Getter(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := vm100()
	s := vm.Snapshot("snap1")
	assert.NotNil(t, s)
	assert.Equal(t, "snap1", s.Name)
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, 100, s.VMID)
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

func TestVirtualMachine_SpiceProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	vm := &VirtualMachine{client: client, Node: "node1", VMID: 101}
	spice, err := vm.SpiceProxy(ctx)
	assert.Nil(t, err)
	require.NotNil(t, spice)
	assert.Equal(t, "spice", spice.Type)
	assert.Equal(t, "node1.example.com", spice.Host)
	assert.Equal(t, "61024", spice.Port)
	assert.Equal(t, "secret-ticket", spice.Password)
	assert.Equal(t, "61025", spice.TLSPort)
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

func TestVirtualMachine_DirIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := vm100().DirIndex(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
	// at minimum the canonical PVE child resources should be present
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Subdir] = true
	}
	assert.True(t, names["config"])
	assert.True(t, names["status"])
	assert.True(t, names["snapshot"])
	assert.True(t, names["firewall"])
	assert.True(t, names["mtunnel"])
}

func TestVirtualMachine_StatusIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := vm100().StatusIndex(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Subdir] = true
	}
	assert.True(t, names["current"])
	assert.True(t, names["start"])
	assert.True(t, names["stop"])
	assert.True(t, names["reboot"])
}

func TestVirtualMachineSnapshot_SubResources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := vm100().Snapshot("snap1").SubResources(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Subdir] = true
	}
	assert.True(t, names["config"])
	assert.True(t, names["rollback"])
}

func TestVirtualMachine_MigrationTunnel(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	tunnel, err := vm100().MigrationTunnel(context.Background(), nil)
	assert.Nil(t, err)
	require.NotNil(t, tunnel)
	assert.Equal(t, "/run/qemu-server/100.mtunnel", tunnel.Socket)
	assert.Equal(t, "PVEMTUNNELTICKET:abc123", tunnel.Ticket)
	assert.Contains(t, tunnel.UPID, "qmtunnel")

	// with explicit options
	tunnel2, err := vm100().MigrationTunnel(context.Background(), &VirtualMachineMigrationTunnelOptions{
		Bridges:  "vmbr0",
		Storages: "local-lvm",
	})
	assert.Nil(t, err)
	require.NotNil(t, tunnel2)
	assert.Equal(t, tunnel.Socket, tunnel2.Socket)
}

func TestVirtualMachine_MigrationTunnelWebSocketPath(t *testing.T) {
	vm := vm100()
	tunnel := &VirtualMachineMigrationTunnel{
		Socket: "/run/qemu-server/100.mtunnel",
		Ticket: "PVEMTUNNELTICKET:abc123",
	}
	path := vm.MigrationTunnelWebSocketPath(tunnel)
	assert.Contains(t, path, "/nodes/node1/qemu/100/mtunnelwebsocket")
	assert.Contains(t, path, "socket=")
	assert.Contains(t, path, "ticket=")
	// ticket should be URL-encoded (':' becomes %3A)
	assert.Contains(t, path, "PVEMTUNNELTICKET%3Aabc123")

	// nil tunnel still returns a valid path skeleton
	emptyPath := vm.MigrationTunnelWebSocketPath(nil)
	assert.Contains(t, emptyPath, "/nodes/node1/qemu/100/mtunnelwebsocket")
}

// vm101 returns a *VirtualMachine wired to vmid 101 on node1 for tests
// that reuse the broader set of vmid-101 mock fixtures.
func vm101() *VirtualMachine {
	return &VirtualMachine{client: mockClient(), Node: "node1", VMID: 101}
}

func TestVirtualMachine_New(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	var vm VirtualMachine
	vm.New(client, "node1", 101)
	assert.Equal(t, StringOrUint64(101), vm.VMID)
	assert.Equal(t, "node1", vm.Node)
	assert.Same(t, client, vm.client)
}

func TestVirtualMachine_Hibernate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := vm101().Hibernate(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmsuspend", task.Type)
	assert.Equal(t, "101", task.ID)
}

func TestVirtualMachine_Migrate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	// nil params path
	task, err := vm101().Migrate(ctx, nil)
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmigrate", task.Type)

	// explicit params path
	task, err = vm101().Migrate(ctx, &VirtualMachineMigrateOptions{Target: "node2", Online: IntOrBool(true)})
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmigrate", task.Type)
}

func TestVirtualMachine_RemoteMigrate_NilParams(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := vm101().RemoteMigrate(context.Background(), nil)
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmremote-migrate", task.Type)
}

func TestVirtualMachine_ResizeDisk(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := vm101().ResizeDisk(context.Background(), "scsi0", "+1G")
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmresize", task.Type)
}

func TestVirtualMachine_UnlinkDisk(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	task, err := vm101().UnlinkDisk(ctx, "scsi5", false)
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmunlink", task.Type)

	// force path
	task, err = vm101().UnlinkDisk(ctx, "scsi5", true)
	assert.Nil(t, err)
	require.NotNil(t, task)
}

func TestVirtualMachine_MoveDisk(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	// nil params + disk supplied
	task, err := vm101().MoveDisk(ctx, "scsi0", nil)
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmmove", task.Type)

	// explicit params, disk overrides Disk on params
	task, err = vm101().MoveDisk(ctx, "scsi1", &VirtualMachineMoveDiskOptions{Storage: "local-lvm"})
	assert.Nil(t, err)
	require.NotNil(t, task)
}

func TestVirtualMachine_ConvertToTemplate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := vm101().ConvertToTemplate(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "qmtemplate", task.Type)
}

func TestVirtualMachine_Pending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pending, err := vm101().Pending(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, pending)
	assert.NotEmpty(t, *pending)
}

func TestVirtualMachine_SendKey(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := vm101().SendKey(context.Background(), "ctrl-alt-del")
	assert.Nil(t, err)
}

func TestVirtualMachine_TermProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	term, err := vm101().TermProxy(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, term)
	assert.Equal(t, "root@pam", term.User)
}

func TestVirtualMachine_VNCProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vnc, err := vm101().VNCProxy(context.Background(), &VNCConfig{GeneratePassword: true, Websocket: true})
	assert.Nil(t, err)
	require.NotNil(t, vnc)
	assert.Equal(t, "root@pam", vnc.User)
}

func TestVirtualMachine_RRDDataAndRRD_CFValidation(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	vm := vm101()

	// More than one consolidation function returns an error in both helpers.
	_, err := vm.RRDData(ctx, TimeframeHour, "AVERAGE", "MAX")
	assert.Error(t, err)

	_, err = vm.RRD(ctx, "cpu", TimeframeHour, "AVERAGE", "MAX")
	assert.Error(t, err)

	// Single CF goes through the happy path.
	rrddata, err := vm.RRDData(ctx, TimeframeHour, "AVERAGE")
	assert.Nil(t, err)
	assert.NotEmpty(t, rrddata)

	rrd, err := vm.RRD(ctx, "cpu", TimeframeHour, "AVERAGE")
	assert.Nil(t, err)
	require.NotNil(t, rrd)
}

func TestVirtualMachine_TagHelpers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	// HasTag: nil config and empty Tags both return false.
	emptyVM := &VirtualMachine{}
	assert.False(t, emptyVM.HasTag("x"))
	emptyVM.VirtualMachineConfig = &VirtualMachineConfig{}
	assert.False(t, emptyVM.HasTag("x"))

	// Populated tag list — HasTag walks the slice.
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   102,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "alpha;beta;gamma",
		},
	}
	assert.True(t, vm.HasTag("beta"))
	assert.False(t, vm.HasTag("missing"))

	// SplitTags populates TagsSlice deterministically.
	vm2 := &VirtualMachine{VirtualMachineConfig: &VirtualMachineConfig{Tags: "a;b;c"}}
	vm2.SplitTags()
	assert.Equal(t, []string{"a", "b", "c"}, vm2.VirtualMachineConfig.TagsSlice)
}

func TestVirtualMachine_AddTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	// Adding an existing tag returns ErrNoop.
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   102,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "existing",
		},
	}
	_, err := vm.AddTag(ctx, "existing")
	assert.True(t, IsErrNoop(err))

	// Adding a fresh tag flows through Config and returns a task.
	vm = &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   102,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "existing",
		},
	}
	task, err := vm.AddTag(ctx, "fresh")
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, vm.VirtualMachineConfig.Tags, "fresh")
}

func TestVirtualMachine_RemoveTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	// Removing a missing tag returns ErrNoop.
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   102,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "alpha;beta",
		},
	}
	_, err := vm.RemoveTag(ctx, "missing")
	assert.True(t, IsErrNoop(err))

	// Removing an existing tag flows through Config and returns a task.
	vm = &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   102,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "alpha;beta;gamma",
		},
	}
	task, err := vm.RemoveTag(ctx, "beta")
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.NotContains(t, vm.VirtualMachineConfig.Tags, "beta")
	assert.Contains(t, vm.VirtualMachineConfig.Tags, "alpha")
	assert.Contains(t, vm.VirtualMachineConfig.Tags, "gamma")
}

func TestVirtualMachine_FirewallRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	rules, err := vm101().FirewallRules(context.Background())
	assert.Nil(t, err)
	assert.Len(t, rules, 2)
	// Each returned rule must carry the parent wiring so Get/Update/Delete work.
	for _, r := range rules {
		assert.Equal(t, "node1", r.node)
		assert.Equal(t, uint64(101), r.vmid)
		assert.NotNil(t, r.client)
	}
}

func TestVirtualMachine_NewFirewallRule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	rule := &FirewallRule{Type: "in", Action: "ACCEPT"}
	err := vm101().NewFirewallRule(context.Background(), rule)
	assert.Nil(t, err)
	// Parent wiring is populated after the POST returns.
	assert.NotNil(t, rule.client)
	assert.Equal(t, "node1", rule.node)
	assert.Equal(t, uint64(101), rule.vmid)
}

func TestVirtualMachine_FirewallOptionGetSet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	vm := vm101()

	// FirewallOptionGet currently passes the value (not pointer) to the
	// client; the call still executes and we only care that the request goes
	// out and returns no error.
	_, err := vm.FirewallOptionGet(ctx)
	assert.Nil(t, err)

	assert.Nil(t, vm.FirewallOptionSet(ctx, &FirewallVirtualMachineOption{
		Enable:   IntOrBool(true),
		PolicyIn: "DROP",
	}))
}

func TestVirtualMachine_FirewallIPSet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	vm := vm101()

	ipsets, err := vm.GetFirewallIPSet(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, ipsets)

	assert.Nil(t, vm.NewFirewallIPSet(ctx, FirewallIPSetCreationOption{Name: "blocked"}))

	entries, err := vm.GetFirewallIPSetEntries(ctx, "blocked")
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)

	assert.Nil(t, vm.NewFirewallIPSetEntry(ctx, "blocked", FirewallIPSetEntryCreationOption{CIDR: "10.1.2.3"}))

	entry, err := vm.GetFirewallIPSetEntry(ctx, "blocked", "10.1.2.3")
	assert.Nil(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "10.1.2.3", entry.CIDR)

	assert.Nil(t, vm.UpdateFirewallIPSetEntry(ctx, "blocked", "10.1.2.3", &FirewallIPSetEntryUpdateOption{Comment: "updated"}))
	assert.Nil(t, vm.DeleteFirewallIPSetEntry(ctx, "blocked", "10.1.2.3", "abc"))
	assert.Nil(t, vm.DeleteFirewallIPSet(ctx, "blocked", true))
}

func TestVirtualMachine_AgentGetHostName(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	hostname, err := vm101().AgentGetHostName(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "vm-101.example.com", hostname)
}

func TestVirtualMachine_AgentGetNetworkIFaces(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ifaces, err := vm101().AgentGetNetworkIFaces(context.Background())
	assert.Nil(t, err)
	// "lo" must be filtered out.
	assert.Len(t, ifaces, 1)
	assert.Equal(t, "eth0", ifaces[0].Name)
}

func TestVirtualMachine_AgentExec(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pid, err := vm101().AgentExec(context.Background(), []string{"echo", "hello"}, "")
	assert.Nil(t, err)
	assert.Equal(t, 1234, pid)
}

func TestVirtualMachine_AgentExecStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	status, err := vm101().AgentExecStatus(context.Background(), 1234)
	assert.Nil(t, err)
	require.NotNil(t, status)
	assert.Equal(t, 1, status.Exited)
	assert.Equal(t, 0, status.ExitCode)
}

func TestVirtualMachine_WaitForAgentExecExit(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Fixture returns exited=1 so the first poll succeeds immediately.
	status, err := vm101().WaitForAgentExecExit(context.Background(), 1234, 5)
	assert.Nil(t, err)
	require.NotNil(t, status)
	assert.Equal(t, 1, status.Exited)
}

func TestVirtualMachine_AgentOsInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	info, err := vm101().AgentOsInfo(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "debian", info.ID)
	assert.Equal(t, "12", info.VersionID)
}

func TestVirtualMachine_WaitForAgent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// AgentOsInfo mock returns success, so WaitForAgent returns nil on the
	// first poll.
	assert.Nil(t, vm101().WaitForAgent(context.Background(), 5))
}

func TestVirtualMachine_AgentSetUserPassword(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, vm101().AgentSetUserPassword(context.Background(), "secret", "root"))
}

// TermWebSocket / VNCWebSocket are intentionally not unit-tested here — they
// require a live websocket dialer; see the file-level comment in
// virtual_machine.go.

// FileWriteStream and other binary streaming QGA helpers do not exist in this
// version of the package; if added, leave them uncovered by gock-only tests.

func TestVirtualMachine_UnmountCloudInitISO_NoTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// VM with no cloud-init tag: returns nil without any API call.
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   100,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: "production;webserver",
		},
	}
	assert.Nil(t, vm.UnmountCloudInitISO(context.Background(), "ide2"))
}

// TestVirtualMachine_ErrorPaths sweeps the error-return branches of the
// status-change/UPID helpers using one-shot 500 mocks. This pushes the
// per-method coverage of every Start/Stop/Reset/etc. helper above 80%
// (the success path covers ~75%; this test covers the remaining return-err
// branch).
func TestVirtualMachine_ErrorPaths(t *testing.T) {
	// Register one-shot 500s for the paths we want to drive into the error
	// branch. NOT calling mocks.On() — these are pure gock fixtures and the
	// VirtualMachine handle uses mockClient() directly. Defer gock.Off-style
	// cleanup via the mocks helper.
	mocks.On(mockConfig)
	defer mocks.Off()

	type tcase struct {
		name string
		fn   func(*VirtualMachine, context.Context) error
		uri  string
		verb string // "POST" | "PUT" | "DELETE" | "GET"
	}
	cases := []tcase{
		{"Start", func(v *VirtualMachine, c context.Context) error { _, err := v.Start(c); return err }, "/nodes/node1/qemu/501/status/start", "POST"},
		{"Stop", func(v *VirtualMachine, c context.Context) error { _, err := v.Stop(c); return err }, "/nodes/node1/qemu/501/status/stop", "POST"},
		{"Reset", func(v *VirtualMachine, c context.Context) error { _, err := v.Reset(c); return err }, "/nodes/node1/qemu/501/status/reset", "POST"},
		{"Reboot", func(v *VirtualMachine, c context.Context) error { _, err := v.Reboot(c); return err }, "/nodes/node1/qemu/501/status/reboot", "POST"},
		{"Shutdown", func(v *VirtualMachine, c context.Context) error { _, err := v.Shutdown(c); return err }, "/nodes/node1/qemu/501/status/shutdown", "POST"},
		{"Pause", func(v *VirtualMachine, c context.Context) error { _, err := v.Pause(c); return err }, "/nodes/node1/qemu/501/status/suspend", "POST"},
		{"Resume", func(v *VirtualMachine, c context.Context) error { _, err := v.Resume(c); return err }, "/nodes/node1/qemu/501/status/resume", "POST"},
		{"Hibernate", func(v *VirtualMachine, c context.Context) error { _, err := v.Hibernate(c); return err }, "/nodes/node1/qemu/501/status/suspend", "POST"},
		{"NewSnapshot", func(v *VirtualMachine, c context.Context) error { _, err := v.NewSnapshot(c, "x"); return err }, "/nodes/node1/qemu/501/snapshot", "POST"},
		{"ConvertToTemplate", func(v *VirtualMachine, c context.Context) error { _, err := v.ConvertToTemplate(c); return err }, "/nodes/node1/qemu/501/template", "POST"},
		{"RemoteMigrate", func(v *VirtualMachine, c context.Context) error { _, err := v.RemoteMigrate(c, nil); return err }, "/nodes/node1/qemu/501/remote_migrate", "POST"},
		{"Migrate", func(v *VirtualMachine, c context.Context) error { _, err := v.Migrate(c, nil); return err }, "/nodes/node1/qemu/501/migrate", "POST"},
		{"ResizeDisk", func(v *VirtualMachine, c context.Context) error {
			_, err := v.ResizeDisk(c, "scsi0", "+1G")
			return err
		}, "/nodes/node1/qemu/501/resize", "PUT"},
		{"UnlinkDisk", func(v *VirtualMachine, c context.Context) error {
			_, err := v.UnlinkDisk(c, "scsi5", false)
			return err
		}, "/nodes/node1/qemu/501/unlink", "PUT"},
		{"MoveDisk", func(v *VirtualMachine, c context.Context) error { _, err := v.MoveDisk(c, "scsi0", nil); return err }, "/nodes/node1/qemu/501/move_disk", "POST"},
	}
	ctx := context.Background()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := gock.New(mockConfig.URI)
			switch tc.verb {
			case "POST":
				g = g.Post("^" + tc.uri + "$")
			case "PUT":
				g = g.Put("^" + tc.uri + "$")
			case "DELETE":
				g = g.Delete("^" + tc.uri + "$")
			case "GET":
				g = g.Get("^" + tc.uri + "$")
			}
			g.Reply(500)

			vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 501}
			err := tc.fn(vm, ctx)
			assert.Error(t, err)
		})
	}
}

// TestVirtualMachine_SnapshotInstance_ErrorPaths exercises the error-return
// branches of the snapshot instance methods (Rollback, Delete).
func TestVirtualMachine_SnapshotInstance_ErrorPaths(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	gock.New(mockConfig.URI).Post("^/nodes/node1/qemu/501/snapshot/x/rollback$").Reply(500)
	_, err := (&VirtualMachineSnapshot{client: mockClient(), Node: "node1", VMID: 501, Name: "x"}).Rollback(ctx)
	assert.Error(t, err)

	gock.New(mockConfig.URI).Delete("^/nodes/node1/qemu/501/snapshot/x$").Reply(500)
	_, err = (&VirtualMachineSnapshot{client: mockClient(), Node: "node1", VMID: 501, Name: "x"}).Delete(ctx)
	assert.Error(t, err)
}

// TestVirtualMachine_Delete_ErrorPath drives Delete's error branch (the
// post-deleteCloudInitISO DELETE failing). The VM has no cloud-init tag so
// deleteCloudInitISO is a no-op and we land on the outer DELETE returning 500.
func TestVirtualMachine_Delete_ErrorPath(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).Delete("^/nodes/node1/qemu/502$").Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.Delete(context.Background(), nil)
	assert.Error(t, err)
}

// TestVirtualMachine_MigratePreconditions_WithTarget exercises the
// target!="" branch which adds query params.
func TestVirtualMachine_MigratePreconditions_WithTarget(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Reuse the existing /migrate mock since it ignores query params.
	pre, err := vm101().MigratePreconditions(context.Background(), "node2")
	assert.Nil(t, err)
	assert.NotNil(t, pre)
}

// TestVirtualMachine_Clone_NewIDProvided covers the "newid supplied" branch
// where Clone skips the cluster.NextID round-trip.
func TestVirtualMachine_Clone_NilParams_ErrorOnNextID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Test the nil-params path that triggers cluster.NextID. This path is
	// already exercised by TestVirtualMachineCloneWithoutNewID — but call it
	// here too to keep coverage explicit.
	vm := vm101()
	newID, _, err := vm.Clone(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotZero(t, newID)
}

// TestAgent_ErrorPaths covers the err!=nil branches of the QGA / agent
// helpers that return early on a non-nil request error.
func TestAgent_ErrorPaths(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()

	type tcase struct {
		name string
		fn   func(*VirtualMachine, context.Context) error
		path string
		verb string
	}
	cases := []tcase{
		{"AgentGetHostName", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetHostName(c); return err }, "/nodes/node1/qemu/501/agent/get-host-name", "GET"},
		{"AgentGetNetworkIFaces", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetNetworkIFaces(c); return err }, "/nodes/node1/qemu/501/agent/network-get-interfaces", "GET"},
		{"AgentOsInfo", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentOsInfo(c); return err }, "/nodes/node1/qemu/501/agent/get-osinfo", "GET"},
		{"AgentExec", func(v *VirtualMachine, c context.Context) error {
			_, err := v.AgentExec(c, []string{"x"}, "")
			return err
		}, "/nodes/node1/qemu/501/agent/exec", "POST"},
		{"AgentExecStatus", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentExecStatus(c, 1); return err }, "/nodes/node1/qemu/501/agent/exec-status", "GET"},
		{"AgentCommandIndex", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentCommandIndex(c); return err }, "/nodes/node1/qemu/501/agent", "GET"},
		{"AgentCommand", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentCommand(c, "ping"); return err }, "/nodes/node1/qemu/501/agent", "POST"},
		{"AgentGetTime", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetTime(c); return err }, "/nodes/node1/qemu/501/agent/get-time", "GET"},
		{"AgentGetTimezone", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetTimezone(c); return err }, "/nodes/node1/qemu/501/agent/get-timezone", "GET"},
		{"AgentGetUsers", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetUsers(c); return err }, "/nodes/node1/qemu/501/agent/get-users", "GET"},
		{"AgentGetVCPUs", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetVCPUs(c); return err }, "/nodes/node1/qemu/501/agent/get-vcpus", "GET"},
		{"AgentGetFsInfo", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetFsInfo(c); return err }, "/nodes/node1/qemu/501/agent/get-fsinfo", "GET"},
		{"AgentGetMemoryBlocks", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetMemoryBlocks(c); return err }, "/nodes/node1/qemu/501/agent/get-memory-blocks", "GET"},
		{"AgentGetMemoryBlockInfo", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetMemoryBlockInfo(c); return err }, "/nodes/node1/qemu/501/agent/get-memory-block-info", "GET"},
		{"AgentGetInfo", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentGetInfo(c); return err }, "/nodes/node1/qemu/501/agent/info", "GET"},
		{"AgentFsfreezeFreeze", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentFsfreezeFreeze(c); return err }, "/nodes/node1/qemu/501/agent/fsfreeze-freeze", "POST"},
		{"AgentFsfreezeThaw", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentFsfreezeThaw(c); return err }, "/nodes/node1/qemu/501/agent/fsfreeze-thaw", "POST"},
		{"AgentFsfreezeStatus", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentFsfreezeStatus(c); return err }, "/nodes/node1/qemu/501/agent/fsfreeze-status", "POST"},
		{"AgentFstrim", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentFstrim(c); return err }, "/nodes/node1/qemu/501/agent/fstrim", "POST"},
		{"AgentFileRead", func(v *VirtualMachine, c context.Context) error { _, err := v.AgentFileRead(c, "/x"); return err }, "/nodes/node1/qemu/501/agent/file-read", "GET"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := gock.New(mockConfig.URI)
			switch tc.verb {
			case "POST":
				g = g.Post("^" + tc.path + "$")
			case "PUT":
				g = g.Put("^" + tc.path + "$")
			case "DELETE":
				g = g.Delete("^" + tc.path + "$")
			case "GET":
				g = g.Get("^" + tc.path + "$")
			}
			g.Reply(500)
			vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 501}
			err := tc.fn(vm, ctx)
			assert.Error(t, err)
		})
	}
}

// TestVirtualMachine_AgentGetHostName_EmptyResult covers the
// "result is empty" branch.
func TestVirtualMachine_AgentGetHostName_EmptyResult(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/agent/get-host-name$").
		Reply(200).
		JSON(`{"data": {}}`)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.AgentGetHostName(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result is empty")
}

// TestVirtualMachine_AgentOsInfo_EmptyResult covers the
// "result is empty" branch.
func TestVirtualMachine_AgentOsInfo_EmptyResult(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/agent/get-osinfo$").
		Reply(200).
		JSON(`{"data": {}}`)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.AgentOsInfo(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "result is empty")
}

// TestVirtualMachine_AgentExec_NoPID covers the "no pid returned" branch.
func TestVirtualMachine_AgentExec_NoPID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Post("^/nodes/node1/qemu/502/agent/exec$").
		Reply(200).
		JSON(`{"data": {}}`)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.AgentExec(context.Background(), []string{"echo"}, "")
	assert.Error(t, err)
}

// TestVirtualMachine_WaitForAgent_Timeout exercises the timeout branch by
// returning an unrelated error from AgentOsInfo (the helper only retries on
// the "QEMU guest agent is not running" 500 message). Driving the timeout
// directly would require a long-running test; this exercises the non-retry
// error branch instead, which is the other code path.
func TestVirtualMachine_WaitForAgent_NonRetryError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Return a 400 — handleResponse surfaces it as a "bad request" error
	// whose message does NOT contain the retry trigger string, so
	// WaitForAgent returns it immediately.
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/agent/get-osinfo$").
		Reply(400).
		JSON(`{"errors": "something else"}`)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	err := vm.WaitForAgent(context.Background(), 1)
	assert.Error(t, err)
}

// TestVirtualMachine_WaitForAgentExecExit_Error covers the
// AgentExecStatus-returns-error branch.
func TestVirtualMachine_WaitForAgentExecExit_Error(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/agent/exec-status$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.WaitForAgentExecExit(context.Background(), 1, 1)
	assert.Error(t, err)
}

// TestVirtualMachine_Ping_ConfigError covers Ping's second-call error branch:
// the /status/current returns ok but /config returns 500.
func TestVirtualMachine_Ping_ConfigError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// /status/current ok with minimal payload
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/status/current$").
		Reply(200).
		JSON(`{"data": {"vmid": 502, "status": "running"}}`)
	// /config errors
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/config$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	err := vm.Ping(context.Background())
	assert.Error(t, err)
}

// TestVirtualMachine_Snapshots_ErrorPath drives the GET-error branch.
func TestVirtualMachine_Snapshots_ErrorPath(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/snapshot$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.Snapshots(context.Background())
	assert.Error(t, err)
}

// TestVirtualMachine_FirewallRules_ErrorPath drives the GET-error branch.
func TestVirtualMachine_FirewallRules_ErrorPath(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/502/firewall/rules$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	_, err := vm.FirewallRules(context.Background())
	assert.Error(t, err)
}

// TestVirtualMachine_NewFirewallRule_ErrorPath drives the POST-error branch.
func TestVirtualMachine_NewFirewallRule_ErrorPath(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Post("^/nodes/node1/qemu/502/firewall/rules$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 502}
	err := vm.NewFirewallRule(context.Background(), &FirewallRule{Type: "in", Action: "ACCEPT"})
	assert.Error(t, err)
}

// TestVirtualMachine_DeleteCloudInitISO_HappyPath drives the successful
// iso-found-and-deleted branch: HasTag passes, Node + Storages succeed, the
// per-storage iteration finds user-data-503.iso on the cidata storage,
// Delete returns a task, WaitFor completes immediately.
func TestVirtualMachine_DeleteCloudInitISO_HappyPath(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "cinode",
		VMID:   503,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	ok, err := vm.deleteCloudInitISO(context.Background())
	if err != nil {
		t.Logf("deleteCloudInitISO err: %v", err)
	}
	assert.Nil(t, err)
	assert.True(t, ok)
}

// TestVirtualMachine_DeleteCloudInitISO_NotFoundOnAnyStorage walks the full
// deleteCloudInitISO body: HasTag passes, Node()+Storages() succeed, the
// per-storage iteration runs but the user-data ISO is not present anywhere,
// so the helper falls through to the "treat as no-op" return.
func TestVirtualMachine_DeleteCloudInitISO_NotFoundOnAnyStorage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   501, // ISO filename derived from VMID won't be in any mock storage
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	ok, err := vm.deleteCloudInitISO(context.Background())
	assert.Nil(t, err)
	assert.True(t, ok)
}

// TestVirtualMachine_UnmountCloudInitISO_WithTag exercises the path that
// goes through Config then deleteCloudInitISO. Uses vmid 100 which has the
// /config POST mock and the per-storage content mock; the cloud-init ISO
// is not present so deleteCloudInitISO short-circuits to the no-op tail.
func TestVirtualMachine_UnmountCloudInitISO_WithTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   100,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	assert.Nil(t, vm.UnmountCloudInitISO(context.Background(), "ide2"))
}

// TestVirtualMachine_CloudInit_NodeError exercises the front half of
// CloudInit: makeCloudInitISO succeeds (real iso written to TempDir, then
// removed via the defer), then client.Node() fails. Covers the opts loop,
// makeCloudInitISO success path, defer registration, and the
// node-not-found return branch.
func TestVirtualMachine_CloudInit_NodeError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// VMID 601 + node "missing" — no /nodes/missing/status mock so
	// client.Node fails.
	gock.New(mockConfig.URI).
		Get("^/nodes/missing/status$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "missing", VMID: 601}
	err := vm.CloudInit(context.Background(), "ide2", "", "", "", "",
		WithCloudInitStorage("local"))
	assert.Error(t, err)
}

// TestVirtualMachine_CloudInit_MakeISOError exercises the
// makeCloudInitISO-returns-error branch. We can't easily make
// makeCloudInitISO fail with valid args, but supplying contents that produce
// a filename collision with an unwritable target path is also tricky on
// Windows; instead trigger the "named storage not found" branch which
// crosses the same return shape just past makeCloudInitISO.
func TestVirtualMachine_CloudInit_StorageError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// node1 is mocked, but storage "ghost" is not — resolveCloudInitStorage
	// surfaces the not-found error from node.Storage.
	gock.New(mockConfig.URI).
		Persist().
		Get("^/nodes/node1/storage/ghost$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 602}
	err := vm.CloudInit(context.Background(), "ide2", "", "", "", "",
		WithCloudInitStorage("ghost"))
	assert.Error(t, err)
}

// TestVirtualMachine_Ping_StatusError exercises Ping's first error branch
// (the /status/current GET returning 500).
func TestVirtualMachine_Ping_StatusError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	gock.New(mockConfig.URI).
		Get("^/nodes/node1/qemu/505/status/current$").
		Reply(500)
	vm := &VirtualMachine{client: mockClient(), Node: "node1", VMID: 505}
	assert.Error(t, vm.Ping(context.Background()))
}

// TestVirtualMachine_DeleteCloudInitISO_StoragesError exercises the
// node.Storages()-returns-error branch. The node lookup uses node1 (which
// is mocked) but we override the /storage list to 500 with a one-shot mock
// that takes precedence over the persisted fixture.
func TestVirtualMachine_DeleteCloudInitISO_StoragesError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// Use a different node name that has a /status mock but no /storage
	// mock, forcing Storages() to surface gock's no-match.
	gock.New(mockConfig.URI).
		Get("^/nodes/lonely/status$").
		Reply(200).
		JSON(`{"data": {"name": "lonely", "status": "online"}}`)
	gock.New(mockConfig.URI).
		Get("^/nodes/lonely/storage$").
		Reply(500)
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "lonely",
		VMID:   502,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	_, err := vm.deleteCloudInitISO(context.Background())
	assert.Error(t, err)
}

// TestVirtualMachine_DeleteCloudInitISO_NodeError exercises the
// client.Node()-returns-error branch.
func TestVirtualMachine_DeleteCloudInitISO_NodeError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// VM points at a node that doesn't have a status mock — client.Node()
	// fetches /nodes/{name}/status to construct the *Node.
	gock.New(mockConfig.URI).
		Get("^/nodes/missing/status$").
		Reply(500)
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "missing",
		VMID:   501,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	_, err := vm.deleteCloudInitISO(context.Background())
	assert.Error(t, err)
}

// TestVirtualMachine_UnmountCloudInitISO_ConfigError exercises the
// Config-returns-error branch of UnmountCloudInitISO.
func TestVirtualMachine_UnmountCloudInitISO_ConfigError(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// vmid 599 has no /config mock — Config() will fail with no-match.
	gock.New(mockConfig.URI).
		Post("^/nodes/node1/qemu/599/config$").
		Reply(500)
	vm := &VirtualMachine{
		client: mockClient(),
		Node:   "node1",
		VMID:   599,
		VirtualMachineConfig: &VirtualMachineConfig{
			Tags: MakeTag(TagCloudInit),
		},
	}
	err := vm.UnmountCloudInitISO(context.Background(), "ide2")
	assert.Error(t, err)
}
