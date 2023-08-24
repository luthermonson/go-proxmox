package proxmox

import (
	"testing"

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
