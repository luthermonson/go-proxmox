package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/diskfs/go-diskfs/backend/file"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

const (
	StatusVirtualMachineRunning = "running"
	StatusVirtualMachineStopped = "stopped"
	StatusVirtualMachinePaused  = "paused"

	UserDataISOFormat = "user-data-%d.iso"
	TagCloudInit      = "cloud-init"
	TagSeperator      = ";"

	volumeIdentifier = "cidata"
	blockSize        = 2048
)

// DefaultAgentWaitInterval is the polling interval when waiting for agent exec commands
var DefaultAgentWaitInterval = 100 * time.Millisecond

func (v *VirtualMachine) New(c *Client, nodeName string, vmid int) {
	v.client = c
	v.Node = nodeName
	v.VMID = StringOrUint64(vmid)
}

func (v *VirtualMachine) Ping(ctx context.Context) error {
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/current", v.Node, v.VMID), &v); err != nil {
		return err
	}
	// New, empty config in case something was deleted
	v.VirtualMachineConfig = &VirtualMachineConfig{}
	return v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", v.Node, v.VMID), &v.VirtualMachineConfig)
}

func (v *VirtualMachine) Config(ctx context.Context, options ...VirtualMachineOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	for _, opt := range options {
		data[opt.Name] = opt.Value
	}
	err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/config", v.Node, v.VMID), data, &upid)
	return NewTask(upid, v.client), err
}

func (v *VirtualMachine) Monitor(ctx context.Context, command string) (s string, err error) {
	data := make(map[string]interface{})
	data["command"] = command
	err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/monitor", v.Node, v.VMID), data, &s)
	return s, err
}

func (v *VirtualMachine) TermProxy(ctx context.Context) (term *Term, err error) {
	return term, v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/termproxy", v.Node, v.VMID), nil, &term)
}

func (v *VirtualMachine) VNCProxy(ctx context.Context, config *VNCConfig) (vnc *VNC, err error) {
	return vnc, v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy?", v.Node, v.VMID), config, &vnc)
}

func (v *VirtualMachine) HasTag(value string) bool {
	if v.VirtualMachineConfig == nil {
		return false
	}

	if v.VirtualMachineConfig.Tags == "" {
		return false
	}

	if v.VirtualMachineConfig.TagsSlice == nil {
		v.SplitTags()
	}

	for _, tag := range v.VirtualMachineConfig.TagsSlice {
		if tag == value {
			return true
		}
	}

	return false
}

func (v *VirtualMachine) AddTag(ctx context.Context, value string) (*Task, error) {
	if v.HasTag(value) {
		return nil, ErrNoop
	}

	if v.VirtualMachineConfig.TagsSlice == nil {
		v.SplitTags()
	}

	v.VirtualMachineConfig.TagsSlice = append(v.VirtualMachineConfig.TagsSlice, value)
	v.VirtualMachineConfig.Tags = strings.Join(v.VirtualMachineConfig.TagsSlice, TagSeperator)

	return v.Config(ctx, VirtualMachineOption{
		Name:  "tags",
		Value: v.VirtualMachineConfig.Tags,
	})
}

func (v *VirtualMachine) RemoveTag(ctx context.Context, value string) (*Task, error) {
	if !v.HasTag(value) {
		return nil, ErrNoop
	}

	if v.VirtualMachineConfig.TagsSlice == nil {
		v.SplitTags()
	}

	for i, tag := range v.VirtualMachineConfig.TagsSlice {
		if tag == value {
			v.VirtualMachineConfig.TagsSlice = append(
				v.VirtualMachineConfig.TagsSlice[:i],
				v.VirtualMachineConfig.TagsSlice[i+1:]...,
			)
		}
	}

	v.VirtualMachineConfig.Tags = strings.Join(v.VirtualMachineConfig.TagsSlice, TagSeperator)
	return v.Config(ctx, VirtualMachineOption{
		Name:  "tags",
		Value: v.VirtualMachineConfig.Tags,
	})
}

func (v *VirtualMachine) SplitTags() {
	v.VirtualMachineConfig.TagsSlice = strings.Split(v.VirtualMachineConfig.Tags, TagSeperator)
}

// CloudInitOption configures optional behavior on VirtualMachine.CloudInit.
// Construct via the With*-prefixed CloudInit option helpers.
type CloudInitOption func(*cloudInitConfig)

type cloudInitConfig struct {
	storage string
}

