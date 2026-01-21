package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (cl *Cluster) SDNSubnets(ctx context.Context, VNetName string) (subnets []*VNetSubnet, err error) {
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", VNetName), &subnets)

	return
}

func (cl *Cluster) SDNApply(ctx context.Context) (*Task, error) {
	var upid UPID
	err := cl.client.Put(ctx, "/cluster/sdn/", nil, &upid)
	return NewTask(upid, cl.client), err
}

func (cl *Cluster) SDNVNets(ctx context.Context) (vnets []*VNet, err error) {
	return vnets, cl.client.Get(ctx, "/cluster/sdn/vnets", &vnets)
}

func (cl *Cluster) SDNVNet(ctx context.Context, name string) (vnet *VNet, err error) {
	return vnet, cl.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", name), &vnet)
}

func (cl *Cluster) NewSDNVNet(ctx context.Context, vnet *VNetOptions) error {
	return cl.client.Post(ctx, "/cluster/sdn/vnets", vnet, nil)
}

func (cl *Cluster) UpdateSDNVNet(ctx context.Context, vnet *VNet) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", vnet.Name), vnet, nil)
}

func (cl *Cluster) DeleteSDNVNet(ctx context.Context, name string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", name), nil)
}

func (cl *Cluster) SDNZones(ctx context.Context, filters ...string) (zones []*SDNZone, err error) {
	u := url.URL{Path: "/cluster/sdn/zones"}

	// filters are variadic because they're optional, munging everything passed into one big string to make
	// a good request and the api will error out if there's an issue
	if f := strings.ReplaceAll(strings.Join(filters, ""), " ", ""); f != "" {
		params := url.Values{}
		params.Add("type", f)
		u.RawQuery = params.Encode()
	}

	return zones, cl.client.Get(ctx, u.String(), &zones)
}

func (cl *Cluster) SDNZone(ctx context.Context, name string) (zone *SDNZone, err error) {
	return zone, cl.client.Get(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", name), &zone)
}

func (cl *Cluster) NewSDNZone(ctx context.Context, zone *SDNZoneOptions) error {
	return cl.client.Post(ctx, "/cluster/sdn/zones", zone, nil)
}

func (cl *Cluster) UpdateSDNZone(ctx context.Context, zone *SDNZoneOptions) error {
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", zone.Name), zone, nil)
}

func (cl *Cluster) DeleteSDNZone(ctx context.Context, name string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", name), nil)
}
