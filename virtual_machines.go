//go:build vms
// +build vms

package proxmox

import "fmt"

func (v *VirtualMachine) Start() (status string, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/start", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Stop() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Suspend() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/suspend", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Reboot() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/reboot", v.Node, v.VMID), nil, &status)
}

func (v *VirtualMachine) Resume() (status *VirtualMachineStatus, err error) {
	return status, v.client.Post(fmt.Sprintf("/nodes/%s/qemu/%d/status/resume", v.Node, v.VMID), nil, &status)
}
