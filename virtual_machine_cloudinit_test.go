package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVirtualMachine_CloudInitPending(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	pending, err := vm100().CloudInitPending(context.Background())
	assert.Nil(t, err)
	assert.Len(t, pending, 2)
	assert.Equal(t, "ipconfig0", pending[0].Key)
	assert.NotEqual(t, pending[0].Value, pending[0].Pending)
}

func TestVirtualMachine_CloudInitRegenerate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	assert.Nil(t, vm100().CloudInitRegenerate(context.Background()))
}

func TestVirtualMachine_CloudInitDump(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	out, err := vm100().CloudInitDump(context.Background(), "user")
	assert.Nil(t, err)
	assert.Contains(t, out, "cloud-config")

	_, err = vm100().CloudInitDump(context.Background(), "")
	assert.NotNil(t, err)
}
