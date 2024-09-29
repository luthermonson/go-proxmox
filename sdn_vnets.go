package proxmox

import (
	"context"
	"fmt"
)

type Vnet struct {
	client    *Client
	Vnet      string
	Zone      string
	Alias     string
	Tag       int
	VlanAware bool
}

type VnetStatus struct {
	client  *Client
	Vnet    string
	Pending bool
	Running bool
}

type VnetOptions []*VnetOption
type VnetOption struct {
	Name  string
	Value interface{}
}

func (c *Client) Vnets(ctx context.Context, pending bool, running bool) (vnets []VnetStatus, err error) {
	return vnets, c.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets?pending=%s&running=%s", pending, running), &vnets)
}

func (c *Client) Vnet(ctx context.Context, name string) (vnet *VnetStatus, err error) {

	if err = c.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", name), &vnet); err != nil {
		return nil, err
	}

	vnet.client = c

	return
}

func (c *Client) NewVnet(ctx context.Context, vnet string, zone string, options ...VnetOption) (ret string, err error) {
	data := make(map[string]interface{})
	data["vnet"] = vnet
	data["zone"] = zone

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err = c.Post(ctx, "/cluster/sdn/vnets", data, &ret)
	return ret, err
}

func (z *Vnet) Update(ctx context.Context, options ...VnetOption) error {
	return z.client.Put(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", z.Vnet), nil, nil)
}

func (z *Vnet) Delete(ctx context.Context) error {
	return z.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", z.Vnet), nil)
}
