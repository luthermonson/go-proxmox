package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNPrefixLists(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	lists, err := cluster.SDNPrefixLists(context.Background(), false, false, false)
	assert.Nil(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "pl1", lists[0].ID)
}

func TestSDNPrefixList_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("pl1")
	assert.Nil(t, l.Read(context.Background()))
	assert.Len(t, l.Entries, 1)

	assert.Nil(t, l.Update(context.Background(), &SDNPrefixListOptions{}))
	assert.Nil(t, l.Delete(context.Background()))
}

func TestCluster_NewSDNPrefixList(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNPrefixList(context.Background(), &SDNPrefixListOptions{ID: "pl3"})
	assert.Nil(t, err)
	assert.NotNil(t, cluster.NewSDNPrefixList(context.Background(), nil))
}

func TestSDNPrefixList_Entries(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("pl1")
	entries, err := l.ListEntries(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "pl1", entries[0].ID)
}

func TestSDNPrefixListEntry_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("pl1")
	e := l.Entry(10)
	assert.Nil(t, e.Read(context.Background()))
	assert.Equal(t, "permit", e.Action)

	assert.Nil(t, e.Update(context.Background(), &SDNPrefixListEntryOptions{Action: "deny"}))
	assert.Nil(t, e.Delete(context.Background()))
}

func TestSDNPrefixList_AddEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("pl1")
	err := l.AddEntry(context.Background(), &SDNPrefixListEntryOptions{Action: "permit", Prefix: "10.0.5.0/24"})
	assert.Nil(t, err)

	assert.NotNil(t, l.AddEntry(context.Background(), nil))
	assert.NotNil(t, l.AddEntry(context.Background(), &SDNPrefixListEntryOptions{Action: "permit"}))
}

func TestCluster_SDNPrefixListsFiltered(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	lists, err := cluster.SDNPrefixLists(context.Background(), true, true, true)
	assert.Nil(t, err)
	assert.Len(t, lists, 1)
	assert.Equal(t, "pl1", lists[0].ID)
}

func TestSDNPrefixList_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("pl1")
	assert.Nil(t, l.Update(context.Background(), nil))
}

func TestSDNPrefixList_EmptyID_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	l := &SDNPrefixList{client: mockClient()}
	assert.Error(t, l.Read(context.Background()))
	assert.Error(t, l.Update(context.Background(), &SDNPrefixListOptions{}))
	assert.Error(t, l.Delete(context.Background()))
	_, err := l.ListEntries(context.Background())
	assert.Error(t, err)
	assert.Error(t, l.AddEntry(context.Background(), &SDNPrefixListEntryOptions{Action: "permit", Prefix: "10.0.0.0/24"}))
}

func TestSDNPrefixList_Read_NotFound(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	l := cluster.SDNPrefixList("missing")
	assert.Error(t, l.Read(context.Background()))
	_, err := l.ListEntries(context.Background())
	assert.Error(t, err)
}

func TestSDNPrefixListEntry_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	e := cluster.SDNPrefixList("pl1").Entry(10)
	assert.Nil(t, e.Update(context.Background(), nil))
}

func TestSDNPrefixListEntry_EmptyID_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	e := &SDNPrefixListEntry{client: mockClient(), Seq: 10}
	assert.Error(t, e.Read(context.Background()))
	assert.Error(t, e.Update(context.Background(), &SDNPrefixListEntryOptions{}))
	assert.Error(t, e.Delete(context.Background()))
}