// WithCloudInitStorage selects a specific Proxmox storage (by name) to upload
// the cloud-init ISO into. The storage must be enabled and accept "iso"
// content. Without this option, CloudInit auto-selects the first enabled
// iso-capable storage on the node, which is non-deterministic across nodes
// with multiple iso-capable storages — see issue #119.
func WithCloudInitStorage(name string) CloudInitOption {
	return func(c *cloudInitConfig) { c.storage = name }
}

// CloudInit takes four yaml docs as a string and make an ISO, upload it to the data store as <vmid>-user-data.iso and will
// mount it as a CD-ROM to be used with nocloud cloud-init. This is NOT how proxmox expects a user to do cloud-init
// which can be found here: https://pve.proxmox.com/wiki/Cloud-Init_Support#:~:text=and%20meta.-,Cloud%2DInit%20specific%20Options,-cicustom%3A%20%5Bmeta
// If you want to use the proxmox implementation you'll need to use the cloudinit APIs https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/qemu/{vmid}/cloudinit
//
// Pass WithCloudInitStorage("name") to control which storage receives the ISO.
// Without it, the first enabled iso-capable storage returned by the node is used.
func (v *VirtualMachine) CloudInit(ctx context.Context, device, userdata, metadata, vendordata, networkconfig string, opts ...CloudInitOption) error {
	var cfg cloudInitConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	isoName := fmt.Sprintf(UserDataISOFormat, v.VMID)
	// create userdata iso file on the local fs
	isofilename, err := makeCloudInitISO(isoName, userdata, metadata, vendordata, networkconfig)
	if err != nil {
		return err
	}

	defer func() {
		if rerr := os.Remove(isofilename); rerr != nil {
			v.client.log.Warnf("failed to remove temp cloud-init iso %s: %v", isofilename, rerr)
		}
	}()

	node, err := v.client.Node(ctx, v.Node)
	if err != nil {
		return err
	}

	storage, err := resolveCloudInitStorage(ctx, node, &cfg)
	if err != nil {
		return err
	}

	task, err := storage.Upload("iso", isofilename)
	if err != nil {
		return err
	}

	// iso should only be < 5mb so wait for it and then mount it
	if err := task.WaitFor(ctx, 5); err != nil {
		return err
	}

	_, err = v.AddTag(ctx, MakeTag(TagCloudInit))
	if err != nil && !IsErrNoop(err) {
		return err
	}

	task, err = v.Config(ctx, VirtualMachineOption{
		Name:  device,
		Value: fmt.Sprintf("%s:iso/%s,media=cdrom", storage.Name, isoName),
	}, VirtualMachineOption{
		Name:  "boot",
		Value: fmt.Sprintf("%s;%s", v.VirtualMachineConfig.Boot, device),
	})
	if err != nil {
		return err
	}

	return task.WaitFor(ctx, 2)
}

// resolveCloudInitStorage picks the *Storage that CloudInit should upload to.
// If cfg.storage is set, the named storage is fetched and validated to accept
// iso content. Otherwise the node's auto-select for iso content is used.
func resolveCloudInitStorage(ctx context.Context, node *Node, cfg *cloudInitConfig) (*Storage, error) {
	if cfg.storage == "" {
		return node.StorageISO(ctx)
	}
	s, err := node.Storage(ctx, cfg.storage)
	if err != nil {
		return nil, fmt.Errorf("cloud-init storage %q: %w", cfg.storage, err)
	}
	if !strings.Contains(s.Content, "iso") {
		return nil, fmt.Errorf("cloud-init storage %q does not accept iso content (got %q)", cfg.storage, s.Content)
	}
	return s, nil
}

func makeCloudInitISO(filename, userdata, metadata, vendordata, networkconfig string) (isopath string, err error) {
	isopath = filepath.Join(os.TempDir(), filename)

	isoFile, err := os.Create(isopath)
	if err != nil {
		return "", err
	}

	if err := isoFile.Close(); err != nil {
		return "", err
	}

	iso, err := file.OpenFromPath(isopath, false)
	if err != nil {
		return "", err
	}

	defer func() {
		if cerr := iso.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	fs, err := iso9660.Create(iso, 0, 0, blockSize, "")
	if err != nil {
		return "", err
	}

	if err = fs.Mkdir("/"); err != nil {
		return "", err
	}

	cifiles := map[string]string{
		"user-data": userdata,
		"meta-data": metadata,
	}
	if vendordata != "" {
		cifiles["vendor-data"] = vendordata
	}
	if networkconfig != "" {
		cifiles["network-config"] = networkconfig
	}

	for name, content := range cifiles {
		rw, err := fs.OpenFile("/"+name, os.O_CREATE|os.O_RDWR)
		if err != nil {
			return "", err
		}

		if _, err = rw.Write([]byte(content)); err != nil {
			return "", err
		}

		// Joliet finalization in go-diskfs v1.9+ requires the file handle to
		// be closed before Finalize so its size is recorded correctly.
		if err = rw.Close(); err != nil {
			return "", err
		}
	}

	if err = fs.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		Joliet:           true,
		VolumeIdentifier: volumeIdentifier,
	}); err != nil {
		return "", err
	}

	return
}

