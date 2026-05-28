package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func diridxNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_Subdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxNode().Subdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "qemu")
	assert.Contains(t, subdirs, "lxc")
	assert.Contains(t, subdirs, "storage")
}

func TestNode_FirewallSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxNode().FirewallSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"rules", "options", "log"}, subdirs)
}

func TestNode_DisksSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxNode().DisksSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "list")
	assert.Contains(t, subdirs, "smart")
	assert.Contains(t, subdirs, "zfs")
}

func TestNodeReplicationJob_Subdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	job := &NodeReplicationJob{client: mockClient(), Node: "node1", ID: "101-0"}
	subdirs, err := job.Subdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"status", "log", "schedule_now"}, subdirs)

	empty := &NodeReplicationJob{client: mockClient(), Node: "node1"}
	_, err = empty.Subdirs(context.Background())
	assert.NotNil(t, err)
}

func TestNodeService_Subdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	svc := &NodeService{client: mockClient(), Node: "node1", Name: "pveproxy"}
	subdirs, err := svc.Subdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"state", "start", "stop", "restart", "reload"}, subdirs)

	empty := &NodeService{client: mockClient(), Node: "node1"}
	_, err = empty.Subdirs(context.Background())
	assert.NotNil(t, err)
}

func TestTask_Subdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task := NewTask("UPID:node1:00000002:00000002:00000002:test:completed:root@pam:", mockClient())
	subdirs, err := task.Subdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"log", "status"}, subdirs)

	empty := &Task{client: mockClient(), Node: "node1"}
	_, err = empty.Subdirs(context.Background())
	assert.NotNil(t, err)
}

func TestStorage_Status(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	storage := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	status, err := storage.Status(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "dir", status.Type)
	assert.Equal(t, 1, status.Active)
	assert.Equal(t, 1, status.Enabled)
	assert.Equal(t, "images,rootdir,vztmpl,backup,iso,snippets", status.Content)
	assert.Equal(t, uint64(60000000000), status.Total)
	assert.Equal(t, uint64(10000000000), status.Used)
	assert.Equal(t, uint64(50000000000), status.Avail)
}
