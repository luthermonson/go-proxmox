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

// NewVirtualMachine fixture to download a tinycore iso and returns a vm, make sure you defer the cleanup if you use it
func NewVirtualMachine(t *testing.T, name string) (*VirtualMachine, error) {
	client := ClientFromLogins()
	node, err := client.Node(td.nodeName)
	require.NoError(t, err)

	isoName := name + ".iso"
	task, err := td.storage.DownloadURL("iso", isoName, tinycoreURL)
	assert.Nil(t, err)
	assert.Nil(t, task.Wait(time.Duration(5*time.Second), time.Duration(5*time.Minute)))

	iso, err := td.storage.ISO(isoName)
	assert.Nil(t, err)

	cluster, err := client.Cluster()
	require.NoError(t, err)

	nextid, err := cluster.NextID()
	require.NoError(t, err)

	task, err = node.NewVirtualMachine(nextid, VirtualMachineOption{Name: "ide2", Value: iso.VolID})
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	vm, err := node.VirtualMachine(nextid)
	require.NoError(t, err)
	task, err = vm.Config(VirtualMachineOption{Name: "name", Value: name})
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	return vm, nil
}

func CleanupVirtualMachine(t *testing.T, vm *VirtualMachine) {
	if vm.VirtualMachineConfig != nil && vm.VirtualMachineConfig.Ide2 != "" {
		s := strings.Split(vm.VirtualMachineConfig.Ide2, ",")
		if len(s) > 2 {
			iso, err := td.storage.ISO(s[0])
			assert.Nil(t, err)
			task, err := iso.Delete()
			require.NoError(t, err)
			require.NoError(t, task.Wait(1*time.Second, 10*time.Second))
		}
	}

	task, err := vm.Stop()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))

	task, err = vm.Delete()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))
}

func TestNode_NewVirtualMachine(t *testing.T) {
	testname := nameGenerator(12)
	vm, err := NewVirtualMachine(t, testname)
	require.NoError(t, err)
	defer CleanupVirtualMachine(t, vm)

	// Start
	task, err := vm.Start()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)

	// Reboot disabled for now doesnt work great w/o the guest agent installed so will uncomment when that's done
	//task, err = vm.Reboot()
	//assert.NoError(t, err)
	//assert.NoError(t, task.Wait(1*time.Second, 60*time.Second))
	//require.NoError(t, vm.Ping())
	//assert.Equal(t, StatusVirtualMachineRunning, vm.Status)

	// Stop
	task, err = vm.Stop()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineStopped, vm.Status)

	// Start again to test hibernating/pause and resumse
	task, err = vm.Start()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)

	// Hibernate
	task, err = vm.Hibernate()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.True(t, vm.IsHibernated())

	task, err = vm.Resume()
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, StatusVirtualMachineRunning, vm.Status)

	// Pause
	task, err = vm.Pause()
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
