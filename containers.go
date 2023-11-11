package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Container) Clone(ctx context.Context, params *ContainerCloneOptions) (newid int, task *Task, err error) {
	var upid UPID

	if params == nil {
		params = &ContainerCloneOptions{}
	}
	if params.NewID <= 0 {
		cluster, err := c.client.Cluster(ctx)
		if err != nil {
			return newid, nil, err
		}
		newid, err := cluster.NextID(ctx)
		if err != nil {
			return newid, nil, err
		}
		params.NewID = newid
	}
	if err := c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/clone", c.Node, c.VMID), params, &upid); err != nil {
		return 0, nil, err
	}
	return newid, NewTask(upid, c.client), nil
}

func (c *Container) Delete(ctx context.Context) (task *Task, err error) {
	var upid UPID
	if err := c.client.Delete(ctx, fmt.Sprintf("/nodes/%s/lxc/%d", c.Node, c.VMID), &upid); err != nil {
		return nil, err
	}

	return NewTask(upid, c.client), nil
}

func (c *Container) Start(ctx context.Context) (status string, err error) {
	return status, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.Node, c.VMID), nil, &status)
}

func (c *Container) Stop(ctx context.Context) (status *ContainerStatus, err error) {
	return status, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.Node, c.VMID), nil, &status)
}

func (c *Container) Suspend(ctx context.Context) (status *ContainerStatus, err error) {
	return status, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/suspend", c.Node, c.VMID), nil, &status)
}

func (c *Container) Reboot(ctx context.Context) (status *ContainerStatus, err error) {
	return status, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.Node, c.VMID), nil, &status)
}

func (c *Container) Resume(ctx context.Context) (status *ContainerStatus, err error) {
	return status, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxc/%d/status/resume", c.Node, c.VMID), nil, &status)
}

func (c *Container) TermProxy(ctx context.Context) (vnc *VNC, err error) {
	return vnc, c.client.Post(ctx, fmt.Sprintf("/nodes/%s/lxk/%d/termproxy", c.Node, c.VMID), nil, &vnc)
}

func (c *Container) VNCWebSocket(vnc *VNC) (chan string, chan string, chan error, func() error, error) {
	p := fmt.Sprintf("/nodes/%s/lxc/%d/vncwebsocket?port=%d&vncticket=%s",
		c.Node, c.VMID, vnc.Port, url.QueryEscape(vnc.Ticket))

	return c.client.VNCWebSocket(p, vnc)
}
