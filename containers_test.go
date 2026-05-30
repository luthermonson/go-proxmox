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
	// Returned snapshots must be wired to their parent container so
	// instance methods can target the right snapshot path.
	for _, s := range snapshots {
		assert.Equal(t, "node1", s.Node)
		assert.Equal(t, 101, s.VMID)
	}
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

func TestContainer_Snapshot_Getter(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	c := &Container{client: client, Node: "node1", VMID: 101}
	s := c.Snapshot("snapshot1")
	assert.NotNil(t, s)
	assert.Equal(t, "snapshot1", s.Name)
	assert.Equal(t, "node1", s.Node)
	assert.Equal(t, 101, s.VMID)
}

func TestContainerSnapshot_SubResources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	entries, err := container.Snapshot("snapshot1").SubResources(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, entries)
}

func TestContainerSnapshot_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Snapshot("snapshot1").Delete(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, task)
}

func TestContainerSnapshot_Rollback(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	container := Container{
		client: client,
		Node:   "node1",
		VMID:   101,
	}
	task, err := container.Snapshot("snapshot1").Rollback(ctx, true)
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

func TestContainerSnapshot_Config(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}
	config, err := c.Snapshot("snapshot1").Config(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "First snapshot", config["description"])
}

func TestContainerSnapshot_UpdateConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	c := &Container{client: client, Node: "node1", VMID: 101}

	// nil options still posts a body successfully (was broken pre-fix)
	assert.Nil(t, c.Snapshot("snapshot1").UpdateConfig(ctx, nil))

	// explicit description
	assert.Nil(t, c.Snapshot("snapshot1").UpdateConfig(ctx, &ContainerSnapshotUpdateOptions{
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

// ----- Lifecycle / status helpers -----

func TestContainer_Ping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := ct100().Ping(context.Background())
	assert.Nil(t, err)
}

func TestContainer_TermProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	term, err := ct100().TermProxy(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, term)
	assert.Equal(t, "root@pam", term.User)
	assert.Contains(t, term.UPID, "vncproxy")
}

func TestContainer_Feature(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	has, err := ct100().Feature(context.Background())
	assert.Nil(t, err)
	assert.True(t, has)
}

func TestContainer_Migrate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := ct100().Migrate(context.Background(), &ContainerMigrateOptions{
		Target: "node2",
		Online: IntOrBool(true),
	})
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, string(task.UPID), "vzmigrate")
}

func TestContainer_Resize(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := ct100().Resize(context.Background(), "rootfs", "+1G")
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, string(task.UPID), "vzresize")
}

func TestContainer_MoveVolume(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := ct100().MoveVolume(context.Background(), &VirtualMachineMoveDiskOptions{
		Disk:    "rootfs",
		Storage: "local-lvm",
	})
	assert.Nil(t, err)
	require.NotNil(t, task)
	assert.Contains(t, string(task.UPID), "vzmovevolume")
}

func TestContainer_VNCProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vnc, err := ct100().VNCProxy(context.Background(), VNCProxyOptions{
		Websocket: "1",
		Height:    768,
		Width:     1024,
	})
	assert.Nil(t, err)
	require.NotNil(t, vnc)
	assert.Equal(t, "root@pam", vnc.User)
	assert.NotEmpty(t, vnc.Ticket)
}

// ----- Tag helpers (SplitTags / AddTag noop paths) -----

func TestContainer_SplitTags(t *testing.T) {
	c := &Container{
		ContainerConfig: &ContainerConfig{Tags: "a;b;c"},
	}
	c.SplitTags()
	assert.Equal(t, []string{"a", "b", "c"}, c.ContainerConfig.TagsSlice)
}

func TestContainer_HasTag_NilConfig(t *testing.T) {
	c := &Container{}
	assert.False(t, c.HasTag("any"))

	// empty tag string also returns false
	c.ContainerConfig = &ContainerConfig{}
	assert.False(t, c.HasTag("any"))
}

func TestContainer_AddTag_NoopWhenPresent(t *testing.T) {
	c := &Container{
		ContainerConfig: &ContainerConfig{Tags: "existing"},
	}
	_, err := c.AddTag(context.Background(), "existing")
	assert.ErrorIs(t, err, ErrNoop)
}

func TestContainer_RemoveTag_NoopWhenAbsent(t *testing.T) {
	c := &Container{
		ContainerConfig: &ContainerConfig{Tags: "existing"},
	}
	_, err := c.RemoveTag(context.Background(), "missing")
	assert.ErrorIs(t, err, ErrNoop)
}

