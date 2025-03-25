package proxmox

import "context"

func (ce *Ceph) Status(ctx context.Context) (*ClusterCephStatus, error) {
	cephStatus := &ClusterCephStatus{}

	if err := ce.client.Get(ctx, "/cluster/ceph/status", cephStatus); !IsNotAuthorized(err) {
		return cephStatus, err
	}

	return cephStatus, nil
}
