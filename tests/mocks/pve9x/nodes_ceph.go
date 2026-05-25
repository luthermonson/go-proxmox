package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

// nodesCeph wires gock fixtures for the per-node /nodes/{node}/ceph/* family.
// Today this only covers /pool/* — extend here when more node-level ceph
// endpoints get Go wrappers.
func nodesCeph() {
	// GET /nodes/node1/ceph/pool - list pools
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/ceph/pool$").
		Reply(200).
		JSON(`{"data": [
			{
				"pool": 1,
				"pool_name": "rbd",
				"type": "replicated",
				"size": 3,
				"min_size": 2,
				"pg_num": 128,
				"pg_num_min": 32,
				"pg_num_final": 128,
				"pg_autoscale_mode": "on",
				"crush_rule": 0,
				"crush_rule_name": "replicated_rule",
				"percent_used": 0.12,
				"bytes_used": 12884901888,
				"target_size": 0,
				"target_size_ratio": 0.0,
				"application_metadata": {"rbd": {}},
				"autoscale_status": {"would_adjust": false}
			},
			{
				"pool": 2,
				"pool_name": "cephfs_metadata",
				"type": "replicated",
				"size": 3,
				"min_size": 2,
				"pg_num": 32,
				"crush_rule": 0,
				"crush_rule_name": "replicated_rule"
			}
		]}`)

	// POST /nodes/node1/ceph/pool - create pool (returns UPID)
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/ceph/pool$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:cephcreatepool:rbd:root@pam:"}`)

	// GET /nodes/node1/ceph/pool/{name} - sub-resource directory index
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/ceph/pool/rbd$").
		Reply(200).
		JSON(`{"data": [{"subdir": "status"}]}`)

	// PUT /nodes/node1/ceph/pool/{name} - update pool (returns UPID)
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/ceph/pool/rbd$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:cephsetpool:rbd:root@pam:"}`)

	// DELETE /nodes/node1/ceph/pool/{name} - destroy pool (returns UPID)
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/ceph/pool/rbd").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:cephdestroypool:rbd:root@pam:"}`)

	// GET /nodes/node1/ceph/pool/{name}/status - pool status
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/ceph/pool/rbd/status$").
		Reply(200).
		JSON(`{"data": {
			"id": 1,
			"name": "rbd",
			"size": 3,
			"min_size": 2,
			"pg_num": 128,
			"pg_num_min": 32,
			"pgp_num": 128,
			"crush_rule": "replicated_rule",
			"pg_autoscale_mode": "on",
			"application": "rbd",
			"application_list": ["rbd"],
			"fast_read": false,
			"hashpspool": true,
			"nodeep-scrub": false,
			"nodelete": false,
			"nopgchange": false,
			"noscrub": false,
			"nosizechange": false,
			"use_gmt_hitset": true,
			"write_fadvise_dontneed": false,
			"target_size": "0",
			"target_size_ratio": 0.0,
			"autoscale_status": {"would_adjust": false},
			"statistics": {"stored": 12884901888, "objects": 4096}
		}}`)
}
