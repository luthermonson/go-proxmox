package integration

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodes(t *testing.T) {
	client := ClientFromLogins()
	nodes, err := client.Nodes(context.TODO())
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(nodes), 1)
	for _, n := range nodes {
		assert.NotEmpty(t, n.Node)
		var node *proxmox.Node
		t.Run("get status for node "+n.Node, func(t *testing.T) {
			var err error
			node, err = client.Node(context.TODO(), n.Node)
			assert.Nil(t, err)
			assert.Equal(t, n.MaxMem, node.Memory.Total)
			assert.Equal(t, n.Disk, node.RootFS.Used)
		})

		t.Run("get VMs for node "+n.Node, func(t *testing.T) {
			_, err := node.VirtualMachines(context.TODO())
			assert.Nil(t, err)
		})

		break // only pull status from one node
	}

	_, err = client.Node(context.TODO(), "doesnt-exist")
	assert.Contains(t, err.Error(), "500 hostname lookup 'doesnt-exist' failed - failed to get address info for: doesnt-exist:")
}

func TestNode(t *testing.T) {
	client := ClientFromLogins()
	node, err := client.Node(context.TODO(), td.nodeName)
	assert.Nil(t, err)
	assert.Equal(t, node.Name, td.nodeName)
}

func TestContainers(t *testing.T) {
	t.Run("get Containers for node "+td.node.Name, func(t *testing.T) {
		_, err := td.node.Containers(context.TODO())
		assert.Nil(t, err)
	})
}

func TestNode_Appliances(t *testing.T) {
	t.Run("get Containers for node "+td.node.Name, func(t *testing.T) {
		aplinfos, err := td.node.Appliances(context.TODO())
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(aplinfos), 1)
	})
}

func TestNode_DownloadAppliance(t *testing.T) {
	var aplinfos proxmox.Appliances
	t.Run("get Containers for node "+td.node.Name, func(t *testing.T) {
		var err error
		aplinfos, err = td.node.Appliances(context.TODO())
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(aplinfos), 1)
	})

	t.Run("download non existing appliance template", func(t *testing.T) {
		_, err := td.node.DownloadAppliance(context.TODO(), "doesnt-exist", td.nodeStorage)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such template"))
	})

	if td.appliancePrefix == "" { // no point if no prefix to check for
		return
	}
	t.Run("download appliance "+td.appliancePrefix, func(t *testing.T) {
		for _, a := range aplinfos {
			if strings.HasPrefix(a.Template, td.appliancePrefix) {
				td.appliance = a // set to use in later tests
				ret, err := td.node.DownloadAppliance(context.TODO(), a.Template, td.nodeStorage)
				assert.Nil(t, err)
				assert.True(t, strings.HasPrefix(ret, fmt.Sprintf("UPID:%s:", td.node.Name)))
			}
		}
	})
}

func TestNode_Storages(t *testing.T) {
	storages, err := td.node.Storages(context.TODO())
	assert.Nil(t, err)
	assert.True(t, len(storages) > 0)

	for _, s := range storages {
		if s.Name == td.nodeStorage {
			assert.True(t, true, "storage exists: "+td.nodeStorage)
			return
		}
	}

	assert.True(t, false, "no storage: "+td.nodeStorage)
}

func TestNode_Storage(t *testing.T) {
	_, err := td.node.Storage(context.TODO(), "doesnt-exist")
	assert.Contains(t, err.Error(), "No such storage.")

	storage, err := td.node.Storage(context.TODO(), td.nodeStorage)
	assert.Nil(t, err)
	assert.Equal(t, td.nodeStorage, storage.Name)
}

func TestNode_TermProxy(t *testing.T) {
	term, err := td.node.TermProxy(context.TODO())
	assert.Nil(t, err)
	send, recv, errs, close, err := td.node.TermWebSocket(term)
	assert.Nil(t, err)
	defer func() {
		assert.NoError(t, close())
	}()

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

	send <- []byte("ls -la\n")
	time.Sleep(1 * time.Second)
	send <- []byte("hostname\n")
	time.Sleep(1 * time.Second)
	send <- []byte("exit\n")
	time.Sleep(1 * time.Second)
}

func TestNode_VncProxy(t *testing.T) {
	assert.NotEqual(t, 0, td.vncVmId)

	vm, err := td.node.VirtualMachine(context.TODO(), td.vncVmId)
	require.NoError(t, err)

	vnc, err := vm.VNCProxy(context.TODO(), nil)
	require.NoError(t, err)

	send, recv, errs, close, err := vm.VNCWebSocket(vnc)
	assert.Nil(t, err)
	defer func() {
		assert.NoError(t, close())
	}()

	go func() {
		for {
			select {
			case msg := <-recv:
				if len(msg) > 0 {
					fmt.Printf("MSG: %s -> %v\n", string(msg), msg)
					if strings.HasPrefix(string(msg), "RFB") {
						send <- msg
					}
					if bytes.Equal(msg, []byte{0x01, 0x02}) {
						fmt.Println("Success!")
					}
				}
			case err := <-errs:
				if err != nil {
					fmt.Println("ERROR: " + err.Error())
					return
				}
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
