package proxmox

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luthermonson/go-proxmox/tests/mocks"
)

func TestVirtualMachine_Ping(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	assert.Nil(t, vm.Ping())
	assert.Equal(t, StringOrUint64(101), vm.VMID)
}

func TestVirtualMachine_RRDData(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	vm := VirtualMachine{
		client: client,
		VMID:   101,
		Node:   "node1",
	}

	rdddata, err := vm.RRDData(HOUR)
	assert.Nil(t, err)
	assert.Len(t, rdddata, 70)
}
