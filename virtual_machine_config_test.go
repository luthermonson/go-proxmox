package proxmox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVirtualMachineConfig_MergeIDEs(t *testing.T) {
	config := &VirtualMachineConfig{
		IDE0: "local:100/vm-100-disk-0.qcow2",
		IDE1: "local:100/vm-100-disk-1.qcow2",
		IDE2: "local:iso/debian-12.iso,media=cdrom",
	}

	ides := config.MergeIDEs()
	assert.NotNil(t, ides)
	assert.Len(t, ides, 3)
	assert.Equal(t, "local:100/vm-100-disk-0.qcow2", ides["ide0"])
	assert.Equal(t, "local:100/vm-100-disk-1.qcow2", ides["ide1"])
	assert.Equal(t, "local:iso/debian-12.iso,media=cdrom", ides["ide2"])

	// Test idempotency - calling again should return the same map
	ides2 := config.MergeIDEs()
	assert.Equal(t, ides, ides2)
}

func TestVirtualMachineConfig_MergeSCSIs(t *testing.T) {
	config := &VirtualMachineConfig{
		SCSI0:  "local-lvm:vm-100-disk-0,size=32G",
		SCSI1:  "local-lvm:vm-100-disk-1,size=64G",
		SCSI10: "local-lvm:vm-100-disk-10,size=128G",
	}

	scsis := config.MergeSCSIs()
	assert.NotNil(t, scsis)
	assert.Len(t, scsis, 3)
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", scsis["scsi0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1,size=64G", scsis["scsi1"])
	assert.Equal(t, "local-lvm:vm-100-disk-10,size=128G", scsis["scsi10"])

	// Test idempotency
	scsis2 := config.MergeSCSIs()
	assert.Equal(t, scsis, scsis2)
}

func TestVirtualMachineConfig_MergeSATAs(t *testing.T) {
	config := &VirtualMachineConfig{
		SATA0: "local-lvm:vm-100-disk-0,size=32G",
		SATA1: "local-lvm:vm-100-disk-1,size=64G",
	}

	satas := config.MergeSATAs()
	assert.NotNil(t, satas)
	assert.Len(t, satas, 2)
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", satas["sata0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1,size=64G", satas["sata1"])
}

func TestVirtualMachineConfig_MergeNets(t *testing.T) {
	config := &VirtualMachineConfig{
		Net0: "virtio=00:11:22:33:44:55,bridge=vmbr0",
		Net1: "virtio=AA:BB:CC:DD:EE:FF,bridge=vmbr1",
	}

	nets := config.MergeNets()
	assert.NotNil(t, nets)
	assert.Len(t, nets, 2)
	assert.Equal(t, "virtio=00:11:22:33:44:55,bridge=vmbr0", nets["net0"])
	assert.Equal(t, "virtio=AA:BB:CC:DD:EE:FF,bridge=vmbr1", nets["net1"])
}

func TestVirtualMachineConfig_MergeVirtIOs(t *testing.T) {
	config := &VirtualMachineConfig{
		VirtIO0: "local-lvm:vm-100-disk-0,size=32G",
		VirtIO1: "local-lvm:vm-100-disk-1,size=64G",
	}

	virtios := config.MergeVirtIOs()
	assert.NotNil(t, virtios)
	assert.Len(t, virtios, 2)
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", virtios["virtio0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1,size=64G", virtios["virtio1"])
}

