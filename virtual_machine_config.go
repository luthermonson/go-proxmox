package proxmox

// MergeDisks returns every disk attached to the VM as a single map keyed by
// its on-the-wire name (e.g. "ide0", "scsi3", "sata1", "virtio7"), combining
// IDEs, SCSIs, SATAs, and VirtIOs.
func (vmc *VirtualMachineConfig) MergeDisks() map[string]string {
	merged := make(map[string]string, len(vmc.IDEs)+len(vmc.SCSIs)+len(vmc.SATAs)+len(vmc.VirtIOs))
	for _, m := range []map[string]string{vmc.IDEs, vmc.SCSIs, vmc.SATAs, vmc.VirtIOs} {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
