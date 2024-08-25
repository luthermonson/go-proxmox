package proxmox

import (
	"context"
	"fmt"
)

type Zone struct {
	client *Client
	Zone   string
	Type   string
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

func (c *Client) Zones(ctx context.Context) (zones []ZoneStatus, err error) {
	return zones, c.Get(ctx, "/cluster/sdn/zones", &zones)
}

func (c *Client) Zone(ctx context.Context, name string) (zone *Zone, err error) {

	if err = c.Get(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", name), &zone); err != nil {
		return nil, err
	}

	zone.client = c

	return
}

func (c *Client) NewZone(ctx context.Context, name string, options ...ZoneOption) (ret string, err error) {
	data := make(map[string]interface{})
	data["zone"] = name

	for _, option := range options {
		data[option.Name] = option.Value
	}

	err = c.Post(ctx, "/cluster/sdn/zones", data, &ret)
	return ret, err
}

func (z *Zone) Delete(ctx context.Context) error {
	return z.client.Delete(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", z.Zone), nil)
}
