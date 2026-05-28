package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestVNet_FirewallIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, err := cluster.SDNVNet(context.Background(), "user1")
	assert.Nil(t, err)
	assert.NotNil(t, v)

	entries, err := v.FirewallIndex(context.Background())
	assert.Nil(t, err)
	assert.Len(t, entries, 2)
}

func TestVNet_FirewallOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, _ := cluster.SDNVNet(context.Background(), "user1")
	opts, err := v.FirewallOptions(context.Background())
	assert.Nil(t, err)
	assert.True(t, bool(opts.Enable))
	assert.Equal(t, "ACCEPT", opts.PolicyForward)

	assert.Nil(t, v.FirewallOptionsUpdate(context.Background(), &SDNVNetFirewallOptionsUpdate{PolicyForward: "DROP"}))
}

func TestVNet_FirewallRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, _ := cluster.SDNVNet(context.Background(), "user1")
	rules, err := v.FirewallRules(context.Background())
	assert.Nil(t, err)
	assert.Len(t, rules, 2)

	rule, err := v.FirewallRule(context.Background(), 0)
	assert.Nil(t, err)
	assert.Equal(t, "ACCEPT", rule.Action)

	assert.Nil(t, v.NewFirewallRule(context.Background(), &SDNVNetFirewallRuleOptions{Type: "in", Action: "ACCEPT"}))
	assert.NotNil(t, v.NewFirewallRule(context.Background(), nil))

	assert.Nil(t, v.FirewallRuleUpdate(context.Background(), 0, &SDNVNetFirewallRuleOptions{Enable: 0}))
	assert.Nil(t, v.FirewallRuleDelete(context.Background(), 0))
}

func TestVNet_IPs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, _ := cluster.SDNVNet(context.Background(), "user1")
	assert.Nil(t, v.CreateIP(context.Background(), &SDNVNetIPOptions{Zone: "zone1", IP: "10.0.0.10"}))
	assert.Nil(t, v.UpdateIP(context.Background(), &SDNVNetIPOptions{Zone: "zone1", IP: "10.0.0.10", VMID: 100}))
	assert.Nil(t, v.DeleteIP(context.Background(), &SDNVNetIPOptions{Zone: "zone1", IP: "10.0.0.10"}))

	assert.NotNil(t, v.CreateIP(context.Background(), nil))
	assert.NotNil(t, v.CreateIP(context.Background(), &SDNVNetIPOptions{IP: "10.0.0.10"}))
}

func TestVNet_Subnets(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	v, _ := cluster.SDNVNet(context.Background(), "user1")
	subs, err := v.Subnets(context.Background())
	assert.Nil(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, "user1", subs[0].VNet)

	s := v.Subnet("zone1-10.0.0.0-24")
	assert.Nil(t, s.Read(context.Background()))
	assert.Equal(t, "10.0.0.0/24", s.CIDR)

	assert.Nil(t, v.NewSubnet(context.Background(), &SDNSubnetOptions{Subnet: "zone1-10.0.1.0-24"}))
	assert.NotNil(t, v.NewSubnet(context.Background(), nil))

	assert.Nil(t, s.Update(context.Background(), &SDNSubnetOptions{Gateway: "10.0.0.2"}))
	assert.Nil(t, s.Delete(context.Background()))
}