func (v *VirtualMachine) TermWebSocket(term *Term) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, term.Port, url.QueryEscape(term.Ticket))

	return v.client.TermWebSocket(p, term)
}

// VNCWebSocket copy/paste when calling to get the channel names right
// send, recv, errors, closer, errors := vm.VNCWebSocket(vnc)
// for this to work you need to first set up a serial terminal on your vm https://pve.proxmox.com/wiki/Serial_Terminal
func (v *VirtualMachine) VNCWebSocket(vnc *VNC) (chan []byte, chan []byte, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return v.client.VNCWebSocket(p, vnc)
}

// SpiceProxy returns SPICE proxy connection info for the VM. Mirrors the
// Container.SpiceProxy surface and serializes the .vv file fields remote-viewer
// expects.
func (v *VirtualMachine) SpiceProxy(ctx context.Context) (spice *SpiceProxy, err error) {
	return spice, v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/spiceproxy", v.Node, v.VMID), nil, &spice)
}

func (v *VirtualMachine) IsRunning() bool {
	return v.Status == StatusVirtualMachineRunning && (v.QMPStatus == "" || v.QMPStatus == StatusVirtualMachineRunning)
}

func (v *VirtualMachine) Start(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/start", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsStopped() bool {
	return v.Status == StatusVirtualMachineStopped && (v.Lock != "suspended")
}

func (v *VirtualMachine) Reset(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/reset", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Shutdown(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/shutdown", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Stop(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsPaused() bool {
	return v.Status == StatusVirtualMachineRunning && v.QMPStatus == StatusVirtualMachinePaused
}

func (v *VirtualMachine) Pause(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsHibernated() bool {
	return v.Status == StatusVirtualMachineStopped && v.Lock == "suspended"
}

func (v *VirtualMachine) Hibernate(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), map[string]string{"todisk": "1"}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Resume(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/resume", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Reboot(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Delete(ctx context.Context) (task *Task, err error) {
	if ok, err := v.deleteCloudInitISO(ctx); err != nil || !ok {
		return nil, err
	}

	var upid UPID
	if err = v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d", v.Node, v.VMID), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

// deleteCloudInitISO scans every enabled iso-capable storage on the VM's node
// for the cloud-init user-data ISO and removes it from the first one that has
// it. Iterating across storages (rather than relying on the auto-selected one,
// as the original code did) keeps deletion correct when CloudInit was called
// with WithCloudInitStorage targeting a non-default storage.
func (v *VirtualMachine) deleteCloudInitISO(ctx context.Context) (ok bool, err error) {
	if !v.HasTag(MakeTag(TagCloudInit)) {
		return true, nil
	}

	node, err := v.client.Node(ctx, v.Node)
	if err != nil {
		return false, err
	}

	storages, err := node.Storages(ctx)
	if err != nil {
		return false, err
	}

	isoFilename := fmt.Sprintf(UserDataISOFormat, v.VMID)
	for _, s := range storages {
		if s.Enabled == 0 || !strings.Contains(s.Content, "iso") {
			continue
		}
		iso, ierr := s.ISO(ctx, isoFilename)
		if ierr != nil {
			// not on this storage; try the next
			continue
		}
		task, terr := iso.Delete(ctx)
		if terr != nil {
			return false, terr
		}
		if werr := task.WaitFor(ctx, 5); werr != nil {
			return false, werr
		}
		return true, nil
	}

	// Not found anywhere — already gone, treat as no-op (matches prior behavior).
	return true, nil
}

func (v *VirtualMachine) Migrate(
	ctx context.Context,
	params *VirtualMachineMigrateOptions,
) (task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &VirtualMachineMigrateOptions{}
	}

	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/migrate", v.Node, v.VMID), params, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Clone(ctx context.Context, params *VirtualMachineCloneOptions) (newid int, task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &VirtualMachineCloneOptions{}
	}

	if params.NewID == 0 {
		cluster, err := v.client.Cluster(ctx)
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

	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/clone", v.Node, v.VMID), params, &upid); err != nil {
		return newid, nil, err
	}

	return newid, NewTask(upid, v.client), nil
}

func (v *VirtualMachine) ResizeDisk(ctx context.Context, disk, size string) (*Task, error) {
	var upid UPID

	if err := v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/resize", v.Node, v.VMID), map[string]string{
		"disk": disk,
		"size": size,
	}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) UnlinkDisk(ctx context.Context, diskID string, force bool) (task *Task, err error) {
	var upid UPID

	params := map[string]string{"idlist": diskID}
	if force {
		params["force"] = "1"
	}
	err = v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/unlink", v.Node, v.VMID), params, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) MoveDisk(ctx context.Context, disk string, params *VirtualMachineMoveDiskOptions) (task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &VirtualMachineMoveDiskOptions{}
	}

	if disk != "" {
		params.Disk = disk
	}

	err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/move_disk", v.Node, v.VMID), params, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) AgentGetHostName(ctx context.Context) (hostname string, err error) {
	node, err := v.client.Node(ctx, v.Node)
	if err != nil {
		return
	}

	var resp struct {
		Result *AgentHostName `json:"result"`
	}
	if err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-host-name", node.Name, v.VMID), &resp); err != nil {
		return
	}
	if resp.Result == nil {
		err = fmt.Errorf("result is empty")
		return
	}
	hostname = resp.Result.HostName
	return
}

