package proxmox

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTask_Wait_Completed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// task is already stopped -> Wait returns nil immediately after first ping.
	task := NewTask(UPID("UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:"), client)
	require.NoError(t, task.Wait(ctx, 10*time.Millisecond, 1*time.Second))
}

func TestTask_WaitFor_Completed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:"), client)
	require.NoError(t, task.WaitFor(ctx, 1))
}

func TestTask_WaitForCompleteStatus(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:"), client)
	status, completed, err := task.WaitForCompleteStatus(ctx, 2, 1)
	require.NoError(t, err)
	assert.True(t, completed)
	assert.True(t, status)
}

func TestTask_Watch(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	task := NewTask(UPID("UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:"), client)
	watch, err := task.Watch(ctx, 0)
	require.NoError(t, err)
	require.NotNil(t, watch)

	// Drain the channel — task is already stopped so the tail loop exits
	// after the first ping and closes the channel.
	var lines []string
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ln, ok := <-watch:
			if !ok {
				assert.GreaterOrEqual(t, len(lines), 2)
				return
			}
			lines = append(lines, ln)
		case <-timeout:
			t.Fatalf("timed out waiting for watcher to close; got %d lines", len(lines))
		}
	}
}

func TestTask_Watch_NoLogs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// This task's log endpoint returns []; Watch retries 3x with a 1s sleep
	// then returns the "no logs available" error. Total wall-time is ~3s.
	task := NewTask(UPID("UPID:node1:000000AA:000000AA:000000AA:emptylog:AA:root@pam:"), client)
	watch, err := task.Watch(ctx, 0)
	assert.Nil(t, watch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no logs available")
}

// --- Wait reports failed tasks as errors ------------------------------------

const (
	upidRunning   = UPID("UPID:node1:00000001:00000001:00000001:test:running:root@pam:")
	upidOK        = UPID("UPID:node1:00000002:00000002:00000002:test:completed:root@pam:")
	upidFailed    = UPID("UPID:node1:00000003:00000003:00000003:test:failed:root@pam:")
	upidWarnings  = UPID("UPID:node1:00000004:00000004:00000004:test:warnings:root@pam:")
	upidHistFail  = UPID("UPID:taskhist:0000A001:0000A001:65000000:qmshutdown:100:root@pam:")
	upidHistOK    = UPID("UPID:taskhist:0000A002:0000A002:65000010:qmstart:100:root@pam:")
	upidHistGone  = UPID("UPID:taskhist:0000A003:0000A003:65000020:qmstart:101:root@pam:")
	shutdownError = "VM quit/powerdown failed - got timeout"
)

func TestTask_Wait_OK(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	task := NewTask(upidOK, client)
	require.NoError(t, task.Wait(context.Background(), 10*time.Millisecond, time.Second))
	assert.True(t, task.IsSuccessful)
	assert.False(t, task.IsFailed)
	assert.Zero(t, task.Warnings)
}

func TestTask_Wait_Failed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	task := NewTask(upidFailed, client)
	err := task.Wait(context.Background(), 10*time.Millisecond, time.Second)
	require.Error(t, err)
	assert.True(t, IsTaskFailed(err))

	var tfe *TaskFailedError
	require.True(t, errors.As(err, &tfe))
	assert.Equal(t, upidFailed, tfe.UPID)
	assert.Equal(t, "node1", tfe.Node)
	assert.Equal(t, "test", tfe.Type)
	assert.Equal(t, "failed", tfe.ID)
	assert.Equal(t, "root@pam", tfe.User)
	assert.Equal(t, "some error occurred", tfe.ExitStatus)
	assert.Contains(t, err.Error(), "some error occurred")
	assert.Contains(t, err.Error(), string(upidFailed))

	// The flags still agree with the error.
	assert.True(t, task.IsCompleted)
	assert.True(t, task.IsFailed)
	assert.False(t, task.IsSuccessful)
}

func TestTask_WaitFor_Failed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	err := NewTask(upidFailed, client).WaitFor(context.Background(), 1)
	assert.True(t, IsTaskFailed(err))
}

func TestTask_Wait_FailedErrorWraps(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	err := NewTask(upidFailed, client).WaitFor(context.Background(), 1)
	wrapped := fmt.Errorf("shutting down vm 100: %w", err)

	var tfe *TaskFailedError
	require.True(t, errors.As(wrapped, &tfe))
	assert.Equal(t, "some error occurred", tfe.ExitStatus)
	assert.True(t, IsTaskFailed(wrapped))
	assert.False(t, IsTaskFailed(ErrTimeout))
	assert.False(t, IsTaskFailed(nil))
}

