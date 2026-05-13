package pve8x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func cluster() {
	gock.New(config.C.URI).
		Get("/cluster/nextid$").
		Reply(200).
		JSON(`{"data": "100"}`)

	gock.New(config.C.URI).
		Get("/cluster/nextid$").
		MatchParam("vmid", "100").
		Reply(200).
		JSON(`{"data": "100"}`)

	gock.New(config.C.URI).
		Get("/cluster/nextid").
		MatchParam("vmid", "200").
		Reply(400).
		JSON(`{"errors":{"vmid":"VM 200 already exists"},"data":null}`)

	gock.New(config.C.URI).
		Get("/cluster/status").
		Reply(200).
		JSON(`{
    "data": [
        {
            "type": "cluster",
            "version": 4,
            "quorate": 1,
            "name": "clustername",
            "id": "cluster",
            "nodes": 4
        },
        {
            "name": "node2",
            "nodeid": 2,
            "id": "node/node2",
            "online": 1,
            "type": "node",
            "ip": "192.168.1.2",
            "local": 0,
            "level": ""
        },
        {
            "name": "node3",
            "nodeid": 3,
            "type": "node",
            "ip": "192.168.1.3",
            "local": 0,
            "id": "node/node3",
            "online": 1,
            "level": ""
        },
        {
            "name": "node1",
            "nodeid": 1,
            "online": 1,
            "id": "node/node1",
            "local": 1,
            "ip": "192.168.1.1",
            "type": "node",
            "level": ""
        },
        {
            "nodeid": 4,
            "name": "node4",
            "level": "",
            "local": 0,
            "type": "node",
            "ip": "192.168.1.4",
            "online": 1,
            "id": "node/node4"
        }
    ]
}`)

	gock.New(config.C.URI).
		Get("^/cluster/resources$").
		MatchParams(map[string]string{
			"type": "node",
		}).
		Reply(200).
		JSON(`{
    "data": [
        {
            "type": "node",
			"id": "node1"
		}
	]
}`)

	gock.New(config.C.URI).
		Get("^/cluster/resources$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "netout": 545248946,
            "type": "qemu",
            "name": "server1",
            "maxcpu": 1,
            "mem": 842551296,
            "netin": 2456116121,
            "maxmem": 1073741824,
            "disk": 0,
            "node": "node2",
            "cpu": 0.0249060195469461,
            "maxdisk": 34359738368,
            "vmid": 100,
            "diskwrite": 6059209728,
            "diskread": 4510777856,
            "status": "running",
            "template": 0,
            "id": "qemu/100",
            "uptime": 874350
        },
        {
            "id": "qemu/101",
            "diskwrite": 0,
            "status": "stopped",
            "diskread": 0,
            "template": 1,
            "uptime": 0,
            "name": "leap154",
            "maxcpu": 4,
            "mem": 0,
            "netin": 0,
            "maxmem": 16777216000,
            "netout": 0,
            "type": "qemu",
            "maxdisk": 68719476736,
            "vmid": 101,
            "disk": 0,
            "node": "node1",
            "cpu": 0
        },
        {
            "netout": 0,
            "type": "qemu",
            "maxcpu": 4,
            "name": "machine-test",
            "maxmem": 8388608000,
            "netin": 0,
            "mem": 0,
            "node": "node1",
            "disk": 0,
            "cpu": 0,
            "maxdisk": 53901000704,
            "vmid": 102,
            "status": "stopped",
            "diskwrite": 0,
            "diskread": 0,
            "template": 0,
            "id": "qemu/102",
            "uptime": 0,
            "tags": "go-proxmox+cloud-init"
        },
        {
            "type": "qemu",
            "netout": 0,
            "netin": 0,
            "mem": 0,
            "maxmem": 8388608000,
            "maxcpu": 4,
            "name": "VM 200",
            "cpu": 0,
            "node": "node1",
            "disk": 0,
            "vmid": 200,
            "maxdisk": 53901000704,
            "template": 0,
            "diskwrite": 0,
            "diskread": 0,
            "status": "stopped",
            "id": "qemu/200",
            "uptime": 0
        },
        {
            "maxdisk": 482713534464,
            "cpu": 0.0054917623564653,
            "node": "node3",
            "disk": 2983723008,
            "maxmem": 16668827648,
            "mem": 1681965056,
            "maxcpu": 4,
            "type": "node",
            "uptime": 872961,
            "level": "",
            "id": "node/node3",
            "cgroup-mode": 2,
            "status": "online"
        },
        {
            "cgroup-mode": 2,
            "status": "online",
            "id": "node/node2",
            "level": "",
            "uptime": 874373,
            "type": "node",
            "mem": 8127873024,
            "maxmem": 33567911936,
            "maxcpu": 12,
            "cpu": 0.00365387809333998,
            "node": "node2",
            "disk": 2797338624,
            "maxdisk": 940166742016
        },
        {
            "status": "online",
            "cgroup-mode": 2,
            "id": "node/node1",
            "level": "",
            "uptime": 872854,
            "type": "node",
            "maxcpu": 8,
            "mem": 2113265664,
            "maxmem": 65919459328,
            "disk": 10486546432,
            "node": "node1",
            "cpu": 0.00336910406788121,
            "maxdisk": 951055941632
        },
        {
            "cgroup-mode": 2,
            "status": "online",
            "id": "node/node4",
            "level": "",
            "uptime": 872920,
            "type": "node",
            "mem": 1698938880,
            "maxmem": 16651702272,
            "maxcpu": 4,
            "cpu": 0.00724094881398252,
            "disk": 2789867520,
            "node": "node4",
            "maxdisk": 482719825920
        },
        {
            "shared": 0,
            "type": "storage",
            "status": "available",
            "plugintype": "zfspool",
            "id": "storage/node3/local-zfs",
            "storage": "local-zfs",
            "node": "node3",
            "disk": 98304,
            "content": "images,rootdir",
            "maxdisk": 479730032640
        },
        {
            "shared": 0,
            "type": "storage",
            "status": "available",
            "id": "storage/node2/local-zfs",
            "plugintype": "zfspool",
            "content": "images,rootdir",
            "disk": 25294921728,
            "storage": "local-zfs",
            "node": "node2",
            "maxdisk": 962664386560
        },
        {
            "maxdisk": 955016175616,
            "node": "node1",
            "storage": "local-zfs",
            "content": "images,rootdir",
            "disk": 14446702592,
            "id": "storage/node1/local-zfs",
            "plugintype": "zfspool",
            "status": "available",
            "shared": 0,
            "type": "storage"
        },
        {
            "type": "storage",
            "shared": 0,
            "status": "available",
            "plugintype": "zfspool",
            "id": "storage/node4/local-zfs",
            "content": "images,rootdir",
            "disk": 98304,
            "storage": "local-zfs",
            "node": "node4",
            "maxdisk": 479930105856
        },
        {
            "maxdisk": 482713534464,
            "content": "backup,vztmpl,iso",
            "disk": 2983723008,
            "storage": "local",
            "node": "node3",
            "plugintype": "dir",
            "id": "storage/node3/local",
            "status": "available",
            "type": "storage",
            "shared": 0
        },
        {
            "maxdisk": 940166742016,
            "node": "node2",
            "storage": "local",
            "disk": 2797338624,
            "content": "backup,vztmpl,iso",
            "id": "storage/node2/local",
            "plugintype": "dir",
            "shared": 0,
            "type": "storage",
            "status": "available"
        },
        {
            "maxdisk": 951055941632,
            "disk": 10486546432,
            "content": "backup,vztmpl,iso",
            "storage": "local",
            "node": "node1",
            "id": "storage/node1/local",
            "plugintype": "dir",
            "status": "available",
            "shared": 0,
            "type": "storage"
        },
        {
            "plugintype": "dir",
            "id": "storage/node4/local",
            "status": "available",
            "shared": 0,
            "type": "storage",
            "maxdisk": 482719825920,
            "storage": "local",
            "node": "node4",
            "content": "backup,vztmpl,iso",
            "disk": 2789867520
        },
        {
            "plugintype": "dir",
            "id": "storage/node3/cloud-init",
            "status": "available",
            "type": "storage",
            "shared": 0,
            "maxdisk": 482713534464,
            "content": "snippets",
            "disk": 2983723008,
            "storage": "cloud-init",
            "node": "node3"
        },
        {
            "plugintype": "dir",
            "id": "storage/node2/cloud-init",
            "status": "available",
            "type": "storage",
            "shared": 0,
            "maxdisk": 940166742016,
            "disk": 2797338624,
            "content": "snippets",
            "node": "node2",
            "storage": "cloud-init"
        },
        {
            "disk": 10486546432,
            "content": "snippets",
            "node": "node1",
            "storage": "cloud-init",
            "maxdisk": 951055941632,
            "status": "available",
            "type": "storage",
            "shared": 0,
            "id": "storage/node1/cloud-init",
            "plugintype": "dir"
        },
        {
            "id": "storage/node4/cloud-init",
            "plugintype": "dir",
            "type": "storage",
            "shared": 0,
            "status": "available",
            "maxdisk": 482719825920,
            "content": "snippets",
            "disk": 2789867520,
            "storage": "cloud-init",
            "node": "node4"
        }
    ]
}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/zones$").
		MatchParams(map[string]string{
			"type": "vxlan",
		}).
		Reply(200).
		JSON(`{
		"data": [
				{"zone":"test1","type":"vxlan","ipam":"pve"}
			]
		}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/zones$").
		Reply(200).
		JSON(`{
		"data": [
				{"zone":"test1","type":"vxlan","ipam":"pve"},
				{"zone":"test2","type":"simple","ipam":"pve"}
			]
		}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/vnets$").
		Reply(200).
		JSON(`{
		"data": [
				{"vnet":"user1","type":"vnet","zone":"test1","vlanaware":1,"tag":10,"alias":"myuser1's network"},
				{"vnet":"user10","type":"vnet","zone":"test1","vlanaware":1,"tag":30},
				{"vnet":"user11","type":"vnet","zone":"test1","vlanaware":1,"tag":31},
				{"vnet":"user2","type":"vnet","zone":"test3","vlanaware":1,"tag":11},
				{"vnet":"user3","type":"vnet","zone":"test1","vlanaware":1,"tag":12}
			]
		}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/vnets/user1$").
		Reply(200).
		JSON(`{
		"data": {"vnet":"user1","type":"vnet","zone":"test1","vlanaware":1,"tag":10,"alias":"myuser1's network"}
		}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/vnets/maxTagVnet$").
		Reply(200).
		JSON(`{
		"data": {"vnet":"maxTagVnet","type":"vnet","zone":"test1","vlanaware":1,"tag":16777215}
		}`)

	// GET /cluster/firewall/groups - List firewall security groups
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/groups$").
		Reply(200).
		JSON(`{
		"data": [
			{
				"group": "test-group",
				"comment": "Test security group"
			},
			{
				"group": "web-servers",
				"comment": "Web server security group"
			}
		]
	}`)

	// GET /cluster/firewall/groups/{group} - Get firewall group rules
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/groups/test-group$").
		Reply(200).
		JSON(`{
		"data": [
			{
				"pos": 0,
				"type": "in",
				"action": "ACCEPT",
				"enable": 1,
				"proto": "tcp",
				"dport": "22",
				"comment": "Allow SSH"
			},
			{
				"pos": 1,
				"type": "in",
				"action": "ACCEPT",
				"enable": 1,
				"proto": "tcp",
				"dport": "80",
				"comment": "Allow HTTP"
			}
		]
	}`)

	// POST /cluster/firewall/groups - Create new firewall group
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/groups$").
		Reply(200).
		JSON(`{
		"data": null
	}`)

	// POST /cluster/firewall/groups/{group} - Create rule in group
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/groups/test-group$").
		Reply(200).
		JSON(`{
		"data": null
	}`)

	// PUT /cluster/firewall/groups/{group}/{pos} - Update rule in group
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/firewall/groups/test-group/[0-9]+$").
		Reply(200).
		JSON(`{
		"data": null
	}`)

	// DELETE /cluster/firewall/groups/{group}/{pos} - Delete rule from group
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/firewall/groups/test-group/[0-9]+$").
		Reply(200).
		JSON(`{
		"data": null
	}`)

	// DELETE /cluster/firewall/groups/{group} - Delete firewall group
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/firewall/groups/test-group$").
		Reply(200).
		JSON(`{
		"data": null
	}`)

	clusterBackup()
	clusterFirewallMain()
	clusterHA()
	clusterReplication()
}

func clusterBackup() {
	// GET /cluster/backup — list all backup schedules
	gock.New(config.C.URI).
		Get("^/cluster/backup$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "id": "backup-1",
            "schedule": "*/30",
            "mode": "snapshot",
            "storage": "local",
            "type": "vzdump",
            "enabled": 1,
            "all": 1,
            "next-run": 1715299200,
            "mailnotification": "always",
            "notes-template": "{{guestname}}"
        },
        {
            "id": "backup-2",
            "schedule": "sat 02:00",
            "mode": "stop",
            "storage": "nfs-backups",
            "type": "vzdump",
            "enabled": 0,
            "vmid": "101,102",
            "mailto": "ops@example.com",
            "prune-backups": "keep-daily=7,keep-weekly=4"
        }
    ]
}`)

	// GET /cluster/backup/{id} — single backup schedule
	gock.New(config.C.URI).
		Get("^/cluster/backup/backup-1$").
		Reply(200).
		JSON(`{
    "data": {
        "id": "backup-1",
        "schedule": "*/30",
        "mode": "snapshot",
        "storage": "local",
        "type": "vzdump",
        "enabled": 1,
        "all": 1
    }
}`)

	// POST /cluster/backup — create
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/backup$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/backup/{id} — update
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/backup/backup-1$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/backup/{id} — delete
	gock.New(config.C.URI).
		Delete("^/cluster/backup/backup-1$").
		Reply(200).
		JSON(`{"data": null}`)
}

// clusterFirewallMain registers mocks for /cluster/firewall/{rules,aliases,ipset,options,macros,refs}.
// Split from cluster() so the function stays under the linter complexity ceiling.
func clusterFirewallMain() {
	// GET /cluster/firewall/rules
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/rules$").
		Reply(200).
		JSON(`{
		"data": [
			{"pos": 0, "type": "in", "action": "ACCEPT", "enable": 1, "proto": "tcp", "dport": "22", "comment": "ssh"},
			{"pos": 1, "type": "in", "action": "DROP",   "enable": 1, "proto": "tcp", "dport": "23", "comment": "telnet"}
		]
	}`)

	// POST /cluster/firewall/rules
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/rules$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/rules/{pos}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/rules/[0-9]+$").
		Reply(200).
		JSON(`{
		"data": {"pos": 0, "type": "in", "action": "ACCEPT", "enable": 1, "proto": "tcp", "dport": "22"}
	}`)

	// PUT /cluster/firewall/rules/{pos}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/firewall/rules/[0-9]+$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/firewall/rules/{pos}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/firewall/rules/[0-9]+$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/aliases
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/aliases$").
		Reply(200).
		JSON(`{
		"data": [
			{"name": "test-alias", "cidr": "10.0.0.0/24", "comment": "primary", "digest": "abc"}
		]
	}`)

	// POST /cluster/firewall/aliases
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/aliases$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/aliases/{name}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/aliases/test-alias$").
		Reply(200).
		JSON(`{
		"data": {"name": "test-alias", "cidr": "10.0.0.0/24", "comment": "primary", "digest": "abc"}
	}`)

	// PUT /cluster/firewall/aliases/{name}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/firewall/aliases/test-alias$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/firewall/aliases/{name}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/firewall/aliases/test-alias$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/ipset
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/ipset$").
		Reply(200).
		JSON(`{
		"data": [
			{"name": "test-ipset", "comment": "test", "digest": "abc"}
		]
	}`)

	// POST /cluster/firewall/ipset
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/ipset$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/ipset/{name}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/ipset/test-ipset$").
		Reply(200).
		JSON(`{
		"data": [
			{"cidr": "10.0.0.1", "comment": "host-a", "nomatch": false, "digest": "abc"},
			{"cidr": "10.0.0.2", "comment": "host-b", "nomatch": true,  "digest": "abc"}
		]
	}`)

	// POST /cluster/firewall/ipset/{name}
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/firewall/ipset/test-ipset$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/firewall/ipset/{name}  (with or without ?force=1)
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/firewall/ipset/test-ipset$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/ipset/{name}/{cidr}
	gock.New(config.C.URI).
		Persist().
		Get(`^/cluster/firewall/ipset/test-ipset/10\.0\.0\.1$`).
		Reply(200).
		JSON(`{
		"data": {"cidr": "10.0.0.1", "comment": "host-a", "nomatch": false, "digest": "abc"}
	}`)

	// PUT /cluster/firewall/ipset/{name}/{cidr}
	gock.New(config.C.URI).
		Persist().
		Put(`^/cluster/firewall/ipset/test-ipset/10\.0\.0\.1$`).
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/firewall/ipset/{name}/{cidr}
	gock.New(config.C.URI).
		Persist().
		Delete(`^/cluster/firewall/ipset/test-ipset/10\.0\.0\.1$`).
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/options
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/options$").
		Reply(200).
		JSON(`{
		"data": {"enable": 0, "ebtables": 1, "policy_in": "DROP", "policy_out": "ACCEPT", "policy_forward": "ACCEPT"}
	}`)

	// PUT /cluster/firewall/options
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/firewall/options$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/firewall/macros
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/macros$").
		Reply(200).
		JSON(`{
		"data": [
			{"macro": "HTTP",  "descr": "HTTP traffic"},
			{"macro": "HTTPS", "descr": "HTTPS traffic"}
		]
	}`)

	// GET /cluster/firewall/refs
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/refs").
		Reply(200).
		JSON(`{
		"data": [
			{"name": "test-alias", "ref": "test-alias", "scope": "dc", "type": "alias", "comment": "primary"},
			{"name": "test-ipset", "ref": "+test-ipset","scope": "dc", "type": "ipset"}
		]
	}`)
}

// clusterHA registers mocks for /cluster/ha/{groups,resources,rules,status}.
// Split from cluster() so the function stays under the linter complexity ceiling.
func clusterHA() {
	// GET /cluster/ha/groups
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/groups$").
		Reply(200).
		JSON(`{"data": [{"group": "test-group", "type": "group", "nodes": "node1,node2", "comment": "test"}]}`)

	// GET /cluster/ha/groups/{group}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": {"group": "test-group", "type": "group", "nodes": "node1,node2", "comment": "test", "nofailback": 0, "restricted": 0}}`)

	// POST /cluster/ha/groups
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/groups$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/ha/groups/{group}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/ha/groups/{group}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/ha/resources
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/resources$").
		Reply(200).
		JSON(`{"data": [{"sid": "vm:100", "type": "vm", "state": "started", "group": "test-group", "max_relocate": 1, "max_restart": 1}]}`)

	// GET /cluster/ha/resources/{sid}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": {"sid": "vm:100", "type": "vm", "state": "started", "group": "test-group", "max_relocate": 1, "max_restart": 1, "comment": "primary db"}}`)

	// POST /cluster/ha/resources
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/ha/resources/{sid}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/ha/resources/{sid}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /cluster/ha/resources/{sid}/migrate
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources/vm:100/migrate$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /cluster/ha/resources/{sid}/relocate
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources/vm:100/relocate$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/ha/rules
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/rules$").
		Reply(200).
		JSON(`{"data": [{"rule": "rule-1", "type": "node-affinity", "resources": "vm:100", "nodes": "node1", "strict": 1}]}`)

	// GET /cluster/ha/rules/{rule}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": {"rule": "rule-1", "type": "node-affinity", "resources": "vm:100", "nodes": "node1", "strict": 1}}`)

	// POST /cluster/ha/rules
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/rules$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/ha/rules/{rule}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/ha/rules/{rule}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/ha/status/current
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/status/current$").
		Reply(200).
		JSON(`{"data": [
			{"id": "node1", "type": "node", "node": "node1", "status": "online", "quorate": 1},
			{"id": "vm:100", "type": "service", "node": "node1", "state": "started", "request_state": "started"}
		]}`)

	// GET /cluster/ha/status/manager_status
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/status/manager_status$").
		Reply(200).
		JSON(`{"data": {
			"manager_status": {"node_status": {"node1": "online"}, "master_node": "node1"},
			"node_status": {"node1": "online", "node2": "online"},
			"service_status": {"vm:100": {"state": "started", "node": "node1"}},
			"quorum": {"quorate": 1}
		}}`)
}

// clusterReplication registers mocks for /cluster/replication/{,id}.
func clusterReplication() {
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/replication$").
		Reply(200).
		JSON(`{"data": [
			{"id": "100-0", "type": "local", "target": "node2", "schedule": "*/15", "guest": 100, "jobnum": 0}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/replication/100-0$").
		Reply(200).
		JSON(`{"data": {"id": "100-0", "type": "local", "target": "node2", "schedule": "*/15", "guest": 100, "jobnum": 0}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/replication$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/replication/100-0$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/replication/100-0$").
		Reply(200).
		JSON(`{"data": null}`)
}
