// +build vms

package proxmox

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	testVirtualMachineID   = 0
)

func TestVMStart(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != "stopped" {
		_, err = vm.Stop()
		require.NoError(t, err)
		vm, err = node.VirtualMachine(testVirtualMachineID)
		require.NoError(t, err)
		require.Equal(t, "stopped", vm.Status)
	}

	_, err = vm.Start()
	assert.NoError(t, err)

	vm, err = node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)
	assert.Equal(t, "running", vm.Status)
}

func TestVMStop(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != "running" {
		_, err = vm.Start()
		require.NoError(t, err)
		vm, err = node.VirtualMachine(testVirtualMachineID)
		require.NoError(t, err)
		require.Equal(t, "running", vm.Status)
	}

	_, err = vm.Stop()
	assert.NoError(t, err)

	vm, err = node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)
	assert.Equal(t, "stopped", vm.Status)
}

func TestVMReboot(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != "running" {
		_, err = vm.Start()
		require.NoError(t, err)
		vm, err = node.VirtualMachine(testVirtualMachineID)
		require.NoError(t, err)
		require.Equal(t, "running", vm.Status)
	}

	_, err = vm.Reboot()
	assert.NoError(t, err)

	vm, err = node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)
	assert.Equal(t, "running", vm.Status)
}

func TestVMSuspendAndResume(t *testing.T) {
	if testVirtualMachineID == 0 || td.nodeName == "" {
		t.Skip()
	}
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	vm, err := node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)

	if vm.Status != "running" {
		_, err = vm.Start()
		require.NoError(t, err)
		vm, err = node.VirtualMachine(testVirtualMachineID)
		require.NoError(t, err)
		require.Equal(t, "running", vm.Status)
	}

	_, err = vm.Suspend()
	assert.NoError(t, err)

	vm, err = node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)
	assert.Equal(t, "suspended", vm.Status)

	_, err = vm.Resume()
	assert.NoError(t, err)

	vm, err = node.VirtualMachine(testVirtualMachineID)
	require.NoError(t, err)
	assert.Equal(t, "running", vm.Status)
}
