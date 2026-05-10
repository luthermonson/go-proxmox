package proxmox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerConfig_UnmarshalJSON_BeyondTen exercises issue #211 for LXC:
// mp0..mp255 and dev0..dev255 in particular routinely exceed 9 in real setups.
func TestContainerConfig_UnmarshalJSON_BeyondTen(t *testing.T) {
	body := []byte(`{
		"mp0": "/srv/data,mp=/data",
		"mp42": "/srv/forty-two,mp=/forty-two",
		"mp255": "/srv/last,mp=/last",
		"net0": "name=eth0,bridge=vmbr0",
		"net20": "name=eth20,bridge=vmbr20",
		"dev15": "/dev/sdb15",
		"unused100": "local-lvm:subvol-100-unused-100"
	}`)

	var cfg ContainerConfig
	assert.NoError(t, json.Unmarshal(body, &cfg))

	assert.Equal(t, "name=eth0,bridge=vmbr0", cfg.Nets["net0"])
	assert.Equal(t, "/srv/forty-two,mp=/forty-two", cfg.Mps["mp42"])
	assert.Equal(t, "/srv/last,mp=/last", cfg.Mps["mp255"])
	assert.Equal(t, "name=eth20,bridge=vmbr20", cfg.Nets["net20"])
	assert.Equal(t, "/dev/sdb15", cfg.Devs["dev15"])
	assert.Equal(t, "local-lvm:subvol-100-unused-100", cfg.Unuseds["unused100"])
}

// TestNode_ContainerConfig_HighIndices is the integration-shaped regression
// test for issue #211 on the LXC side: routes mp42/mp255, dev15, net20,
// unused100 through node.Container(ctx, 102) and the gock mock.
func TestNode_ContainerConfig_HighIndices(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	ct, err := node.Container(ctx, 102)
	assert.Nil(t, err)
	assert.NotNil(t, ct)
	assert.NotNil(t, ct.ContainerConfig)

	cfg := ct.ContainerConfig

	assert.Equal(t, "/srv/data,mp=/data", cfg.Mps["mp0"])
	assert.Equal(t, "/srv/forty-two,mp=/forty-two", cfg.Mps["mp42"])
	assert.Equal(t, "/srv/last,mp=/last", cfg.Mps["mp255"])
	assert.Equal(t, "name=eth0,bridge=vmbr0", cfg.Nets["net0"])
	assert.Equal(t, "name=eth20,bridge=vmbr20", cfg.Nets["net20"])
	assert.Equal(t, "/dev/sdb15", cfg.Devs["dev15"])
	assert.Equal(t, "local-lvm:subvol-102-unused-100", cfg.Unuseds["unused100"])
}

func TestContainer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}
	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)
	assert.NotEmpty(t, container, container.ContainerConfig)
}

func TestContainers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node := Node{
		client: client,
		Name:   "node1",
	}
	containers, err := node.Containers(ctx)
	assert.Nil(t, err)
	assert.Len(t, containers, 3)
}

func TestContainerClone(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	cloneOptions := ContainerCloneOptions{
		NewID: 102,
	}
	newid, _, err := container.Clone(ctx, &cloneOptions)
	assert.Equal(t, cloneOptions.NewID, newid)
	assert.Nil(t, err)
}

func TestContainerCloneWithoutNewID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	cloneOptions := ContainerCloneOptions{}
	newid, _, err := container.Clone(ctx, &cloneOptions)
	assert.Equal(t, 100, newid)
	assert.Nil(t, err)
}

func TestContainerDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(ctx, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

// TestContainerDelete_ForceParam verifies issue #149: passing
// ContainerDeleteOptions{Force: true} actually puts force=1 on the wire as a
// query parameter to DELETE /nodes/{node}/lxc/{vmid}. The gock mock only
// matches when the parameter is present, so a regression where the option is
// silently dropped (as DeleteFirewallIPSet currently does, by passing the map
// as the response target instead of via DeleteWithParams) makes this test fail
// with "cannot match any request".
func TestContainerDelete_ForceParam(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Delete("^/nodes/node1/lxc/101$").
		MatchParam("force", "1").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdestroy:101:root@pam:"}`)

	client := mockClient()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(context.Background(), &ContainerDeleteOptions{Force: true})
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

// TestContainerDelete_AllOptionsOnWire is the equivalent for purge and
// destroy-unreferenced-disks: all three options must appear in the query
// string when set, with the spec-correct hyphenated key for the last one.
func TestContainerDelete_AllOptionsOnWire(t *testing.T) {
	defer gock.Off()

	gock.New(TestURI).
		Delete("^/nodes/node1/lxc/101$").
		MatchParams(map[string]string{
			"force":                      "1",
			"purge":                      "1",
			"destroy-unreferenced-disks": "1",
		}).
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdestroy:101:root@pam:"}`)

	client := mockClient()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Delete(context.Background(), &ContainerDeleteOptions{
		Force:                    true,
		Purge:                    true,
		DestroyUnreferencedDisks: true,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Config(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerStart(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Start(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerStop(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Stop(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerSuspend(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Suspend(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerReboot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Reboot(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerResume(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Resume(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerShutdown(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Shutdown(ctx, false, 60)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerTemplate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	err := container.Template(ctx)
	assert.Nil(t, err)
}

func TestContainerSnapshots(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	snapshots, err := container.Snapshots(ctx)
	assert.Nil(t, err)
	assert.Len(t, snapshots, 3)
}

func TestContainerNewSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.NewSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerGetSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	snapshot, err := container.GetSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotNil(t, snapshot)
}

func TestContainerDeleteSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.DeleteSnapshot(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerRollbackSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.RollbackSnapshot(ctx, "snapshot1", true)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerInterfaces(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	interfaces, err := container.Interfaces(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, interfaces)
	assert.Equal(t, interfaces, ContainerInterfaces{{HWAddr: "00:00:00:00:00:00", Inet: "127.0.0.1/8", Name: "lo", Inet6: "::1/128"}, {Inet6: "fe80::be24:11ff:fe89:6707/64", Name: "eth0", HWAddr: "bc:24:11:89:67:07", Inet: "192.168.3.95/22"}})
}

func TestContainerTagsSlice(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.NotEmpty(t, container.ContainerConfig.TagsSlice)
}

func TestContainer_AddTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	_, err = container.AddTag(ctx, "newTag")
	assert.Nil(t, err)
	assert.True(t, container.HasTag("newTag"))
}

func TestContainer_HasTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.True(t, container.HasTag("tag1"))
	assert.False(t, container.HasTag("not_there"))
}

func TestContainer_RemoveTag(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	container, err := node.Container(ctx, 101)
	assert.Nil(t, err)

	assert.True(t, container.HasTag("tag1"))
	_, err = container.RemoveTag(ctx, "tag1")
	assert.Nil(t, err)
	assert.False(t, container.HasTag("tag1"))
}

func TestContainerRRDData(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	data, err := c.RRDData(ctx, TimeframeHour)
	assert.Nil(t, err)
	assert.Len(t, data, 2)
	assert.Equal(t, uint64(1715299200), data[0].Time)
}

func TestContainerPending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	pending, err := c.Pending(ctx)
	assert.Nil(t, err)
	assert.Len(t, pending, 3)
	assert.Equal(t, "memory", pending[0].Key)
	assert.EqualValues(t, 2048, pending[0].Pending)
	assert.Equal(t, "swap", pending[2].Key)
	assert.Equal(t, 1, pending[2].Delete)
}

func TestContainerRRD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	rrd, err := c.RRD(ctx, "cpu", TimeframeHour)
	assert.Nil(t, err)
	require.NotNil(t, rrd)
	assert.Contains(t, rrd.Filename, "101.png")
}

func TestContainerRemoteMigrate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	task, err := c.RemoteMigrate(ctx, &ContainerRemoteMigrateOptions{
		TargetEndpoint: "apitoken=PVEAPIToken=user@pam!tok=secret host=target.example.com fingerprint=AA:BB",
		TargetBridge:   "vmbr0=vmbr0",
		TargetStorage:  "local=local",
		TargetVMID:     201,
		Online:         IntOrBool(true),
	})
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, string(task.UPID), "vzremote-migrate")
}

func TestContainerSpiceProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	spice, err := c.SpiceProxy(ctx)
	assert.Nil(t, err)
	require.NotNil(t, spice)
	assert.Equal(t, "spice", spice.Type)
	assert.Equal(t, "node1.example.com", spice.Host)
	assert.Equal(t, "61024", spice.Port)
	assert.Equal(t, "secret-ticket", spice.Password)
	assert.Equal(t, "61025", spice.TLSPort)
}

func TestContainerFirewallLog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	entries, err := c.FirewallLog(ctx)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	// PVE returns [n, "text"] tuples; verify the custom UnmarshalJSON flattens them
	assert.Equal(t, 42, entries[0].LineNum)
	assert.Contains(t, entries[0].Text, "policy DROP")
	assert.Equal(t, 43, entries[1].LineNum)
	assert.Contains(t, entries[1].Text, "policy ACCEPT")
}

func TestContainerFirewallRefs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	refs, err := c.FirewallRefs(ctx)
	assert.Nil(t, err)
	assert.Len(t, refs, 2)
	assert.Equal(t, "alias", refs[0].Type)
	assert.Equal(t, "lan", refs[0].Name)
	assert.Equal(t, "ipset", refs[1].Type)
	assert.Equal(t, "blocked", refs[1].Name)
}

func TestContainerGetSnapshotConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	config, err := c.GetSnapshotConfig(ctx, "snapshot1")
	assert.Nil(t, err)
	assert.Equal(t, "First snapshot", config["description"])
}

func TestContainerUpdateSnapshot(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}

	// nil options still posts a body successfully (was broken pre-fix)
	assert.Nil(t, c.UpdateSnapshot(ctx, "snapshot1", nil))

	// explicit description
	assert.Nil(t, c.UpdateSnapshot(ctx, "snapshot1", &ContainerSnapshotUpdateOptions{
		Description: "updated by tests",
	}))
}

func ct100() *Container {
	return &Container{client: mockClient(), Node: "node1", VMID: 100}
}

func TestContainer_DirIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := ct100().DirIndex(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Subdir] = true
	}
	assert.True(t, names["config"])
	assert.True(t, names["status"])
	assert.True(t, names["snapshot"])
	assert.True(t, names["firewall"])
	assert.True(t, names["mtunnel"])
}

func TestContainer_StatusIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := ct100().StatusIndex(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Subdir] = true
	}
	assert.True(t, names["current"])
	assert.True(t, names["start"])
	assert.True(t, names["stop"])
}

