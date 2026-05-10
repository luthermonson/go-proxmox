package proxmox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luthermonson/go-proxmox/tests/mocks"
)

func TestVirtualMachineConfig_MergeDisks(t *testing.T) {
	cfg := &VirtualMachineConfig{
		IDEs:    map[string]string{"ide0": "local:100/vm-100-disk-0.qcow2"},
		SCSIs:   map[string]string{"scsi0": "local-lvm:vm-100-disk-1,size=32G"},
		SATAs:   map[string]string{"sata0": "local-lvm:vm-100-disk-2,size=64G"},
		VirtIOs: map[string]string{"virtio0": "local-lvm:vm-100-disk-3,size=128G"},
	}

	disks := cfg.MergeDisks()
	assert.NotNil(t, disks)
	assert.Len(t, disks, 4)
	assert.Equal(t, "local:100/vm-100-disk-0.qcow2", disks["ide0"])
	assert.Equal(t, "local-lvm:vm-100-disk-1,size=32G", disks["scsi0"])
	assert.Equal(t, "local-lvm:vm-100-disk-2,size=64G", disks["sata0"])
	assert.Equal(t, "local-lvm:vm-100-disk-3,size=128G", disks["virtio0"])
}

func TestVirtualMachineConfig_MergeDisks_Empty(t *testing.T) {
	disks := (&VirtualMachineConfig{}).MergeDisks()
	assert.NotNil(t, disks)
	assert.Len(t, disks, 0)
}

// TestVirtualMachineConfig_UnmarshalJSON_BeyondTen exercises issue #211: the
// Proxmox API can return device indices well past 9 (net0..net31, scsi0..scsi30,
// unused0..unused255 etc.). The maps must capture every index from the raw JSON.
func TestVirtualMachineConfig_UnmarshalJSON_BeyondTen(t *testing.T) {
	body := []byte(`{
		"net0": "virtio=00:00:00:00:00:00,bridge=vmbr0",
		"net15": "virtio=00:00:00:00:00:15,bridge=vmbr15",
		"net31": "virtio=00:00:00:00:00:31,bridge=vmbr31",
		"scsi0": "local-lvm:vm-100-disk-0,size=32G",
		"scsi30": "local-lvm:vm-100-disk-30,size=32G",
		"unused15": "local-lvm:vm-100-unused-15",
		"unused255": "local-lvm:vm-100-unused-255",
		"hostpci15": "0000:0f:00.0",
		"ipconfig20": "ip=10.0.0.20/24"
	}`)

	var cfg VirtualMachineConfig
	assert.NoError(t, json.Unmarshal(body, &cfg))

	assert.Equal(t, "virtio=00:00:00:00:00:00,bridge=vmbr0", cfg.Nets["net0"])
	assert.Equal(t, "virtio=00:00:00:00:00:15,bridge=vmbr15", cfg.Nets["net15"])
	assert.Equal(t, "virtio=00:00:00:00:00:31,bridge=vmbr31", cfg.Nets["net31"])
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", cfg.SCSIs["scsi0"])
	assert.Equal(t, "local-lvm:vm-100-disk-30,size=32G", cfg.SCSIs["scsi30"])
	assert.Equal(t, "local-lvm:vm-100-unused-15", cfg.Unuseds["unused15"])
	assert.Equal(t, "local-lvm:vm-100-unused-255", cfg.Unuseds["unused255"])
	assert.Equal(t, "0000:0f:00.0", cfg.HostPCIs["hostpci15"])
	assert.Equal(t, "ip=10.0.0.20/24", cfg.IPConfigs["ipconfig20"])
}