// TestContainer_AddTag_NilTagsSlice covers the SplitTags() lazy-init branch in
// AddTag — the existing test enters via node.Container which pre-populates
// TagsSlice through HasTag, so the explicit nil branch on AddTag itself was
// otherwise unreached.
func TestContainer_AddTag_NilTagsSlice(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	c := &Container{
		client:          mockClient(),
		Node:            "node1",
		VMID:            101,
		ContainerConfig: &ContainerConfig{Tags: "preexisting"},
	}
	// HasTag of an absent value short-circuits before priming TagsSlice, so
	// the nil branch in AddTag fires on the next line.
	_, err := c.AddTag(context.Background(), "fresh")
	assert.Nil(t, err)
	assert.Contains(t, c.ContainerConfig.TagsSlice, "fresh")
	assert.Contains(t, c.ContainerConfig.TagsSlice, "preexisting")
}

// TestContainer_RemoveTag_NilTagsSlice mirrors AddTag — exercises the
// SplitTags() lazy-init when TagsSlice is nil on entry.
func TestContainer_RemoveTag_NilTagsSlice(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	c := &Container{
		client:          mockClient(),
		Node:            "node1",
		VMID:            101,
		ContainerConfig: &ContainerConfig{Tags: "keep;drop"},
	}
	_, err := c.RemoveTag(context.Background(), "drop")
	assert.Nil(t, err)
	assert.False(t, c.HasTag("drop"))
	assert.True(t, c.HasTag("keep"))
}

// TestContainer_RRDData_TooManyCFs covers the variadic guard rejecting >1 cf.
func TestContainer_RRDData_TooManyCFs(t *testing.T) {
	c := &Container{client: mockClient(), Node: "node1", VMID: 100}
	_, err := c.RRDData(context.Background(), TimeframeHour, AVERAGE, MAX)
	assert.Error(t, err)
}

// TestContainer_RRD_TooManyCFs mirrors RRDData for the RRD endpoint.
func TestContainer_RRD_TooManyCFs(t *testing.T) {
	c := &Container{client: mockClient(), Node: "node1", VMID: 100}
	_, err := c.RRD(context.Background(), "cpu", TimeframeHour, AVERAGE, MAX)
	assert.Error(t, err)
}

// ----- Error paths: cover post-call `if err != nil { return nil, err }`
// branches by routing each mutating endpoint at a deliberately-unmocked VMID
// (999). Gock has no matcher so the request fails, exercising the early
// return that the happy-path tests skip.

func errCt() *Container {
	return &Container{client: mockClient(), Node: "node1", VMID: 999}
}

func TestContainer_ErrorPaths_Mutations(t *testing.T) {
	defer gock.Off()
	ctx := context.Background()

	// Delete (DeleteWithParams)
	_, err := errCt().Delete(ctx, &ContainerDeleteOptions{Force: true})
	assert.Error(t, err)

	// Lifecycle (Post)
	_, err = errCt().Start(ctx)
	assert.Error(t, err)
	_, err = errCt().Stop(ctx)
	assert.Error(t, err)
	_, err = errCt().Suspend(ctx)
	assert.Error(t, err)
	_, err = errCt().Reboot(ctx)
	assert.Error(t, err)
	_, err = errCt().Resume(ctx)
	assert.Error(t, err)
	_, err = errCt().Shutdown(ctx, false, 30)
	assert.Error(t, err)

	// Clone with explicit NewID (skips the NextID branch but still hits the post error)
	_, _, err = errCt().Clone(ctx, &ContainerCloneOptions{NewID: 1000})
	assert.Error(t, err)

	// Migrate / Resize / MoveVolume
	_, err = errCt().Migrate(ctx, &ContainerMigrateOptions{Target: "node2"})
	assert.Error(t, err)
	_, err = errCt().Resize(ctx, "rootfs", "+1G")
	assert.Error(t, err)
	_, err = errCt().MoveVolume(ctx, &VirtualMachineMoveDiskOptions{Disk: "rootfs", Storage: "local"})
	assert.Error(t, err)

	// Snapshots / NewSnapshot / Snapshot ops
	_, err = errCt().Snapshots(ctx)
	assert.Error(t, err)
	_, err = errCt().NewSnapshot(ctx, "snap1")
	assert.Error(t, err)
	_, err = errCt().Snapshot("snap1").Rollback(ctx, true)
	assert.Error(t, err)
	_, err = errCt().Snapshot("snap1").Delete(ctx)
	assert.Error(t, err)

	// Firewall rules / NewFirewallRule
	_, err = errCt().FirewallRules(ctx)
	assert.Error(t, err)
	assert.Error(t, errCt().NewFirewallRule(ctx, &FirewallRule{Type: "in", Action: "ACCEPT"}))

	// RemoteMigrate
	_, err = errCt().RemoteMigrate(ctx, &ContainerRemoteMigrateOptions{TargetEndpoint: "x"})
	assert.Error(t, err)
}