func TestContainer_MigratePreconditions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pre, err := ct100().MigratePreconditions(context.Background(), "")
	assert.Nil(t, err)
	require.NotNil(t, pre)
	assert.True(t, pre.Running)
	assert.Contains(t, pre.AllowedNodes, "node2")

	// with target — same mock fields, just verify the call shape
	pre2, err := ct100().MigratePreconditions(context.Background(), "node2")
	assert.Nil(t, err)
	require.NotNil(t, pre2)
}

func TestContainer_MigrationTunnel(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	tunnel, err := ct100().MigrationTunnel(context.Background(), nil)
	assert.Nil(t, err)
	require.NotNil(t, tunnel)
	assert.Equal(t, "/run/pve/100.mtunnel", tunnel.Socket)
	assert.Equal(t, "PVEMTUNNELTICKET:lxc-abc123", tunnel.Ticket)
	assert.Contains(t, tunnel.UPID, "vzmtunnel")

	// with explicit options
	tunnel2, err := ct100().MigrationTunnel(context.Background(), &ContainerMigrationTunnelOptions{
		Bridges:  "vmbr0",
		Storages: "local-lvm",
	})
	assert.Nil(t, err)
	require.NotNil(t, tunnel2)
}

func TestContainer_MigrationTunnelWebSocketPath(t *testing.T) {
	c := ct100()
	tunnel := &ContainerMigrationTunnel{
		Socket: "/run/pve/100.mtunnel",
		Ticket: "PVEMTUNNELTICKET:lxc-abc123",
	}
	path := c.MigrationTunnelWebSocketPath(tunnel)
	assert.Contains(t, path, "/nodes/node1/lxc/100/mtunnelwebsocket")
	assert.Contains(t, path, "socket=")
	assert.Contains(t, path, "ticket=")
	// ticket should be URL-encoded (':' becomes %3A)
	assert.Contains(t, path, "PVEMTUNNELTICKET%3Alxc-abc123")

	// nil tunnel still returns a valid path skeleton
	emptyPath := c.MigrationTunnelWebSocketPath(nil)
	assert.Contains(t, emptyPath, "/nodes/node1/lxc/100/mtunnelwebsocket")
}
