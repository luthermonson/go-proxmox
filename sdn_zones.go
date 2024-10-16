package proxmox

import (
	"context"
	"fmt"
	"strconv"
)

type Zone struct {
	client                   *Client
	Zone                     string `json:"zone"`
	Type                     string `json:"type,omitempty"`
	AdvertiseSubnets         bool   `json:"advertise-subnets,omitempty"`
	Bridge                   string `json:"bridge,omitempty"`
	BridgeDisableMacLearning bool   `json:"bridge-disable-mac-learning,omitempty"`
	Controller               string `json:"controller,omitempty"`
	Dhcp                     string `json:"dhcp,omitempty"`
	DisableArpNdSuppression  bool   `json:"disable-arp-nd-suppression,omitempty"`
	Dns                      string `json:"dns,omitempty"`
	DnsZone                  string `json:"dnszone,omitempty"`
	DpId                     int    `json:"dp-id,omitempty"`
	Exitnodes                string `json:"exitnodes,omitempty"`
	ExitnodesLocalRouting    bool   `json:"exitnodes-local-routing,omitempty"`
	ExitnodesPrimary         string `json:"exitnodes-primary,omitempty"`
	Ipam                     string `json:"ipam,omitempty"`
	Mac                      string `json:"mac,omitempty"`
	Mtu                      int    `json:"mtu,omitempty"`
	Nodes                    string `json:"nodes,omitempty"`
	Peers                    string `json:"peers,omitempty"`
	Reversedns               string `json:"reversedns,omitempty"`
	RtImport                 string `json:"rt-import,omitempty"`
	Tag                      int    `json:"tag,omitempty"`
	VlanProtocol             string `json:"vlan-protocol,omitempty"`
	VrfVxlan                 int    `json:"vrf-vxlan,omitempty"`
	VxlanPort                int    `json:"vxlan-port,omitempty"`
	Digest                   string `json:"digest,omitempty"`
}

type ZoneStatus struct {
	Zone    string
	Pending bool
	Running bool
}

type ZoneOptions []*ZoneOption
type ZoneOption struct {
	Name  string
	Value interface{}
}

type ZoneConfig struct {
	Zone                     string `json:"zone"`
	Type                     string `json:"type,omitempty"`
	AdvertiseSubnets         bool   `json:"advertise-subnets,omitempty"`
	Bridge                   string `json:"bridge,omitempty"`
	BridgeDisableMacLearning bool   `json:"bridge-disable-mac-learning,omitempty"`
	Controller               string `json:"controller,omitempty"`
	Dhcp                     string `json:"dhcp,omitempty"`
	DisableArpNdSuppression  bool   `json:"disable-arp-nd-suppression,omitempty"`
	Dns                      string `json:"dns,omitempty"`
	DnsZone                  string `json:"dnszone,omitempty"`
	DpId                     int    `json:"dp-id,omitempty"`
	Exitnodes                string `json:"exitnodes,omitempty"`
	ExitnodesLocalRouting    bool   `json:"exitnodes-local-routing,omitempty"`
	ExitnodesPrimary         string `json:"exitnodes-primary,omitempty"`
	Ipam                     string `json:"ipam,omitempty"`
	Mac                      string `json:"mac,omitempty"`
	Mtu                      int    `json:"mtu,omitempty"`
	Nodes                    string `json:"nodes,omitempty"`
	Peers                    string `json:"peers,omitempty"`
	Reversedns               string `json:"reversedns,omitempty"`
	RtImport                 string `json:"rt-import,omitempty"`
	Tag                      int    `json:"tag,omitempty"`
	VlanProtocol             string `json:"vlan-protocol,omitempty"`
	VrfVxlan                 int    `json:"vrf-vxlan,omitempty"`
	VxlanPort                int    `json:"vxlan-port,omitempty"`
}

func (c *Client) Zones(ctx context.Context, pending bool, running bool) (zones []ZoneStatus, err error) {
	pendingStr := strconv.FormatBool(pending)
	runningStr := strconv.FormatBool(running)
	return zones, c.Get(ctx, fmt.Sprintf("/cluster/sdn/zones?pending=%v&running=%v", pendingStr, runningStr), &zones)
}

func (c *Client) Zone(ctx context.Context, name string) (zone *Zone, err error) {

	if err = c.Get(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", name), &zone); err != nil {
		return nil, err
	}

	zone.client = c

	return
}

func (c *Client) NewZone(ctx context.Context, config ZoneConfig) (zone *Zone, err error) {

	if err = c.Post(ctx, "/cluster/sdn/zones", config, nil); err != nil {
		return nil, err
	}

	return c.Zone(ctx, config.Zone)
}

func (z *Zone) Update(ctx context.Context, config ZoneConfig) error {
	return z.client.Put(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", z.Zone), config, nil)
}

func (z *Zone) Delete(ctx context.Context) error {
	return z.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", z.Zone), nil)
}
