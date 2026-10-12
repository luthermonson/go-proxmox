package proxmox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// TaskRunning is the Status PVE reports for a task that has not finished.
	TaskRunning = "running"
	// TaskStopped is the Status PVE reports for a finished task. Whether it
	// succeeded is carried by ExitStatus.
	TaskStopped = "stopped"
	// TaskExitOK is the ExitStatus of a task that finished cleanly.
	TaskExitOK = "OK"

	// taskWarningsPrefix starts the ExitStatus of a task that finished but
	// logged warnings, e.g. "WARNINGS: 2". PVE counts these as successes.
	taskWarningsPrefix = "WARNINGS:"

	// taskHistoryLimit caps how many entries the "no such task" fallback
	// reads from the node's task index. The query is already narrowed by
	// start time, type and VMID, so this only matters on very busy nodes.
	taskHistoryLimit = 500
)

var DefaultWaitInterval = 1 * time.Second

// TaskFailedError is returned by Task.Wait, Task.WaitFor and
// Task.WaitForCompleteStatus when the task finished with an exit status other
// than "OK" (or "WARNINGS: n", which PVE counts as success). ExitStatus holds
// PVE's message verbatim, e.g. "VM quit/powerdown failed - got timeout".
//
// Match it with errors.As or IsTaskFailed:
//
//	var tfe *proxmox.TaskFailedError
//	if errors.As(err, &tfe) {
//		log.Printf("%s on %s failed: %s", tfe.Type, tfe.Node, tfe.ExitStatus)
//	}
type TaskFailedError struct {
	UPID       UPID
	Node       string
	Type       string
	ID         string
	User       string
	ExitStatus string
}

func (e *TaskFailedError) Error() string {
	status := e.ExitStatus
	if status == "" {
		status = "unknown exit status"
	}
	what := e.Type
	if e.ID != "" {
		what += " " + e.ID
	}
	if what == "" {
		return fmt.Sprintf("task %s failed: %s", e.UPID, status)
	}
	if e.Node != "" {
		what += " on " + e.Node
	}
	return fmt.Sprintf("task %s (%s) failed: %s", e.UPID, what, status)
}

// IsTaskFailed reports whether err is, or wraps, a *TaskFailedError.
func IsTaskFailed(err error) bool {
	var tfe *TaskFailedError
	return errors.As(err, &tfe)
}

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

// Ping refreshes the task's status from /nodes/{node}/tasks/{upid}/status and
// recomputes IsRunning, IsCompleted, IsSuccessful, IsFailed and Warnings. On
// error the task keeps the state from its previous successful Ping.
//
// Ping only reports transport/API errors; a task that finished unsuccessfully
// is not an error here. Use Wait (or check IsFailed) for that.
func (t *Task) Ping(ctx context.Context) error {
	client, upid, node, typ, id, user := t.client, t.UPID, t.Node, t.Type, t.ID, t.User
	if err := t.client.Get(ctx, fmt.Sprintf("/nodes/%s/tasks/%s/status", t.Node, t.UPID), t); err != nil {
		return err
	}

	// The status payload is copied over the whole struct; keep the identity
	// we already had if the response left any of it out.
	if t.client == nil {
		t.client = client
	}
	if t.UPID == "" {
		t.UPID = upid
	}
	if t.Node == "" {
		t.Node = node
	}
	if t.Type == "" {
		t.Type = typ
	}
	if t.ID == "" {
		t.ID = id
	}
	if t.User == "" {
		t.User = user
	}

	t.updateState()
	return nil
}

// updateState derives the Is* flags and Warnings from Status and ExitStatus.
func (t *Task) updateState() {
	t.IsCompleted = t.Status == TaskStopped
	t.IsRunning = !t.IsCompleted
	t.Warnings = 0
	t.IsSuccessful = false
	t.IsFailed = false
	if !t.IsCompleted {
		return
	}

	switch {
	case t.ExitStatus == TaskExitOK:
		t.IsSuccessful = true
	case strings.HasPrefix(t.ExitStatus, taskWarningsPrefix):
		t.IsSuccessful = true
		n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(t.ExitStatus, taskWarningsPrefix)))
		if err != nil || n < 1 {
			n = 1
		}
		t.Warnings = n
	default:
		t.IsFailed = true
	}
}

