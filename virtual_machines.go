// +build vms

package proxmox

import "fmt"

func (v *VirtualMachine) Start() (status string, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%s/status/start", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Stop() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%s/status/stop", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Suspend() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%s/status/suspend", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Reboot() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%s/status/reboot", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Resume() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%s/status/resume", v.Node, v.VMID), nil, &status)
}
