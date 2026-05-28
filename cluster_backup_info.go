package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Wrappers for /cluster/backup-info + /cluster/backup/{id}/included_volumes —
// cluster-wide backup coverage audit helpers.

// BackupInfoSubdirs enumerates the children of /cluster/backup-info
// ("not-backed-up" today). ACL-filtered.
//
// GET /cluster/backup-info
func (cl *Cluster) BackupInfoSubdirs(ctx context.Context) ([]string, error) {
	return cl.backupInfoDiridx(ctx, "/cluster/backup-info")
}

// GuestsNotInBackup returns guests that aren't covered by any backup job.
//
// GET /cluster/backup-info/not-backed-up
func (cl *Cluster) GuestsNotInBackup(ctx context.Context) (guests []*BackupGuestEntry, err error) {
	err = cl.client.Get(ctx, "/cluster/backup-info/not-backed-up", &guests)
	return
}

// IncludedVolumes returns a tree of guests and their volumes' inclusion
// status for the given backup job ID.
//
// GET /cluster/backup/{id}/included_volumes
func (b *ClusterBackup) IncludedVolumes(ctx context.Context) (root *BackupIncludedVolumesRoot, err error) {
	if b.ID == "" {
		return nil, errors.New("backup id can not be empty")
	}
	root = &BackupIncludedVolumesRoot{}
	err = b.client.Get(ctx, fmt.Sprintf("/cluster/backup/%s/included_volumes", b.ID), root)
	return
}
