package proxmox

import (
	"context"
	"fmt"
)

func (cl *Cluster) SDNSubnets(ctx context.Context, VNetName string) (subnets []*VNetSubnet, err error) {
	err = cl.client.Get(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", VNetName), &subnets)

	return
}
