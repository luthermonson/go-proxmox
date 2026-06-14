package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (c *Container) Clone(ctx context.Context, params *ContainerCloneOptions) (newid int, task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &ContainerCloneOptions{}
	}
	if params.NewID <= 0 {
		cluster, err := c.client.Cluster(ctx)
		if err != nil {
			return newid, nil, err
		}
		newid, err = cluster.NextID(ctx)
		if err != nil {
			return newid, nil, err
		}
		params.NewID = newid
	} else {
		newid = params.NewID
	}
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/clone", c.Node, c.VMID), params, &upid); err != nil {
		return 0, nil, err
	}
	return newid, NewTask(upid, c.client), nil
}

// Delete removes the container. Pass a non-nil *ContainerDeleteOptions to
// force-delete a running container (Force), purge it from related
// configurations (Purge), or destroy unreferenced disks
// (DestroyUnreferencedDisks). nil applies the API defaults.
func (c *Container) Delete(ctx context.Context, params *ContainerDeleteOptions) (task *Task, err error) {
	var upid UPID
	if err := c.client.DeleteWithParams(ctx, fmt.Sprintf("/nodes/%s/lxc/%d", c.Node, c.VMID), params, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, c.client), nil
}

func (c *Container) Ping(ctx context.Context) error {
	return c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/current", c.Node, c.VMID), &c)
}

// Config sets ContainerOptions for Container
// see https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/lxc/{vmid}/config for available attributes
func (c *Container) Config(ctx context.Context, options ...ContainerOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	for _, option := range options {
		data[option.Name] = option.Value
	}
	err := c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/config", c.Node, c.VMID), data, &upid)
	return NewTask(upid, c.client), err
}

func (c *Container) Start(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.Node, c.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Stop(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.Node, c.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Suspend(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/suspend", c.Node, c.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Reboot(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.Node, c.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Resume(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/resume", c.Node, c.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Shutdown(ctx context.Context, force bool, timeout int) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/shutdown", c.Node, c.VMID), map[string]interface{}{"forceStop": force, "timeout": timeout}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) TermProxy(ctx context.Context) (term *Term, err error) {
	return term, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/termproxy", c.Node, c.VMID), nil, &term)
}

func (c *Container) TermWebSocket(term *Term) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/lxc/%d/vncwebsocket?port=%d&vncticket=%s",
		c.Node, c.VMID, term.Port, url.QueryEscape(term.Ticket))

	return c.client.TermWebSocket(p, term)
}

func (c *Container) VNCWebSocket(vnc *VNC) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/lxc/%d/vncwebsocket?port=%d&vncticket=%s",
		c.Node, c.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return c.client.VNCWebSocket(p, vnc)
}

func (c *Container) Feature(ctx context.Context) (hasFeature bool, err error) {
	var feature struct {
		HasFeature bool `json:"hasFeature"`
	}
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/feature", c.Node, c.VMID), &feature)
	return feature.HasFeature, err
}

func (c *Container) Interfaces(ctx context.Context) (interfaces ContainerInterfaces, err error) {
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/interfaces", c.Node, c.VMID), &interfaces)
	return interfaces, err
}

func (c *Container) Migrate(ctx context.Context, params *ContainerMigrateOptions) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/migrate", c.Node, c.VMID), params, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) Resize(ctx context.Context, disk, size string) (task *Task, err error) {
	var upid UPID
	if err := c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/resize", c.Node, c.VMID), map[string]interface{}{"disk": disk, "size": size}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) MoveVolume(ctx context.Context, params *VirtualMachineMoveDiskOptions) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/move_volume", c.Node, c.VMID), params, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) RRDData(ctx context.Context, timeframe Timeframe, consolidationFunction ...ConsolidationFunction) (rrddata []*RRDData, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/lxc/%d/rrddata", c.Node, c.VMID)}

	// consolidation functions are variadic because they're optional, but Proxmox only allows one cf parameter
	params := url.Values{}
	if len(consolidationFunction) > 0 {
		if len(consolidationFunction) != 1 {
			return nil, fmt.Errorf("only one consolidation function allowed")
		}

		params.Add("cf", string(consolidationFunction[0]))
	}

	params.Add("timeframe", string(timeframe))
	u.RawQuery = params.Encode()

	err = c.client.Get(ctx, u.String(), &rrddata)
	return
}

