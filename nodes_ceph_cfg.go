package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// CephCfg returns the directory index for /nodes/{node}/ceph/cfg. PVE
// exposes no useful payload on the index itself; the real data lives on the
// child endpoints (db, raw, value).
func (n *Node) CephCfg(ctx context.Context) (entries []map[string]interface{}, err error) {
	return entries, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/cfg", n.Name), &entries)
}

// CephCfgDB returns the Ceph mon config database — every key/value the
// cluster has stored in the centralised KV config, scoped by section and
// optional mask.
func (n *Node) CephCfgDB(ctx context.Context) (entries []*CephCfgDBEntry, err error) {
	return entries, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/cfg/db", n.Name), &entries)
}

// CephCfgRaw returns the raw text contents of /etc/pve/ceph.conf as a
// single string — exactly what the on-disk file looks like.
func (n *Node) CephCfgRaw(ctx context.Context) (raw string, err error) {
	return raw, n.client.Get(ctx, fmt.Sprintf("/nodes/%s/ceph/cfg/raw", n.Name), &raw)
}

// CephCfgValue resolves specific config values from either ceph.conf or the
// mon config DB. configKeys is a string of `<section>:<key>` items separated
// by `;`, `,`, or space (e.g. "global:auth_cluster_required;osd:osd_pool_default_size").
// PVE normalises underscores in section and key names to hyphens in the
// response, so callers should not key off the requested spelling.
func (n *Node) CephCfgValue(ctx context.Context, configKeys string) (values CephCfgValue, err error) {
	if configKeys == "" {
		return nil, errors.New("config-keys is required")
	}
	q := url.Values{}
	q.Set("config-keys", configKeys)
	path := fmt.Sprintf("/nodes/%s/ceph/cfg/value?%s", n.Name, q.Encode())
	return values, n.client.Get(ctx, path, &values)
}
