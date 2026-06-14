package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) Nodes(ctx context.Context) (ns NodeStatuses, err error) {
	return ns, c.Get(ctx, "/nodes", &ns)
}

func (c *Client) Node(ctx context.Context, name string) (*Node, error) {
	node := &Node{
		Name:   name,
		client: c,
	}

	// requires (/, Sys.Audit), do not error out if no access to still get the node
	if err := node.Status(ctx); !IsNotAuthorized(err) {
		return node, err
	}

	return node, nil
}

func (n *Node) New(c *Client, name string) *Node {
	node := &Node{
		Name:   name,
		client: c,
	}

	return node
}

func (n *Node) Status(ctx context.Context) error {
	return n.client.Get(ctx, fmt.Sprintf("/nodes/%s/status", n.Name), n)
}

func (n *Node) Version(ctx context.Context) (version *Version, err error) {
	return version, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/version", n.Name), &version)
}

func (n *Node) Report(ctx context.Context) (report string, err error) {
	return report, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/report", n.Name), &report)
}

func (n *Node) TermProxy(ctx context.Context) (term *Term, err error) {
	return term, n.client.Post(ctx, fmt.Sprintf("/nodes/%s/termproxy", n.Name), nil, &term)
}

func (n *Node) TermWebSocket(term *Term) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/vncwebsocket?port=%d&vncticket=%s",
		n.Name, term.Port, url.QueryEscape(term.Ticket))

	return n.client.TermWebSocket(p, term)
}

// VNCWebSocket send, recv, errors, closer, error
func (n *Node) VNCWebSocket(vnc *VNC) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/vncwebsocket?port=%d&vncticket=%s",
		n.Name, vnc.Port, url.QueryEscape(vnc.Ticket))

	return n.client.VNCWebSocket(p, vnc)
}

func (n *Node) VirtualMachines(ctx context.Context) (vms VirtualMachines, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu", n.Name), &vms); err != nil {
		return nil, err
	}

	for _, v := range vms {
		v.client = n.client
		v.Node = n.Name
	}

	return vms, nil
}

func (n *Node) NewVirtualMachine(ctx context.Context, vmid int, options ...VirtualMachineOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	data["vmid"] = vmid

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu", n.Name), data, &upid)
	return NewTask(upid, n.client), err
}

func (n *Node) VirtualMachine(ctx context.Context, vmid int) (*VirtualMachine, error) {
	vm := &VirtualMachine{
		client: n.client,
		Node:   n.Name,
	}

	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/current", n.Name, vmid), &vm); nil != err {
		return nil, err
	}

	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", n.Name, vmid), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	return vm, nil
}

func (n *Node) Containers(ctx context.Context) (c Containers, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc", n.Name), &c); err != nil {
		return
	}

	for _, container := range c {
		container.client = n.client
		container.Node = n.Name
	}

	return
}

func (n *Node) Container(ctx context.Context, vmid int) (*Container, error) {
	c := &Container{
		client: n.client,
		Node:   n.Name,
	}

	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/current", n.Name, vmid), &c); err != nil {
		return nil, err
	}

	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/config", n.Name, vmid), &c.ContainerConfig); err != nil {
		return nil, err
	}

	return c, nil
}

func (n *Node) NewContainer(ctx context.Context, vmid int, options ...ContainerOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	if vmid <= 0 {
		cluster, err := n.client.Cluster(ctx)
		if err != nil {
			return nil, err
		}
		vmid, err = cluster.NextID(ctx)
		if err != nil {
			return nil, err
		}
	}
	data["vmid"] = vmid

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc", n.Name), data, &upid)
	return NewTask(upid, n.client), err
}

func (n *Node) Appliances(ctx context.Context) (appliances Appliances, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/aplinfo", n.Name), &appliances)
	if err != nil {
		return appliances, err
	}

	for _, t := range appliances {
		t.client = n.client
		t.Node = n.Name
	}

	return appliances, nil
}

