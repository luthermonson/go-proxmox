//go:build nodes
// +build nodes

package proxmox

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testVirtualMachineID = 113
)

func TestNode_NewVirtualMachine(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	isoName := nameGenerator(12) + ".iso"
	task, err := td.storage.DownloadURL("iso", isoName, tinycoreURL)
	assert.Nil(t, err)
	assert.Nil(t, task.Wait(time.Duration(5*time.Second), time.Duration(5*time.Minute)))

	iso, err := td.storage.ISO(isoName)
	assert.Nil(t, err)

	cluster, err := client.Cluster()
	require.NoError(t, err)

	nextid, err := cluster.NextID()
	require.NoError(t, err)

	task, err = node.NewVirtualMachine(nextid)
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	vm, err := node.VirtualMachine(nextid)
	require.NoError(t, err)
	task, err = vm.Delete()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	assert.True(t, strings.HasSuffix(iso.Path, isoName))
	assert.Nil(t, iso.Delete())
}

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
