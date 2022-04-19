package proxmox

import (
	"fmt"
	"net/url"
)

func (c *Client) Nodes() (ns NodeStatuses, err error) {
	return ns, c.Get("/nodes", &ns)
}

func (c *Client) Node(name string) (*Node, error) {
	var node Node
	if err := c.Get(fmt.Sprintf("/nodes/%s/status", name), &node); err != nil {
		return nil, err
	}
	node.Name = name
	node.client = c

	return &node, nil
}

func (n *Node) Version() (version *Version, err error) {
	return version, n.client.Get("/nodes/%s/version", &version)
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

	//var vmconf VirtualMachineConfig
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/config", n.Name, vmid), &vm.VirtualMachineConfig); err != nil {
		return nil, err
	}

	//vm.VirtualMachineConfig = &vmconf

	return vm, nil
}

func (n *Node) Containers() (c Containers, err error) {
	if err := n.client.Get(fmt.Sprintf("/nodes/%s/lxc", n.Name), &c); err != nil {
		return nil, err
	}

	for _, container := range c {
		container.client = n.client
		container.Node = n.Name
	}

	return c, nil
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

//networks
func (n *Node) Networks() (networks NodeNetworks, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/network", n.Name), &networks)
	if err != nil {
		return nil, err
	}

	for _, v := range networks {
		v.client = n.client
		v.Node = n.Name
		v.NodeApi = n
	}

	return
}
func (n *Node) Network(iface string) (network *NodeNetwork, err error) {

	err = n.client.Get(fmt.Sprintf("/nodes/%s/network/%s", n.Name, iface), &network)
	if err != nil {
		return nil, err
	}

	if nil != network {
		network.client = n.client
		network.Node = n.Name
		network.NodeApi = n
		network.Iface = iface
	}

	return network, nil
}

func (n *Node) NewNetwork(network *NodeNetwork) (task *Task, err error) {

	err = n.client.Post(fmt.Sprintf("/nodes/%s/network", n.Name), network, network)
	if nil != err {
		return
	}

	network.client = n.client
	network.Node = n.Name
	network.NodeApi = n
	return n.NetworkReload()
}
func (n *Node) NetworkReload() (*Task, error) {
	var upid UPID
	err := n.client.Put(fmt.Sprintf("/nodes/%s/network", n.Name), nil, &upid)
	if err != nil {
		return nil, err
	}

	return NewTask(upid, n.client), nil
}
