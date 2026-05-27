package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// This file fills in the "loose" /nodes/{node}/* gaps that don't warrant
// their own dedicated file: node config + /etc/hosts, power management
// (reboot/shutdown), console launchers, log readers, command execute,
// finished-task listing, and pending-network revert.

// --- /nodes/{node}/config -------------------------------------------------

// GetConfig returns node-level config (acme/acmedomain[0-5], description,
// location, ballooning-target, startall-onboot-delay, wakeonlan, digest).
// PVE encodes substructures (acme, location, wakeonlan) as property strings.
func (n *Node) GetConfig(ctx context.Context) (cfg *NodeConfig, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/config", n.Name), &cfg)
	return
}

// GetConfigProperty fetches just one property from the node config
// ("acme"|"acmedomain0..5"|"ballooning-target"|"description"|"location"
// |"startall-onboot-delay"|"wakeonlan"). The returned config has only that
// field populated; others are zero-valued.
func (n *Node) GetConfigProperty(ctx context.Context, property string) (cfg *NodeConfig, err error) {
	if property == "" {
		return n.GetConfig(ctx)
	}
	q := url.Values{}
	q.Set("property", property)
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/config?%s", n.Name, q.Encode()), &cfg)
	return
}

// UpdateConfig sets node configuration options. Passing nil opts is a no-op
// on the server. Use opts.Delete to unset specific keys.
func (n *Node) UpdateConfig(ctx context.Context, opts *NodeConfigOptions) error {
	if opts == nil {
		opts = &NodeConfigOptions{}
	}
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/config", n.Name), opts, nil)
}

// --- /nodes/{node}/hosts --------------------------------------------------

// Hosts returns the full contents of /etc/hosts plus a digest for optimistic
// concurrency. Pass the digest back to UpdateHosts to refuse the write on
// concurrent modification.
func (n *Node) Hosts(ctx context.Context) (hosts *NodeHosts, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/hosts", n.Name), &hosts)
	return
}

// UpdateHosts overwrites /etc/hosts on the node. digest is optional — if
// non-empty, PVE refuses the write when the current file differs from it.
func (n *Node) UpdateHosts(ctx context.Context, data, digest string) error {
	body := map[string]string{"data": data}
	if digest != "" {
		body["digest"] = digest
	}
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/hosts", n.Name), body, nil)
}

// --- /nodes/{node}/status (power) -----------------------------------------

// Reboot reboots the node. Requires Sys.PowerMgmt on /nodes/{node}.
// PVE returns null — no task to track.
func (n *Node) Reboot(ctx context.Context) error {
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/status", n.Name),
		map[string]string{"command": "reboot"}, nil)
}

// Shutdown powers off the node. Requires Sys.PowerMgmt on /nodes/{node}.
func (n *Node) Shutdown(ctx context.Context) error {
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/status", n.Name),
		map[string]string{"command": "shutdown"}, nil)
}

// --- /nodes/{node}/{journal,syslog,netstat} -------------------------------

// NodeJournalOptions filters the systemd journal read. Pass either Since/
// Until OR StartCursor/EndCursor (PVE rejects mixing). LastEntries conflicts
// with any range.
type NodeJournalOptions struct {
	Since       int64
	Until       int64
	StartCursor string
	EndCursor   string
	LastEntries int
}

// Journal returns systemd journal lines as plain strings. PVE caps line
// count server-side; use LastEntries or a Since/Until range to bound output.
func (n *Node) Journal(ctx context.Context, opts *NodeJournalOptions) (lines []string, err error) {
	path := fmt.Sprintf("/nodes/%s/journal", n.Name)
	if opts != nil {
		q := url.Values{}
		if opts.Since > 0 {
			q.Set("since", strconv.FormatInt(opts.Since, 10))
		}
		if opts.Until > 0 {
			q.Set("until", strconv.FormatInt(opts.Until, 10))
		}
		if opts.StartCursor != "" {
			q.Set("startcursor", opts.StartCursor)
		}
		if opts.EndCursor != "" {
			q.Set("endcursor", opts.EndCursor)
		}
		if opts.LastEntries > 0 {
			q.Set("lastentries", strconv.Itoa(opts.LastEntries))
		}
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}
	err = n.client.Get(ctx, path, &lines)
	return
}

// NodeSyslogOptions filters the classic syslog reader. Since/Until are
// "YYYY-MM-DD[ HH:MM[:SS]]" strings per PVE; Service filters to one unit.
type NodeSyslogOptions struct {
	Start   int
	Limit   int
	Since   string
	Until   string
	Service string
}

// Syslog returns rsyslog lines as {n, t} entries — n is the line number, t
// the text. For systemd journal, use Journal instead.
func (n *Node) Syslog(ctx context.Context, opts *NodeSyslogOptions) (entries []*LogEntry, err error) {
	path := fmt.Sprintf("/nodes/%s/syslog", n.Name)
	if opts != nil {
		q := url.Values{}
		if opts.Start > 0 {
			q.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Since != "" {
			q.Set("since", opts.Since)
		}
		if opts.Until != "" {
			q.Set("until", opts.Until)
		}
		if opts.Service != "" {
			q.Set("service", opts.Service)
		}
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}
	err = n.client.Get(ctx, path, &entries)
	return
}

// Netstat returns per-VM/CT tap interface counters. PVE leaves the response
// shape loose — wrap each entry as a generic map.
func (n *Node) Netstat(ctx context.Context) (entries []map[string]any, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/netstat", n.Name), &entries)
	return
}

// --- /nodes/{node}/execute (batch API call) -------------------------------

// NodeExecuteCommand is one entry in the Execute batch. Args carries the
// API parameters as PVE-native key/value pairs.
type NodeExecuteCommand struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Args   map[string]interface{} `json:"args,omitempty"`
}

