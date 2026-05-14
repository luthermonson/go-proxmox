package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_APT(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	entries, err := node.APT(context.Background())
	require.NoError(t, err)
	require.Len(t, entries, 4)
	ids := []string{entries[0].ID, entries[1].ID, entries[2].ID, entries[3].ID}
	assert.Contains(t, ids, "changelog")
	assert.Contains(t, ids, "repositories")
	assert.Contains(t, ids, "update")
	assert.Contains(t, ids, "versions")
}

func TestNode_APTUpdates(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	updates, err := node.APTUpdates(context.Background())
	require.NoError(t, err)
	require.Len(t, updates, 1)
	assert.Equal(t, "pve-manager", updates[0].Package)
	assert.Equal(t, "9.1-2", updates[0].Version)
	assert.Equal(t, "9.1-1", updates[0].OldVersion)
	assert.Equal(t, "Proxmox", updates[0].Origin)
}

func TestNode_APTUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	task, err := node.APTUpdate(context.Background(), true, false)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "node1", task.Node)
	assert.Equal(t, "aptupdate", task.Type)
}

func TestNode_APTChangelog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	cl, err := node.APTChangelog(context.Background(), "pve-manager", "")
	require.NoError(t, err)
	assert.Contains(t, cl, "pve-manager")
	assert.Contains(t, cl, "urgency=medium")
}

func TestNode_APTRepositories(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	repos, err := node.APTRepositories(context.Background())
	require.NoError(t, err)
	require.NotNil(t, repos)
	assert.Equal(t, "abcdef0123456789", repos.Digest)
	require.Len(t, repos.Files, 1)
	assert.Equal(t, "/etc/apt/sources.list", repos.Files[0].Path)
	assert.Equal(t, "list", repos.Files[0].FileType)
	require.Len(t, repos.Files[0].Repositories, 1)
	assert.True(t, repos.Files[0].Repositories[0].Enabled)
	assert.Contains(t, repos.Files[0].Repositories[0].URIs, "http://deb.debian.org/debian")

	require.Len(t, repos.Infos, 1)
	assert.Equal(t, "warning", repos.Infos[0].Kind)

	// Standard-repos status is *bool to distinguish "configured but disabled" (false)
	// from "not present" (nil).
	require.GreaterOrEqual(t, len(repos.StandardRepos), 2)
	var ent, nosub *APTStandardRepo
	for _, sr := range repos.StandardRepos {
		switch sr.Handle {
		case "enterprise":
			ent = sr
		case "no-subscription":
			nosub = sr
		}
	}
	require.NotNil(t, ent)
	require.NotNil(t, ent.Status)
	assert.False(t, *ent.Status)
	require.NotNil(t, nosub)
	assert.Nil(t, nosub.Status)
}

func TestNode_APTChangeRepository(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}
	err := node.APTChangeRepository(context.Background(), "/etc/apt/sources.list.d/pve-enterprise.list", 0, false, "abcdef0123456789")
	assert.NoError(t, err)
}

func TestNode_APTAddRepository(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}
	err := node.APTAddRepository(context.Background(), "no-subscription", "abcdef0123456789")
	assert.NoError(t, err)
}

func TestNode_APTVersions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()

	node := &Node{client: mockClient(), Name: "node1"}

	versions, err := node.APTVersions(context.Background())
	require.NoError(t, err)
	require.Len(t, versions, 2)
	assert.Equal(t, "pve-manager", versions[0].Package)
	assert.Equal(t, "Installed", versions[0].CurrentState)
	assert.Equal(t, "9.1-1", versions[0].ManagerVersion)
	assert.Equal(t, "proxmox-ve", versions[1].Package)
	assert.NotEmpty(t, versions[1].RunningKernel)
}
