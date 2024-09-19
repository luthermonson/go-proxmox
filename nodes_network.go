package proxmox

import (
	"context"
	"fmt"
)

func (n *Node) NewNetwork(ctx context.Context, network *NodeNetwork) (task *Task, err error) {
	err = n.client.Post(ctx, fmt.Sprintf("/nodes/%s/network", n.Name), network, network)
	if nil != err {
		return
	}

	network.client = n.client
	network.Node = n.Name
	network.NodeAPI = n
	return n.NetworkReload(ctx)
}

func (n *Node) Network(ctx context.Context, iface string) (network *NodeNetwork, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/network/%s", n.Name, iface), &network)
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

func (n *Node) Networks(ctx context.Context) (networks NodeNetworks, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/network", n.Name), &networks)
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

func (n *Node) NetworkReload(ctx context.Context) (*Task, error) {
	var upid UPID
	err := n.client.Put(ctx, fmt.Sprintf("/nodes/%s/network", n.Name), nil, &upid)
	if err != nil {
		return nil, err
	}

	return NewTask(upid, n.client), nil
}

func (nw *NodeNetwork) Update(ctx context.Context) error {
	if nw.Iface == "" {
		return nil
	}
	return nw.client.Put(ctx, fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), nw, nil)
}

func (nw *NodeNetwork) Delete(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if nw.Iface == "" {
		return
	}
	err = nw.client.Delete(ctx, fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), &upid)
	if err != nil {
		return
	}

	return nw.NodeAPI.NetworkReload(ctx)
}