func (c *Container) Template(ctx context.Context) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/template", c.Node, c.VMID), nil, nil)
}

func (c *Container) VNCProxy(ctx context.Context, vncOptions VNCProxyOptions) (vnc *VNC, err error) {
	return vnc, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/vncproxy", c.Node, c.VMID), vncOptions, &vnc)
}

// Snapshots lists every snapshot of the container. Each returned
// *ContainerSnapshot has its client/Node/VMID pre-populated so callers can
// invoke instance methods (Rollback, Delete, Config, …) directly.
func (c *Container) Snapshots(ctx context.Context) (snapshots []*ContainerSnapshot, err error) {
	if err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot", c.Node, c.VMID), &snapshots); err != nil {
		return nil, err
	}
	for _, s := range snapshots {
		s.client = c.client
		s.Node = c.Node
		s.VMID = int(c.VMID)
	}
	return snapshots, nil
}

// NewSnapshot creates a snapshot of the container and returns the worker
// task. The returned snapshot's metadata is reachable via c.Snapshot(name)
// once the task completes.
func (c *Container) NewSnapshot(ctx context.Context, snapName string) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot", c.Node, c.VMID), map[string]interface{}{"snapname": snapName}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

// Snapshot constructs a handle to a single snapshot by name. It performs no
// API call; the returned *ContainerSnapshot is wired to the container's
// client and node/vmid so its instance methods (Rollback, Delete, Config,
// UpdateConfig, SubResources) target the correct snapshot path.
func (c *Container) Snapshot(name string) *ContainerSnapshot {
	return &ContainerSnapshot{
		client: c.client,
		Node:   c.Node,
		VMID:   int(c.VMID),
		Name:   name,
	}
}

// Rollback rolls the container back to this snapshot. When start is true the
// container is started after the rollback completes.
func (s *ContainerSnapshot) Rollback(ctx context.Context, start bool) (task *Task, err error) {
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/rollback", s.Node, s.VMID, s.Name), map[string]interface{}{"start": start}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Delete removes this snapshot from the container. Returns the worker task.
func (s *ContainerSnapshot) Delete(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := s.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s", s.Node, s.VMID, s.Name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Config reads this snapshot's metadata (description, parent, etc.). PVE
// returns a free-form map since snapshot configs include arbitrary device
// keys snapshotted at the time of creation.
func (s *ContainerSnapshot) Config(ctx context.Context) (config map[string]interface{}, err error) {
	return config, s.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/config", s.Node, s.VMID, s.Name), &config)
}

// UpdateConfig updates this snapshot's metadata. PVE only allows changing
// the description on this endpoint; pass nil options to clear it.
func (s *ContainerSnapshot) UpdateConfig(ctx context.Context, options *ContainerSnapshotUpdateOptions) error {
	if options == nil {
		options = &ContainerSnapshotUpdateOptions{}
	}
	return s.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/config", s.Node, s.VMID, s.Name), options, nil)
}

// SubResources returns the per-snapshot directory index
// (GET /nodes/{node}/lxc/{vmid}/snapshot/{snapname}) — one entry per
// sub-resource (config, rollback) on this snapshot.
func (s *ContainerSnapshot) SubResources(ctx context.Context) (entries []*ContainerSnapshotIndexEntry, err error) {
	err = s.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s", s.Node, s.VMID, s.Name), &entries)
	return
}

func (c *Container) Firewall(ctx context.Context) (firewall *Firewall, err error) {
	return firewall, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall", c.Node, c.VMID), &firewall)
}

func (c *Container) GetFirewallAliases(ctx context.Context) (aliases []*FirewallAlias, err error) {
	return aliases, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/aliases", c.Node, c.VMID), &aliases)
}

func (c *Container) NewFirewallAlias(ctx context.Context, alias *FirewallAlias) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/aliases", c.Node, c.VMID), alias, nil)
}

func (c *Container) GetFirewallAlias(ctx context.Context, name string) (alias *FirewallAlias, err error) {
	return alias, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/aliases/%s", c.Node, c.VMID, name), &alias)
}

func (c *Container) UpdateFirewallAlias(ctx context.Context, name string, alias *FirewallAlias) error {
	return c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/aliases/%s", c.Node, c.VMID, name), alias, nil)
}

func (c *Container) DeleteFirewallAlias(ctx context.Context, name string) error {
	return c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/aliases/%s", c.Node, c.VMID, name), nil)
}

func (c *Container) GetFirewallIPSet(ctx context.Context) (ipsets []*FirewallIPSet, err error) {
	return ipsets, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset", c.Node, c.VMID), &ipsets)
}

func (c *Container) NewFirewallIPSet(ctx context.Context, ipset FirewallIPSetCreationOption) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset", c.Node, c.VMID), ipset, nil)
}

