package proxmox

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	StatusVirtualMachineRunning = "running"
	StatusVirtualMachineStopped = "stopped"
	StatusVirtualMachinePaused  = "paused"
)

func (v *VirtualMachine) Ping() error {
	return v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", v.Node, v.VMID), &v)
}

func (v *VirtualMachine) Config(options ...VirtualMachineOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	for _, opt := range options {
		data[opt.Name] = opt.Value
	}
	err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/config", v.Node, v.VMID), data, &upid)
	return NewTask(upid, v.client), err
}

func (v *VirtualMachine) TermProxy() (vnc *VNC, err error) {
	return vnc, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/termproxy", v.Node, v.VMID), nil, &vnc)
}

// VNCWebSocket copy/paste when calling to get the channel names right
// send, recv, errors, closer, errors := vm.VNCWebSocket(vnc)
// for this to work you need to first setup a serial terminal on your vm https://pve.proxmox.com/wiki/Serial_Terminal
func (v *VirtualMachine) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?port=%d&vncticket=%s",
		v.Node, v.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return v.client.VNCWebSocket(p, vnc)
}

func (v *VirtualMachine) IsRunning() bool {
	return v.Status == StatusVirtualMachineRunning && v.QMPStatus == StatusVirtualMachineRunning
}

func (v *VirtualMachine) Start() (*Task, error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/start", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsStopped() bool {
	return v.Status == StatusVirtualMachineStopped && v.QMPStatus == StatusVirtualMachineStopped
}

func (v *VirtualMachine) Reset() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/reset", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Shutdown() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/shutdown", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Stop() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsPaused() bool {
	return v.Status == StatusVirtualMachineRunning && v.QMPStatus == StatusVirtualMachinePaused
}

func (v *VirtualMachine) Pause() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) IsHibernated() bool {
	return v.Status == StatusVirtualMachineStopped && v.QMPStatus == StatusVirtualMachineStopped && v.Lock == "suspended"
}

func (v *VirtualMachine) Hibernate() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), map[string]string{"todisk": "1"}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Resume() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/resume", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Reboot() (task *Task, err error) {
	var upid UPID
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", v.Node, v.VMID), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Delete() (task *Task, err error) {
	var upid UPID
	if err := v.client.Delete(fmt.Sprintf("/nodes/%s/qemu/%d", v.Node, v.VMID), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Migrate(target, targetstorage string) (task *Task, err error) {
	var upid UPID
	params := map[string]string{
		"target": target,
	}
	if targetstorage != "" {
		params["targetstorage"] = targetstorage
	}
	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/migrate", v.Node, v.VMID), params, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) Clone(name, target string) (newid int, task *Task, err error) {
	var upid UPID
	cluster, err := v.client.Cluster()
	if err != nil {
		return newid, nil, err
	}

	newid, err = cluster.NextID()
	if err != nil {
		return newid, nil, err
	}

	if err := v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/clone", v.Node, v.VMID), map[string]string{
		"newid":  strconv.Itoa(newid),
		"name":   name,
		"target": target,
	}, &upid); err != nil {
		return newid, nil, err
	}

	return newid, NewTask(upid, v.client), nil
}

func (v *VirtualMachine) ResizeDisk(disk, size string) (err error) {
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/resize", v.Node, v.VMID), map[string]string{
		"disk": disk,
		"size": size,
	}, nil)
	if err != nil {
		return
	}

	return
}

func (v *VirtualMachine) UnlinkDisk(diskID string, force bool) (task *Task, err error) {
	var upid UPID

	params := map[string]string{"idlist": diskID}
	if force {
		params["force"] = "1"
	}
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/unlink", v.Node, v.VMID), params, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) MoveDisk(disk, storage string) (task *Task, err error) {
	var upid UPID

	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/move_disk", v.Node, v.VMID), map[string]string{
		"disk":    disk,
		"storage": storage,
	}, &upid)
	if err != nil {
		return
	}

	return NewTask(upid, v.client), nil
}

func (v *VirtualMachine) AgentGetNetworkIFaces() (iFaces []*AgentNetworkIface, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	networks := map[string][]*AgentNetworkIface{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/network-get-interfaces", node.Name, v.VMID), &networks)
	if err != nil {
		return
	}
	if result, ok := networks["result"]; ok {
		for _, iface := range result {
			if "lo" == iface.Name {
				continue
			}
			iFaces = append(iFaces, iface)
		}
	}

	return

}

func (v *VirtualMachine) AgentOsInfo() (info *AgentOsInfo, err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}
	results := map[string]*AgentOsInfo{}
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-osinfo", node.Name, v.VMID), &results)

	if err != nil {
		return
	}
	info, ok := results["result"]
	if !ok {
		err = fmt.Errorf("result is empty")
	}
	return

}
func (v *VirtualMachine) AgentSetUserPassword(password string, username string) (err error) {
	node, err := v.client.Node(v.Node)
	if err != nil {
		return
	}

	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", node.Name, v.VMID), map[string]string{"password": password, "username": username}, nil)

	return

}

func (v *VirtualMachine) FirewallOptionGet() (firewallOption *FirewallVirtualMachineOption, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption)
	return
}
func (v *VirtualMachine) FirewallOptionSet(firewallOption *FirewallVirtualMachineOption) (err error) {
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", v.Node, v.VMID), firewallOption, nil)
	return
}

func (v *VirtualMachine) FirewallGetRules() (rules []*FirewallRule, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), &rules)
	return
}

func (v *VirtualMachine) FirewallRulesCreate(rule *FirewallRule) (err error) {
	err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", v.Node, v.VMID), rule, nil)
	return
}
func (v *VirtualMachine) FirewallRulesUpdate(rule *FirewallRule) (err error) {
	err = v.client.Put(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rule.Pos), rule, nil)
	return
}
func (v *VirtualMachine) FirewallRulesDelete(rulePos int) (err error) {
	err = v.client.Delete(fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", v.Node, v.VMID, rulePos), nil)
	return
}

func (v *VirtualMachine) NewSnapshot(name string) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), map[string]string{"snapname": name}, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}
func (v *VirtualMachine) Snapshots() (snapshots []*Snapshot, err error) {
	err = v.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot", v.Node, v.VMID), &snapshots)
	return
}

func (v *VirtualMachine) SnapshotRollback(name string) (task *Task, err error) {
	var upid UPID
	if err = v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/snapshot/%s/rollback", v.Node, v.VMID, name), nil, &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, v.client), nil
}