// TestVirtualMachineConfig_UnmarshalJSON_PrefixCollisions guards against the
// regression that closed PR #217 would have introduced: routing keys via
// strings.HasPrefix puts "scsihw" into the SCSIs map and the bare "numa"
// scalar into the Numas map. The prefix-then-pure-digits routing skips both.
func TestVirtualMachineConfig_UnmarshalJSON_PrefixCollisions(t *testing.T) {
	body := []byte(`{
		"scsihw": "virtio-scsi-pci",
		"scsi0": "local-lvm:vm-100-disk-0,size=32G",
		"numa": 1,
		"numa0": "cpus=0-1,memory=2048"
	}`)

	var cfg VirtualMachineConfig
	assert.NoError(t, json.Unmarshal(body, &cfg))

	if assert.NotNil(t, cfg.SCSIHW) {
		assert.Equal(t, "virtio-scsi-pci", *cfg.SCSIHW)
	}
	_, hasSCSIHW := cfg.SCSIs["scsihw"]
	assert.False(t, hasSCSIHW, "SCSIHW must not be routed into SCSIs")
	assert.Equal(t, "local-lvm:vm-100-disk-0,size=32G", cfg.SCSIs["scsi0"])

	assert.Equal(t, IntOrBool(true), cfg.Numa)
	_, hasBareNuma := cfg.Numas["numa"]
	assert.False(t, hasBareNuma, "bare numa scalar must not be routed into Numas")
	assert.Equal(t, "cpus=0-1,memory=2048", cfg.Numas["numa0"])
}

func TestIndexedDeviceKey(t *testing.T) {
	cases := []struct {
		in     string
		prefix string
		ok     bool
	}{
		{"net0", "net", true},
		{"net31", "net", true},
		{"scsi30", "scsi", true},
		{"unused255", "unused", true},
		{"ipconfig20", "ipconfig", true},
		{"scsihw", "", false},
		{"numa", "", false},
		{"net", "", false},
		{"net0a", "", false},
		{"123", "", false},
		{"", "", false},
		{"a1b", "", false},
	}
	for _, tc := range cases {
		prefix, ok := indexedDeviceKey(tc.in)
		assert.Equal(t, tc.ok, ok, "ok for %q", tc.in)
		assert.Equal(t, tc.prefix, prefix, "prefix for %q", tc.in)
	}
}

// TestNode_VirtualMachineConfig_HighIndices is the integration-shaped regression
// test for issue #211: it goes through node.VirtualMachine(ctx, 102), which hits
// the gock mock returning net15..net31, scsi30, unused255, hostpci15, ipconfig20,
// plus prefix-collision keys (scsihw, bare numa).
func TestNode_VirtualMachineConfig_HighIndices(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	vm, err := node.VirtualMachine(ctx, 102)
	assert.Nil(t, err)
	assert.NotNil(t, vm)
	assert.NotNil(t, vm.VirtualMachineConfig)

	cfg := vm.VirtualMachineConfig

	assert.Equal(t, "virtio=00:00:00:00:00:00,bridge=vmbr0", cfg.Nets["net0"])
	assert.Equal(t, "virtio=00:00:00:00:00:15,bridge=vmbr15", cfg.Nets["net15"])
	assert.Equal(t, "virtio=00:00:00:00:00:31,bridge=vmbr31", cfg.Nets["net31"])
	assert.Equal(t, "local-lvm:vm-102-disk-0,size=32G", cfg.SCSIs["scsi0"])
	assert.Equal(t, "local-lvm:vm-102-disk-30,size=32G", cfg.SCSIs["scsi30"])
	assert.Equal(t, "local-lvm:vm-102-unused-15", cfg.Unuseds["unused15"])
	assert.Equal(t, "local-lvm:vm-102-unused-255", cfg.Unuseds["unused255"])
	assert.Equal(t, "0000:0f:00.0", cfg.HostPCIs["hostpci15"])
	assert.Equal(t, "ip=10.0.0.20/24,gw=10.0.0.1", cfg.IPConfigs["ipconfig20"])
	assert.Equal(t, "cpus=0-1,memory=2048", cfg.Numas["numa0"])

	// scsihw must remain in the SCSIHW scalar — never in SCSIs.
	if assert.NotNil(t, cfg.SCSIHW) {
		assert.Equal(t, "virtio-scsi-pci", *cfg.SCSIHW)
	}
	_, hasSCSIHW := cfg.SCSIs["scsihw"]
	assert.False(t, hasSCSIHW)

	// Bare numa scalar must remain in Numa — never in Numas.
	assert.Equal(t, IntOrBool(true), cfg.Numa)
	_, hasBareNuma := cfg.Numas["numa"]
	assert.False(t, hasBareNuma)
}