func (v *VirtualMachine) AgentGetNetworkIFaces(ctx context.Context) (iFaces []*AgentNetworkIface, err error) {
	var resp struct {
		Result []*AgentNetworkIface `json:"result"`
	}
	if err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", v.Node, v.VMID), &resp); err != nil {
		return
	}
	for _, iface := range resp.Result {
		if iface.Name == "lo" {
			continue
		}
		iFaces = append(iFaces, iface)
	}
	return
}

func (v *VirtualMachine) WaitForAgent(ctx context.Context, seconds int) error {
	timeout := time.After(time.Duration(seconds) * time.Second)
	ticker := time.NewTicker(DefaultWaitInterval)
	defer ticker.Stop()

	for {
		_, err := v.AgentOsInfo(ctx)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "500 QEMU guest agent is not running") {
			return err
		}

		select {
		case <-timeout:
			return ErrTimeout
		case <-ticker.C:
		}
	}
}

func (v *VirtualMachine) AgentExec(ctx context.Context, command []string, inputData string) (pid int, err error) {
	tmpdata := map[string]interface{}{}
	err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", v.Node, v.VMID),
		map[string]interface{}{
			"command":    command,
			"input-data": inputData,
		},
		&tmpdata)

	p := tmpdata["pid"]
	if p == nil {
		return 0, fmt.Errorf("no pid returned from agent exec command")
	}
	pid = int(p.(float64))
	return
}

func (v *VirtualMachine) AgentExecStatus(ctx context.Context, pid int) (status *AgentExecStatus, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec-status?pid=%d", v.Node, v.VMID, pid), &status)
	if err != nil {
		return nil, err
	}

	return
}

func (v *VirtualMachine) WaitForAgentExecExit(ctx context.Context, pid, seconds int) (*AgentExecStatus, error) {
	timeout := time.After(time.Duration(seconds) * time.Second)
	ticker := time.NewTicker(DefaultAgentWaitInterval)
	defer ticker.Stop()

	for {
		status, err := v.AgentExecStatus(ctx, pid)
		if err != nil {
			return nil, err
		}
		if status.Exited != 0 {
			return status, nil
		}

		select {
		case <-timeout:
			return nil, ErrTimeout
		case <-ticker.C:
		}
	}
}

func (v *VirtualMachine) AgentOsInfo(ctx context.Context) (info *AgentOsInfo, err error) {
	var resp struct {
		Result *AgentOsInfo `json:"result"`
	}
	if err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-osinfo", v.Node, v.VMID), &resp); err != nil {
		return
	}
	if resp.Result == nil {
		err = fmt.Errorf("result is empty")
		return
	}
	info = resp.Result
	return
}

func (v *VirtualMachine) AgentSetUserPassword(ctx context.Context, password string, username string) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", v.Node, v.VMID), map[string]string{"password": password, "username": username}, nil)
}

