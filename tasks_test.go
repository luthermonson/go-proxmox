//go:build nodes
// +build nodes

package proxmox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTask(t *testing.T) {
	upid := "UPID:test:002F0193:09CCCA13:61CC858A:tasktype:taskid:root@pam:"
	task := NewTask(upid, td.client)
	assert.Equal(t, task.Node, "test")
	assert.Equal(t, task.Type, "tasktype")
	assert.Equal(t, task.ID, "taskid")
	assert.Equal(t, task.User, "root@pam")
}

func TestTask_JsonUnmarshalNoEndTime(t *testing.T) {
	// no duration from a not completed task
	data := `{"pstart":165231870,"type":"testtype","status":"teststatus","id":"test.iso","node":"testnode","user":"root@pam","pid":3161937,"upid":"UPID:i7:00303F51:09D93CFE:61CCA568:download:8fd77349e9f6.iso:root@pam:","starttime":1641020400}`
	starttime := time.Date(2022, time.January, 01, 0, 0, 0, 0, time.Local)

	var task Task
	assert.Nil(t, json.Unmarshal([]byte(data), &task))
	assert.Equal(t, "root@pam", task.User)
	assert.Equal(t, "teststatus", task.Status)
	assert.Equal(t, "testtype", task.Type)
	assert.Equal(t, "test.iso", task.ID)
	assert.Equal(t, "testnode", task.Node)
	assert.Equal(t, starttime, task.StartTime)
	assert.True(t, task.EndTime.IsZero()) // empty endtime
	assert.Equal(t, float64(0), task.Duration.Seconds())
}

func TestTask_JsonUnmarshalWithEndTime(t *testing.T) {
	// with an endtime from the cluster wide /cluster/tasks status log and will calc duration
	data := `{"pstart":165231870,"type":"testtype","status":"teststatus","id":"test.iso","node":"testnode","user":"root@pam","pid":3161937,"upid":"UPID:i7:00303F51:09D93CFE:61CCA568:download:8fd77349e9f6.iso:root@pam:","starttime":1641020400, "endtime":1641020460}`
	starttime := time.Date(2022, time.January, 01, 0, 0, 0, 0, time.Local)
	endtime := time.Date(2022, time.January, 01, 0, 1, 0, 0, time.Local)

	var task Task
	assert.Nil(t, json.Unmarshal([]byte(data), &task))
	assert.Equal(t, "root@pam", task.User)
	assert.Equal(t, "teststatus", task.Status)
	assert.Equal(t, "testtype", task.Type)
	assert.Equal(t, "test.iso", task.ID)
	assert.Equal(t, "testnode", task.Node)
	assert.Equal(t, starttime, task.StartTime)
	assert.Equal(t, endtime, task.EndTime)
	assert.Equal(t, float64(60), task.Duration.Seconds())
}

// TestTask will start a download of a large iso, tail the logs and cancel it
func TestTask(t *testing.T) {
	// download ubuntu iso for long running task to test against
	isoName := randomString(12) + ".iso"
	task, err := td.storage.DownloadURL("iso", isoName, ubuntuURL)
	assert.Nil(t, err)

	// test ping and wait, big iso should take more than 15s
	go func() {
		timeout := task.Wait(time.Duration(5*time.Second), time.Duration(30*time.Second))
		assert.Contains(t, timeout.Error(), "timeout while waiting for task")
		assert.Nil(t, task.Stop())
	}()

	log, err := task.Log(0, 50)
	assert.Nil(t, err)
	assert.Contains(t, log[0], ubuntuURL)

	watch, err := task.Watch(0)
	assert.Nil(t, err)
	for {
		select {
		case ln, ok := <-watch:
			if !ok {
				watch = nil
				break
			}
			task.client.log.Debugf("%s", ln)
		}
		if watch == nil {
			break
		}
	}
}
