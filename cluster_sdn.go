package proxmox

import (
	"context"
	"fmt"
)

func (cl *Cluster) SDNSubnets(ctx context.Context, VnetName string) (subnets []*VnetSubnet, err error) {
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", VnetName), &subnets)

	return
}
