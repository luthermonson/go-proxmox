//go:build nodes
// +build nodes

package proxmox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testVirtualMachineID = 113
)

func TestVirtualMachine_Start(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != StatusVirtualMachineStopped {
		task, err := vm.Stop()
		require.NoError(t, err)
		require.NoError(t, task.Wait(1*time.Second, 5*time.Second))
		require.NoError(t, vm.Ping())
		require.Equal(t, "stopped", vm.Status)
	}

	task, err := vm.Start()
	assert.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)
}

func TestVirtualMachine_Stop(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != StatusVirtualMachineRunning {
		task, err := vm.Start()
		require.NoError(t, err)
		require.NoError(t, task.Wait(1*time.Second, 15*time.Second))
		require.NoError(t, vm.Ping())
		require.Equal(t, StatusVirtualMachineRunning, vm.Status)
	}

	task, err := vm.Stop()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineStopped, vm.Status)
}

func TestVirtualMachine_Reboot(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != StatusVirtualMachineRunning {
		task, err := vm.Start()
		require.NoError(t, err)
		assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
		require.NoError(t, vm.Ping())
		require.Equal(t, StatusVirtualMachineRunning, vm.Status)
	}

	task, err := vm.Reboot()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 30*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)
}

func TestVirtualMachine_Hibernate(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != StatusVirtualMachineRunning {
		task, err := vm.Start()
		require.NoError(t, err)
		assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
		require.NoError(t, vm.Ping())
		require.Equal(t, StatusVirtualMachineRunning, vm.Status)
	}

	task, err := vm.Hibernate()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.True(t, vm.IsHibernated())

	task, err = vm.Resume()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)
}

func TestVirtualMachine_Pause(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != StatusVirtualMachineRunning {
		task, err := vm.Start()
		require.NoError(t, err)
		assert.NoError(t, task.Wait(1*time.Second, 5*time.Second))
		require.NoError(t, vm.Ping())
		require.Equal(t, StatusVirtualMachineRunning, vm.Status)
	}

	task, err := vm.Pause()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.True(t, vm.IsPaused())

	task, err = vm.Resume()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)
}
