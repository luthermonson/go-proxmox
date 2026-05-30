package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func looseNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_GetConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cfg, err := looseNode().GetConfig(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "primary hypervisor", cfg.Description)
	assert.Equal(t, 80, cfg.BallooningTarget)
	assert.Equal(t, "abc123", cfg.Digest)
}

func TestNode_GetConfigProperty(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cfg, err := looseNode().GetConfigProperty(context.Background(), "description")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestNode_UpdateConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	target := 75
	err := looseNode().UpdateConfig(context.Background(), &NodeConfigOptions{
		Description:      "updated",
		BallooningTarget: &target,
		Delete:           "wakeonlan",
	})
	assert.Nil(t, err)

	assert.Nil(t, looseNode().UpdateConfig(context.Background(), nil))
}

func TestNode_Hosts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	hosts, err := looseNode().Hosts(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, hosts.Data, "node1.example.com")
	assert.Equal(t, "hostsdigest123", hosts.Digest)
}

func TestNode_UpdateHosts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := looseNode().UpdateHosts(context.Background(),
		"127.0.0.1 localhost\n10.0.0.1 node1\n", "hostsdigest123")
	assert.Nil(t, err)

	assert.Nil(t, looseNode().UpdateHosts(context.Background(), "data", ""))
}

func TestNode_RebootAndShutdown(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, looseNode().Reboot(context.Background()))
	assert.Nil(t, looseNode().Shutdown(context.Background()))
}

func TestNode_Journal(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	lines, err := looseNode().Journal(context.Background(), &NodeJournalOptions{
		LastEntries: 100,
	})
	assert.Nil(t, err)
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[0], "systemd[1]")

	lines, err = looseNode().Journal(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, lines)
}

func TestNode_Syslog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := looseNode().Syslog(context.Background(), &NodeSyslogOptions{
		Limit:   50,
		Service: "pveproxy",
	})
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, 1, entries[0].N)
}

func TestNode_Netstat(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := looseNode().Netstat(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "tap100i0", entries[0]["dev"])
}

