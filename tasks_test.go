package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	upid := NewTask("", &Client{})
	assert.Nil(t, upid)

	task := NewTask(UPID("UPID:nodename:00388B23:02D69651:63C4F6AF:tasktype:100:root@pam:"), &Client{})
	assert.Equal(t, "nodename", task.Node)
	assert.Equal(t, "100", task.ID)
	assert.Equal(t, "tasktype", task.Type)
	assert.Equal(t, "root@pam", task.User)
	assert.Equal(t, UPID("UPID:nodename:00388B23:02D69651:63C4F6AF:tasktype:100:root@pam:"), task.UPID)
}

func TestTask_Ping_Running(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000001:00000001:00000001:test:running:root@pam:"), client)
	err := task.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "running", task.Status)
	assert.True(t, task.IsRunning)
	assert.False(t, task.IsCompleted)
	assert.False(t, task.IsSuccessful)
	assert.False(t, task.IsFailed)
}

func TestTask_Ping_Completed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000002:00000002:00000002:test:completed:root@pam:"), client)
	err := task.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "stopped", task.Status)
	assert.Equal(t, "OK", task.ExitStatus)
	assert.False(t, task.IsRunning)
	assert.True(t, task.IsCompleted)
	assert.True(t, task.IsSuccessful)
	assert.False(t, task.IsFailed)
}

func TestTask_Ping_Failed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000003:00000003:00000003:test:failed:root@pam:"), client)
	err := task.Ping(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "stopped", task.Status)
	assert.Equal(t, "some error occurred", task.ExitStatus)
	assert.False(t, task.IsRunning)
	assert.True(t, task.IsCompleted)
	assert.False(t, task.IsSuccessful)
	assert.True(t, task.IsFailed)
}

func TestTask_Stop(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000001:00000001:00000001:test:running:root@pam:"), client)
	err := task.Stop(ctx)
	assert.Nil(t, err)
}

func TestTask_Log(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000002:00000002:00000002:test:completed:root@pam:"), client)
	log, err := task.Log(ctx, 0, 50)
	assert.Nil(t, err)
	assert.NotNil(t, log)
	assert.Len(t, log, 5)
	assert.Equal(t, "task started", log[0])
	assert.Equal(t, "task completed successfully", log[4])
}

func TestTask_Log_WithOffset(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000002:00000002:00000002:test:completed:root@pam:"), client)
	log, err := task.Log(ctx, 5, 50)
	assert.Nil(t, err)
	assert.NotNil(t, log)
	assert.Len(t, log, 0) // No more logs after offset 5
}

func TestTask_Log_Running(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000001:00000001:00000001:test:running:root@pam:"), client)
	log, err := task.Log(ctx, 0, 50)
	assert.Nil(t, err)
	assert.NotNil(t, log)
	assert.Len(t, log, 2)
	assert.Equal(t, "task started", log[0])
	assert.Equal(t, "processing...", log[1])
}

