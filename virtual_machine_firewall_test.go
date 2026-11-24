package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func vm100() *VirtualMachine {
	return &VirtualMachine{client: mockClient(), Node: "node1", VMID: 100}
}

func TestVirtualMachine_FirewallSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	// The endpoint is a directory index: it returns an array of child-link
	// objects, which the old Firewall method tried to decode into a single
	// *Firewall struct and always failed. FirewallSubdirs returns the list.
	subdirs, err := vm100().FirewallSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "rules")
	assert.Contains(t, subdirs, "aliases")
	assert.Contains(t, subdirs, "options")
}

func TestVirtualMachine_FirewallRule_Get(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	rule := vm100().FirewallRule(0)
	assert.NotNil(t, rule)
	err := rule.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "ACCEPT", rule.Action)
}

func TestVirtualMachine_FirewallLog(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := vm100().FirewallLog(context.Background(), 0, 50, 0, 0)
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, 0, entries[0].LineNum)
	assert.Contains(t, entries[0].Text, "block")
}

func TestVirtualMachine_FirewallRefs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	refs, err := vm100().FirewallRefs(context.Background(), "")
	assert.Nil(t, err)
	assert.Len(t, refs, 2)
}

func TestVirtualMachine_FirewallAliases(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	vm := vm100()
	ctx := context.Background()

	aliases, err := vm.GetFirewallAliases(ctx)
	assert.Nil(t, err)
	assert.Len(t, aliases, 1)
	assert.Equal(t, "internal", aliases[0].Name)

	alias, err := vm.GetFirewallAlias(ctx, "internal")
	assert.Nil(t, err)
	assert.Equal(t, "10.0.0.0/8", alias.Cidr)

	assert.Nil(t, vm.NewFirewallAlias(ctx, &FirewallAlias{Name: "internal", Cidr: "10.0.0.0/8"}))
	assert.Nil(t, vm.UpdateFirewallAlias(ctx, "internal", &FirewallAlias{Cidr: "10.0.0.0/16"}))
	assert.Nil(t, vm.DeleteFirewallAlias(ctx, "internal"))
}
