package proxmox

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) Nodes() (ns NodeStatuses, err error) {
	return ns, c.Get("/nodes", &ns)
}

func (c *Client) Node(name string) (node *Node, err error) {
	if err := c.Get(fmt.Sprintf("/nodes/%s/status", name), &node); err != nil {
		return nil, err
	}
	node.Name = name
	node.client = c

	return
}

func (n *Node) Version() (version *Version, err error) {
	return version, n.client.Get(fmt.Sprintf("/nodes/%s/version", n.Name), &version)
}

func (n *Node) TermProxy() (vnc *VNC, err error) {
	return vnc, n.client.Post(fmt.Sprintf("/nodes/%s/termproxy", n.Name), nil, &vnc)
}

// VNCWebSocket send, recv, errors, closer, error
func (n *Node) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/vncwebsocket?port=%d&vncticket=%s",
		n.Name, vnc.Port, url.QueryEscape(vnc.Ticket))

	return n.client.VNCWebSocket(p, vnc)
}

func (n *Node) VirtualMachines() (vms VirtualMachines, err error) {
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu", n.Name), &vms); err != nil {
		return nil, err
	}

	for _, v := range vms {
		v.client = n.client
		v.Node = n.Name
	}

	return vms, nil
}

func (n *Node) NewVirtualMachine(vmid int, options ...VirtualMachineOption) (*Task, error) {
	var upid UPID
	data := make(map[string]interface{})
	data["vmid"] = vmid

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err := n.client.Post(fmt.Sprintf("/nodes/%s/qemu", n.Name), data, &upid)
	return NewTask(upid, n.client), err
}

func (n *Node) VirtualMachine(vmid int) (*VirtualMachine, error) {
	vm := &VirtualMachine{
		client: n.client,
		Node:   n.Name,
	}

	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", n.Name, vmid), &vm); nil != err {
		return nil, err
	}

	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/config", n.Name, vmid), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	return vm, nil
}

func (n *Node) Containers() (c Containers, err error) {
	if err = n.client.Get(fmt.Sprintf("/nodes/%s/lxc", n.Name), &c); err != nil {
		return
	}

	for _, container := range c {
		container.client = n.client
		container.Node = n.Name
	}

	return
}

func (n *Node) Container(vmid int) (*Container, error) {
	var c Container
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/lxc/%d/status/current", n.Name, vmid), &c); err != nil {
		return nil, err
	}
	c.client = n.client
	c.Node = n.Name

	return &c, nil
}

func (n *Node) Appliances() (appliances Appliances, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/aplinfo", n.Name), &appliances)
	if err != nil {
		return appliances, err
	}

	for _, t := range appliances {
		t.client = n.client
		t.Node = n.Name
	}

	return appliances, nil
}

func (n *Node) DownloadAppliance(template, storage string) (ret string, err error) {
	return ret, n.client.Post(fmt.Sprintf("/nodes/%s/aplinfo", n.Name), map[string]string{
		"template": template,
		"storage":  storage,
	}, &ret)
}

func (n *Node) VzTmpls(storage string) (templates VzTmpls, err error) {
	return templates, n.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content?content=vztmpl", n.Name, storage), &templates)
}

func (n *Node) VzTmpl(template, storage string) (*VzTmpl, error) {
	templates, err := n.VzTmpls(storage)
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

func (n *Node) Storages() (storages Storages, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/storage", n.Name), &storages)
	if err != nil {
		return
	}

	for _, s := range storages {
		s.Node = n.Name
		s.client = n.client
	}

	return
}

func (n *Node) Storage(name string) (storage *Storage, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/status", n.Name, name), &storage)
	if err != nil {
		return
	}

	storage.Node = n.Name
	storage.client = n.client
	storage.Name = name

	return
}

func (n *Node) StorageISO() (*Storage, error) {
	return n.findStorageByContent("iso")
}

func (n *Node) StorageVZTmpl() (*Storage, error) {
	return n.findStorageByContent("vztmpl")
}

func (n *Node) StorageBackup() (*Storage, error) {
	return n.findStorageByContent("backup")
}

func (n *Node) StorageRootDir() (*Storage, error) {
	return n.findStorageByContent("rootdir")
}

func (n *Node) StorageImages() (*Storage, error) {
	return n.findStorageByContent("images")
}

// findStorageByContent takes iso/backup/vztmpl/rootdir/images and returns the storage that type of content should be on
func (n *Node) findStorageByContent(content string) (storage *Storage, err error) {
	storages, err := n.Storages()
	if err != nil {
		return nil, err
	}

	for _, storage := range storages {
		if strings.Contains(storage.Content, content) {
			storage.Node = n.Name
			storage.client = n.client
			return storage, nil
		}
	}

	return nil, ErrNotFound
}

func (n *Node) FirewallOptionGet() (firewallOption *FirewallNodeOption, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/firewall/options", n.Name), firewallOption)
	return
}

func (n *Node) FirewallOptionSet(firewallOption *FirewallNodeOption) error {
	return n.client.Put(fmt.Sprintf("/nodes/%s/firewall/options", n.Name), firewallOption, nil)
}

func (n *Node) FirewallGetRules() (rules []*FirewallRule, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), &rules)
	return
}

func (n *Node) FirewallRulesCreate(rule *FirewallRule) error {
	return n.client.Post(fmt.Sprintf("/nodes/%s/firewall/rules", n.Name), rule, nil)
}

func (n *Node) FirewallRulesUpdate(rule *FirewallRule) error {
	return n.client.Put(fmt.Sprintf("/nodes/%s/firewall/rules/%d", n.Name, rule.Pos), rule, nil)
}

func (n *Node) FirewallRulesDelete(rulePos int) error {
	return n.client.Delete(fmt.Sprintf("/nodes/%s/firewall/rules/%d", n.Name, rulePos), nil)
}
