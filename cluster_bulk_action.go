package proxmox

import (
	"context"
)

// Wrappers for /cluster/bulk-action — fleet-wide guest start/shutdown/suspend
// /migrate operations. Each action returns a UPID for the worker task that
// dispatches per-guest subtasks; poll the returned *Task for completion.

// BulkActionSubdirs enumerates the children of /cluster/bulk-action ("guest"
// today). ACL-filtered.
//
// GET /cluster/bulk-action
func (cl *Cluster) BulkActionSubdirs(ctx context.Context) ([]string, error) {
	return cl.bulkActionDiridx(ctx, "/cluster/bulk-action")
}

// BulkActionGuestSubdirs enumerates the children of /cluster/bulk-action/guest
// ("start", "shutdown", "suspend", "migrate"). ACL-filtered.
//
// GET /cluster/bulk-action/guest
func (cl *Cluster) BulkActionGuestSubdirs(ctx context.Context) ([]string, error) {
	return cl.bulkActionDiridx(ctx, "/cluster/bulk-action/guest")
}

// BulkStartGuests starts (or resumes) all guests matching opts.VMIDs, or every
// guest cluster-wide when opts.VMIDs is empty. Returns the UPID for the
// worker task.
//
// POST /cluster/bulk-action/guest/start
func (cl *Cluster) BulkStartGuests(ctx context.Context, opts *BulkStartOptions) (*Task, error) {
	if opts == nil {
		opts = &BulkStartOptions{}
	}
	var upid UPID
	if err := cl.client.Post(ctx, "/cluster/bulk-action/guest/start", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// BulkShutdownGuests shuts down all guests matching opts.VMIDs, or every guest
// cluster-wide when opts.VMIDs is empty. Returns the UPID for the worker task.
//
// POST /cluster/bulk-action/guest/shutdown
func (cl *Cluster) BulkShutdownGuests(ctx context.Context, opts *BulkShutdownOptions) (*Task, error) {
	if opts == nil {
		opts = &BulkShutdownOptions{}
	}
	var upid UPID
	if err := cl.client.Post(ctx, "/cluster/bulk-action/guest/shutdown", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// BulkSuspendGuests suspends all guests matching opts.VMIDs, or every guest
// cluster-wide when opts.VMIDs is empty. Returns the UPID for the worker task.
//
// POST /cluster/bulk-action/guest/suspend
func (cl *Cluster) BulkSuspendGuests(ctx context.Context, opts *BulkSuspendOptions) (*Task, error) {
	if opts == nil {
		opts = &BulkSuspendOptions{}
	}
	var upid UPID
	if err := cl.client.Post(ctx, "/cluster/bulk-action/guest/suspend", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}

// BulkMigrateGuests migrates guests to the target node. opts.Target is
// required by the schema; opts.VMIDs filters which guests are moved. Returns
// the UPID for the worker task.
//
// POST /cluster/bulk-action/guest/migrate
func (cl *Cluster) BulkMigrateGuests(ctx context.Context, opts *BulkMigrateOptions) (*Task, error) {
	if opts == nil {
		opts = &BulkMigrateOptions{}
	}
	var upid UPID
	if err := cl.client.Post(ctx, "/cluster/bulk-action/guest/migrate", opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, cl.client), nil
}
