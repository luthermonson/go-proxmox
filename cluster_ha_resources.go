package proxmox

import (
	"context"
	"fmt"
)

func (cl *Cluster) CreateHAResource(ctx context.Context, haResource HAResource) error {
	if err := cl.client.Post(ctx, "/cluster/ha/resources", &haResource, nil); err != nil {
		return err
	}

	return nil
}

func (cl *Cluster) DeleteHAResource(ctx context.Context, sid SID) error {
	if err := cl.client.Delete(ctx, fmt.Sprintf("/cluster/ha/resources/%s:%d", sid.Type, sid.ID), nil); err != nil {
		return err
	}

	return nil
}

func (cl *Cluster) GetHAResource(ctx context.Context, sid SID) (HAResource, error) {
	var haResource HAResource
	if err := cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/resources/%s:%d", sid.Type, sid.ID), &haResource); err != nil {
		return haResource, err
	}

	return haResource, nil
}

func (cl *Cluster) ListHAResources(ctx context.Context, resourceType HAResourceType) ([]SID, error) {
	var haResources []SID
	if err := cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/resources?type=%s", resourceType), &haResources); err != nil {
		return nil, err
	}

	return haResources, nil
}
