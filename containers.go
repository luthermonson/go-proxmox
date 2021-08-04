package proxmox

import (
	"encoding/json"
	"fmt"
)

func (c *Container) Start() (status string, err error) {
	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%s/status/start", c.Node, c.VMID), []byte{}, &status)
}

func (c *Container) Stop() (status ContainerStatus, err error) {
	json, err := json.Marshal(c)
	if err != nil {
		return status, err
	}

	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%s/status/stop", c.Node, c.VMID), json, &status)
}

func (c *Container) Suspend() (status ContainerStatus, err error) {
	json, err := json.Marshal(c)
	if err != nil {
		return status, err
	}

	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%s/status/suspend", c.Node, c.VMID), json, &status)
}

func (c *Container) Reboot() (status ContainerStatus, err error) {
	json, err := json.Marshal(c)
	if err != nil {
		return status, err
	}

	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%s/status/reboot", c.Node, c.VMID), json, &status)
}

func (c *Container) Resume() (status ContainerStatus, err error) {
	json, err := json.Marshal(c)
	if err != nil {
		return status, err
	}

	return status, c.client.Post(fmt.Sprintf("/nodes/%s/lxc/%s/status/resume", c.Node, c.VMID), json, &status)
}
