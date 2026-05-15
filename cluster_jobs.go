package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// Jobs lists the resource-type directory under /cluster/jobs (just
// "realm-sync" today, but PVE may grow other recurring job kinds).
func (cl *Cluster) Jobs(ctx context.Context) (entries []*ClusterJobIndexEntry, err error) {
	err = cl.client.Get(ctx, "/cluster/jobs", &entries)
	return
}

// ScheduleAnalyze simulates a systemd-style calendar event and returns the
// next iterations firings. iterations defaults to 10 server-side when 0;
// startTime in UNIX seconds (0 = "now").
func (cl *Cluster) ScheduleAnalyze(ctx context.Context, schedule string, iterations, startTime int) (preview []*ClusterScheduleEvent, err error) {
	if schedule == "" {
		err = errors.New("schedule is required")
		return
	}
	q := url.Values{}
	q.Set("schedule", schedule)
	if iterations > 0 {
		q.Set("iterations", strconv.Itoa(iterations))
	}
	if startTime > 0 {
		q.Set("starttime", strconv.Itoa(startTime))
	}
	err = cl.client.Get(ctx, "/cluster/jobs/schedule-analyze?"+q.Encode(), &preview)
	return
}

// --- realm-sync jobs --------------------------------------------------------

// RealmSyncJobs lists configured realm-sync (LDAP/AD/OIDC) jobs.
func (cl *Cluster) RealmSyncJobs(ctx context.Context) (jobs []*ClusterRealmSyncJob, err error) {
	err = cl.client.Get(ctx, "/cluster/jobs/realm-sync", &jobs)
	return
}

// RealmSyncJob reads a single realm-sync job by id.
func (cl *Cluster) RealmSyncJob(ctx context.Context, id string) (job *ClusterRealmSyncJob, err error) {
	if id == "" {
		err = errors.New("realm-sync job id is required")
		return
	}
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/jobs/realm-sync/%s", id), &job)
	return
}

// NewRealmSyncJob creates a realm-sync job. PVE puts the id in the URL (not
// the body) and requires opts.Schedule.
func (cl *Cluster) NewRealmSyncJob(ctx context.Context, id string, opts *ClusterRealmSyncJobOptions) error {
	if id == "" {
		return errors.New("realm-sync job id is required")
	}
	if opts == nil || opts.Schedule == "" {
		return errors.New("realm-sync job schedule is required")
	}
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/jobs/realm-sync/%s", id), opts, nil)
}

// UpdateRealmSyncJob mutates an existing job. Pass opts.Delete (comma list)
// to reset specific keys back to PVE defaults.
func (cl *Cluster) UpdateRealmSyncJob(ctx context.Context, id string, opts *ClusterRealmSyncJobOptions) error {
	if id == "" {
		return errors.New("realm-sync job id is required")
	}
	if opts == nil {
		opts = &ClusterRealmSyncJobOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/jobs/realm-sync/%s", id), opts, nil)
}

// DeleteRealmSyncJob removes a realm-sync job.
func (cl *Cluster) DeleteRealmSyncJob(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("realm-sync job id is required")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/jobs/realm-sync/%s", id), nil)
}
