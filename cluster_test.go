package proxmox

import (
	"context"
	"net/http"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/luthermonson/go-proxmox/tests/mocks/capture"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCluster(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 4, cluster.Version)
	assert.Equal(t, "cluster", cluster.ID)
	for _, n := range cluster.Nodes {
		assert.Contains(t, n.ID, "node/node")
		assert.Equal(t, "node", n.Type)
	}
}

func TestNextID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	nextid, err := cluster.NextID(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 100, nextid)
}

func TestCheckID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	checkIDFree, err := cluster.CheckID(ctx, 100)
	assert.Nil(t, err)
	assert.Equal(t, true, checkIDFree)
	checkIDTaken, err := cluster.CheckID(ctx, 200)
	assert.Nil(t, err)
	assert.Equal(t, false, checkIDTaken)
	checkIDInUse, err := cluster.CheckID(ctx, 300)
	assert.Nil(t, err)
	assert.Equal(t, false, checkIDInUse)
	_, err = cluster.CheckID(ctx, 99)
	assert.True(t, IsBadRequest(err))
}

func TestCluster_Resources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	rs, err := cluster.Resources(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 20, len(rs))

	// type param test
	rs, err = cluster.Resources(ctx, "node")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))
}

func TestCluster_SDNZones(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	zones, err := cluster.SDNZones(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(zones))
	assert.Equal(t, CSV{"host1", "host2"}, zones[0].Nodes)
	assert.Equal(t, CSV{"203.0.113.184", "203.0.113.185"}, zones[0].Peers)

	// type param test
	zones, err = cluster.SDNZones(ctx, "vxlan")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(zones))
	assert.Equal(t, "vxlan", zones[0].Type)
	assert.Equal(t, "test1", zones[0].Name)
	assert.Equal(t, "pve", zones[0].IPAM)
}

func TestCluster_SDNZone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	zone, err := cluster.SDNZone(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "test1", zone.Name)
	assert.Equal(t, "vxlan", zone.Type)
	assert.Equal(t, CSV{"host1", "host2"}, zone.Nodes)
	assert.Equal(t, CSV{"203.0.113.184", "203.0.113.185"}, zone.Peers)
}

func TestCluster_Backups(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	backups, err := cluster.Backups(ctx)
	assert.Nil(t, err)
	assert.Len(t, backups, 2)

	assert.Equal(t, "backup-1", backups[0].ID)
	assert.Equal(t, "*/30", backups[0].Schedule)
	assert.Equal(t, "snapshot", backups[0].Mode)
	assert.Equal(t, IntOrBool(true), backups[0].Enabled)
	assert.Equal(t, IntOrBool(true), backups[0].All)

	assert.Equal(t, "backup-2", backups[1].ID)
	assert.Equal(t, IntOrBool(false), backups[1].Enabled)
	assert.Equal(t, "101,102", backups[1].VMID)
	assert.Equal(t, "keep-daily=7,keep-weekly=4", backups[1].PruneBackups)

	// every returned backup must carry the client back so receivers (Update/Delete) work
	for _, b := range backups {
		assert.NotNil(t, b.client, "ClusterBackup.client must be wired for %s", b.ID)
	}
}

func TestCluster_Backup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// empty id guard
	_, err = cluster.Backup(ctx, "")
	assert.Error(t, err)

	backup, err := cluster.Backup(ctx, "backup-1")
	assert.Nil(t, err)
	assert.Equal(t, "backup-1", backup.ID)
	assert.Equal(t, "snapshot", backup.Mode)
	assert.Equal(t, "local", backup.Storage)
	assert.NotNil(t, backup.client)
}

func TestCluster_NewBackup(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	assert.Nil(t, cluster.NewBackup(ctx, &ClusterBackupOptions{
		Schedule: "daily 02:00",
		Mode:     "snapshot",
		Storage:  "local",
		All:      true,
	}))

	// nil opts is allowed; method should still POST without panicking
	assert.Nil(t, cluster.NewBackup(ctx, nil))
}

func TestClusterBackup_Update(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// empty id guard — no HTTP request issued
	empty := ClusterBackup{client: client}
	assert.Error(t, empty.Update(ctx, &ClusterBackupOptions{}))

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	backup, err := cluster.Backup(ctx, "backup-1")
	assert.Nil(t, err)

	assert.Nil(t, backup.Update(ctx, &ClusterBackupOptions{
		Schedule: "weekly",
	}))
	assert.Nil(t, backup.Update(ctx, nil)) // nil opts ok
}