func (n *Node) DownloadAppliance(ctx context.Context, template, storage string) (ret string, err error) {
	return ret, n.client.Post(ctx, fmt.Sprintf("/nodes/%s/aplinfo", n.Name), map[string]string{
		"template": template,
		"storage":  storage,
	}, &ret)
}

func (n *Node) VzTmpls(ctx context.Context, storage string) (templates VzTmpls, err error) {
	return templates, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content?content=vztmpl", n.Name, storage), &templates)
}

func (n *Node) VzTmpl(ctx context.Context, template, storage string) (*VzTmpl, error) {
	templates, err := n.VzTmpls(ctx, storage)
	if err != nil {
		return nil, err
	}

	volid := fmt.Sprintf("%s:vztmpl/%s", storage, template)
	for _, t := range templates {
		if t.VolID == volid {
			return t, nil
		}
	}

	return nil, fmt.Errorf("could not find vztmpl: %s", template)
}

func (n *Node) Storages(ctx context.Context) (storages Storages, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage", n.Name), &storages)
	if err != nil {
		return
	}

	for _, s := range storages {
		s.Node = n.Name
		s.client = n.client
	}

	return
}

func (n *Node) Storage(ctx context.Context, name string) (storage *Storage, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/status", n.Name, name), &storage)
	if err != nil {
		return
	}

	storage.Node = n.Name
	storage.client = n.client
	storage.Name = name

	return
}

func (n *Node) StorageDownloadURL(ctx context.Context, options *StorageDownloadURLOptions) (ret string, err error) {
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/storage/%s/download-url", n.Name, options.Storage), options, &ret)
	return ret, err
}

func (n *Node) StorageISO(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "iso")
}

func (n *Node) StorageVZTmpl(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "vztmpl")
}

func (n *Node) StorageBackup(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "backup")
}

// StorageSnippets returns a storage configured for the "snippets" content
// type. Note that Proxmox does not expose a REST upload endpoint for
// snippets — they must be written to the storage path directly (e.g. via
// SCP/SFTP). This helper is for read-side discovery (e.g. resolving the
// storage's path so a caller can write to it out-of-band).
func (n *Node) StorageSnippets(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "snippets")
}

func (n *Node) StorageRootDir(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "rootdir")
}

func (n *Node) StorageImages(ctx context.Context) (*Storage, error) {
	return n.findStorageByContent(ctx, "images")
}

// findStorageByContent takes iso/backup/vztmpl/rootdir/images and returns the storage that type of content should be on
func (n *Node) findStorageByContent(ctx context.Context, content string) (storage *Storage, err error) {
	storages, err := n.Storages(ctx)
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if storage.Enabled == 0 {
			continue
		}

		if strings.Contains(storage.Content, content) {
			storage.Node = n.Name
			storage.client = n.client
			return storage, nil
		}
	}

	return nil, ErrNotFound
}

func (n *Node) FirewallOptionGet(ctx context.Context) (firewallOption *FirewallNodeOption, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/firewall/options", n.Name), &firewallOption)
	return
}

func (n *Node) FirewallOptionSet(ctx context.Context, firewallOption *FirewallNodeOption) error {
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/firewall/options", n.Name), firewallOption, nil)
}

// FirewallRules lists firewall rules for the node. Returned rules carry the
// parent context required to call (*FirewallRule).Get/Update/Delete.
func (n *Node) FirewallRules(ctx context.Context) (rules []*FirewallRule, err error) {
	if err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), &rules); err != nil {
		return nil, err
	}
	for _, r := range rules {
		r.client = n.client
		r.kind = fwRuleKindNode
		r.node = n.Name
	}
	return rules, nil
}

