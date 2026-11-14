package proxmox

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStorageRef_MakeNoRequests pins that constructing a storage ref touches
// nothing and only wires the identifying fields, leaving the status fields
// zero (the storage class-3 behaviour, mirroring the VM/container refs).
func TestStorageRef_MakeNoRequests(t *testing.T) {
	client, rec := newRefsClient(t)

	s := client.NodeRef("node1").StorageRef("local")

	assert.Empty(t, rec.calls(), "constructing a storage ref must not touch the API")
	assert.False(t, gock.HasUnmatchedRequest())

	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, "local", s.Name)
	assert.Same(t, client, s.client)
	// Status fields stay zero until Ping runs.
	assert.Zero(t, s.Enabled)
	assert.Empty(t, s.Content)
	assert.Zero(t, s.Total)
}

// TestStorageRef_GetContentIsOneRequest is the headline goal: node -> storage
// -> content is exactly one GET, nothing spent on traversal.
func TestStorageRef_GetContentIsOneRequest(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/storage/local/content$").Reply(200).
		JSON(`{"data": [{"volid": "local:iso/foo.iso", "content": "iso", "size": 123}]}`)

	content, err := client.NodeRef("node1").StorageRef("local").GetContent(context.Background())
	require.NoError(t, err)
	require.Len(t, content, 1)
	assert.Equal(t, []string{"GET /nodes/node1/storage/local/content"}, rec.calls())
}

// TestStorageRef_URLOnlyMethods covers that the methods addressing the storage
// by node+name in the URL issue their one request and no lazy status fetch.
func TestStorageRef_URLOnlyMethods(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		call func(*Storage) error
		want string
	}{
		{"GetContent", func(s *Storage) error { _, err := s.GetContent(ctx); return err }, "GET /nodes/node1/storage/local/content"},
		{"DeleteContent", func(s *Storage) error {
			_, err := s.DeleteContent(ctx, "local:iso/foo.iso")
			return err
		}, "DELETE /nodes/node1/storage/local/content/local:iso/foo.iso"},
		{"ISO", func(s *Storage) error { _, err := s.ISO(ctx, "foo.iso"); return err }, "GET /nodes/node1/storage/local/content/local:iso/foo.iso"},
		{"VzTmpl", func(s *Storage) error { _, err := s.VzTmpl(ctx, "foo.tar.gz"); return err }, "GET /nodes/node1/storage/local/content/local:vztmpl/foo.tar.gz"},
		{"Identity", func(s *Storage) error { _, err := s.Identity(ctx); return err }, "GET /nodes/node1/storage/local/identity"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, rec := newRefsClient(t)
			gock.New(refsTestURI).Get("^/nodes/node1/storage/local/content$").Persist().Reply(200).JSON(`{"data": []}`)
			gock.New(refsTestURI).Get("^/nodes/node1/storage/local/content/").Persist().Reply(200).JSON(`{"data": {}}`)
			gock.New(refsTestURI).Get("^/nodes/node1/storage/local/identity$").Persist().Reply(200).JSON(`{"data": {}}`)
			gock.New(refsTestURI).Delete("^/nodes/node1/storage/local/content/").Persist().Reply(200).JSON(upid("imgdel", "local"))

			s := client.NodeRef("node1").StorageRef("local")
			require.NoError(t, tc.call(s))
			assert.Equal(t, []string{tc.want}, rec.calls())
		})
	}
}

// TestStorageRef_Ping fills the handle from the status endpoint and shows
// the identity survives the refresh.
func TestStorageRef_Ping(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/storage/local/status$").Reply(200).
		JSON(`{"data": {"enabled": 1, "active": 1, "content": "iso,vztmpl,backup", "total": 1000, "used": 400, "avail": 600, "type": "dir"}}`)

	s := client.NodeRef("node1").StorageRef("local")
	require.NoError(t, s.Ping(context.Background()))
	assert.Equal(t, []string{"GET /nodes/node1/storage/local/status"}, rec.calls())

	assert.Equal(t, 1, s.Enabled)
	assert.Equal(t, "iso,vztmpl,backup", s.Content)
	assert.Equal(t, uint64(1000), s.Total)
	assert.Equal(t, uint64(600), s.Avail)
	// Identity and client survive the refresh.
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, "local", s.Name)
	assert.Same(t, client, s.client)
}

// TestStorageRef_PingFailureLeavesHandle guards that a failed refresh does
// not blow away the handle's identity (same contract as VM Ping).
func TestStorageRef_PingFailureLeavesHandle(t *testing.T) {
	client, _ := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/storage/local/status$").Reply(500)

	s := client.NodeRef("node1").StorageRef("local")
	require.Error(t, s.Ping(context.Background()))
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, "local", s.Name)
	assert.Zero(t, s.Enabled)
}

// TestNodeStorage_Deprecated keeps the deprecated fetch-on-construct path
// working: it still fetches status and wires identity.
func TestNodeStorage_Deprecated(t *testing.T) {
	client, rec := newRefsClient(t)
	gock.New(refsTestURI).Get("^/nodes/node1/storage/local/status$").Reply(200).
		JSON(`{"data": {"enabled": 1, "content": "iso"}}`)

	s, err := client.NodeRef("node1").Storage(context.Background(), "local")
	require.NoError(t, err)
	assert.Equal(t, []string{"GET /nodes/node1/storage/local/status"}, rec.calls())
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, "local", s.Name)
	assert.Equal(t, 1, s.Enabled)
}
