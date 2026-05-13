package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_BackupJobs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	jobs, err := cluster.BackupJobs(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, jobs)
	assert.Equal(t, "backup-job-1", jobs[0].ID)
}

func TestCluster_BackupJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	job, err := cluster.BackupJob(ctx, "backup-job-1")
	assert.Nil(t, err)
	assert.Equal(t, "backup-job-1", job.ID)
	assert.NotNil(t, job.Schedule)
	assert.Equal(t, "sat 02:00", *job.Schedule)
	assert.NotNil(t, job.Enabled)
	assert.Equal(t, IntOrBool(true), *job.Enabled)
}

func TestCluster_NewBackupJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewBackupJob(ctx, &BackupJobOptions{
		Schedule: Ptr("sat 02:00"),
		Storage:  "local",
		Mode:     Ptr("snapshot"),
		All:      IntOrBool(true),
	})
	assert.Nil(t, err)
}

func TestCluster_BackupJobUpdateAndDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.BackupJobUpdate(ctx, "backup-job-1", &BackupJobUpdateOption{
		Enabled: Ptr(IntOrBool(false)),
	}))
	assert.Nil(t, cluster.BackupJobDelete(ctx, "backup-job-1"))
}

func TestCluster_BackupJobIncludedVolumes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	vols, err := cluster.BackupJobIncludedVolumes(ctx, "backup-job-1")
	assert.Nil(t, err)
	assert.NotEmpty(t, vols.Children)
}

func TestCluster_BackupInfoNotBackedUp(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	guests, err := cluster.BackupInfoNotBackedUp(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, guests)
	assert.Equal(t, 200, guests[0].VMID)
}
