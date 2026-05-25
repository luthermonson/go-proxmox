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

func TestNode_CephOSD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	items, err := cephOSDNode().CephOSD(context.Background(), 0)
	assert.Nil(t, err)
	assert.Len(t, items, 5)
	assert.Equal(t, "metadata", items[0]["name"])
}

func TestNode_DeleteCephOSD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	task, err := cephOSDNode().DeleteCephOSD(context.Background(), 0, true)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroyosd", task.Type)

	task, err = cephOSDNode().DeleteCephOSD(context.Background(), 0, false)
	assert.Nil(t, err)
	assert.Equal(t, "cephdestroyosd", task.Type)
}

func TestNode_CephOSDIn(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSDIn(context.Background(), 0)
	assert.Nil(t, err)
}

func TestNode_CephOSDOut(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSDOut(context.Background(), 0)
	assert.Nil(t, err)
}

func TestNode_CephOSDScrub(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := cephOSDNode().CephOSDScrub(context.Background(), 0, false)
	assert.Nil(t, err)

	err = cephOSDNode().CephOSDScrub(context.Background(), 0, true)
	assert.Nil(t, err)
}

func TestNode_CephOSDLVInfo(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	info, err := cephOSDNode().CephOSDLVInfo(context.Background(), 0, "")
	assert.Nil(t, err)
	assert.Equal(t, "osd-block-abcd1234", info.LVName)
	assert.Equal(t, "ceph-vg", info.VGName)

	info, err = cephOSDNode().CephOSDLVInfo(context.Background(), 0, "db")
	assert.Nil(t, err)
	assert.NotNil(t, info)
}

func TestNode_CephOSDMetadata(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	details, err := cephOSDNode().CephOSDMetadata(context.Background(), 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, details.OSD.ID)
	assert.Equal(t, "node1", details.OSD.Hostname)
	assert.Equal(t, "bluestore", details.OSD.OSDObjectStore)
	assert.Len(t, details.Devices, 1)
	assert.Equal(t, "block", details.Devices[0].Device)
}
