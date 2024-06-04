package proxmox

import (
	"context"
	"fmt"
)

func (cl *Cluster) HAResourceCreate(ctx context.Context, haResource HAResource) error {
	if err := cl.client.Post(ctx, "/cluster/ha/resources", &haResource, nil); err != nil {
		return err
	}

	return nil
}

func (cl *Cluster) HAResourceDelete(ctx context.Context, sid SID) error {
	if err := cl.client.Delete(ctx, fmt.Sprintf("/cluster/ha/resources/%s:%d", sid.Type, sid.ID), nil); err != nil {
		return err
	}

	return nil
}