// FirewallRule returns a *FirewallRule wired to the node's firewall scope at
// the given position. The returned instance is a lazy handle — call Get(ctx)
// to populate it from /firewall/rules/{pos}.
func (n *Node) FirewallRule(pos int) *FirewallRule {
	return &FirewallRule{
		client: n.client,
		kind:   fwRuleKindNode,
		node:   n.Name,
		Pos:    pos,
	}
}

// NewFirewallRule creates a firewall rule on the node. After a successful
// POST the rule is wired with parent context so subsequent
// Update/Delete/Get calls route correctly. Note: PVE's POST does not return
// the assigned position; callers that need it should re-list via FirewallRules.
func (n *Node) NewFirewallRule(ctx context.Context, rule *FirewallRule) error {
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), rule, nil); err != nil {
		return err
	}
	rule.client = n.client
	rule.kind = fwRuleKindNode
	rule.node = n.Name
	return nil
}

// FirewallGetRule fetches one rule by position. Companion to
// FirewallGetRules, mirroring the per-rule getter on Container/VirtualMachine.
func (n *Node) FirewallGetRule(ctx context.Context, rulePos int) (rule *FirewallRule, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/firewall/rules/%d", n.Name, rulePos), &rule)
	return
}

// NodeFirewallLogOptions filters the host-firewall log read. All optional.
type NodeFirewallLogOptions struct {
	Start int
	Limit int
	Since int64 // unix epoch
	Until int64 // unix epoch
}

// FirewallLog returns the host firewall's iptables/nftables log entries.
// Each LogEntry is {n: line-number, t: text}.
func (n *Node) FirewallLog(ctx context.Context, opts *NodeFirewallLogOptions) (entries []*LogEntry, err error) {
	path := fmt.Sprintf("/nodes/%s/firewall/log", n.Name)
	if opts != nil {
		q := url.Values{}
		if opts.Start > 0 {
			q.Set("start", fmt.Sprintf("%d", opts.Start))
		}
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Since > 0 {
			q.Set("since", fmt.Sprintf("%d", opts.Since))
		}
		if opts.Until > 0 {
			q.Set("until", fmt.Sprintf("%d", opts.Until))
		}
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}
	err = n.client.Get(ctx, path, &entries)
	return
}

func (n *Node) UploadCustomCertificate(ctx context.Context, cert *CustomCertificate) error {
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/certificates/custom", n.Name), cert, nil)
}

func (n *Node) DeleteCustomCertificate(ctx context.Context) error {
	return n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/certificates/custom", n.Name), nil)
}

func (n *Node) GetCustomCertificates(ctx context.Context) (certs *NodeCertificates, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/certificates/info", n.Name), &certs)
	return
}

// ListCertificateSubresources enumerates the children of
// /nodes/{node}/certificates (typically "info", "custom", "acme"). The PVE
// schema only documents this as a directory index, so we collapse the
// {"name": ...} link objects into a flat []string.
func (n *Node) ListCertificateSubresources(ctx context.Context) ([]string, error) {
	var items []struct {
		Name string `json:"name"`
	}
	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/certificates", n.Name), &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Name)
	}
	return out, nil
}

// ListACMECertificateSubresources enumerates the children of
// /nodes/{node}/certificates/acme (typically just "certificate").
func (n *Node) ListACMECertificateSubresources(ctx context.Context) ([]string, error) {
	var items []struct {
		Name string `json:"name"`
	}
	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/certificates/acme", n.Name), &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Name)
	}
	return out, nil
}