// Execute runs a batch of API calls in order. Root-only on PVE. Returns the
// responses in submission order — each entry is the same envelope PVE would
// return from the single endpoint.
func (n *Node) Execute(ctx context.Context, commands []*NodeExecuteCommand) (results []map[string]any, err error) {
	if len(commands) == 0 {
		return nil, errors.New("at least one command is required")
	}
	body := map[string]interface{}{"commands": commands}
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/execute", n.Name), body, &results)
	return
}

// --- /nodes/{node}/{vncshell,spiceshell} ----------------------------------

// NodeConsoleCmd narrows the shell command. Empty defaults to "login"
// (requires root@pam).
type NodeConsoleCmd string

const (
	NodeConsoleLogin       NodeConsoleCmd = "login"
	NodeConsoleUpgrade     NodeConsoleCmd = "upgrade"
	NodeConsoleCephInstall NodeConsoleCmd = "ceph_install"
)

// NodeVNCShellOptions configures the noVNC console launcher. WebSocket
// enables noVNC-style transport; Width/Height are pixel dimensions
// (16-4096 / 16-2160). CmdOpts are null-separated arguments to Cmd.
type NodeVNCShellOptions struct {
	Cmd       NodeConsoleCmd
	CmdOpts   string
	Width     int
	Height    int
	WebSocket bool
}

// VNCShell opens a VNC proxy to the node — typically a serial-like login
// shell. Returns ticket + port for the websocket follow-up call.
func (n *Node) VNCShell(ctx context.Context, opts *NodeVNCShellOptions) (vnc *VNC, err error) {
	body := map[string]interface{}{}
	if opts != nil {
		if opts.Cmd != "" {
			body["cmd"] = string(opts.Cmd)
		}
		if opts.CmdOpts != "" {
			body["cmd-opts"] = opts.CmdOpts
		}
		if opts.Width > 0 {
			body["width"] = opts.Width
		}
		if opts.Height > 0 {
			body["height"] = opts.Height
		}
		if opts.WebSocket {
			body["websocket"] = 1
		}
	}
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/vncshell", n.Name), body, &vnc)
	return
}

// NodeSpiceShellOptions configures the SPICE console launcher. Proxy
// overrides the SPICE proxy hostname (defaults to the node itself).
type NodeSpiceShellOptions struct {
	Cmd     NodeConsoleCmd
	CmdOpts string
	Proxy   string
}

// SpiceShell opens a SPICE proxy. Returned object can be fed directly to
// remote-viewer.
func (n *Node) SpiceShell(ctx context.Context, opts *NodeSpiceShellOptions) (proxy *SpiceProxy, err error) {
	body := map[string]interface{}{}
	if opts != nil {
		if opts.Cmd != "" {
			body["cmd"] = string(opts.Cmd)
		}
		if opts.CmdOpts != "" {
			body["cmd-opts"] = opts.CmdOpts
		}
		if opts.Proxy != "" {
			body["proxy"] = opts.Proxy
		}
	}
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/spiceshell", n.Name), body, &proxy)
	return
}

// --- /nodes/{node}/tasks (list) -------------------------------------------

// NodeTasksOptions filters the finished-task index. All fields are optional.
type NodeTasksOptions struct {
	Errors       bool
	Limit        int
	Since        int64
	Until        int64
	Source       string // "archive" | "active" | "all"
	Start        int
	StatusFilter string // comma-separated task statuses
	TypeFilter   string // task type, e.g. "vzdump"
	UserFilter   string
	VMID         int
}

// Tasks lists finished tasks on this node. Pre-populates each *Task with
// client + node so callers can chain Wait/Log/Stop. Unlike Task.Wait which
// targets one UPID, this lists archived/active tasks for monitoring.
func (n *Node) Tasks(ctx context.Context, opts *NodeTasksOptions) (tasks []*Task, err error) {
	path := fmt.Sprintf("/nodes/%s/tasks", n.Name)
	if opts != nil {
		q := url.Values{}
		if opts.Errors {
			q.Set("errors", "1")
		}
		if opts.Limit > 0 {
			q.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Since > 0 {
			q.Set("since", strconv.FormatInt(opts.Since, 10))
		}
		if opts.Until > 0 {
			q.Set("until", strconv.FormatInt(opts.Until, 10))
		}
		if opts.Source != "" {
			q.Set("source", opts.Source)
		}
		if opts.Start > 0 {
			q.Set("start", strconv.Itoa(opts.Start))
		}
		if opts.StatusFilter != "" {
			q.Set("statusfilter", opts.StatusFilter)
		}
		if opts.TypeFilter != "" {
			q.Set("typefilter", opts.TypeFilter)
		}
		if opts.UserFilter != "" {
			q.Set("userfilter", opts.UserFilter)
		}
		if opts.VMID > 0 {
			q.Set("vmid", strconv.Itoa(opts.VMID))
		}
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}
	if err = n.client.Get(ctx, path, &tasks); err != nil {
		return
	}
	for _, t := range tasks {
		t.client = n.client
	}
	return
}

// --- /nodes/{node}/network (revert pending) -------------------------------

// RevertNetworkChanges discards any pending /etc/network/interfaces.new
// staged by network create/update calls but not yet reloaded.
func (n *Node) RevertNetworkChanges(ctx context.Context) error {
	return n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/network", n.Name), nil)
}
