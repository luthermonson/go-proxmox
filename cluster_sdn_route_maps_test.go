package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNRouteMaps(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	maps, err := cluster.SDNRouteMaps(context.Background(), false)
	assert.Nil(t, err)
	assert.Len(t, maps, 2)
	assert.Equal(t, "rm1", maps[0].ID)
}

func TestCluster_SDNRouteMapEntries(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNRouteMapEntries(context.Background(), false, false)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
}

func TestCluster_SDNRouteMapEntriesFor(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNRouteMapEntriesFor(context.Background(), "rm1", false, false)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "rm1", entries[0].RouteMapID)

	_, err = cluster.SDNRouteMapEntriesFor(context.Background(), "", false, false)
	assert.NotNil(t, err)
}

func TestSDNRouteMapEntry_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	e := cluster.SDNRouteMapEntry("rm1", 10)
	assert.Nil(t, e.Read(context.Background()))
	assert.Equal(t, "permit", e.Action)
	assert.Len(t, e.Match, 1)

	assert.Nil(t, e.Update(context.Background(), &SDNRouteMapEntryOptions{Action: "deny"}))
	assert.Nil(t, e.Delete(context.Background()))
}

func TestCluster_NewSDNRouteMapEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNRouteMapEntry(context.Background(), &SDNRouteMapEntryOptions{RouteMapID: "rm1", Order: 30, Action: "permit"})
	assert.Nil(t, err)

	assert.NotNil(t, cluster.NewSDNRouteMapEntry(context.Background(), nil))
	assert.NotNil(t, cluster.NewSDNRouteMapEntry(context.Background(), &SDNRouteMapEntryOptions{RouteMapID: "rm1"}))
}

func TestCluster_SDNRouteMapsRunning(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	maps, err := cluster.SDNRouteMaps(context.Background(), true)
	assert.Nil(t, err)
	assert.Len(t, maps, 1)
}

func TestCluster_SDNRouteMapEntriesPending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNRouteMapEntries(context.Background(), true, false)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
}

func TestCluster_SDNRouteMapEntriesForPending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	entries, err := cluster.SDNRouteMapEntriesFor(context.Background(), "rm1", true, false)
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "rm1", entries[0].RouteMapID)
}

func TestSDNRouteMapEntry_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	e := cluster.SDNRouteMapEntry("rm1", 10)
	assert.Nil(t, e.Update(context.Background(), nil))
}

func TestSDNRouteMapEntry_EmptyID_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	e := &SDNRouteMapEntry{client: mockClient(), Order: 10}
	assert.Error(t, e.Read(context.Background()))
	assert.Error(t, e.Update(context.Background(), &SDNRouteMapEntryOptions{}))
	assert.Error(t, e.Delete(context.Background()))
}

func TestSDNRouteMapEntry_Read_NotFound(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	e := cluster.SDNRouteMapEntry("missing", 99)
	assert.Error(t, e.Read(context.Background()))
}
