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

func (n *Node) AplInfo() (aplinfos AplInfos, err error) {
	return aplinfos, n.client.Get(fmt.Sprintf("/nodes/%s/aplinfo", n.Name), &aplinfos)
}
