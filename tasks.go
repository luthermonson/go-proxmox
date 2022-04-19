package proxmox

import (
	"fmt"
	"strings"
	"time"
)

const (
	TaskRunning = "running"
)

var DefaultWaitInterval = 1 * time.Second

func NewTask(upid UPID, client *Client) *Task {
	if upid == "" {
		return nil
	}

	task := &Task{
		UPID:   upid,
		client: client,
	}

	sp := strings.Split(string(task.UPID), ":")
	if len(sp) == 0 || len(sp) < 7 {
		return task
	}

	task.Node = sp[1]
	task.Type = sp[5]
	task.ID = sp[6]
	task.User = sp[7]

	return task
}

func (t *Task) Ping() error {
	tmp := NewTask(t.UPID, t.client)
	err := t.client.Get(fmt.Sprintf("/nodes/%s/tasks/%s/status", t.Node, t.UPID), t)
	if nil != err || nil == t {
		t = tmp
	}
	if nil == t.client {
		t.client = tmp.client
	}

	if "stopped" == t.Status {
		t.IsCompleted = true
	} else {
		t.IsRunning = true
	}
	if t.IsCompleted {
		if "OK" == t.ExitStatus {
			t.IsSuccessful = true
		} else {
			t.IsFailed = true
		}
	}
	return err
}

func (t *Task) Stop() error {
	return t.client.Delete(fmt.Sprintf("/nodes/%s/tasks/%s", t.Node, t.UPID), nil)
}

func (t *Task) Log(start, limit int) (l Log, err error) {
	return l, t.client.Get(fmt.Sprintf("/nodes/%s/tasks/%s/log?start=%d&limit=%d", t.Node, t.UPID, start, limit), &l)
}

func (t *Task) Watch(start int) (chan string, error) {
	t.client.log.Debugf("starting watcher on %s", t.UPID)
	watch := make(chan string)

	log, err := t.Log(start, 50)
	if err != nil {
		return watch, err
	}

	for i := 0; i < 3; i++ {
		// retry 3 times if the log has no entries
		t.client.log.Debugf("no logs for %s found, retrying %d of 3 times", t.UPID, i)
		if len(log) > 0 {
			break
		}
		time.Sleep(1 * time.Second)

		log, err = t.Log(start, 50)
		if err != nil {
			return watch, err
		}
	}

	if len(log) == 0 {
		return watch, fmt.Errorf("no logs available for %s", t.UPID)
	}

	go func() {
		t.client.log.Debugf("logs found for task %s", t.UPID)
		for _, ln := range log {
			watch <- ln
		}
		t.client.log.Debugf("watching task %s", t.UPID)
		err := tasktail(len(log), watch, t)
		if err != nil {
			t.client.log.Errorf("error watching logs: %s", err)
		}
	}()

	t.client.log.Debugf("returning watcher for %s", t.UPID)
	return watch, nil
}

func tasktail(start int, watch chan string, task *Task) error {
	for {
		task.client.log.Debugf("tailing log for task %s", task.UPID)
		if err := task.Ping(); err != nil {
			return err
		}

		if task.Status != TaskRunning {
			task.client.log.Debugf("task %s is no longer running, closing down watcher", task.UPID)
			close(watch)
			return nil
		}

		logs, err := task.Log(start, 50)
		if err != nil {
			return err
		}
		for _, ln := range logs {
			watch <- ln
		}
		start = start + len(logs)
		time.Sleep(2 * time.Second)
	}
}

func (t *Task) WaitFor(seconds int) error {
	return t.Wait(DefaultWaitInterval, time.Duration(seconds)*time.Second)
}

func (t *Task) Wait(interval, max time.Duration) error {
	// ping it quick to fill in all the details we need in case they're not there
	t.Ping()
	t.client.log.Debugf("waiting for %s, checking every %fs for %fs", t.UPID, interval.Seconds(), max.Seconds())

	timeout := time.After(max)
	for {
		select {
		case <-timeout:
			t.client.log.Debugf("timed out waiting for task %s for %fs", t.UPID, max.Seconds())
			return ErrTimeout
		default:
			if err := t.Ping(); err != nil {
				return err
			}

			if t.Status != TaskRunning {
				t.client.log.Debugf("task %s has completed with status %s", t.UPID, t.Status)
				return nil
			}
			t.client.log.Debugf("waiting on task %s sleeping for %fs", t.UPID, interval.Seconds())
		}
		time.Sleep(interval)
	}
}

func (t *Task) WaitForCompleteStatus(timesNum int, steps ...int) (status bool, completed bool, err error) {
	step := 1
	if len(steps) > 0 && steps[0] > 1 {
		step = steps[0]
	}
	timeout := time.After(time.Duration(step*timesNum) * time.Second)

	for {
		select {
		case <-timeout:
			return
		default:
			err = t.Ping()
			if nil != err {
				t.client.log.Debugf("task %s ping error %+v", t.UPID, err)
				break
			}
			completed = t.IsCompleted

			if completed {
				status = t.IsSuccessful
				return
			}
		}

		time.Sleep(time.Duration(step) * time.Second)

	}
	return
}