func (c *Container) DeleteFirewallIPSet(ctx context.Context, name string, force bool) error {
	return c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s", c.Node, c.VMID, name), map[string]interface{}{"force": force})
}

func (c *Container) GetFirewallIPSetEntries(ctx context.Context, name string) (entries []*FirewallIPSetEntry, err error) {
	return entries, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s", c.Node, c.VMID, name), &entries)
}

func (c *Container) NewFirewallIPSetEntry(ctx context.Context, name string, entry FirewallIPSetEntryCreationOption) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s", c.Node, c.VMID, name), entry, nil)
}

func (c *Container) DeleteFirewallIPSetEntry(ctx context.Context, name string, cidr string, digest string) error {
	return c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s/%s", c.Node, c.VMID, name, cidr), map[string]interface{}{
		"digest": digest,
	})
}

func (c *Container) GetFirewallIPSetEntry(ctx context.Context, name string, cidr string) (entry *FirewallIPSetEntry, err error) {
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s/%s", c.Node, c.VMID, name, cidr), &entry)
	return
}

func (c *Container) UpdateFirewallIPSetEntry(ctx context.Context, name string, cidr string, entry *FirewallIPSetEntryUpdateOption) error {
	return c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s/%s", c.Node, c.VMID, name, cidr), entry, nil)
}

// FirewallRules lists firewall rules for the container. Returned rules carry
// the parent context required to call (*FirewallRule).Get/Update/Delete.
func (c *Container) FirewallRules(ctx context.Context) (rules []*FirewallRule, err error) {
	if err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules", c.Node, c.VMID), &rules); err != nil {
		return nil, err
	}
	for _, r := range rules {
		r.client = c.client
		r.kind = fwRuleKindLXC
		r.node = c.Node
		r.vmid = uint64(c.VMID)
	}
	return rules, nil
}

// FirewallRule returns a *FirewallRule wired to the container's firewall
// scope at the given position. The returned instance is a lazy handle — call
// Get(ctx) to populate it from /firewall/rules/{pos}.
func (c *Container) FirewallRule(pos int) *FirewallRule {
	return &FirewallRule{
		client: c.client,
		kind:   fwRuleKindLXC,
		node:   c.Node,
		vmid:   uint64(c.VMID),
		Pos:    pos,
	}
}

// NewFirewallRule creates a firewall rule on the container. After a
// successful POST the rule is wired with parent context so subsequent
// Update/Delete/Get calls route correctly. Note: PVE's POST does not return
// the assigned position; callers that need it should re-list via FirewallRules.
func (c *Container) NewFirewallRule(ctx context.Context, rule *FirewallRule) error {
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules", c.Node, c.VMID), rule, nil); err != nil {
		return err
	}
	rule.client = c.client
	rule.kind = fwRuleKindLXC
	rule.node = c.Node
	rule.vmid = uint64(c.VMID)
	return nil
}

func (c *Container) GetFirewallOptions(ctx context.Context) (options *FirewallVirtualMachineOption, err error) {
	return options, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/options", c.Node, c.VMID), &options)
}

func (c *Container) UpdateFirewallOptions(ctx context.Context, options *FirewallVirtualMachineOption) error {
	return c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/options", c.Node, c.VMID), options, nil)
}

