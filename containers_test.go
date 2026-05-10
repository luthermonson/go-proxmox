package proxmox

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestContainer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}
	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)
	assert.NotEmpty(t, container, container.ContainerConfig)
}

func TestContainers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}
	containers, err := node.Containers(ctx)
	assert.Nil(t, err)
	assert.Len(t, containers, 3)
}

func TestContainerClone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	cloneOptions := ContainerCloneOptions{
		NewID: 102,
	}
	newid, _, err := container.Clone(ctx, &cloneOptions)
	assert.Equal(t, cloneOptions.NewID, newid)
	assert.Nil(t, err)
}

func TestContainerCloneWithoutNewID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	cloneOptions := ContainerCloneOptions{}
	newid, _, err := container.Clone(ctx, &cloneOptions)
	assert.Equal(t, 100, newid)
	assert.Nil(t, err)
}

func TestContainerDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(ctx, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

// TestContainerDelete_ForceParam verifies issue #149: passing
// ContainerDeleteOptions{Force: true} actually puts force=1 on the wire as a
// query parameter to DELETE /nodes/{node}/lxc/{vmid}. The gock mock only
// matches when the parameter is present, so a regression where the option is
// silently dropped (as DeleteFirewallIPSet currently does, by passing the map
// as the response target instead of via DeleteWithParams) makes this test fail
// with "cannot match any request".
func TestContainerDelete_ForceParam(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Delete("^/nodes/node1/lxc/101$").
		MatchParam("force", "1").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdestroy:101:root@pam:"}`)

	client := mockClient()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(context.Background(), &ContainerDeleteOptions{Force: true})
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

// TestContainerDelete_AllOptionsOnWire is the equivalent for purge and
// destroy-unreferenced-disks: all three options must appear in the query
// string when set, with the spec-correct hyphenated key for the last one.
func TestContainerDelete_AllOptionsOnWire(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Delete("^/nodes/node1/lxc/101$").
		MatchParams(map[string]string{
			"force":                      "1",
			"purge":                      "1",
			"destroy-unreferenced-disks": "1",
		}).
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdestroy:101:root@pam:"}`)

	client := mockClient()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(context.Background(), &ContainerDeleteOptions{
		Force:                    true,
		Purge:                    true,
		DestroyUnreferencedDisks: true,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Config(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerStart(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Start(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerStop(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Stop(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerSuspend(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Suspend(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerReboot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Reboot(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerResume(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Resume(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerShutdown(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Shutdown(ctx, false, 60)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerTemplate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	err := container.Template(ctx)
	assert.Nil(t, err)
}

func TestContainerSnapshots(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	snapshots, err := container.Snapshots(ctx)
	assert.Nil(t, err)
	assert.Len(t, snapshots, 3)
}

func TestContainerNewSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.NewSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerGetSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	snapshot, err := container.GetSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotNil(t, snapshot)
}

func TestContainerDeleteSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.DeleteSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerRollbackSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.RollbackSnapshot(ctx, "snapshot1", true)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerInterfaces(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	interfaces, err := container.Interfaces(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, interfaces)
	assert.Equal(t, interfaces, ContainerInterfaces{{HWAddr: "00:00:00:00:00:00", Inet: "127.0.0.1/8", Name: "lo", Inet6: "::1/128"}, {Inet6: "fe80::be24:11ff:fe89:6707/64", Name: "eth0", HWAddr: "bc:24:11:89:67:07", Inet: "192.168.3.95/22"}})
}

func TestContainerTagsSlice(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.NotEmpty(t, container.ContainerConfig.TagsSlice)
}

func TestContainer_AddTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	_, err = container.AddTag(ctx, "newTag")
	assert.Nil(t, err)
	assert.True(t, container.HasTag("newTag"))
}

func TestContainer_HasTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.True(t, container.HasTag("tag1"))
	assert.False(t, container.HasTag("not_there"))
}

func TestContainer_RemoveTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.True(t, container.HasTag("tag1"))
	_, err = container.RemoveTag(ctx, "tag1")
	assert.Nil(t, err)
	assert.False(t, container.HasTag("tag1"))
}
