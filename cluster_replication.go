package proxmox

import (
	"context"
	"fmt"
)

// ReplicationJobs lists configured storage replication jobs (the per-guest
// snapshot-based replication, distinct from cluster backup jobs).
func (cl *Cluster) ReplicationJobs(ctx context.Context) (jobs []*ReplicationJob, err error) {
	err = cl.client.Get(ctx, "/cluster/replication", &jobs)
	return
}

func (cl *Cluster) ReplicationJob(ctx context.Context, id string) (job *ReplicationJob, err error) {
	job = &ReplicationJob{}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/replication/%s", id), job)
	return
}

func (cl *Cluster) NewReplicationJob(ctx context.Context, opts *ReplicationJobOptions) error {
	return cl.client.Post(ctx, "/cluster/replication", opts, nil)
}

func (cl *Cluster) ReplicationJobUpdate(ctx context.Context, id string, opts *ReplicationJobUpdateOption) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/replication/%s", id), opts, nil)
}

// ReplicationJobDelete marks the job for removal. force=true removes the
// jobconfig entry without cleaning up replicated state; keep=true leaves the
// replicated data at the target. The default (force=false, keep=false) is the
// safe path — PVE schedules a cleanup pass and removes target data.
func (cl *Cluster) ReplicationJobDelete(ctx context.Context, id string, force, keep bool) error {
	path := fmt.Sprintf("/cluster/replication/%s", id)
	query := ""
	if force {
		query = "force=1"
	}
	if keep {
		if query != "" {
			query += "&"
		}
		query += "keep=1"
	}
	if query != "" {
		path += "?" + query
	}
	return cl.client.Delete(ctx, path, nil)
}
