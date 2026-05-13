package proxmox

import (
	"context"
	"fmt"
)

// BackupJobs lists all scheduled vzdump backup jobs. These are cluster-level
// scheduled jobs — distinct from `VirtualMachineBackupOptions`, which is the
// per-call vzdump configuration on a node (POST /nodes/{node}/vzdump). PVE
// stores backup jobs in /etc/pve/jobs.cfg and reuses many of the same param
// names, but the lifecycle (CRUD on /cluster/backup/{id}) is independent.
func (cl *Cluster) BackupJobs(ctx context.Context) (jobs []*BackupJob, err error) {
	err = cl.client.Get(ctx, "/cluster/backup", &jobs)
	return
}

func (cl *Cluster) BackupJob(ctx context.Context, id string) (job *BackupJob, err error) {
	job = &BackupJob{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/backup/%s", id), job)
	return
}

func (cl *Cluster) NewBackupJob(ctx context.Context, opts *BackupJobOptions) error {
	return cl.client.Post(ctx, "/cluster/backup", opts, nil)
}

func (cl *Cluster) BackupJobUpdate(ctx context.Context, id string, opts *BackupJobUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/backup/%s", id), opts, nil)
}

func (cl *Cluster) BackupJobDelete(ctx context.Context, id string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/backup/%s", id), nil)
}

// BackupJobIncludedVolumes returns the tree of guests covered by the job and
// the per-disk backup status. The response is shaped for ExtJS tree views in
// the PVE UI, so the result is loosely typed.
func (cl *Cluster) BackupJobIncludedVolumes(ctx context.Context, id string) (volumes *BackupIncludedVolumes, err error) {
	volumes = &BackupIncludedVolumes{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/backup/%s/included_volumes", id), volumes)
	return
}

// BackupInfoNotBackedUp returns the guests that aren't covered by any backup
// job — the inverse view of "what's at risk". The /cluster/backup-info parent
// path is just a directory index and is intentionally not wrapped.
func (cl *Cluster) BackupInfoNotBackedUp(ctx context.Context) (guests []*NotBackedUpGuest, err error) {
	err = cl.client.Get(ctx, "/cluster/backup-info/not-backed-up", &guests)
	return
}
