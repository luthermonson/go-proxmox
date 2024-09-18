//go:build nodes
// +build nodes

package integration

import (
	"context"
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
	node, err := client.Node(context.TODO(), td.nodeName)
	require.NoError(t, err)

	isoName := name + ".iso"
	task, err := td.storage.DownloadURL(context.TODO(), "iso", isoName, tinycoreURL)
	assert.Nil(t, err)
	assert.Nil(t, task.Wait(context.TODO(), time.Duration(5*time.Second), time.Duration(5*time.Minute)))

	iso, err := td.storage.ISO(context.TODO(), isoName)
	assert.Nil(t, err)

	cluster, err := client.Cluster(context.TODO())
	require.NoError(t, err)

	nextid, err := cluster.NextID(context.TODO())
	require.NoError(t, err)

	task, err = node.NewVirtualMachine(context.TODO(), nextid, proxmox.VirtualMachineOption{Name: "cdrom", Value: iso.VolID})
	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 10*time.Second))

	vm, err := node.VirtualMachine(context.TODO(), nextid)
	require.NoError(t, err)
	task, err = vm.Config(context.TODO(),
		proxmox.VirtualMachineOption{Name: "name", Value: name},
		proxmox.VirtualMachineOption{Name: "serial0", Value: "socket"},
	)

	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 10*time.Second))

	return vm
}

func CleanupVirtualMachine(t *testing.T, vm *proxmox.VirtualMachine) {
	task, err := vm.Stop(context.TODO())
	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 30*time.Second))

	if vm.VirtualMachineConfig != nil && vm.VirtualMachineConfig.IDE2 != "" {
		s := strings.Split(vm.VirtualMachineConfig.IDE2, ",")
		if len(s) > 2 {
			iso, err := td.storage.ISO(context.TODO(), filepath.Base(s[0]))
			assert.Nil(t, err)
			task, err := iso.Delete(context.TODO())
			require.NoError(t, err)
			require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 10*time.Second))
		}
	}

	task, err = vm.Delete(context.TODO())
	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 30*time.Second))
}

func TestNode_NewVirtualMachine(t *testing.T) {
	testname := nameGenerator(12)
	vm := NewVirtualMachine(t, testname)
	require.NotNil(t, vm)
	defer CleanupVirtualMachine(t, vm)

	// Start
	task, err := vm.Start(context.TODO())
	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 10*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

	// TODO while this connects it doesn't do anything because the vm isn't setup to use the serial0 socket
	term, err := vm.TermProxy(context.TODO())
	require.NoError(t, err)
	send, recv, errs, close, err := vm.TermWebSocket(term)
	defer close()

	go func() {
		for {
			select {
			case msg := <-recv:
				if len(msg) > 0 {
					fmt.Println("MSG: " + string(msg))
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
	send <- []byte("\n")
	time.Sleep(2 * time.Second)
	send <- []byte("ls -la\n")
	time.Sleep(2 * time.Second)
	send <- []byte("hostname\n")
	time.Sleep(2 * time.Second)

	// Reboot disabled for now doesn't work great w/o the guest agent installed so will uncomment when that's done
	//task, err = vm.Reboot()
	//assert.NoError(t, err)
	//assert.NoError(t, task.Wait(1*time.Second, 60*time.Second))
	//require.NoError(t, vm.Ping())
	//assert.Equal(t, StatusVirtualMachineRunning, vm.Status)

	// Stop
	task, err = vm.Stop(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(context.TODO(), 1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.Equal(t, proxmox.StatusVirtualMachineStopped, vm.Status)

	// Start again to test hibernating/pause and resumse
	task, err = vm.Start(context.TODO())
	require.NoError(t, err)
	require.NoError(t, task.Wait(context.TODO(), 1*time.Second, 30*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

	// Hibernate
	task, err = vm.Hibernate(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(context.TODO(), 1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.True(t, vm.IsHibernated())

	task, err = vm.Resume(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(context.TODO(), 1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)

	// Pause
	task, err = vm.Pause(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(context.TODO(), 1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.True(t, vm.IsPaused())

	task, err = vm.Resume(context.TODO())
	assert.NoError(t, err)
	assert.NoError(t, task.Wait(context.TODO(), 1*time.Second, 15*time.Second))
	require.NoError(t, vm.Ping(context.TODO()))
	assert.Equal(t, proxmox.StatusVirtualMachineRunning, vm.Status)
}
