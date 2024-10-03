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

func (c *Container) Delete(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d", c.Node, c.VMID), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, c.client), nil
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
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/resize", c.Node, c.VMID), map[string]interface{}{"disk": disk, "size": size}, &upid); err != nil {
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

func (c *Container) RRDData(ctx context.Context, timeframe Timeframe, consolidationFunction ConsolidationFunction) (rrddata []*RRDData, err error) {
	u := url.URL{Path: fmt.Sprintf("/lxc/%s/qemu/%d/rrddata", c.Node, c.VMID)}

	// consolidation functions are variadic because they're optional, putting everything into one string and sending that
	params := url.Values{}
	if len(consolidationFunction) > 0 {

		f := ""
		for _, cf := range consolidationFunction {
			f = f + string(cf)
		}
		params.Add("cf", f)
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

func (c *Container) Snapshots(ctx context.Context) (snapshots []*ContainerSnapshot, err error) {
	err = c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot", c.Node, c.VMID), &snapshots)
	return
}

func (c *Container) NewSnapshot(ctx context.Context, snapName string) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot", c.Node, c.VMID), map[string]interface{}{"snapname": snapName}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) GetSnapshot(ctx context.Context, snapshot string) (snap []*ContainerSnapshot, err error) {
	return snap, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s", c.Node, c.VMID, snapshot), &snap)
}

func (c *Container) DeleteSnapshot(ctx context.Context, snapshot string) (task *Task, err error) {
	var upid UPID
	if err := c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s", c.Node, c.VMID, snapshot), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) RollbackSnapshot(ctx context.Context, snapshot string, start bool) (task *Task, err error) {
	var upid UPID
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/rollback", c.Node, c.VMID, snapshot), map[string]interface{}{"start": start}, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, c.client), nil
}

func (c *Container) GetSnapshotConfig(ctx context.Context, snapshot string) (config map[string]interface{}, err error) {
	return config, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/config", c.Node, c.VMID, snapshot), &config)
}

func (c *Container) UpdateSnapshot(ctx context.Context, snapshot string) error {
	return c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/snapshot/%s/config", c.Node, c.VMID, snapshot), nil, nil)
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

func (c *Container) NewFirewallIPSet(ctx context.Context, ipset *FirewallIPSet) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset", c.Node, c.VMID), ipset, nil)
}

func (c *Container) DeleteFirewallIPSet(ctx context.Context, name string, force bool) error {
	return c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/ipset/%s", c.Node, c.VMID, name), map[string]interface{}{"force": force})
}

func (c *Container) FirewallRules(ctx context.Context) (rules []*FirewallRule, err error) {
	return rules, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules", c.Node, c.VMID), &rules)
}

func (c *Container) NewFirewallRule(ctx context.Context, rule *FirewallRule) error {
	return c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules", c.Node, c.VMID), rule, nil)
}

func (c *Container) GetFirewallRule(ctx context.Context, rulePos int) (rule *FirewallRule, err error) {
	return rule, c.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules/%d", c.Node, c.VMID, rulePos), &rule)
}

func (c *Container) UpdateFirewallRule(ctx context.Context, rulePos int, rule *FirewallRule) error {
	return c.client.Put(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules/%d", c.Node, c.VMID, rulePos), rule, nil)
}

func (c *Container) DeleteFirewallRule(ctx context.Context, rulePos int) error {
	return c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules/%d", c.Node, c.VMID, rulePos), nil)
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
