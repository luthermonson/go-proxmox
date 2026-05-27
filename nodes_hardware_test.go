package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func hwNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

// --- /nodes/{node}/scan -----------------------------------------------------

func TestNode_ScanIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	methods, err := hwNode().ScanIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, methods, "zfs")
	assert.Contains(t, methods, "pbs")
}

func TestNode_ScanZFS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pools, err := hwNode().ScanZFS(context.Background())
	assert.Nil(t, err)
	assert.Len(t, pools, 2)
	assert.Equal(t, "rpool", pools[0].Pool)
}

func TestNode_ScanLVM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vgs, err := hwNode().ScanLVM(context.Background())
	assert.Nil(t, err)
	assert.Len(t, vgs, 2)
	assert.Equal(t, "pve", vgs[0].VG)
}

func TestNode_ScanLVMThin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pools, err := hwNode().ScanLVMThin(context.Background(), "pve")
	assert.Nil(t, err)
	assert.Len(t, pools, 1)
	assert.Equal(t, "data", pools[0].LV)

	_, err = hwNode().ScanLVMThin(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_ScanNFS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	exports, err := hwNode().ScanNFS(context.Background(), "nfs.example.com")
	assert.Nil(t, err)
	assert.Len(t, exports, 2)
	assert.Equal(t, "/exports/backup", exports[0].Path)

	_, err = hwNode().ScanNFS(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_ScanCIFS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	shares, err := hwNode().ScanCIFS(context.Background(), &ScanCIFSOptions{
		Server:   "cifs.example.com",
		Username: "u",
		Password: "p",
		Domain:   "WORKGROUP",
	})
	assert.Nil(t, err)
	assert.Len(t, shares, 2)
	assert.Equal(t, "backup", shares[0].Share)

	_, err = hwNode().ScanCIFS(context.Background(), nil)
	assert.NotNil(t, err)
	_, err = hwNode().ScanCIFS(context.Background(), &ScanCIFSOptions{})
	assert.NotNil(t, err)
}

func TestNode_ScanPBS(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	stores, err := hwNode().ScanPBS(context.Background(), &ScanPBSOptions{
		Server:      "pbs.example.com",
		Username:    "user@pbs",
		Password:    "p",
		Fingerprint: "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99",
		Port:        8007,
	})
	assert.Nil(t, err)
	assert.Len(t, stores, 2)
	assert.Equal(t, "main", stores[0].Store)

	_, err = hwNode().ScanPBS(context.Background(), &ScanPBSOptions{Server: "x"})
	assert.NotNil(t, err)
}

func TestNode_ScanISCSI(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	targets, err := hwNode().ScanISCSI(context.Background(), "192.0.2.10:3260")
	assert.Nil(t, err)
	assert.Len(t, targets, 2)

	_, err = hwNode().ScanISCSI(context.Background(), "")
	assert.NotNil(t, err)
}

// --- /nodes/{node}/capabilities --------------------------------------------

func TestNode_CapabilitiesIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subs, err := hwNode().CapabilitiesIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"qemu"}, subs)
}

func TestNode_QEMUCapabilitiesIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subs, err := hwNode().QEMUCapabilitiesIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"cpu", "cpu-flags", "machines", "migration"}, subs)
}

func TestNode_QEMUCPUModels(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	models, err := hwNode().QEMUCPUModels(context.Background(), "x86_64")
	assert.Nil(t, err)
	assert.Len(t, models, 3)
	assert.True(t, models[2].Custom)
	assert.True(t, models[1].Abstract)
}

func TestNode_QEMUCPUFlags(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	flags, err := hwNode().QEMUCPUFlags(context.Background(), "x86_64", "kvm")
	assert.Nil(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "aes", flags[0].Name)
	assert.Len(t, flags[0].SupportedOn, 2)
}

func TestNode_QEMUMachineTypes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	types, err := hwNode().QEMUMachineTypes(context.Background(), "")
	assert.Nil(t, err)
	assert.Len(t, types, 2)
	assert.Equal(t, "q35", types[0].Type)
	assert.Equal(t, "pveX backport", types[1].Changes)
}

func TestNode_QEMUMigrationCapabilities(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	caps, err := hwNode().QEMUMigrationCapabilities(context.Background())
	assert.Nil(t, err)
	assert.True(t, caps.HasDbusVMState)
}

// --- /nodes/{node}/hardware ------------------------------------------------

func TestNode_HardwareIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	types, err := hwNode().HardwareIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"pci", "usb"}, types)
}

func TestNode_ListPCIDevices(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	devices, err := hwNode().ListPCIDevices(context.Background(), &HardwarePCIOptions{
		ClassBlacklist: []string{"05", "06"},
		Terse:          false,
	})
	assert.Nil(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "0000:01:00.0", devices[0].ID)
	assert.True(t, devices[0].MdevCapable)
	assert.Equal(t, "node1", devices[0].Node)
}

func TestNode_ListUSBDevices(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	devices, err := hwNode().ListUSBDevices(context.Background())
	assert.Nil(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "Linux Foundation", devices[0].Manufacturer)
	assert.Equal(t, "5000", devices[1].Speed)
}

func TestNode_PCIDevice_IndexAndMdev(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	d := hwNode().PCIDevice("0000:01:00.0")
	assert.Equal(t, "node1", d.Node)
	assert.Equal(t, "0000:01:00.0", d.ID)

	methods, err := d.Index(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"mdev"}, methods)

	mdevs, err := d.Mdev(context.Background())
	assert.Nil(t, err)
	assert.Len(t, mdevs, 2)
	assert.Equal(t, "nvidia-256", mdevs[0].Type)
	assert.Equal(t, 2, mdevs[0].Available)

	empty := hwNode().PCIDevice("")
	_, err = empty.Index(context.Background())
	assert.NotNil(t, err)
	_, err = empty.Mdev(context.Background())
	assert.NotNil(t, err)
}
