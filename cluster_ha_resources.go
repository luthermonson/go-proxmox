package proxmox

import (
	"context"
)

func (cl *Cluster) HAResourceCreate(ctx context.Context, haResource HAResource) error {
	if err := cl.client.Post(ctx, "/cluster/ha/resources", &haResource, nil); err != nil {
		return err
	}

	return nil
}
