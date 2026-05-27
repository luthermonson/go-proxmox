package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func cephOSDNode() *Node {
	return &Node{client: mockClient(), Name: "node1"}
}

func TestNode_CephOSDs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	tree, err := cephOSDNode().CephOSDs(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, "noout", tree.Flags)
	assert.Equal(t, "default", tree.Root["name"])
}

func TestNode_CreateCephOSD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephOSDNode().CreateCephOSD(context.Background(), &CephOSDCreateOptions{Dev: "/dev/sda"})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "cephcreateosd", task.Type)

	_, err = cephOSDNode().CreateCephOSD(context.Background(), nil)
	assert.NotNil(t, err)

	_, err = cephOSDNode().CreateCephOSD(context.Background(), &CephOSDCreateOptions{})
	assert.NotNil(t, err)
}

func TestCephOSD_SubResources(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	items, err := cephOSDNode().CephOSD(0).SubResources(context.Background())
	assert.Nil(t, err)
	assert.Len(t, items, 5)
	assert.Equal(t, "metadata", items[0]["name"])
}

func TestCephOSD_Delete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephOSDNode().CephOSD(0).Delete(context.Background(), true)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroyosd", task.Type)

	task, err = cephOSDNode().CephOSD(0).Delete(context.Background(), false)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroyosd", task.Type)
}

func TestCephOSD_In(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSD(0).In(context.Background())
	assert.Nil(t, err)
}

func TestCephOSD_Out(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSD(0).Out(context.Background())
	assert.Nil(t, err)
}

func TestCephOSD_Scrub(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSD(0).Scrub(context.Background(), false)
	assert.Nil(t, err)

	err = cephOSDNode().CephOSD(0).Scrub(context.Background(), true)
	assert.Nil(t, err)
}

func TestCephOSD_LVInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	info, err := cephOSDNode().CephOSD(0).LVInfo(context.Background(), "")
	assert.Nil(t, err)
	assert.Equal(t, "osd-block-abcd1234", info.LVName)
	assert.Equal(t, "ceph-vg", info.VGName)

	info, err = cephOSDNode().CephOSD(0).LVInfo(context.Background(), "db")
	assert.Nil(t, err)
	assert.NotNil(t, info)
}

func TestCephOSD_Metadata(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	details, err := cephOSDNode().CephOSD(0).Metadata(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 0, details.OSD.ID)
	assert.Equal(t, "node1", details.OSD.Hostname)
	assert.Equal(t, "bluestore", details.OSD.OSDObjectStore)
	assert.Len(t, details.Devices, 1)
	assert.Equal(t, "block", details.Devices[0].Device)
}
