package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

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
	_, _, err := container.Clone(ctx, &cloneOptions)
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
	task, err := container.Delete(ctx)
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
	_, err := container.Config(ctx)
	assert.Nil(t, err)
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
