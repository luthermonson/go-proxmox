package proxmox

import (
	"reflect"
	"strconv"
	"strings"
)

func (vmc *VirtualMachineConfig) mergeIndexedDevices(prefix string) map[string]string {
	deviceMap := make(map[string]string)
	t := reflect.TypeOf(*vmc)
	v := reflect.ValueOf(*vmc)
	count := v.NumField()

	for i := 0; i < count; i++ {
		fn := t.Field(i).Name
		fv := v.Field(i).String()
		if fv == "" {
			continue
		}
		if strings.HasPrefix(fn, prefix) {
			// Ignore non-numeric suffixes like SCSIHW
			suffix := strings.TrimPrefix(fn, prefix)
			if _, err := strconv.Atoi(suffix); err != nil {
				continue
			}
			deviceMap[strings.ToLower(fn)] = fv
		}
	}

	return deviceMap
}

func (vmc *VirtualMachineConfig) MergeIDEs() map[string]string {
	if nil == vmc.IDEs {
		vmc.IDEs = vmc.mergeIndexedDevices("IDE")
	}
	return vmc.IDEs
}

func (vmc *VirtualMachineConfig) MergeSCSIs() map[string]string {
	if nil == vmc.SCSIs {
		vmc.SCSIs = vmc.mergeIndexedDevices("SCSI")
	}
	return vmc.SCSIs
}

func (vmc *VirtualMachineConfig) MergeSATAs() map[string]string {
	if nil == vmc.SATAs {
		vmc.SATAs = vmc.mergeIndexedDevices("SATA")
	}
	return vmc.SATAs
}

func (vmc *VirtualMachineConfig) MergeNets() map[string]string {
	if nil == vmc.Nets {
		vmc.Nets = vmc.mergeIndexedDevices("Net")
	}
	return vmc.Nets
}

func (vmc *VirtualMachineConfig) MergeVirtIOs() map[string]string {
	if nil == vmc.VirtIOs {
		vmc.VirtIOs = vmc.mergeIndexedDevices("VirtIO")
	}
	return vmc.VirtIOs
}

func (vmc *VirtualMachineConfig) MergeUnuseds() map[string]string {
	if nil == vmc.Unuseds {
		vmc.Unuseds = vmc.mergeIndexedDevices("Unused")
	}
	return vmc.Unuseds
}

func (vmc *VirtualMachineConfig) MergeSerials() map[string]string {
	if nil == vmc.Serials {
		vmc.Serials = vmc.mergeIndexedDevices("Serial")
	}
	return vmc.Serials
}

func (vmc *VirtualMachineConfig) MergeUSBs() map[string]string {
	if nil == vmc.USBs {
		vmc.USBs = vmc.mergeIndexedDevices("USB")
	}
	return vmc.USBs
}

func (vmc *VirtualMachineConfig) MergeHostPCIs() map[string]string {
	if nil == vmc.HostPCIs {
		vmc.HostPCIs = vmc.mergeIndexedDevices("HostPCI")
	}
	return vmc.HostPCIs
}

func (vmc *VirtualMachineConfig) MergeNumas() map[string]string {
	if nil == vmc.Numas {
		vmc.Numas = vmc.mergeIndexedDevices("Numa")
	}
	return vmc.Numas
}

func (vmc *VirtualMachineConfig) MergeParallels() map[string]string {
	if nil == vmc.Parallels {
		vmc.Parallels = vmc.mergeIndexedDevices("Parallel")
	}
	return vmc.Parallels
}

func (vmc *VirtualMachineConfig) MergeIPConfigs() map[string]string {
	if nil == vmc.IPConfigs {
		vmc.IPConfigs = vmc.mergeIndexedDevices("IPConfig")
	}
	return vmc.IPConfigs
}

func (vmc *VirtualMachineConfig) MergeDisks() map[string]string {
	mergedDisks := make(map[string]string)

	for k, v := range vmc.MergeIDEs() {
		mergedDisks[k] = v
	}

	for k, v := range vmc.MergeSCSIs() {
		mergedDisks[k] = v
	}

	for k, v := range vmc.MergeSATAs() {
		mergedDisks[k] = v
	}

	for k, v := range vmc.MergeVirtIOs() {
		mergedDisks[k] = v
	}
	return mergedDisks
}
