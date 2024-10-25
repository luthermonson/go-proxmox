package proxmox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

func (cl *Cluster) ListHAResources(ctx context.Context, resourceType HAResourceType) ([]HAResource, error) {
	var haResources []haResource
	if err := cl.client.Get(ctx, fmt.Sprintf("/cluster/ha/resources?type=%s", resourceType), &haResources); err != nil {
		return nil, err
	}

	return fillHAResources(haResources)
}

func fillHAResources(resources []haResource) ([]HAResource, error) {
	haResources := make([]HAResource, 0, len(resources))
	for _, resource := range resources {
		sid, err := parseSid(resource.Sid)
		if err != nil {
			return nil, fmt.Errorf("parse sid: %w", err)
		}

		haResources = append(haResources, HAResource{
			ID:          sid.ID,
			Group:       &resource.Group,
			Comment:     &resource.Comment,
			MaxRelocate: &resource.MaxRelocate,
			MaxRestart:  &resource.MaxRestart,
			State:       &resource.State,
		})
	}

	return haResources, nil
}

func parseSid(sid string) (SID, error) {
	sidParts := strings.Split(sid, ":")
	if len(sidParts) != 2 {
		return SID{}, errors.New("invalid sid parts count")
	}

	vmid, err := strconv.Atoi(sidParts[1])
	if err != nil {
		return SID{}, fmt.Errorf("parse sid vmid: %w", err)
	}

	return SID{
		Type: HAResourceType(sidParts[0]),
		ID:   vmid,
	}, nil
}
