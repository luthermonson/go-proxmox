package proxmox

import (
	"fmt"
)

func (nw *NodeNetwork) Delete() (task *Task, err error) {
	var upid UPID
	if "" == nw.Iface {
		return
	}
	err = nw.client.Delete(fmt.Sprintf("/nodes/%s/network/%s", nw.Node, nw.Iface), &upid)
	if err != nil {
		return
	}

	return nw.NodeApi.NetworkReload()
}