// Tag Management Helpers

// HasTag returns if a Tag is present in TagSlice
func (c *Container) HasTag(value string) bool {
	if c.ContainerConfig == nil {
		return false
	}

	if c.ContainerConfig.Tags == "" {
		return false
	}

	if c.ContainerConfig.TagsSlice == nil {
		c.SplitTags()
	}

	for _, tag := range c.ContainerConfig.TagsSlice {
		if tag == value {
			return true
		}
	}

	return false
}

// AddTag appends the passed value to TagsSlice and updates Tags via c.Config
// If accurate state is important, then reassign the value of Container after the task
// has completed.
func (c *Container) AddTag(ctx context.Context, value string) (*Task, error) {
	if c.HasTag(value) {
		return nil, ErrNoop
	}

	if c.ContainerConfig.TagsSlice == nil {
		c.SplitTags()
	}

	c.ContainerConfig.TagsSlice = append(c.ContainerConfig.TagsSlice, value)
	c.ContainerConfig.Tags = strings.Join(c.ContainerConfig.TagsSlice, TagSeperator)
	c.Tags = c.ContainerConfig.Tags // Keep the parent object up to date

	return c.Config(ctx, ContainerOption{
		Name:  "tags",
		Value: c.ContainerConfig.Tags,
	})
}

// RemoveTag removes the passed value from TagsSlice and updates Tags via c.Config
// If accurate state is important, then reassign the value of Container after the task
// has completed.
func (c *Container) RemoveTag(ctx context.Context, value string) (*Task, error) {
	if !c.HasTag(value) {
		return nil, ErrNoop
	}

	if c.ContainerConfig.TagsSlice == nil {
		c.SplitTags()
	}

	for i, tag := range c.ContainerConfig.TagsSlice {
		if tag == value {
			c.ContainerConfig.TagsSlice = append(
				c.ContainerConfig.TagsSlice[:i],
				c.ContainerConfig.TagsSlice[i+1:]...,
			)
		}
	}

	c.ContainerConfig.Tags = strings.Join(c.ContainerConfig.TagsSlice, TagSeperator)
	c.Tags = c.ContainerConfig.Tags // keep the parent object up to date
	return c.Config(ctx, ContainerOption{
		Name:  "tags",
		Value: c.ContainerConfig.Tags,
	})
}

// SplitTags sets ContainerConfig TagsSlice my splitting the value of ContainerConfig.Tags with TagSeparator
func (c *Container) SplitTags() {
	c.ContainerConfig.TagsSlice = strings.Split(c.ContainerConfig.Tags, TagSeperator)
}

// Pending returns the container's staged config changes — keys whose desired
// value differs from the running value. Empty list when the container's live
// config matches its on-disk config.
func (c *Container) Pending(ctx context.Context) (pending []*ContainerPending, err error) {
	return pending, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/pending", c.Node, c.VMID), &pending)
}

// RRD asks PVE to render a single timeframe PNG on the server and returns its
// on-disk filename (the file lives in PVE's rrdcached dir). Most callers want
// RRDData instead for numeric series; this exists for API parity.
func (c *Container) RRD(ctx context.Context, ds string, timeframe Timeframe, consolidationFunction ...ConsolidationFunction) (rrd *ContainerRRD, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/lxc/%d/rrd", c.Node, c.VMID)}
	params := url.Values{}
	if len(consolidationFunction) > 0 {
		if len(consolidationFunction) != 1 {
			return nil, fmt.Errorf("only one consolidation function allowed")
		}
		params.Add("cf", string(consolidationFunction[0]))
	}
	params.Add("ds", ds)
	params.Add("timeframe", string(timeframe))
	u.RawQuery = params.Encode()
	err = c.client.Get(ctx, u.String(), &rrd)
	return
}

// RemoteMigrate triggers a cross-cluster migration of this container. The
// target endpoint is an API-token bundle string per PVE's pvesh docs.
func (c *Container) RemoteMigrate(ctx context.Context, params *ContainerRemoteMigrateOptions) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/remote_migrate", c.Node, c.VMID), params, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