// OrderACMECertificate orders a new ACME certificate using the node's
// configured ACME account/plugin. force=true overwrites an existing custom
// or ACME certificate. POST /nodes/{node}/certificates/acme/certificate.
func (n *Node) OrderACMECertificate(ctx context.Context, force bool) (*Task, error) {
	body := map[string]any{}
	if force {
		body["force"] = 1
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/certificates/acme/certificate", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// RenewACMECertificate renews the node's ACME certificate. PVE skips renewal
// when the cert is more than 30 days from expiry — force=true overrides that
// check. PUT /nodes/{node}/certificates/acme/certificate.
func (n *Node) RenewACMECertificate(ctx context.Context, force bool) (*Task, error) {
	body := map[string]any{}
	if force {
		body["force"] = 1
	}
	var upid UPID
	if err := n.client.Put(ctx, fmt.Sprintf("/nodes/%s/certificates/acme/certificate", n.Name), body, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// RevokeACMECertificate revokes the node's ACME certificate with the issuing
// CA. DELETE /nodes/{node}/certificates/acme/certificate.
func (n *Node) RevokeACMECertificate(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/certificates/acme/certificate", n.Name), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

func (n *Node) Vzdump(ctx context.Context, params *VirtualMachineBackupOptions) (task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &VirtualMachineBackupOptions{}
	}

	if err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/vzdump", n.Name), params, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

func (n *Node) VzdumpExtractConfig(ctx context.Context, volume string) (*VzdumpConfig, error) {
	var vzdumpExtractedConfig string

	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/vzdump/extractconfig?volume=%s", n.Name, volume), &vzdumpExtractedConfig); err != nil {
		return nil, err
	}

	return n.parseVzdumpConfig(vzdumpExtractedConfig)
}

func (n *Node) parseVzdumpConfig(vzdumpExtractedConfig string) (*VzdumpConfig, error) {
	vzdumpFields := strings.Split(vzdumpExtractedConfig, StringSeparator)

	configFields := make(map[string]any)

	for _, field := range vzdumpFields {
		if field != "" {
			newStr := strings.SplitN(field, FieldSeparator, 2)
			if len(newStr) == 2 {
				configFields[newStr[0]] = strings.Trim(newStr[1], SpaceSeparator)
			}
		}
	}

	jsonData, err := json.Marshal(configFields)
	if err != nil {
		return nil, fmt.Errorf("cannot present vzdump config as json string : %w", err)
	}

	vzdumpCfg := &VzdumpConfig{}
	if err := json.Unmarshal(jsonData, vzdumpCfg); err != nil {
		return nil, fmt.Errorf("cannot parse data for vzdump config : %w", err)
	}

	return vzdumpCfg, nil
}

// ---- /nodes/{node}/services --------------------------------------------------

// Services returns the list of services on the node (pveproxy, pvedaemon,
// corosync, ssh, etc.). GET /nodes/{node}/services. Each returned
// *NodeService is pre-populated with client and Node so callers can chain
// instance methods (Start, Stop, Restart, Reload, State) directly.
func (n *Node) Services(ctx context.Context) (services []*NodeService, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/services", n.Name), &services)
	if err != nil {
		return
	}
	for _, s := range services {
		s.client = n.client
		s.Node = n.Name
		// PVE returns "service" as the canonical id and usually mirrors it in
		// "name"; make sure Name is populated either way so instance methods
		// always have a path segment.
		if s.Name == "" {
			s.Name = s.Service
		}
	}
	return
}

// Service returns a handle for a single service without making an API call.
// Use State to populate the handle from /nodes/{node}/services/{name}/state,
// or just call Start/Stop/Restart/Reload directly.
func (n *Node) Service(name string) *NodeService {
	return &NodeService{
		client: n.client,
		Node:   n.Name,
		Name:   name,
	}
}

// State refreshes the service handle from
// GET /nodes/{node}/services/{name}/state. The /services/{name} root is just
// a directory index and is intentionally not wrapped — state is the only
// useful read on a specific service.
func (s *NodeService) State(ctx context.Context) error {
	// Preserve the caller's identifying fields — the API response will
	// repopulate Service/Name but we want our cached Node and client to
	// survive the round-trip.
	client, node, name := s.client, s.Node, s.Name
	if err := client.Get(ctx, fmt.Sprintf("/nodes/%s/services/%s/state", node, name), s); err != nil {
		return err
	}
	s.client = client
	s.Node = node
	if s.Name == "" {
		s.Name = name
	}
	return nil
}

// Start issues POST /nodes/{node}/services/{name}/start. Returns a Task
// because PVE does service control asynchronously.
func (s *NodeService) Start(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/services/%s/start", s.Node, s.Name), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Stop issues POST /nodes/{node}/services/{name}/stop.
func (s *NodeService) Stop(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/services/%s/stop", s.Node, s.Name), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Restart issues POST /nodes/{node}/services/{name}/restart — a hard
// restart. Use Reload for graceful restart of services that support it.
func (s *NodeService) Restart(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/services/%s/restart", s.Node, s.Name), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Reload issues POST /nodes/{node}/services/{name}/reload, which PVE
// documents as "falls back to restart if reload isn't supported".
func (s *NodeService) Reload(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := s.client.Post(ctx, fmt.Sprintf("/nodes/%s/services/%s/reload", s.Node, s.Name), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, s.client), nil
}

// Time returns the node's current time and timezone configuration.
// GET /nodes/{node}/time. The "time" and "localtime" fields are unix epoch
// seconds — see NodeTime.
func (n *Node) Time(ctx context.Context) (t *NodeTime, err error) {
	t = &NodeTime{}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/time", n.Name), t)
	return
}

// SetTimezone sets the node's timezone. Valid names come from
// /usr/share/zoneinfo/zone.tab. PUT /nodes/{node}/time.
func (n *Node) SetTimezone(ctx context.Context, timezone string) error {
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/time", n.Name), map[string]string{"timezone": timezone}, nil)
}

// ---- /nodes/{node}/subscription ----------------------------------------------

// Subscription reads the node's subscription state — license level, status,
// next-due-date, etc. GET /nodes/{node}/subscription.
func (n *Node) Subscription(ctx context.Context) (sub *Subscription, err error) {
	sub = &Subscription{}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/subscription", n.Name), sub)
	return
}

// SetSubscription registers a Proxmox VE subscription key on the node.
// PUT /nodes/{node}/subscription.
func (n *Node) SetSubscription(ctx context.Context, key string) error {
	return n.client.Put(ctx, fmt.Sprintf("/nodes/%s/subscription", n.Name), map[string]string{"key": key}, nil)
}

// RefreshSubscription asks the node to re-validate the cached subscription
// status against Proxmox's servers. force=true bypasses the local cache and
// always hits the upstream. POST /nodes/{node}/subscription.
func (n *Node) RefreshSubscription(ctx context.Context, force bool) error {
	body := map[string]any{}
	if force {
		body["force"] = 1
	}
	return n.client.Post(ctx, fmt.Sprintf("/nodes/%s/subscription", n.Name), body, nil)
}

// DeleteSubscription removes the subscription key from the node, returning it
// to community-edition status. DELETE /nodes/{node}/subscription.
func (n *Node) DeleteSubscription(ctx context.Context) error {
	return n.client.Delete(ctx, fmt.Sprintf("/nodes/%s/subscription", n.Name), nil)
}

// ---- node-wide mass operations -----------------------------------------------

// StartAll starts every VM and container on the node honoring the configured
// startup order. Pass NodeStartAllOptions{Force: IntOrBool(true)} to bypass
// the order, or VMs to limit the set of guests started.
func (n *Node) StartAll(ctx context.Context, opts *NodeStartAllOptions) (*Task, error) {
	if opts == nil {
		opts = &NodeStartAllOptions{}
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/startall", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// StopAll stops every VM and container on the node. ForceStop defaults to 1
// (true) server-side; pass a populated NodeStopAllOptions to override timeout
// or restrict which guests are stopped.
func (n *Node) StopAll(ctx context.Context, opts *NodeStopAllOptions) (*Task, error) {
	if opts == nil {
		opts = &NodeStopAllOptions{}
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/stopall", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// SuspendAll suspends every VM on the node. (LXC containers are not suspended
// by this endpoint — PVE only honors VMs here.)
func (n *Node) SuspendAll(ctx context.Context, opts *NodeSuspendAllOptions) (*Task, error) {
	if opts == nil {
		opts = &NodeSuspendAllOptions{}
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/suspendall", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// MigrateAll migrates every VM and container on the node to opts.Target.
// Target is required.
func (n *Node) MigrateAll(ctx context.Context, opts *NodeMigrateAllOptions) (*Task, error) {
	if opts == nil || opts.Target == "" {
		return nil, errors.New("migrateall target can not be empty")
	}
	var upid UPID
	if err := n.client.Post(ctx, fmt.Sprintf("/nodes/%s/migrateall", n.Name), opts, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, n.client), nil
}

// WakeOnLAN sends a Wake-on-LAN magic packet to the node and returns the MAC
// address that was woken. PVE looks up the WoL MAC from the cluster config —
// the node's wakeonlan setting (see datacenter.cfg) must be configured.
func (n *Node) WakeOnLAN(ctx context.Context) (mac string, err error) {
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/wakeonlan", n.Name), nil, &mac)
	return
}

// ---- /nodes/{node}/replication ----------------------------------------------

// Replications returns replication-job status from the node's perspective. The
// cluster-wide job *configuration* lives at /cluster/replication; this endpoint
// reports per-node runtime state (last sync, last duration, fail count, etc.)
// for each job that targets or originates on this node. guest filters to a
// specific VMID; pass 0 for all.
func (n *Node) Replications(ctx context.Context, guest int) (jobs []*NodeReplicationJob, err error) {
	path := fmt.Sprintf("/nodes/%s/replication", n.Name)
	if guest != 0 {
		path += fmt.Sprintf("?guest=%d", guest)
	}
	if err = n.client.Get(ctx, path, &jobs); err != nil {
		return nil, err
	}
	for _, j := range jobs {
		j.client = n.client
		j.Node = n.Name
	}
	return jobs, nil
}

// Replication returns a handle to a single replication job on this node. It
// does not perform an API call; use the returned job's methods to query or
// act on the underlying /nodes/{node}/replication/{id}/* endpoints.
func (n *Node) Replication(id string) *NodeReplicationJob {
	return &NodeReplicationJob{
		client: n.client,
		Node:   n.Name,
		ID:     id,
	}
}

// Status refreshes runtime state for this replication job in-place.
// GET /nodes/{node}/replication/{id}/status. The /replication/{id} root is
// just a tree index and is intentionally not wrapped.
func (r *NodeReplicationJob) Status(ctx context.Context) error {
	return r.client.Get(ctx, fmt.Sprintf("/nodes/%s/replication/%s/status", r.Node, r.ID), r)
}

// Log returns the job's log lines. start/limit are optional pagination — pass
// 0 for default. PVE returns a list of {n, t} entries where n is line number
// and t is the line text.
func (r *NodeReplicationJob) Log(ctx context.Context, start, limit int) (entries []*LogEntry, err error) {
	path := fmt.Sprintf("/nodes/%s/replication/%s/log", r.Node, r.ID)
	q := ""
	if start > 0 {
		q = fmt.Sprintf("start=%d", start)
	}
	if limit > 0 {
		if q != "" {
			q += "&"
		}
		q += fmt.Sprintf("limit=%d", limit)
	}
	if q != "" {
		path += "?" + q
	}
	err = r.client.Get(ctx, path, &entries)
	return
}

// ScheduleNow asks PVE to run this replication job as soon as possible
// (bypassing its schedule). POST /nodes/{node}/replication/{id}/schedule_now —
// returns a Task UPID.
func (r *NodeReplicationJob) ScheduleNow(ctx context.Context) (*Task, error) {
	var upid UPID
	if err := r.client.Post(ctx, fmt.Sprintf("/nodes/%s/replication/%s/schedule_now", r.Node, r.ID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, r.client), nil
}