// TestClusterBackup_UpdateSendsExplicitFalse checks that the options PVE
// defaults to true can be turned off: a disabled job, remove=0 to keep old
// backups and stdexcludes=0 reach the API instead of being dropped.
func TestClusterBackup_UpdateSendsExplicitFalse(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	require.NoError(t, err)
	backup, err := cluster.Backup(ctx, "backup-1")
	require.NoError(t, err)

	require.NoError(t, backup.Update(ctx, &ClusterBackupOptions{
		Enabled:     Ptr(false),
		Remove:      Ptr(false),
		StdExcludes: Ptr(false),
	}))
	require.NoError(t, cluster.NewBackup(ctx, &ClusterBackupOptions{
		Schedule: "daily 02:00",
		Storage:  "local",
		All:      true,
		Enabled:  Ptr(false),
	}))

	var bodies []map[string]any
	for _, r := range capture.Requests {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			bodies = append(bodies, r.Body)
		}
	}
	require.Len(t, bodies, 2)
	assert.Equal(t, false, bodies[0]["enabled"])
	assert.Equal(t, false, bodies[0]["remove"])
	assert.Equal(t, false, bodies[0]["stdexcludes"])
	assert.Equal(t, false, bodies[1]["enabled"])
}

// TestClusterBackup_UpdateSendsZeroCounts checks that zstd=0 (half of the
// cores), ionice=0, lockwait=0 and stopwait=0 reach the API instead of being
// dropped, which left the job on the PVE defaults 1, 7, 180 and 10.
func TestClusterBackup_UpdateSendsZeroCounts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	require.NoError(t, err)
	backup, err := cluster.Backup(ctx, "backup-1")
	require.NoError(t, err)

	require.NoError(t, backup.Update(ctx, &ClusterBackupOptions{
		IoNice:   Ptr(uint(0)),
		LockWait: Ptr(uint(0)),
		StopWait: Ptr(uint(0)),
		Zstd:     Ptr(uint(0)),
	}))

	var body map[string]any
	for _, r := range capture.Requests {
		if r.Method == http.MethodPut {
			body = r.Body
		}
	}
	require.NotNil(t, body)
	for _, k := range []string{"ionice", "lockwait", "stopwait", "zstd"} {
		v, present := body[k]
		assert.Truef(t, present, "%q missing from the PUT body", k)
		assert.EqualValues(t, 0, v)
	}
}

func TestClusterBackup_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	// empty id guard
	empty := ClusterBackup{client: client}
	assert.Error(t, empty.Delete(ctx))

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)
	backup, err := cluster.Backup(ctx, "backup-1")
	assert.Nil(t, err)

	assert.Nil(t, backup.Delete(ctx))
}

func TestCluster_Tasks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	tasks, err := cluster.Tasks(ctx)
	assert.Nil(t, err)
	assert.Len(t, tasks, 2)
	// every returned task must carry the client back
	for _, task := range tasks {
		assert.NotNil(t, task.client, "Task.client must be wired for %s", task.UPID)
	}
}

func TestCluster_Ceph(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	ceph, err := cluster.Ceph(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ceph)
	assert.NotNil(t, ceph.client)
}

func TestCluster_New(t *testing.T) {
	client := mockClient()
	cl := &Cluster{}
	got := cl.New(client)
	assert.NotNil(t, got)
	assert.Same(t, client, got.client)
}

func TestCluster_SDNVNets(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	// json unmarshaling tests
	vnets, err := cluster.SDNVNets(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(vnets))

	// vnet name test
	vnet, err := cluster.SDNVNet(ctx, "user1")
	assert.Nil(t, err)
	assert.Equal(t, "user1", vnet.Name)
	assert.Equal(t, "vnet", vnet.Type)
	assert.Equal(t, "test1", vnet.Zone)
	assert.Equal(t, "myuser1's network", vnet.Alias)
	assert.Equal(t, 1, vnet.VlanAware)
	assert.Equal(t, uint32(10), vnet.Tag)

	// VNet Tag max value (VXLAN VNI range is 0-16777215)
	vnetMaxTag, err := cluster.SDNVNet(ctx, "maxTagVnet")
	assert.Nil(t, err)
	assert.Equal(t, "maxTagVnet", vnetMaxTag.Name)
	assert.Equal(t, uint32(16777215), vnetMaxTag.Tag)
}