// ----- Firewall: top-level + aliases -----

func TestContainer_Firewall(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	fw, err := ct100().Firewall(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, fw)
	assert.NotEmpty(t, fw.Rules)
	assert.NotEmpty(t, fw.Aliases)
}

func TestContainer_FirewallAliases(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	c := ct100()

	aliases, err := c.GetFirewallAliases(ctx)
	assert.Nil(t, err)
	assert.Len(t, aliases, 1)
	assert.Equal(t, "lan", aliases[0].Name)

	alias, err := c.GetFirewallAlias(ctx, "lan")
	assert.Nil(t, err)
	require.NotNil(t, alias)
	assert.Equal(t, "192.168.0.0/16", alias.Cidr)

	assert.Nil(t, c.NewFirewallAlias(ctx, &FirewallAlias{Name: "lan", Cidr: "192.168.0.0/16"}))
	assert.Nil(t, c.UpdateFirewallAlias(ctx, "lan", &FirewallAlias{Cidr: "192.168.0.0/16", Comment: "updated"}))
	assert.Nil(t, c.DeleteFirewallAlias(ctx, "lan"))
}

// ----- Firewall: IPSet collection + entries -----

func TestContainer_FirewallIPSet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	c := ct100()

	sets, err := c.GetFirewallIPSet(ctx)
	assert.Nil(t, err)
	assert.Len(t, sets, 1)
	assert.Equal(t, "blocked", sets[0].Name)

	assert.Nil(t, c.NewFirewallIPSet(ctx, FirewallIPSetCreationOption{Name: "blocked"}))
	assert.Nil(t, c.DeleteFirewallIPSet(ctx, "blocked", true))
}

func TestContainer_FirewallIPSetEntries(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	c := ct100()

	entries, err := c.GetFirewallIPSetEntries(ctx, "blocked")
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "10.0.0.1", entries[0].CIDR)

	entry, err := c.GetFirewallIPSetEntry(ctx, "blocked", "10.0.0.1")
	assert.Nil(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "10.0.0.1", entry.CIDR)

	assert.Nil(t, c.NewFirewallIPSetEntry(ctx, "blocked", FirewallIPSetEntryCreationOption{CIDR: "10.0.0.1"}))
	assert.Nil(t, c.UpdateFirewallIPSetEntry(ctx, "blocked", "10.0.0.1", &FirewallIPSetEntryUpdateOption{Comment: "updated"}))
	assert.Nil(t, c.DeleteFirewallIPSetEntry(ctx, "blocked", "10.0.0.1", ""))
}

// ----- Firewall: rules collection + per-rule instance -----

func TestContainer_FirewallRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	c := ct100()
	rules, err := c.FirewallRules(context.Background())
	assert.Nil(t, err)
	assert.Len(t, rules, 2)
	// Returned rules must be wired with parent context so instance methods
	// resolve to the container's /firewall/rules/{pos} path.
	for _, r := range rules {
		assert.NotNil(t, r)
		assert.Equal(t, "ACCEPT", rules[0].Action)
	}
}

func TestContainer_FirewallRule_Getter(t *testing.T) {
	c := ct100()
	r := c.FirewallRule(0)
	assert.NotNil(t, r)
	assert.Equal(t, 0, r.Pos)
}

func TestContainer_FirewallRule_GetUpdateDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	c := ct100()

	rule := c.FirewallRule(0)
	assert.Nil(t, rule.Get(ctx))
	assert.Equal(t, "ACCEPT", rule.Action)
	assert.Equal(t, "in", rule.Type)
	assert.True(t, rule.IsEnable())

	rule.Comment = "updated via test"
	assert.Nil(t, rule.Update(ctx))
	assert.Nil(t, rule.Delete(ctx))
}

func TestContainer_NewFirewallRule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	c := ct100()
	rule := &FirewallRule{Type: "in", Action: "ACCEPT", Enable: 1}
	assert.Nil(t, c.NewFirewallRule(context.Background(), rule))
	// After a successful POST, the rule must be wired with parent context so
	// follow-up Get/Update/Delete on the returned instance route correctly.
	assert.Nil(t, rule.Get(context.Background()))
}

// ----- Firewall: options -----

func TestContainer_FirewallOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	c := ct100()

	opts, err := c.GetFirewallOptions(ctx)
	assert.Nil(t, err)
	require.NotNil(t, opts)
	assert.Equal(t, "DROP", opts.PolicyIn)

	assert.Nil(t, c.UpdateFirewallOptions(ctx, &FirewallVirtualMachineOption{
		PolicyIn:  "ACCEPT",
		PolicyOut: "ACCEPT",
	}))
}