func TestNode_VirtualMachineConfig_AllMergedMapsPopulated(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	vm, err := node.VirtualMachine(ctx, 101)
	assert.Nil(t, err)
	assert.NotNil(t, vm)
	assert.NotNil(t, vm.VirtualMachineConfig)

	cfg := vm.VirtualMachineConfig

	// IDEs
	assert.NotNil(t, cfg.IDEs)
	assert.Len(t, cfg.IDEs, 2)
	assert.Equal(t, "local:100/vm-101-disk-0.qcow2", cfg.IDEs["ide0"])
	assert.Equal(t, "local:iso/debian-12.iso,media=cdrom", cfg.IDEs["ide2"])

	// SCSIs
	assert.NotNil(t, cfg.SCSIs)
	assert.Len(t, cfg.SCSIs, 1)
	assert.Equal(t, "local-lvm:vm-101-disk-0,size=32G", cfg.SCSIs["scsi0"])

	// SATAs
	assert.NotNil(t, cfg.SATAs)
	assert.Len(t, cfg.SATAs, 1)
	assert.Equal(t, "local-lvm:vm-101-disk-1,size=32G", cfg.SATAs["sata0"])

	// VirtIOs
	assert.NotNil(t, cfg.VirtIOs)
	assert.Len(t, cfg.VirtIOs, 1)
	assert.Equal(t, "local-lvm:vm-101-disk-2,size=64G", cfg.VirtIOs["virtio0"])

	// Unuseds
	assert.NotNil(t, cfg.Unuseds)
	assert.Len(t, cfg.Unuseds, 1)
	assert.Equal(t, "local-lvm:vm-101-unused", cfg.Unuseds["unused0"])

	// Nets
	assert.NotNil(t, cfg.Nets)
	assert.Len(t, cfg.Nets, 1)
	assert.Equal(t, "virtio=BC:24:11:2E:C5:4A,bridge=vmbr0", cfg.Nets["net0"])

	// Numas
	assert.NotNil(t, cfg.Numas)
	assert.Len(t, cfg.Numas, 1)
	assert.Equal(t, "cpus=0-1,memory=2048", cfg.Numas["numa0"])

	// HostPCIs
	assert.NotNil(t, cfg.HostPCIs)
	assert.Len(t, cfg.HostPCIs, 1)
	assert.Equal(t, "0000:01:00.0", cfg.HostPCIs["hostpci0"])

	// Serials
	assert.NotNil(t, cfg.Serials)
	assert.Len(t, cfg.Serials, 1)
	assert.Equal(t, "socket", cfg.Serials["serial0"])

	// USBs
	assert.NotNil(t, cfg.USBs)
	assert.Len(t, cfg.USBs, 1)
	assert.Equal(t, "host=1234:5678", cfg.USBs["usb0"])

	// Parallels
	assert.NotNil(t, cfg.Parallels)
	assert.Len(t, cfg.Parallels, 1)
	assert.Equal(t, "/dev/parport0", cfg.Parallels["parallel0"])

	// IPConfigs
	assert.NotNil(t, cfg.IPConfigs)
	assert.Len(t, cfg.IPConfigs, 1)
	assert.Equal(t, "ip=192.168.1.10/24,gw=192.168.1.1", cfg.IPConfigs["ipconfig0"])
}
