package proxmox

import "fmt"

func (n *Node) NewNetwork(network *NodeNetwork) (task *Task, err error) {
	err = n.client.Post(fmt.Sprintf("/nodes/%s/network", n.Name), network, network)
	if nil != err {
		return
	}

	network.client = n.client
	network.Node = n.Name
	network.NodeAPI = n
	return n.NetworkReload()
}

func (n *Node) Network(iface string) (network *NodeNetwork, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/network/%s", n.Name, iface), &network)
	if err != nil {
		return
	}

	if nil != network {
		network.client = n.client
		network.Node = n.Name
		network.NodeAPI = n
		network.Iface = iface
	}

	return
}

func (n *Node) Networks() (networks NodeNetworks, err error) {
	err = n.client.Get(fmt.Sprintf("/nodes/%s/network", n.Name), &networks)
	if err != nil {
		return nil, err
	}

	for _, v := range networks {
		v.client = n.client
		v.Node = n.Name
		v.NodeAPI = n
	}

	return
}

func (n *Node) NetworkReload() (*Task, error) {
	var upid UPID
	err := n.client.Put(fmt.Sprintf("/nodes/%s/network", n.Name), nil, &upid)
	if err != nil {
		return nil, err
	}

	return NewTask(upid, n.client), nil
}

func (nw *NodeNetwork) Update() error {
	if "" == nw.Iface {
		return nil
	}
	return nw.client.Put(fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), nw, nil)
}

func (nw *NodeNetwork) Delete() (task *Task, err error) {
	var upid UPID
	if "" == nw.Iface {
		return
	}
	err = nw.client.Delete(fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), &upid)
	if err != nil {
		return
	}

	return nw.NodeAPI.NetworkReload()
}