func TestNode_Execute(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	results, err := looseNode().Execute(context.Background(), []*NodeExecuteCommand{
		{Method: "GET", Path: "/version"},
		{Method: "GET", Path: "/cluster/status"},
	})
	assert.Nil(t, err)
	assert.Len(t, results, 2)

	_, err = looseNode().Execute(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestNode_VNCShell(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vnc, err := looseNode().VNCShell(context.Background(), &NodeVNCShellOptions{
		Cmd:       NodeConsoleLogin,
		Width:     1024,
		Height:    768,
		WebSocket: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "vncpw", vnc.Password)
	assert.NotEmpty(t, vnc.Ticket)
}

func TestNode_SpiceShell(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	proxy, err := looseNode().SpiceShell(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, "node1.example.com", proxy.Host)
	assert.Equal(t, "spice", proxy.Type)
}

func TestNode_SpiceShell_AllOpts(t *testing.T) {
	// Exercise the cmd/cmd-opts/proxy branches.
	mocks.On(mockConfig)
	defer mocks.Off()
	proxy, err := looseNode().SpiceShell(context.Background(), &NodeSpiceShellOptions{
		Cmd:     NodeConsoleUpgrade,
		CmdOpts: "--yes",
		Proxy:   "proxy.example.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, proxy)
}

func TestNode_VNCShell_DefaultOpts(t *testing.T) {
	// Nil opts path — the conditional `if opts != nil` short-circuits.
	mocks.On(mockConfig)
	defer mocks.Off()
	vnc, err := looseNode().VNCShell(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, vnc)
}

func TestNode_Journal_AllBranches(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	_, err := looseNode().Journal(context.Background(), &NodeJournalOptions{
		Since:       1715000000,
		Until:       1715100000,
		StartCursor: "cur-start",
		EndCursor:   "cur-end",
	})
	assert.Nil(t, err)
}

func TestNode_Syslog_AllBranches(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	_, err := looseNode().Syslog(context.Background(), &NodeSyslogOptions{
		Start:   1,
		Limit:   10,
		Since:   "2025-01-01",
		Until:   "2025-12-31",
		Service: "pveproxy",
	})
	assert.Nil(t, err)

	// Empty opts also exercises the "no query params" path.
	_, err = looseNode().Syslog(context.Background(), &NodeSyslogOptions{})
	assert.Nil(t, err)

	// nil opts skips the conditional entirely.
	_, err = looseNode().Syslog(context.Background(), nil)
	assert.Nil(t, err)
}

func TestNode_Tasks_AllBranches(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	_, err := looseNode().Tasks(context.Background(), &NodeTasksOptions{
		Errors:       true,
		Limit:        10,
		Since:        1715000000,
		Until:        1715100000,
		Source:       "all",
		Start:        5,
		StatusFilter: "OK",
		TypeFilter:   "vzdump",
		UserFilter:   "root@pam",
		VMID:         100,
	})
	assert.Nil(t, err)
}

func TestNode_GetConfigProperty_EmptyDelegates(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cfg, err := looseNode().GetConfigProperty(context.Background(), "")
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
}

func TestNode_FirewallLog_AllBranches(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	_, err := looseNode().FirewallLog(context.Background(), &NodeFirewallLogOptions{
		Start: 1, Limit: 10, Since: 1715000000, Until: 1715100000,
	})
	assert.Nil(t, err)
}

func TestNode_RRD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	img, err := looseNode().RRD(context.Background(), "cpu,mem", TimeframeHour, AVERAGE)
	assert.Nil(t, err)
	assert.Equal(t, "rrd-node-graph.png", img.Filename)

	_, err = looseNode().RRD(context.Background(), "", TimeframeHour, "")
	assert.NotNil(t, err)
}

func TestNode_RRDData(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	data, err := looseNode().RRDData(context.Background(), TimeframeHour, "")
	assert.Nil(t, err)
	assert.Len(t, data, 2)
}

func TestNode_QueryURLMetadata(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	verify := false
	meta, err := looseNode().QueryURLMetadata(context.Background(),
		"https://example.com/debian.iso", &verify)
	assert.Nil(t, err)
	assert.Equal(t, "debian-12.iso", meta.Filename)
	assert.Equal(t, int64(654311424), meta.Size)

	_, err = looseNode().QueryURLMetadata(context.Background(), "", nil)
	assert.NotNil(t, err)
}

func TestNode_QueryOCIRepoTags(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	tags, err := looseNode().QueryOCIRepoTags(context.Background(), "docker.io/library/alpine")
	assert.Nil(t, err)
	assert.Contains(t, tags, "latest")
	assert.Len(t, tags, 4)

	_, err = looseNode().QueryOCIRepoTags(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_VzdumpDefaults(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	defaults, err := looseNode().VzdumpDefaults(context.Background(), "backup")
	assert.Nil(t, err)
	assert.Equal(t, "snapshot", defaults["mode"])
	assert.Equal(t, "zstd", defaults["compress"])
}

func TestNode_FirewallGetRule(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	rule, err := looseNode().FirewallGetRule(context.Background(), 0)
	assert.Nil(t, err)
	assert.Equal(t, "ACCEPT", rule.Action)
	assert.Equal(t, "ssh", rule.Comment)
}

func TestNode_FirewallLog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := looseNode().FirewallLog(context.Background(), &NodeFirewallLogOptions{Limit: 100})
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Contains(t, entries[0].T, "DROP")

	entries, err = looseNode().FirewallLog(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, entries)
}

func TestNode_Tasks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	tasks, err := looseNode().Tasks(context.Background(), &NodeTasksOptions{
		Limit:  10,
		Source: "all",
		VMID:   100,
	})
	assert.Nil(t, err)
	assert.Len(t, tasks, 2)
	assert.NotNil(t, tasks[0].client) // wired up for chaining
	assert.Equal(t, "vzdump", tasks[0].Type)
}

func TestNode_RevertNetworkChanges(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, looseNode().RevertNetworkChanges(context.Background()))
}

func TestStorage_Identity(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	s := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	id, err := s.Identity(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "local", id.ID)
	assert.Equal(t, "dir", id.Type)
}

func TestStorage_RRD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	s := &Storage{client: mockClient(), Node: "node1", Name: "local"}
	img, err := s.RRD(context.Background(), "total,used", TimeframeWeek, MAX)
	assert.Nil(t, err)
	assert.Equal(t, "rrd-storage-graph.png", img.Filename)

	_, err = s.RRD(context.Background(), "", TimeframeHour, "")
	assert.NotNil(t, err)
}