// failure returns a *TaskFailedError if the task finished unsuccessfully.
func (t *Task) failure() error {
	if !t.IsCompleted || t.IsSuccessful {
		return nil
	}
	return &TaskFailedError{
		UPID:       t.UPID,
		Node:       t.Node,
		Type:       t.Type,
		ID:         t.ID,
		User:       t.User,
		ExitStatus: t.ExitStatus,
	}
}

// poll pings the task once and reports whether it has finished. If the status
// endpoint no longer knows the UPID (PVE answers "no such task" once the
// task's status file is gone), it falls back to the node's task history to
// recover the final status.
func (t *Task) poll(ctx context.Context) (bool, error) {
	err := t.Ping(ctx)
	if err == nil {
		return t.IsCompleted, nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return false, ctxErr
	}
	if !isNoSuchTask(err) {
		return false, err
	}

	t.client.log.Debugf("task %s not found via status endpoint, checking node task history", t.UPID)
	found, herr := t.statusFromHistory(ctx)
	if herr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return false, ctxErr
		}
		return false, fmt.Errorf("%w (task history lookup also failed: %w)", err, herr)
	}
	if !found {
		return false, err
	}
	return true, nil
}

// isNoSuchTask reports whether err is PVE's answer for an unknown UPID. PVE
// raises it as a parameter error ({"errors":{"upid":"no such task"}}), and the
// message survives into the error text however the response is wrapped.
func isNoSuchTask(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "no such task")
}

// statusFromHistory looks the task up in /nodes/{node}/tasks and, if it is
// listed as finished, fills in Status/ExitStatus from that entry. It reports
// whether a final status was found.
func (t *Task) statusFromHistory(ctx context.Context) (bool, error) {
	if t.Node == "" {
		return false, nil
	}

	opts := &NodeTasksOptions{Source: "all", Limit: taskHistoryLimit, TypeFilter: t.Type}
	if start, ok := upidStartTime(t.UPID); ok && start > 1 {
		opts.Since = start - 1
	}
	if vmid, err := strconv.Atoi(t.ID); err == nil && vmid > 0 {
		opts.VMID = vmid
	}

	node := &Node{client: t.client, Name: t.Node}
	history, err := node.Tasks(ctx, opts)
	if err != nil {
		return false, err
	}

	for _, h := range history {
		if h.UPID != t.UPID {
			continue
		}
		// In the task index "status" is the exit status; it is empty while
		// the task is still active.
		if h.Status == "" || h.Status == TaskRunning {
			return false, nil
		}
		t.ExitStatus = h.Status
		t.Status = TaskStopped
		if !h.StartTime.IsZero() {
			t.StartTime = h.StartTime
		}
		if !h.EndTime.IsZero() {
			t.EndTime = h.EndTime
		}
		if !t.StartTime.IsZero() && !t.EndTime.IsZero() {
			t.Duration = t.EndTime.Sub(t.StartTime)
		}
		t.updateState()
		return true, nil
	}
	return false, nil
}

