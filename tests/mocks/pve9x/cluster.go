package pve9x

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
				{"vnet":"user1","type":"vnet","zone":"test1","vlanaware":1,"tag":10},
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
		"data": {"vnet":"user1","type":"vnet","zone":"test1","vlanaware":1,"tag":10}
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
}
