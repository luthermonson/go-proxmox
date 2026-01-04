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
	return v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/status/current", v.Node, v.VMID), &v)
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

// CloudInit takes four yaml docs as a string and make an ISO, upload it to the data store as <vmid>-user-data.iso and will
// mount it as a CD-ROM to be used with nocloud cloud-init. This is NOT how proxmox expects a user to do cloud-init
// which can be found here: https://pve.proxmox.com/wiki/Cloud-Init_Support#:~:text=and%20meta.-,Cloud%2DInit%20specific%20Options,-cicustom%3A%20%5Bmeta
// If you want to use the proxmox implementation you'll need to use the cloudinit APIs https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/qemu/{vmid}/cloudinit
func (v *VirtualMachine) CloudInit(ctx context.Context, device, userdata, metadata, vendordata, networkconfig string) error {
	isoName := fmt.Sprintf(UserDataISOFormat, v.VMID)
	// create userdata iso file on the local fs
	isofilename, err := makeCloudInitISO(isoName, userdata, metadata, vendordata, networkconfig)
	if err != nil {
		return err
	}

	defer func() {
		// _ = os.Remove(iso.Name())
	}()

	node, err := v.client.Node(ctx, v.Node)
	if err != nil {
		return err
	}

	storage, err := node.StorageISO(ctx)
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
		err = iso.Close()
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

	for filename, content := range cifiles {
		rw, err := fs.OpenFile("/"+filename, os.O_CREATE|os.O_RDWR)
		if err != nil {
			return "", err
		}

		if _, err = rw.Write([]byte(content)); err != nil {
			return "", err
		}
	}

	if err = fs.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
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

func (v *VirtualMachine) deleteCloudInitISO(ctx context.Context) (ok bool, err error) {
	if v.HasTag(MakeTag(TagCloudInit)) {
		node, err := v.client.Node(ctx, v.Node)
		if err != nil {
			return false, err
		}
		isoStorage, err := node.StorageISO(ctx)
		if err != nil {
			return false, err
		}

		var iso *ISO
		iso, err = isoStorage.ISO(ctx, fmt.Sprintf(UserDataISOFormat, v.VMID))
		if err != nil {
			// skipping, iso not found return no error.
			return true, nil
		}
		task, err := iso.Delete(ctx)
		if err != nil {
			return false, err
		}
		if err := task.WaitFor(ctx, 5); err != nil {
			return false, err
		}
	}

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

func (v *VirtualMachine) AgentGetNetworkIFaces(ctx context.Context) (iFaces []*AgentNetworkIface, err error) {
	networks := map[string][]*AgentNetworkIface{}
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", v.Node, v.VMID), &networks)
	if err != nil {
		return
	}
	if result, ok := networks["result"]; ok {
		for _, iface := range result {
			if iface.Name == "lo" {
				continue
			}
			iFaces = append(iFaces, iface)
		}
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
	results := map[string]*AgentOsInfo{}
	err = v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-osinfo", v.Node, v.VMID), &results)
	if err != nil {
		return
	}

	var ok bool
	info, ok = results["result"]
	if !ok {
		err = fmt.Errorf("result is empty")
	}

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