// SpiceProxy returns SPICE proxy connection info for the container. PVE
// supports SPICE primarily for QEMU VMs; LXC support depends on the
// container's console configuration and may error on most setups.
func (c *Container) SpiceProxy(ctx context.Context) (spice *SpiceProxy, err error) {
	return spice, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/spiceproxy", c.Node, c.VMID), nil, &spice)
}

// FirewallLog returns the per-container firewall log entries.
func (c *Container) FirewallLog(ctx context.Context) (entries []*FirewallLogEntry, err error) {
	return entries, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/log", c.Node, c.VMID), &entries)
}

// FirewallRefs returns the IPSets and aliases referenceable from this
// container's firewall rules (cluster-level + node-level + container-level).
func (c *Container) FirewallRefs(ctx context.Context) (refs []*FirewallRef, err error) {
	return refs, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/refs", c.Node, c.VMID), &refs)
}

// DirIndex returns the per-container directory index
// (GET /nodes/{node}/lxc/{vmid}) — one entry per sub-resource (config, status,
// snapshot, firewall, …). The sub-resources themselves are wrapped as their
// own methods on *Container; this is mostly useful for discovery.
func (c *Container) DirIndex(ctx context.Context) (entries []*ContainerDirIndexEntry, err error) {
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d", c.Node, c.VMID), &entries)
	return
}

// StatusIndex returns the container status directory index
// (GET /nodes/{node}/lxc/{vmid}/status) — one entry per status sub-command
// (current, start, stop, …). The operations are wrapped as Start/Stop/etc.
// on *Container.
func (c *Container) StatusIndex(ctx context.Context) (entries []*ContainerStatusIndexEntry, err error) {
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status", c.Node, c.VMID), &entries)
	return
}

// MigratePreconditions returns the precondition check for migrating this
// container (GET /nodes/{node}/lxc/{vmid}/migrate). If target is non-empty,
// PVE evaluates the check against that specific target node; otherwise it
// reports which nodes are allowed/not-allowed and any local-disk constraints.
func (c *Container) MigratePreconditions(ctx context.Context, target string) (preconditions *ContainerMigratePreconditions, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/lxc/%d/migrate", c.Node, c.VMID)}
	if target != "" {
		params := url.Values{}
		params.Add("target", target)
		u.RawQuery = params.Encode()
	}
	err = c.client.Get(ctx, u.String(), &preconditions)
	return
}

// MigrationTunnel opens a migration tunnel for this container
// (POST /nodes/{node}/lxc/{vmid}/mtunnel) and returns the Unix socket path,
// authentication ticket, and worker UPID.
//
// PVE marks this endpoint as "for internal use by VM migration" — callers
// should generally use Migrate or the higher-level migration flow rather
// than wiring this up directly. Wrapped here for full API surface coverage.
func (c *Container) MigrationTunnel(ctx context.Context, options *ContainerMigrationTunnelOptions) (tunnel *ContainerMigrationTunnel, err error) {
	if options == nil {
		options = &ContainerMigrationTunnelOptions{}
	}
	err = c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/mtunnel", c.Node, c.VMID), options, &tunnel)
	return
}

// MigrationTunnelWebSocketPath returns the path callers can pass to a
// websocket dialer to upgrade the migration tunnel returned by
// MigrationTunnel (GET /nodes/{node}/lxc/{vmid}/mtunnelwebsocket).
//
// PVE marks this endpoint as "for internal use by VM migration"; this helper
// just builds the path with the correct query string. There is no generic
// Client helper for migration-tunnel websockets — dial it directly with the
// authenticated cookies/headers if you need to consume it.
func (c *Container) MigrationTunnelWebSocketPath(tunnel *ContainerMigrationTunnel) string {
	q := url.Values{}
	if tunnel != nil {
		q.Set("socket", tunnel.Socket)
		q.Set("ticket", tunnel.Ticket)
	}
	return fmt.Sprintf("/nodes/%s/lxc/%d/mtunnelwebsocket?%s", c.Node, c.VMID, q.Encode())
}