func (v *VirtualMachine) SendKey(ctx context.Context, key string) error {
	data := map[string]interface{}{"key": key}
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/sendkey", v.Node, v.VMID), data, nil)
}

func (v *VirtualMachine) GetFirewallIPSet(ctx context.Context) (ipsets []*FirewallIPSet, err error) {
	return ipsets, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", v.Node, v.VMID), &ipsets)
}

func (v *VirtualMachine) NewFirewallIPSet(ctx context.Context, ipset FirewallIPSetCreationOption) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", v.Node, v.VMID), ipset, nil)
}

func (v *VirtualMachine) DeleteFirewallIPSet(ctx context.Context, name string, force bool) error {
	return v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", v.Node, v.VMID, name), map[string]interface{}{"force": force})
}

func (v *VirtualMachine) GetFirewallIPSetEntries(ctx context.Context, name string) (entries []*FirewallIPSetEntry, err error) {
	return entries, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", v.Node, v.VMID, name), &entries)
}

func (v *VirtualMachine) NewFirewallIPSetEntry(ctx context.Context, name string, entry FirewallIPSetEntryCreationOption) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", v.Node, v.VMID, name), entry, nil)
}

func (v *VirtualMachine) DeleteFirewallIPSetEntry(ctx context.Context, name string, cidr string, digest string) error {
	return v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s/%s", v.Node, v.VMID, name, cidr), map[string]interface{}{
		"digest": digest,
	})
}

func (v *VirtualMachine) GetFirewallIPSetEntry(ctx context.Context, name string, cidr string) (entry *FirewallIPSetEntry, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s/%s", v.Node, v.VMID, name, cidr), &entry)
	return
}

func (v *VirtualMachine) UpdateFirewallIPSetEntry(ctx context.Context, name string, cidr string, entry *FirewallIPSetEntryUpdateOption) error {
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s/%s", v.Node, v.VMID, name, cidr), entry, nil)
}

func (v *VirtualMachine) FirewallOptionGet(ctx context.Context) (firewallOption *FirewallVirtualMachineOption, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption)
	return
}

func (v *VirtualMachine) FirewallOptionSet(ctx context.Context, firewallOption *FirewallVirtualMachineOption) error {
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption, nil)
}

func (v *VirtualMachine) FirewallGetRules(ctx context.Context) (rules []*FirewallRule, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), &rules)
	return
}

func (v *VirtualMachine) FirewallRulesCreate(ctx context.Context, rule *FirewallRule) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), rule, nil)
}

func (v *VirtualMachine) FirewallRulesUpdate(ctx context.Context, rule *FirewallRule) error {
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rule.Pos), rule, nil)
}

func (v *VirtualMachine) FirewallRulesDelete(ctx context.Context, rulePos int) error {
	return v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rulePos), nil)
}

func (v *VirtualMachine) NewSnapshot(ctx context.Context, name string) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), map[string]string{"snapname": name}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Snapshots(ctx context.Context) (snapshots []*Snapshot, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), &snapshots)
	return
}

func (v *VirtualMachine) SnapshotRollback(ctx context.Context, name string) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s/rollback", v.Node, v.VMID, name), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) DeleteSnapshot(ctx context.Context, snapshot string) (task *Task, err error) {
	var upid UPID
	if err := v.client.Delete(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s", v.Node, v.VMID, snapshot), &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, v.client), nil
}

// GetSnapshotConfig reads a snapshot's metadata (description, parent, etc.).
// PVE returns a free-form map since snapshot configs include arbitrary
// device keys snapshotted at the time of creation.
func (v *VirtualMachine) GetSnapshotConfig(ctx context.Context, snapshot string) (config map[string]interface{}, err error) {
	return config, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s/config", v.Node, v.VMID, snapshot), &config)
}

// VirtualMachineSnapshotUpdateOptions is the body for PUT
// /qemu/{vmid}/snapshot/{name}/config. PVE only lets you change the
// description through this endpoint.
type VirtualMachineSnapshotUpdateOptions struct {
	Description string `json:"description,omitempty"`
}

// UpdateSnapshot updates a VM snapshot's metadata. PVE only allows changing
// the description on this endpoint; pass nil options to clear it.
func (v *VirtualMachine) UpdateSnapshot(ctx context.Context, snapshot string, options *VirtualMachineSnapshotUpdateOptions) error {
	if options == nil {
		options = &VirtualMachineSnapshotUpdateOptions{}
	}
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s/config", v.Node, v.VMID, snapshot), options, nil)
}

