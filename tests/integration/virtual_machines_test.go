//go:build nodes
// +build nodes

package integration

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luthermonson/go-proxmox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewVirtualMachine fixture to download a tinycore iso and returns a vm, make sure you defer the cleanup if you use it
func NewVirtualMachine(t *testing.T, name string) *proxmox.VirtualMachine {
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

	task, err = node.NewVirtualMachine(nextid, proxmox.VirtualMachineOption{Name: "cdrom", Value: iso.VolID})
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	vm, err := node.VirtualMachine(nextid)
	require.NoError(t, err)
	task, err = vm.Config(
		proxmox.VirtualMachineOption{Name: "name", Value: name},
		proxmox.VirtualMachineOption{Name: "serial0", Value: "socket"},
	)

	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))

	return vm
}

func CleanupVirtualMachine(t *testing.T, vm *proxmox.VirtualMachine) {
	task, err := vm.Stop()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))

	if vm.VirtualMachineConfig != nil && vm.VirtualMachineConfig.IDE2 != "" {
		s := strings.Split(vm.VirtualMachineConfig.IDE2, ",")
		if len(s) > 2 {
			iso, err := td.storage.ISO(filepath.Base(s[0]))
			assert.Nil(t, err)
			task, err := iso.Delete()
			require.NoError(t, err)
			require.NoError(t, task.Wait(1*time.Second, 10*time.Second))
		}
	}

	task, err = vm.Delete()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))
}

func TestNode_NewVirtualMachine(t *testing.T) {
	testname := nameGenerator(12)
	vm := NewVirtualMachine(t, testname)
	require.NotNil(t, vm)
	defer CleanupVirtualMachine(t, vm)

	// Start
	task, err := vm.Start()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 10*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

	// TODO while this connects it doesn't do anything because the vm isn't setup to use the serial0 socket
	vnc, err := vm.TermProxy()
	require.NoError(t, err)
	send, recv, errs, close, err := vm.VNCWebSocket(vnc)
	defer close()

	go func() {
		for {
			select {
			case msg := <-recv:
				if msg != "" {
					fmt.Println("MSG: " + msg)
				}
			case err := <-errs:
				if err != nil {
					fmt.Println("ERROR: " + err.Error())
					return
				}
			}
		}
	}()

	time.Sleep(2 * time.Second)
	send <- "\n"
	time.Sleep(2 * time.Second)
	send <- "ls -la"
	time.Sleep(2 * time.Second)
	send <- "hostname"
	time.Sleep(2 * time.Second)

	// Reboot disabled for now doesn't work great w/o the guest agent installed so will uncomment when that's done
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
	assert.Equal(t, proxmox.StatusVirtualMachineStopped, vm.Status)

	// Start again to test hibernating/pause and resumse
	task, err = vm.Start()
	require.NoError(t, err)
	require.NoError(t, task.Wait(1*time.Second, 30*time.Second))
	require.NoError(t, vm.Ping())
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

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
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

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
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)
}
