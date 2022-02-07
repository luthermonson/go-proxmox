package proxmox

import (
	"fmt"
	"net/url"
)

func (c *Container) Start() (status string, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.Node, c.VMID), nil, &status)
}

func (c *Container) Stop() (status *ContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.Node, c.VMID), nil, &status)
}

func (c *Container) Suspend() (status *ContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/suspend", c.Node, c.VMID), nil, &status)
}

func (c *Container) Reboot() (status *ContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.Node, c.VMID), nil, &status)
}

func (c *Container) Resume() (status *ContainerStatus, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%d/status/resume", c.Node, c.VMID), nil, &status)
}

func (c *Container) TermProxy() (vnc *VNC, err error) {
	return vnc, c.client.Post(fmt.Sprintf("/nodes/%s/lxk/%d/termproxy", c.Node, c.VMID), nil, &vnc)
}

func (c *Container) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/lxc/%d/vncwebsocket?port=%d&vncticket=%s",
		c.Node, c.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return c.client.VNCWebSocket(p, vnc)
}
