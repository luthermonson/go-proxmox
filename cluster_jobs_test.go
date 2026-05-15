package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_Jobs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.Jobs(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
}

func TestCluster_ScheduleAnalyze(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	preview, err := cluster.ScheduleAnalyze(context.Background(), "daily", 5, 0)
	assert.Nil(t, err)
	assert.Len(t, preview, 2)
	assert.NotZero(t, preview[0].Timestamp)

	_, err = cluster.ScheduleAnalyze(context.Background(), "", 0, 0)
	assert.NotNil(t, err)
}

func TestCluster_RealmSyncJobs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	jobs, err := cluster.RealmSyncJobs(context.Background())
	assert.Nil(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "ldap-sync", jobs[0].ID)
}

func TestCluster_RealmSyncJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	job, err := cluster.RealmSyncJob(context.Background(), "ldap-sync")
	assert.Nil(t, err)
	assert.Equal(t, "ldap1", job.Realm)

	_, err = cluster.RealmSyncJob(context.Background(), "")
	assert.NotNil(t, err)
}

func TestCluster_NewRealmSyncJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewRealmSyncJob(context.Background(), "ldap-sync", &ClusterRealmSyncJobOptions{Schedule: "daily", Realm: "ldap1"})
	assert.Nil(t, err)

	err = cluster.NewRealmSyncJob(context.Background(), "", nil)
	assert.NotNil(t, err)

	err = cluster.NewRealmSyncJob(context.Background(), "ldap-sync", &ClusterRealmSyncJobOptions{})
	assert.NotNil(t, err)
}

func TestCluster_UpdateRealmSyncJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.UpdateRealmSyncJob(context.Background(), "ldap-sync", &ClusterRealmSyncJobOptions{Comment: "updated"})
	assert.Nil(t, err)

	err = cluster.UpdateRealmSyncJob(context.Background(), "", nil)
	assert.NotNil(t, err)
}

func TestCluster_DeleteRealmSyncJob(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.DeleteRealmSyncJob(context.Background(), "ldap-sync")
	assert.Nil(t, err)

	err = cluster.DeleteRealmSyncJob(context.Background(), "")
	assert.NotNil(t, err)
}
