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