func TestTask_Wait_Warnings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	task := NewTask(upidWarnings, client)
	require.NoError(t, task.Wait(context.Background(), 10*time.Millisecond, time.Second))
	assert.Equal(t, "WARNINGS: 2", task.ExitStatus)
	assert.Equal(t, 2, task.Warnings)
	assert.True(t, task.IsSuccessful)
	assert.False(t, task.IsFailed)
}

func TestTask_Ping_Warnings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	task := NewTask(upidWarnings, client)
	require.NoError(t, task.Ping(context.Background()))
	assert.True(t, task.IsCompleted)
	assert.True(t, task.IsSuccessful)
	assert.False(t, task.IsFailed)
	assert.Equal(t, 2, task.Warnings)
}

func TestTask_Wait_Timeout(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	err := NewTask(upidRunning, client).Wait(context.Background(), 10*time.Millisecond, 50*time.Millisecond)
	assert.ErrorIs(t, err, ErrTimeout)
	assert.False(t, IsTaskFailed(err))
}

func TestTask_Wait_ContextCanceled(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	// The poll interval is far longer than the cancel delay, so returning
	// promptly proves Wait selects on ctx.Done() instead of sleeping.
	start := time.Now()
	err := NewTask(upidRunning, client).Wait(ctx, 10*time.Second, time.Minute)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestTask_Wait_ContextDeadline(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := NewTask(upidRunning, client).Wait(ctx, 10*time.Second, time.Minute)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestTask_Wait_AlreadyCanceled(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewTask(upidRunning, client).Wait(ctx, 10*time.Millisecond, time.Second)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestTask_WaitForCompleteStatus_Failed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	status, completed, err := NewTask(upidFailed, client).WaitForCompleteStatus(context.Background(), 2, 1)
	assert.True(t, completed)
	assert.False(t, status)
	var tfe *TaskFailedError
	require.True(t, errors.As(err, &tfe))
	assert.Equal(t, "some error occurred", tfe.ExitStatus)
}

func TestTask_WaitForCompleteStatus_Warnings(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	status, completed, err := NewTask(upidWarnings, client).WaitForCompleteStatus(context.Background(), 2, 1)
	require.NoError(t, err)
	assert.True(t, completed)
	assert.True(t, status)
}

func TestTask_WaitForCompleteStatus_ContextCanceled(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	status, completed, err := NewTask(upidRunning, client).WaitForCompleteStatus(ctx, 60, 2)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, completed)
	assert.False(t, status)
}

func TestTask_Wait_NoSuchTask_HistoryFailed(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	// The status endpoint has forgotten the task; the node task index still
	// records the failed shutdown, so Wait must report it, not succeed.
	task := NewTask(upidHistFail, client)
	err := task.Wait(context.Background(), 10*time.Millisecond, time.Second)

	var tfe *TaskFailedError
	require.True(t, errors.As(err, &tfe), "got %v", err)
	assert.Equal(t, upidHistFail, tfe.UPID)
	assert.Equal(t, "taskhist", tfe.Node)
	assert.Equal(t, "qmshutdown", tfe.Type)
	assert.Equal(t, "100", tfe.ID)
	assert.Equal(t, shutdownError, tfe.ExitStatus)
	assert.Equal(t, TaskStopped, task.Status)
	assert.True(t, task.IsFailed)
}

func TestTask_Wait_NoSuchTask_HistoryOK(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	task := NewTask(upidHistOK, client)
	require.NoError(t, task.Wait(context.Background(), 10*time.Millisecond, time.Second))
	assert.Equal(t, TaskExitOK, task.ExitStatus)
	assert.True(t, task.IsSuccessful)
	assert.Equal(t, 8*time.Second, task.Duration)
}

func TestTask_Wait_NoSuchTask_NotInHistory(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	// Gone from both the status endpoint and the history: surface the
	// original "no such task" error rather than guessing.
	err := NewTask(upidHistGone, client).Wait(context.Background(), 10*time.Millisecond, time.Second)
	require.Error(t, err)
	assert.False(t, IsTaskFailed(err))
	assert.Contains(t, err.Error(), "no such task")
}

func TestTaskFailedError_Error(t *testing.T) {
	err := &TaskFailedError{
		UPID:       upidHistFail,
		Node:       "taskhist",
		Type:       "qmshutdown",
		ID:         "100",
		ExitStatus: shutdownError,
	}
	assert.Equal(t, "task "+string(upidHistFail)+" (qmshutdown 100 on taskhist) failed: "+shutdownError, err.Error())

	bare := &TaskFailedError{UPID: "UPID:x"}
	assert.Equal(t, "task UPID:x failed: unknown exit status", bare.Error())
}

func TestUPIDStartTime(t *testing.T) {
	start, ok := upidStartTime(upidHistFail)
	require.True(t, ok)
	assert.Equal(t, int64(0x65000000), start)

	_, ok = upidStartTime("garbage")
	assert.False(t, ok)
}
