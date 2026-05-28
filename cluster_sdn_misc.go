package proxmox

import (
	"context"
	"errors"
	"net/url"
)

// SDNIndex returns the directory entries under /cluster/sdn.
//
// GET /cluster/sdn
func (cl *Cluster) SDNIndex(ctx context.Context) (entries []map[string]any, err error) {
	err = cl.client.Get(ctx, "/cluster/sdn", &entries)
	return
}

// SDNLock acquires the global SDN configuration lock. The returned token must
// be passed as LockToken on subsequent mutating SDN endpoints (controllers,
// fabrics, IPAMs, etc.) and on SDNRollback/SDNReleaseLock.
//
// allowPending lets the lock be acquired even if there are pending changes.
//
// POST /cluster/sdn/lock
func (cl *Cluster) SDNLock(ctx context.Context, allowPending bool) (SDNLockToken, error) {
	body := map[string]any{}
	if allowPending {
		body["allow-pending"] = 1
	}
	var token string
	if err := cl.client.Post(ctx, "/cluster/sdn/lock", body, &token); err != nil {
		return "", err
	}
	return SDNLockToken(token), nil
}

// SDNReleaseLock releases the global SDN configuration lock. Pass force=true
// to release without providing the matching token (admin override).
//
// DELETE /cluster/sdn/lock
func (cl *Cluster) SDNReleaseLock(ctx context.Context, token SDNLockToken, force bool) error {
	path := "/cluster/sdn/lock"
	q := url.Values{}
	if token != "" {
		q.Set("lock-token", string(token))
	}
	if force {
		q.Set("force", "1")
	}
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	return cl.client.Delete(ctx, path, nil)
}

// SDNRollback discards pending SDN configuration changes. token may be empty
// when no lock is held. releaseLock controls whether the lock is released on
// success (PVE default: true).
//
// POST /cluster/sdn/rollback
func (cl *Cluster) SDNRollback(ctx context.Context, token SDNLockToken, releaseLock bool) error {
	body := map[string]any{}
	if token != "" {
		body["lock-token"] = string(token)
	}
	if !releaseLock {
		// PVE default is release-lock=1; only emit the override.
		body["release-lock"] = 0
	}
	return cl.client.Post(ctx, "/cluster/sdn/rollback", body, nil)
}

// SDNDryRun returns the diff (FRR + /etc/network/interfaces.d/sdn) between the
// current and pending SDN configuration on a specific node.
//
// GET /cluster/sdn/dry-run
func (cl *Cluster) SDNDryRun(ctx context.Context, node string) (*SDNDryRun, error) {
	if node == "" {
		return nil, errors.New("node is required")
	}
	q := url.Values{}
	q.Set("node", node)
	out := &SDNDryRun{}
	if err := cl.client.Get(ctx, "/cluster/sdn/dry-run?"+q.Encode(), out); err != nil {
		return nil, err
	}
	return out, nil
}