func TestVirtualMachineConfig_MergeUnuseds(t *testing.T) {
	config := &VirtualMachineConfig{
		Unused0: "local-lvm:vm-100-disk-0",
		Unused1: "local-lvm:vm-100-disk-1",
	}

	unuseds := config.MergeUnuseds()
	assert.NotNil(t, unuseds)
	assert.Len(t, unuseds, 2)
	assert.Equal(t, "local-lvm:vm-100-disk-0", unuseds["unused0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1", unuseds["unused1"])
}

func TestVirtualMachineConfig_MergeSerials(t *testing.T) {
	config := &VirtualMachineConfig{
		Serial0: "socket",
		Serial1: "/dev/ttyS1",
	}

	serials := config.MergeSerials()
	assert.NotNil(t, serials)
	assert.Len(t, serials, 2)
	assert.Equal(t, "socket", serials["serial0"])
	assert.Equal(t, "/dev/ttyS1", serials["serial1"])
}

func TestVirtualMachineConfig_MergeUSBs(t *testing.T) {
	config := &VirtualMachineConfig{
		USB0: "host=1234:5678",
		USB1: "host=8765:4321",
	}

	usbs := config.MergeUSBs()
	assert.NotNil(t, usbs)
	assert.Len(t, usbs, 2)
	assert.Equal(t, "host=1234:5678", usbs["usb0"])
	assert.Equal(t, "host=8765:4321", usbs["usb1"])
}

func TestVirtualMachineConfig_MergeHostPCIs(t *testing.T) {
	config := &VirtualMachineConfig{
		HostPCI0: "0000:01:00.0",
		HostPCI1: "0000:02:00.0,pcie=1",
	}

	hostpcis := config.MergeHostPCIs()
	assert.NotNil(t, hostpcis)
	assert.Len(t, hostpcis, 2)
	assert.Equal(t, "0000:01:00.0", hostpcis["hostpci0"])
	assert.Equal(t, "0000:02:00.0,pcie=1", hostpcis["hostpci1"])
}

func TestVirtualMachineConfig_MergeNumas(t *testing.T) {
	config := &VirtualMachineConfig{
		Numa0: "cpus=0-1,memory=2048",
		Numa1: "cpus=2-3,memory=2048",
	}

	numas := config.MergeNumas()
	assert.NotNil(t, numas)
	assert.Len(t, numas, 2)
	assert.Equal(t, "cpus=0-1,memory=2048", numas["numa0"])
	assert.Equal(t, "cpus=2-3,memory=2048", numas["numa1"])
}

func TestVirtualMachineConfig_MergeParallels(t *testing.T) {
	config := &VirtualMachineConfig{
		Parallel0: "/dev/parport0",
	}

	parallels := config.MergeParallels()
	assert.NotNil(t, parallels)
	assert.Len(t, parallels, 1)
	assert.Equal(t, "/dev/parport0", parallels["parallel0"])
}

func TestVirtualMachineConfig_MergeIPConfigs(t *testing.T) {
	config := &VirtualMachineConfig{
		IPConfig0: "ip=192.168.1.10/24,gw=192.168.1.1",
		IPConfig1: "ip=10.0.0.10/24,gw=10.0.0.1",
	}

	ipconfigs := config.MergeIPConfigs()
	assert.NotNil(t, ipconfigs)
	assert.Len(t, ipconfigs, 2)
	assert.Equal(t, "ip=192.168.1.10/24,gw=192.168.1.1", ipconfigs["ipconfig0"])
	assert.Equal(t, "ip=10.0.0.10/24,gw=10.0.0.1", ipconfigs["ipconfig1"])
}

func TestVirtualMachineConfig_MergeDisks(t *testing.T) {
	config := &VirtualMachineConfig{
		IDE0:    "local:100/vm-100-disk-0.qcow2",
		SCSI0:   "local-lvm:vm-100-disk-1,size=32G",
		SATA0:   "local-lvm:vm-100-disk-2,size=64G",
		VirtIO0: "local-lvm:vm-100-disk-3,size=128G",
	}

	disks := config.MergeDisks()
	assert.NotNil(t, disks)
	assert.Len(t, disks, 4)
	assert.Equal(t, "local:100/vm-100-disk-0.qcow2", disks["ide0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1,size=32G", disks["scsi0"])
	assert.Equal(t, "local-lvm:vm-100-disk-2,size=64G", disks["sata0"])
	assert.Equal(t, "local-lvm:vm-100-disk-3,size=128G", disks["virtio0"])
}

func TestVirtualMachineConfig_MergeDisks_Empty(t *testing.T) {
	config := &VirtualMachineConfig{}

	disks := config.MergeDisks()
	assert.NotNil(t, disks)
	assert.Len(t, disks, 0)
}

func TestVirtualMachineConfig_MergeIDEs_Empty(t *testing.T) {
	config := &VirtualMachineConfig{}

	ides := config.MergeIDEs()
	assert.NotNil(t, ides)
	assert.Len(t, ides, 0)
}

func TestVirtualMachineConfig_MergeIndexedDevices_IgnoresNonNumeric(t *testing.T) {
	// This tests that fields like SCSIHW (which starts with SCSI but has non-numeric suffix)
	// are correctly ignored by the merge function
	config := &VirtualMachineConfig{
		SCSI0:   "local-lvm:vm-100-disk-0,size=32G",
		SCSIHW: "virtio-scsi-pci",
	}

	scsis := config.MergeSCSIs()
	assert.NotNil(t, scsis)
	assert.Len(t, scsis, 1) // Should only have SCSI0, not SCSIHW
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", scsis["scsi0"])
	_, hasSCSIHW := scsis["scsihw"]
	assert.False(t, hasSCSIHW, "SCSIHW should not be included in merge")
}
