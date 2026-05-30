package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func diskNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_Disks(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	disks, err := diskNode().Disks(context.Background(), false, false, "")
	assert.Nil(t, err)
	assert.Len(t, disks, 2)
	assert.Equal(t, "/dev/sda", disks[0].DevPath)
	assert.Equal(t, "ssd", disks[0].Type)
}

func TestNode_DiskSMART(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	smart, err := diskNode().DiskSMART(context.Background(), "/dev/sda", false)
	assert.Nil(t, err)
	assert.Equal(t, "PASSED", smart.Health)

	_, err = diskNode().DiskSMART(context.Background(), "", false)
	assert.NotNil(t, err)
}

func TestNode_DiskInitGPT(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DiskInitGPT(context.Background(), "/dev/sda", "")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "diskinit", task.Type)

	_, err = diskNode().DiskInitGPT(context.Background(), "", "")
	assert.NotNil(t, err)
}

func TestNode_DiskWipe(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DiskWipe(context.Background(), "/dev/sda")
	assert.Nil(t, err)
	assert.Equal(t, "wipedisk", task.Type)
}

func TestNode_Directories(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	dirs, err := diskNode().Directories(context.Background())
	assert.Nil(t, err)
	assert.Len(t, dirs, 1)
	assert.Equal(t, "/mnt/pve/iso", dirs[0].Path)
}

func TestNode_NewDirectory(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().NewDirectory(context.Background(), &NodeDirectoryOptions{Name: "iso", Device: "/dev/sda1"})
	assert.Nil(t, err)
	assert.Equal(t, "mkdir", task.Type)

	_, err = diskNode().NewDirectory(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestNode_DeleteDirectory(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DeleteDirectory(context.Background(), "iso", true, true)
	assert.Nil(t, err)
	assert.Equal(t, "rmdir", task.Type)

	_, err = diskNode().DeleteDirectory(context.Background(), "", false, false)
	assert.NotNil(t, err)
}

func TestNode_LVMs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	lvm, err := diskNode().LVMs(context.Background())
	assert.Nil(t, err)
	assert.NotEmpty(t, lvm.Children)
	assert.Equal(t, "pve", lvm.Children[0].Name)
}

func TestNode_NewLVM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().NewLVM(context.Background(), &NodeLVMOptions{Name: "pve", Device: "/dev/sda"})
	assert.Nil(t, err)
	assert.Equal(t, "mklvm", task.Type)
}

func TestNode_DeleteLVM(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DeleteLVM(context.Background(), "pve", false, false)
	assert.Nil(t, err)
	assert.Equal(t, "rmlvm", task.Type)
}

func TestNode_LVMThins(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	thins, err := diskNode().LVMThins(context.Background())
	assert.Nil(t, err)
	assert.Len(t, thins, 1)
	assert.Equal(t, "data", thins[0].LV)
}

func TestNode_NewLVMThin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().NewLVMThin(context.Background(), &NodeLVMThinOptions{Name: "data", Device: "/dev/sdb"})
	assert.Nil(t, err)
	assert.Equal(t, "mklvmthin", task.Type)
}

func TestNode_DeleteLVMThin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DeleteLVMThin(context.Background(), "data", "pve", true, false)
	assert.Nil(t, err)
	assert.Equal(t, "rmlvmthin", task.Type)

	_, err = diskNode().DeleteLVMThin(context.Background(), "data", "", false, false)
	assert.NotNil(t, err)
}

func TestNode_ZFSPools(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pools, err := diskNode().ZFSPools(context.Background())
	assert.Nil(t, err)
	assert.Len(t, pools, 1)
	assert.Equal(t, "rpool", pools[0].Name)
}

func TestNode_ZFSPool(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pool, err := diskNode().ZFSPool(context.Background(), "rpool")
	assert.Nil(t, err)
	assert.Equal(t, "ONLINE", pool.State)
	assert.NotEmpty(t, pool.Children)
}

func TestNode_NewZFSPool(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().NewZFSPool(context.Background(), &NodeZFSPoolOptions{Name: "rpool", Devices: "/dev/sda /dev/sdb", RaidLevel: "mirror"})
	assert.Nil(t, err)
	assert.Equal(t, "mkzfs", task.Type)

	_, err = diskNode().NewZFSPool(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestNode_DeleteZFSPool(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DeleteZFSPool(context.Background(), "rpool", true, true)
	assert.Nil(t, err)
	assert.Equal(t, "rmzfs", task.Type)

	_, err = diskNode().DeleteZFSPool(context.Background(), "", false, false)
	assert.NotNil(t, err)
}

func TestNode_Disks_AllQueryBranches(t *testing.T) {
	// Hit every conditional in the query-string builder.
	mocks.On(mockConfig)
	defer mocks.Off()

	_, err := diskNode().Disks(context.Background(), true, true, "unused")
	assert.Nil(t, err)
}

func TestNode_DiskSMART_HealthOnly(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	_, err := diskNode().DiskSMART(context.Background(), "/dev/sda", true)
	assert.Nil(t, err)
}

func TestNode_DiskInitGPT_WithUUID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DiskInitGPT(context.Background(), "/dev/sda",
		"01234567-89ab-cdef-0123-456789abcdef")
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestNode_DiskWipe_Validation(t *testing.T) {
	_, err := diskNode().DiskWipe(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_NewLVM_Validation(t *testing.T) {
	_, err := diskNode().NewLVM(context.Background(), nil)
	assert.NotNil(t, err)
	_, err = diskNode().NewLVM(context.Background(), &NodeLVMOptions{Name: "x"})
	assert.NotNil(t, err)
}

func TestNode_DeleteLVM_Validation(t *testing.T) {
	_, err := diskNode().DeleteLVM(context.Background(), "", false, false)
	assert.NotNil(t, err)
}

func TestNode_DeleteDirectory_QueryBranches(t *testing.T) {
	// cleanupConfig but no cleanupDisks (and a query-string branch already
	// covered by the existing TestNode_DeleteDirectory, which sets both).
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := diskNode().DeleteDirectory(context.Background(), "iso", false, true)
	assert.Nil(t, err)
	assert.NotNil(t, task)
}

func TestNode_NewLVMThin_Validation(t *testing.T) {
	_, err := diskNode().NewLVMThin(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestNode_DeleteLVMThin_Validation(t *testing.T) {
	_, err := diskNode().DeleteLVMThin(context.Background(), "", "", false, false)
	assert.NotNil(t, err)
}

func TestNode_ZFSPool_Validation(t *testing.T) {
	_, err := diskNode().ZFSPool(context.Background(), "")
	assert.NotNil(t, err)
}

func TestNode_NewZFSPool_PartialValidation(t *testing.T) {
	_, err := diskNode().NewZFSPool(context.Background(), &NodeZFSPoolOptions{Name: "n"})
	assert.NotNil(t, err)
}
