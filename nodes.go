package proxmox

import "fmt"

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

func (n *Node) VirtualMachines() (vms VirtualMachines, err error) {
	return vms, n.client.Get(fmt.Sprintf("/nodes/%s/qemu", n.Name), &vms)
}

func (n *Node) VirtualMachine(vmid int) (vm *VirtualMachine, err error) {
	return vm, n.client.Get(fmt.Sprintf("/nodes/%s/qemu/%d/status/current", n.Name, vmid), &vm)
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

func (n *Node) Container(vmid int) (c *Container, err error) {
	c.client = n.client
	c.Node = n.Name

	return c, n.client.Get(fmt.Sprintf("/nodes/%s/lxc/%d/status/current", n.Name, vmid), &c)
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

func (n *Node) VzTmpls(storage string) (templates VzTpls, err error) {
	return templates, n.client.Get(fmt.Sprintf("/nodes/%s/storage/%s/content?content=vztmpl", n.Name, storage), &templates)
}

func (n *Node) VzTmpl(template, storage string) (*VzTpl, error) {
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
	return storages, n.client.Get(fmt.Sprintf("/nodes/%s/storage", n.Name), &storages)
}

// TODO https://192.168.1.6:8006/api2/extjs/nodes/i7/storage/local/content//local:vztmpl/alpine-3.11-default_20200425_amd64.tar.xz?delay=5
func (n *Node) DeleteFile() (ret string, err error) {
	return ret, err
}