// RRDData takes a timeframe enum and an optional consolidation function
// usage: vm.RRDData(HOURLY) or vm.RRDData(HOURLY, AVERAGE)
func (v *VirtualMachine) RRDData(ctx context.Context, timeframe Timeframe, consolidationFunction ...ConsolidationFunction) (rrddata []*RRDData, err error) {
	u := url.URL{Path: fmt.Sprintf("/nodes/%s/qemu/%d/rrddata", v.Node, v.VMID)}

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

	err = v.client.Get(ctx, u.String(), &rrddata)
	return
}

func (v *VirtualMachine) ConvertToTemplate(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/template", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}
	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) UnmountCloudInitISO(ctx context.Context, device string) error {
	if !v.HasTag(MakeTag(TagCloudInit)) {
		return nil
	}

	_, err := v.Config(ctx, VirtualMachineOption{
		Name:  device,
		Value: "none,media=cdrom",
	})
	if err != nil {
		return err
	}

	if _, err = v.deleteCloudInitISO(ctx); err != nil {
		return err
	}
	return nil
}

func (v *VirtualMachine) Pending(ctx context.Context) (pending *PendingConfiguration, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/pending", v.Node, v.VMID), &pending)
	return
}

// DirIndex returns the per-VM directory index
// (GET /nodes/{node}/qemu/{vmid}) — one entry per child resource (config,
// status, snapshot, firewall, agent, …). Mostly useful for discovery; the
// actual resources are wrapped as their own methods on *VirtualMachine.
func (v *VirtualMachine) DirIndex(ctx context.Context) (entries []*VirtualMachineDirIndexEntry, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d", v.Node, v.VMID), &entries)
	return
}

// StatusIndex returns the VM status directory index
// (GET /nodes/{node}/qemu/{vmid}/status) — one entry per status sub-command
// (current, start, stop, reboot, …). The actual operations are wrapped as
// Start/Stop/Reboot/etc. on *VirtualMachine.
func (v *VirtualMachine) StatusIndex(ctx context.Context) (entries []*VirtualMachineStatusIndexEntry, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status", v.Node, v.VMID), &entries)
	return
}

// SnapshotIndex returns the per-snapshot directory index
// (GET /nodes/{node}/qemu/{vmid}/snapshot/{snapname}) — one entry per
// sub-resource (config, rollback) on the named snapshot.
func (v *VirtualMachine) SnapshotIndex(ctx context.Context, snapshot string) (entries []*VirtualMachineSnapshotIndexEntry, err error) {
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s", v.Node, v.VMID, snapshot), &entries)
	return
}

// MigrationTunnel opens a migration tunnel for this VM
// (POST /nodes/{node}/qemu/{vmid}/mtunnel) and returns the Unix socket path,
// authentication ticket, and worker UPID.
//
// PVE marks this endpoint as "for internal use by VM migration" — callers
// should generally use Migrate or the higher-level migration flow rather
// than wiring this up directly. It is wrapped here only for full API
// surface coverage.
func (v *VirtualMachine) MigrationTunnel(ctx context.Context, options *VirtualMachineMigrationTunnelOptions) (tunnel *VirtualMachineMigrationTunnel, err error) {
	if options == nil {
		options = &VirtualMachineMigrationTunnelOptions{}
	}
	err = v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/mtunnel", v.Node, v.VMID), options, &tunnel)
	return
}

// MigrationTunnelWebSocketPath returns the path callers can pass to a
// websocket dialer (or to Client.VNCWebSocket-style helpers) to upgrade the
// migration tunnel returned by MigrationTunnel
// (GET /nodes/{node}/qemu/{vmid}/mtunnelwebsocket).
//
// PVE marks this endpoint as "for internal use by VM migration"; this
// helper just builds the path with the correct query string. There is no
// generic Client helper for migration-tunnel websockets — the protocol
// differs from the VNC/term tunnels — so dial it directly with the
// authenticated cookies/headers if you need to consume it.
func (v *VirtualMachine) MigrationTunnelWebSocketPath(tunnel *VirtualMachineMigrationTunnel) string {
	q := url.Values{}
	if tunnel != nil {
		q.Set("socket", tunnel.Socket)
		q.Set("ticket", tunnel.Ticket)
	}
	return fmt.Sprintf("/nodes/%s/qemu/%d/mtunnelwebsocket?%s", v.Node, v.VMID, q.Encode())
}