// upidStartTime extracts the start time (hex unix seconds, 5th field) from a
// UPID of the form UPID:node:pid:pstart:starttime:type:id:user:.
func upidStartTime(upid UPID) (int64, bool) {
	sp := strings.Split(string(upid), ":")
	if len(sp) < 5 {
		return 0, false
	}
	v, err := strconv.ParseInt(sp[4], 16, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// sleepCtx waits for d or until ctx is done, returning ctx.Err() in the
// latter case.
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (t *Task) Stop(ctx context.Context) error {
	return t.client.Delete(ctx, fmt.Sprintf("/nodes/%s/tasks/%s", t.Node, t.UPID), nil)
}

func (t *Task) Log(ctx context.Context, start, limit int) (l Log, err error) {
	return l, t.client.Get(ctx, fmt.Sprintf("/nodes/%s/tasks/%s/log?start=%d&limit=%d", t.Node, t.UPID, start, limit), &l)
}

func (t *Task) Watch(ctx context.Context, start int) (chan string, error) {
	t.client.log.Debugf("starting watcher on %s", t.UPID)

	log, err := t.Log(ctx, start, 50)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		// retry 3 times if the log has no entries
		t.client.log.Debugf("no logs for %s found, retrying %d of 3 times", t.UPID, i)
		if len(log) > 0 {
			break
		}
		if err := sleepCtx(ctx, 1*time.Second); err != nil {
			return nil, err
		}

		log, err = t.Log(ctx, start, 50)
		if err != nil {
			return nil, err
		}
	}

	if len(log) == 0 {
		return nil, fmt.Errorf("no logs available for %s", t.UPID)
	}

	watch := make(chan string)

	go func() {
		defer close(watch)
		t.client.log.Debugf("logs found for task %s", t.UPID)
		for _, ln := range log {
			select {
			case watch <- ln:
			case <-ctx.Done():
				return
			}
		}
		t.client.log.Debugf("watching task %s", t.UPID)
		err := tasktail(ctx, len(log), watch, t)
		if err != nil {
			t.client.log.Errorf("error watching logs: %s", err)
		}
	}()

	t.client.log.Debugf("returning watcher for %s", t.UPID)
	return watch, nil
}

func tasktail(ctx context.Context, start int, watch chan string, task *Task) error {
	for {
		task.client.log.Debugf("tailing log for task %s", task.UPID)
		if err := task.Ping(ctx); err != nil {
			return err
		}

		if task.Status != TaskRunning {
			task.client.log.Debugf("task %s is no longer running, closing down watcher", task.UPID)
			return nil
		}

		logs, err := task.Log(ctx, start, 50)
		if err != nil {
			return err
		}
		for _, ln := range logs {
			select {
			case watch <- ln:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		start += len(logs)
		if err := sleepCtx(ctx, 2*time.Second); err != nil {
			return err
		}
	}
}

// WaitFor is Wait with DefaultWaitInterval and a timeout in seconds.
func (t *Task) WaitFor(ctx context.Context, seconds int) error {
	return t.Wait(ctx, DefaultWaitInterval, time.Duration(seconds)*time.Second)
}

// Wait polls the task every interval until it finishes, max elapses, or ctx is
// done. It returns:
//
//   - nil if the task finished with exit status "OK" or "WARNINGS: n" (check
//     t.Warnings for the latter);
//   - a *TaskFailedError if it finished with any other exit status;
//   - ErrTimeout if it was still running after max;
//   - ctx.Err() if ctx was cancelled or its deadline passed;
//   - the API error if the status could not be read. When PVE no longer knows
//     the UPID ("no such task"), the node's task history is consulted first.
func (t *Task) Wait(ctx context.Context, interval, max time.Duration) error {
	if interval <= 0 {
		interval = DefaultWaitInterval
	}
	t.client.log.Debugf("waiting for %s, checking every %fs for %fs", t.UPID, interval.Seconds(), max.Seconds())

	timeout := time.NewTimer(max)
	defer timeout.Stop()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		done, err := t.poll(ctx)
		if err != nil {
			return err
		}
		if done {
			t.client.log.Debugf("task %s has completed with exit status %q", t.UPID, t.ExitStatus)
			return t.failure()
		}

		t.client.log.Debugf("waiting on task %s, next check in %fs", t.UPID, interval.Seconds())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			t.client.log.Debugf("timed out waiting for task %s for %fs", t.UPID, max.Seconds())
			return ErrTimeout
		case <-ticker.C:
		}
	}
}

// WaitForCompleteStatus polls the task every steps[0] seconds (default 1) for
// up to timesNum polls. completed reports whether the task finished and status
// whether it succeeded. If it finished unsuccessfully err is a
// *TaskFailedError. If the polls run out, err is nil or the last status error.
// If ctx is done, err is ctx.Err().
func (t *Task) WaitForCompleteStatus(ctx context.Context, timesNum int, steps ...int) (status bool, completed bool, err error) {
	step := 1
	if len(steps) > 0 && steps[0] > 1 {
		step = steps[0]
	}
	timeout := time.NewTimer(time.Duration(step*timesNum) * time.Second)
	defer timeout.Stop()
	ticker := time.NewTicker(time.Duration(step) * time.Second)
	defer ticker.Stop()

	for {
		var done bool
		done, err = t.poll(ctx)
		switch {
		case err != nil && ctx.Err() != nil:
			return status, completed, ctx.Err()
		case err != nil:
			t.client.log.Debugf("task %s ping error %+v", t.UPID, err)
		case done:
			completed = true
			status = t.IsSuccessful
			err = t.failure()
			return
		}

		select {
		case <-ctx.Done():
			return status, completed, ctx.Err()
		case <-timeout.C:
			return
		case <-ticker.C:
		}
	}
}
