package proxmox

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

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
