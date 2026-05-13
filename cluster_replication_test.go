package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_ReplicationJobs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	jobs, err := cluster.ReplicationJobs(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, jobs)
	assert.Equal(t, "100-0", jobs[0].ID)
	assert.Equal(t, "node2", jobs[0].Target)
}

func TestCluster_ReplicationJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	job, err := cluster.ReplicationJob(ctx, "100-0")
	assert.Nil(t, err)
	assert.Equal(t, "100-0", job.ID)
	assert.NotNil(t, job.Schedule)
	assert.Equal(t, "*/15", *job.Schedule)
}

func TestCluster_NewReplicationJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewReplicationJob(ctx, &ReplicationJobOptions{
		ID:     "100-0",
		Target: "node2",
		Type:   "local",
	})
	assert.Nil(t, err)
}

func TestCluster_ReplicationJobUpdateAndDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.ReplicationJobUpdate(ctx, "100-0", &ReplicationJobUpdateOption{
		Disable: IntOrBool(true),
	}))
	// force=false, keep=true exercises the query-param path
	assert.Nil(t, cluster.ReplicationJobDelete(ctx, "100-0", false, true))
}
