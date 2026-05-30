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
				{"zone":"test1","type":"vxlan","ipam":"pve","nodes":"host1,host2","peers":"203.0.113.184,203.0.113.185"},
				{"zone":"test2","type":"simple","ipam":"pve"}
			]
		}`)

	gock.New(config.C.URI).
		Get("^/cluster/sdn/zones/test1$").
		Reply(200).
		JSON(`{
		"data": {"zone":"test1","type":"vxlan","ipam":"pve","nodes":"host1,host2","peers":"203.0.113.184,203.0.113.185"}
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

	// --- /cluster/metrics/server ---------------------------------------------

	// GET /cluster/metrics/server — list configured metric servers
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/metrics/server$").
		Reply(200).
		JSON(`{
    "data": [
        {"id": "influx1", "type": "influxdb", "server": "metrics.example.com", "port": 8086, "disable": 0},
        {"id": "graphite1", "type": "graphite", "server": "graphite.example.com", "port": 2003, "disable": 0}
    ]
}`)

	// GET /cluster/metrics/server/{id} — single metric server config
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/metrics/server/influx1$").
		Reply(200).
		JSON(`{
    "data": {
        "id": "influx1",
        "type": "influxdb",
        "server": "metrics.example.com",
        "port": 8086,
        "influxdbproto": "http",
        "bucket": "proxmox",
        "organization": "ops",
        "token": "secret",
        "disable": 0,
        "verify-certificate": 1,
        "digest": "abc123"
    }
}`)

	// POST /cluster/metrics/server/{id} — create
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/metrics/server/influx1$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/metrics/server/{id} — update
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/metrics/server/influx1$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/metrics/server/{id} — delete
	gock.New(config.C.URI).
		Delete("^/cluster/metrics/server/influx1$").
		Reply(200).
		JSON(`{"data": null}`)

	// --- /cluster/acme -------------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/directories$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "Let's Encrypt V2", "url": "https://acme-v02.api.letsencrypt.org/directory"},
        {"name": "Let's Encrypt V2 Staging", "url": "https://acme-staging-v02.api.letsencrypt.org/directory"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/challenge-schema$").
		Reply(200).
		JSON(`{
    "data": [
        {"id": "dns-01", "name": "dns-01 challenge", "type": "dns", "schema": {}},
        {"id": "http-01", "name": "http-01 challenge", "type": "standalone", "schema": {}}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/tos").
		Reply(200).
		JSON(`{"data": "https://letsencrypt.org/documents/LE-SA-v1.4-April-3-2024.pdf"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/meta").
		Reply(200).
		JSON(`{
    "data": {
        "caaIdentities": ["letsencrypt.org"],
        "externalAccountRequired": 0,
        "termsOfService": "https://letsencrypt.org/documents/LE-SA-v1.4-April-3-2024.pdf",
        "website": "https://letsencrypt.org"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/account$").
		Reply(200).
		JSON(`{"data": [{"name": "default"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/account/default$").
		Reply(200).
		JSON(`{
    "data": {
        "directory": "https://acme-v02.api.letsencrypt.org/directory",
        "location": "https://acme-v02.api.letsencrypt.org/acme/acct/123456",
        "tos": "https://letsencrypt.org/documents/LE-SA-v1.4-April-3-2024.pdf",
        "account": {"status": "valid", "contact": ["mailto:admin@example.com"]}
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/acme/account$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:acme-register:default:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/acme/account/default$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:acme-update:default:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/acme/account/default$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:acme-deactivate:default:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/plugins$").
		Reply(200).
		JSON(`{
    "data": [
        {"plugin": "cloudflare", "type": "dns", "api": "cf", "data": "Y2YtdG9rZW49c2VjcmV0", "disable": 0, "validation-delay": 30}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme/plugins/cloudflare$").
		Reply(200).
		JSON(`{
    "data": {
        "plugin": "cloudflare",
        "type": "dns",
        "api": "cf",
        "data": "Y2YtdG9rZW49c2VjcmV0",
        "disable": 0,
        "validation-delay": 30,
        "digest": "abc123"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/acme/plugins$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/acme/plugins/cloudflare$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/acme/plugins/cloudflare$").
		Reply(200).
		JSON(`{"data": null}`)

	// --- /cluster/mapping ----------------------------------------------------

	// GET /cluster/mapping — directory index
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "dir"},
        {"name": "pci"},
        {"name": "usb"}
    ]
}`)

	// GET /cluster/mapping/dir — list directory mappings
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/dir$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "id": "shared-iso",
            "description": "Shared ISO directory",
            "map": ["node=node1,path=/srv/iso", "node=node2,path=/srv/iso"]
        }
    ]
}`)

	// GET /cluster/mapping/dir/{id}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/dir/shared-iso$").
		Reply(200).
		JSON(`{
    "data": {
        "id": "shared-iso",
        "description": "Shared ISO directory",
        "map": ["node=node1,path=/srv/iso", "node=node2,path=/srv/iso"],
        "digest": "d1"
    }
}`)

	// POST /cluster/mapping/dir
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/mapping/dir$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/mapping/dir/{id}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/mapping/dir/shared-iso$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/mapping/dir/{id}
	gock.New(config.C.URI).
		Delete("^/cluster/mapping/dir/shared-iso$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/mapping/pci — list PCI mappings
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/pci$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "id": "gpu0",
            "description": "Tesla T4",
            "map": ["node=node1,path=0000:01:00.0,id=10de:1eb8,iommugroup=12"],
            "mdev": 0
        }
    ]
}`)

	// GET /cluster/mapping/pci/{id}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/pci/gpu0$").
		Reply(200).
		JSON(`{
    "data": {
        "id": "gpu0",
        "description": "Tesla T4",
        "map": ["node=node1,path=0000:01:00.0,id=10de:1eb8,iommugroup=12"],
        "mdev": 0,
        "live-migration-capable": 0,
        "digest": "p1"
    }
}`)

	// POST /cluster/mapping/pci
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/mapping/pci$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/mapping/pci/{id}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/mapping/pci/gpu0$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/mapping/pci/{id}
	gock.New(config.C.URI).
		Delete("^/cluster/mapping/pci/gpu0$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /cluster/mapping/usb — list USB mappings
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/usb$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "id": "yubikey",
            "description": "YubiKey 5",
            "map": ["node=node1,id=1050:0407,path=1-1"]
        }
    ]
}`)

	// GET /cluster/mapping/usb/{id}
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/mapping/usb/yubikey$").
		Reply(200).
		JSON(`{
    "data": {
        "id": "yubikey",
        "description": "YubiKey 5",
        "map": ["node=node1,id=1050:0407,path=1-1"],
        "digest": "u1"
    }
}`)

	// POST /cluster/mapping/usb
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/mapping/usb$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /cluster/mapping/usb/{id}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/mapping/usb/yubikey$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /cluster/mapping/usb/{id}
	gock.New(config.C.URI).
		Delete("^/cluster/mapping/usb/yubikey$").
		Reply(200).
		JSON(`{"data": null}`)

	// --- /cluster/notifications ---------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "endpoints"},
        {"name": "matchers"},
        {"name": "targets"},
        {"name": "matcher-fields"},
        {"name": "matcher-field-values"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/matcher-fields$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "type"},
        {"name": "hostname"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/matcher-field-values$").
		Reply(200).
		JSON(`{
    "data": [
        {"field": "type", "value": "vzdump", "comment": "Backup notifications"},
        {"field": "type", "value": "system"}
    ]
}`)

	// Targets
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/targets$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "mail-to-root", "type": "sendmail", "origin": "builtin", "disable": 0},
        {"name": "gotify1", "type": "gotify", "origin": "user-created", "disable": 0}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/targets/mail-to-root/test$").
		Reply(200).
		JSON(`{"data": null}`)

	// Matchers
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/matchers$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "name": "default-matcher",
            "mode": "all",
            "target": ["mail-to-root"],
            "origin": "builtin"
        }
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/matchers/default-matcher$").
		Reply(200).
		JSON(`{
    "data": {
        "name": "default-matcher",
        "mode": "all",
        "match-severity": ["warning", "error"],
        "target": ["mail-to-root"],
        "digest": "m1"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/matchers$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/notifications/matchers/default-matcher$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/notifications/matchers/default-matcher$").
		Reply(200).
		JSON(`{"data": null}`)

	// Gotify endpoints
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/gotify$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "gotify1", "server": "https://gotify.example.com", "disable": 0}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/gotify/gotify1$").
		Reply(200).
		JSON(`{
    "data": {
        "name": "gotify1",
        "server": "https://gotify.example.com",
        "disable": 0,
        "digest": "g1"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/endpoints/gotify$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/notifications/endpoints/gotify/gotify1$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/notifications/endpoints/gotify/gotify1$").
		Reply(200).
		JSON(`{"data": null}`)

	// Sendmail endpoints
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/sendmail$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "mail-to-root", "mailto-user": ["root@pam"], "origin": "builtin"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/sendmail/mail-to-root$").
		Reply(200).
		JSON(`{
    "data": {
        "name": "mail-to-root",
        "mailto-user": ["root@pam"],
        "from-address": "root@pve",
        "digest": "s1"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/endpoints/sendmail$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/notifications/endpoints/sendmail/mail-to-root$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/notifications/endpoints/sendmail/mail-to-root$").
		Reply(200).
		JSON(`{"data": null}`)

	// SMTP endpoints
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/smtp$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "name": "smtp1",
            "server": "smtp.example.com",
            "port": 587,
            "mode": "starttls",
            "from-address": "alerts@example.com",
            "username": "alerts"
        }
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/smtp/smtp1$").
		Reply(200).
		JSON(`{
    "data": {
        "name": "smtp1",
        "server": "smtp.example.com",
        "port": 587,
        "mode": "starttls",
        "from-address": "alerts@example.com",
        "mailto": ["ops@example.com"],
        "digest": "st1"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/endpoints/smtp$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/notifications/endpoints/smtp/smtp1$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/notifications/endpoints/smtp/smtp1$").
		Reply(200).
		JSON(`{"data": null}`)

	// Webhook endpoints
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/webhook$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "wh1", "url": "https://hook.example.com/alert", "method": "post"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints/webhook/wh1$").
		Reply(200).
		JSON(`{
    "data": {
        "name": "wh1",
        "url": "https://hook.example.com/alert",
        "method": "post",
        "body": "eyJoZWxsbyI6IndvcmxkIn0=",
        "digest": "w1"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/notifications/endpoints/webhook$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/notifications/endpoints/webhook/wh1$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/notifications/endpoints/webhook/wh1$").
		Reply(200).
		JSON(`{"data": null}`)

	// --- /cluster/jobs -------------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/jobs$").
		Reply(200).
		JSON(`{"data": [{"subdir": "realm-sync"}, {"subdir": "schedule-analyze"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/jobs/schedule-analyze").
		Reply(200).
		JSON(`{"data": [
			{"timestamp": 1715731200, "utc": "2026-05-15 00:00:00"},
			{"timestamp": 1715817600, "utc": "2026-05-16 00:00:00"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/jobs/realm-sync$").
		Reply(200).
		JSON(`{"data": [
			{"id": "ldap-sync", "realm": "ldap1", "schedule": "daily", "enabled": 1, "enable-new": 1, "scope": "both"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/jobs/realm-sync/ldap-sync$").
		Reply(200).
		JSON(`{"data": {
			"id": "ldap-sync",
			"realm": "ldap1",
			"schedule": "daily",
			"enabled": 1,
			"enable-new": 1,
			"scope": "both",
			"remove-vanished": "none",
			"comment": "daily LDAP sync"
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/jobs/realm-sync/ldap-sync$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/jobs/realm-sync/ldap-sync$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/cluster/jobs/realm-sync/ldap-sync$").
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
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/groups$").
		Reply(200).
		JSON(`{"data": [{"group": "test-group", "type": "group", "nodes": "node1,node2", "comment": "test"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": {"group": "test-group", "type": "group", "nodes": "node1,node2", "comment": "test", "nofailback": 0, "restricted": 0}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/groups$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/groups/test-group$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/resources$").
		Reply(200).
		JSON(`{"data": [{"sid": "vm:100", "type": "vm", "state": "started", "group": "test-group", "max_relocate": 1, "max_restart": 1}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": {"sid": "vm:100", "type": "vm", "state": "started", "group": "test-group", "max_relocate": 1, "max_restart": 1, "comment": "primary db"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/resources/vm:100$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources/vm:100/migrate$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/resources/vm:100/relocate$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/rules$").
		Reply(200).
		JSON(`{"data": [{"rule": "rule-1", "type": "node-affinity", "resources": "vm:100", "nodes": "node1", "strict": 1}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": {"rule": "rule-1", "type": "node-affinity", "resources": "vm:100", "nodes": "node1", "strict": 1}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/rules$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/ha/rules/rule-1$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/status/current$").
		Reply(200).
		JSON(`{"data": [
			{"id": "node1", "type": "node", "node": "node1", "status": "online", "quorate": 1},
			{"id": "vm:100", "type": "service", "node": "node1", "state": "started", "request_state": "started"}
		]}`)

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

	// --- diridx endpoints (see cluster_diridx.go) ---------------------------

	// GET /cluster — cluster root diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"replication"},
			{"subdir":"metrics"},
			{"subdir":"config"},
			{"subdir":"firewall"},
			{"subdir":"backup"},
			{"subdir":"backupinfo"},
			{"subdir":"ha"},
			{"subdir":"acme"},
			{"subdir":"ceph"},
			{"subdir":"jobs"},
			{"subdir":"sdn"},
			{"subdir":"log"},
			{"subdir":"resources"},
			{"subdir":"tasks"},
			{"subdir":"options"},
			{"subdir":"status"},
			{"subdir":"nextid"}
		]}`)

	// GET /cluster/acme — acme diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/acme$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"plugins"},
			{"subdir":"account"},
			{"subdir":"tos"},
			{"subdir":"meta"},
			{"subdir":"directories"},
			{"subdir":"challenge-schema"}
		]}`)

	// GET /cluster/firewall — cluster firewall diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"groups"},
			{"subdir":"rules"},
			{"subdir":"ipset"},
			{"subdir":"aliases"},
			{"subdir":"options"},
			{"subdir":"macros"},
			{"subdir":"refs"}
		]}`)

	// GET /cluster/sdn — cluster sdn diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"vnets"},
			{"subdir":"zones"},
			{"subdir":"controllers"},
			{"subdir":"ipams"},
			{"subdir":"dns"},
			{"subdir":"fabrics"},
			{"subdir":"subnets"}
		]}`)

	// GET /cluster/ceph — cluster ceph diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ceph$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"metadata"},
			{"subdir":"status"},
			{"subdir":"flags"}
		]}`)

	// GET /cluster/config — cluster config diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"nodes"},
			{"subdir":"join"},
			{"subdir":"totem"},
			{"subdir":"qdevice"},
			{"subdir":"apiversion"}
		]}`)

	// GET /cluster/ha — cluster ha diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"groups"},
			{"subdir":"resources"},
			{"subdir":"status"},
			{"subdir":"rules"}
		]}`)

	// GET /cluster/ha/status — cluster ha status diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ha/status$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"current"},
			{"subdir":"manager_status"}
		]}`)

	// GET /cluster/qemu — cluster qemu diridx (subdirs are VMIDs)
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/qemu$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"100"},
			{"subdir":"101"},
			{"subdir":"200"}
		]}`)

	// --- /cluster/sdn apply + zones/vnets mutations + filtered list -----------

	// PUT /cluster/sdn/ — apply pending config (returns UPID)
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/?$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:sdnapply:cluster:root@pam:"}`)

	// POST /cluster/sdn/zones — create
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/zones$").
		Reply(200).
		JSON(`{"data":null}`)

	// PUT /cluster/sdn/zones/{name} — update (mock by test1)
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/zones/test1$").
		Reply(200).
		JSON(`{"data":null}`)

	// DELETE /cluster/sdn/zones/{name}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/zones/test1$").
		Reply(200).
		JSON(`{"data":null}`)

	// POST /cluster/sdn/vnets — create
	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/vnets$").
		Reply(200).
		JSON(`{"data":null}`)

	// PUT /cluster/sdn/vnets/{name}
	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/vnets/user1$").
		Reply(200).
		JSON(`{"data":null}`)

	// DELETE /cluster/sdn/vnets/{name}
	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/vnets/user1$").
		Reply(200).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/vnets/user1/subnets — already registered below, leave as-is

	// GET /cluster/sdn/controllers?type=evpn — filtered list
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/controllers$").
		MatchParam("type", "evpn").
		Reply(200).
		JSON(`{"data":[{"controller":"ctrl1","type":"evpn","asn":65000}]}`)

	// GET /cluster/sdn/dns?type=powerdns — filtered list
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/dns$").
		MatchParam("type", "powerdns").
		Reply(200).
		JSON(`{"data":[{"dns":"pdns1","type":"powerdns"}]}`)

	// GET /cluster/sdn/ipams?type=pve — filtered list
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams$").
		MatchParam("type", "pve").
		Reply(200).
		JSON(`{"data":[{"ipam":"pve","type":"pve"}]}`)

	// GET /cluster/sdn/fabrics/fabric?pending=1&running=1 — filtered list
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/fabric$").
		MatchParam("pending", "1").
		Reply(200).
		JSON(`{"data":[{"id":"fab1","protocol":"openfabric","ip_prefix":"10.0.0.0/24","state":"new"}]}`)

	// GET /cluster/sdn/prefix-lists?pending=1&verbose=1
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists$").
		MatchParam("verbose", "1").
		Reply(200).
		JSON(`{"data":[{"id":"pl1","entries":[{"seq":10,"action":"permit","prefix":"10.0.0.0/24"}]}]}`)

	// GET /cluster/sdn/route-maps?running=1
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps$").
		MatchParam("running", "1").
		Reply(200).
		JSON(`{"data":[{"id":"rm1"}]}`)

	// GET /cluster/sdn/route-maps/entries?pending=1
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries$").
		MatchParam("pending", "1").
		Reply(200).
		JSON(`{"data":[{"route-map-id":"rm1","order":10,"action":"permit"}]}`)

	// GET /cluster/sdn/route-maps/entries/rm1?pending=1
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries/rm1$").
		MatchParam("pending", "1").
		Reply(200).
		JSON(`{"data":[{"route-map-id":"rm1","order":10,"action":"permit"}]}`)

	// --- /cluster/sdn error fixtures (404/401) for negative paths -------------

	// GET /cluster/sdn/controllers/missing -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/controllers/missing$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/dns/missing -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/dns/missing$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/ipams/missing -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams/missing$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/fabrics/fabric/missing -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/fabric/missing$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/prefix-lists/missing -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists/missing$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/prefix-lists/missing/entries -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists/missing/entries$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/route-maps/entries/missing/entry/99 -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries/missing/entry/99$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/vnets/missing/subnets -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/missing/subnets$").
		Reply(500).
		JSON(`{"data":null}`)

	// GET /cluster/sdn/vnets/missing/firewall/options -> 404
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/missing/firewall/options$").
		Reply(500).
		JSON(`{"data":null}`)

	// --- /cluster/sdn lock/rollback/dry-run -----------------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/lock$").
		Reply(200).
		JSON(`{"data":"tok-abc123"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/lock").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/rollback$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/dry-run").
		Reply(200).
		JSON(`{"data":{
			"frr-diff":"+router bgp 65000\n+ network 10.0.0.0/24\n",
			"interfaces-diff":"+auto vnet1\n+iface vnet1 inet manual\n"
		}}`)

	// --- /cluster/sdn/controllers ---------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/controllers$").
		Reply(200).
		JSON(`{"data":[
			{"controller":"ctrl1","type":"evpn","asn":65000,"peers":"10.0.0.1,10.0.0.2"},
			{"controller":"ctrl2","type":"bgp","asn":65001}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/controllers/ctrl1$").
		Reply(200).
		JSON(`{"data":{"controller":"ctrl1","type":"evpn","asn":65000,"peers":"10.0.0.1,10.0.0.2","peer-group-name":"VTEP"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/controllers$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/controllers/ctrl1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/controllers/ctrl1$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/dns -----------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/dns$").
		Reply(200).
		JSON(`{"data":[
			{"dns":"pdns1","type":"powerdns"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/dns/pdns1$").
		Reply(200).
		JSON(`{"data":{"dns":"pdns1","type":"powerdns","url":"https://pdns.example.com/api/v1","ttl":3600}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/dns$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/dns/pdns1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/dns/pdns1$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/ipams ---------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams$").
		Reply(200).
		JSON(`{"data":[
			{"ipam":"pve","type":"pve"},
			{"ipam":"netbox1","type":"netbox"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams/pve$").
		Reply(200).
		JSON(`{"data":{"ipam":"pve","type":"pve"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams/pve/status$").
		Reply(200).
		JSON(`{"data":[
			{"hostname":"vm100","ip":"10.0.0.10","mac":"aa:bb:cc:dd:ee:01","subnet":"10.0.0.0-24","vmid":"100","vnet":"vnet1","zone":"zone1"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/ipams$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/ipams/pve$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/ipams/pve$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/fabrics -------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"fabric"},
			{"subdir":"node"},
			{"subdir":"all"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/all$").
		Reply(200).
		JSON(`{"data":{
			"fabrics":[{"id":"fab1","protocol":"openfabric","ip_prefix":"10.0.0.0/24"}],
			"nodes":[{"fabric_id":"fab1","node_id":"node1","ip":"10.0.0.1"}]
		}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/fabric$").
		Reply(200).
		JSON(`{"data":[
			{"id":"fab1","protocol":"openfabric","ip_prefix":"10.0.0.0/24"},
			{"id":"fab2","protocol":"ospf","area":"0.0.0.0"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/fabric/fab1$").
		Reply(200).
		JSON(`{"data":{"id":"fab1","protocol":"openfabric","ip_prefix":"10.0.0.0/24","hello_interval":3,"csnp_interval":10}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/fabrics/fabric$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/fabrics/fabric/fab1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/fabrics/fabric/fab1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/node$").
		Reply(200).
		JSON(`{"data":[
			{"fabric_id":"fab1","node_id":"node1","ip":"10.0.0.1"},
			{"fabric_id":"fab1","node_id":"node2","ip":"10.0.0.2"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/node/fab1$").
		Reply(200).
		JSON(`{"data":[
			{"fabric_id":"fab1","node_id":"node1","ip":"10.0.0.1"},
			{"fabric_id":"fab1","node_id":"node2","ip":"10.0.0.2"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/fabrics/node/fab1/node1$").
		Reply(200).
		JSON(`{"data":{"fabric_id":"fab1","node_id":"node1","ip":"10.0.0.1","interfaces":["name=eth0,ip=10.0.0.1/24"]}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/fabrics/node/fab1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/fabrics/node/fab1/node1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/fabrics/node/fab1/node1$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/prefix-lists --------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists$").
		Reply(200).
		JSON(`{"data":[
			{"id":"pl1"},
			{"id":"pl2"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists/pl1$").
		Reply(200).
		JSON(`{"data":{"id":"pl1","entries":[{"seq":10,"action":"permit","prefix":"10.0.0.0/24"}]}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/prefix-lists$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/prefix-lists/pl1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/prefix-lists/pl1$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists/pl1/entries$").
		Reply(200).
		JSON(`{"data":[
			{"seq":10,"action":"permit","prefix":"10.0.0.0/24"},
			{"seq":20,"action":"deny","prefix":"10.0.1.0/24"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/prefix-lists/pl1/entries/10$").
		Reply(200).
		JSON(`{"data":{"seq":10,"action":"permit","prefix":"10.0.0.0/24"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/prefix-lists/pl1/entries$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/prefix-lists/pl1/entries/10$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/prefix-lists/pl1/entries/10$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/route-maps ----------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps$").
		Reply(200).
		JSON(`{"data":[
			{"id":"rm1"},
			{"id":"rm2"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries$").
		Reply(200).
		JSON(`{"data":[
			{"route-map-id":"rm1","order":10,"action":"permit"},
			{"route-map-id":"rm1","order":20,"action":"deny"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries/rm1$").
		Reply(200).
		JSON(`{"data":[
			{"route-map-id":"rm1","order":10,"action":"permit","match":["key=metric,value=100"]}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/route-maps/entries/rm1/entry/10$").
		Reply(200).
		JSON(`{"data":{"route-map-id":"rm1","order":10,"action":"permit","match":["key=metric,value=100"],"set":["key=local-preference,value=200"]}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/route-maps/entries$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/route-maps/entries/rm1/entry/10$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/route-maps/entries/rm1/entry/10$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/vnets/{vnet}/firewall -----------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/firewall$").
		Reply(200).
		JSON(`{"data":[{"subdir":"options"},{"subdir":"rules"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/firewall/options$").
		Reply(200).
		JSON(`{"data":{"enable":1,"policy_forward":"ACCEPT","log_level_forward":"info"}}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/vnets/user1/firewall/options$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/firewall/rules$").
		Reply(200).
		JSON(`{"data":[
			{"pos":0,"type":"in","action":"ACCEPT","enable":1,"source":"10.0.0.0/24"},
			{"pos":1,"type":"out","action":"DROP","enable":1}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/firewall/rules/0$").
		Reply(200).
		JSON(`{"data":{"pos":0,"type":"in","action":"ACCEPT","enable":1,"source":"10.0.0.0/24"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/vnets/user1/firewall/rules$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/vnets/user1/firewall/rules/0$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/vnets/user1/firewall/rules/0$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/vnets/{vnet}/ips ----------------------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/vnets/user1/ips$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/vnets/user1/ips$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/vnets/user1/ips").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/sdn/vnets/{vnet}/subnets ------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/subnets$").
		Reply(200).
		JSON(`{"data":[
			{"id":"zone1-10.0.0.0-24","cidr":"10.0.0.0/24","gateway":"10.0.0.1","type":"subnet","vnet":"user1","zone":"zone1"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/vnets/user1/subnets/zone1-10.0.0.0-24$").
		Reply(200).
		JSON(`{"data":{"id":"zone1-10.0.0.0-24","cidr":"10.0.0.0/24","gateway":"10.0.0.1","type":"subnet","vnet":"user1","zone":"zone1"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/sdn/vnets/user1/subnets$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/sdn/vnets/user1/subnets/zone1-10.0.0.0-24$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/sdn/vnets/user1/subnets/zone1-10.0.0.0-24$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/config/* ----------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config/apiversion$").
		Reply(200).
		JSON(`{"data":1}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config/join$").
		Reply(200).
		JSON(`{"data":{
			"config_digest":"abc123",
			"preferred_node":"node1",
			"nodelist":[
				{"name":"node1","nodeid":1,"pve_addr":"10.0.0.1","pve_fp":"AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99","quorum_votes":1,"ring0_addr":"10.0.0.1"}
			],
			"totem":{"transport":"knet","token":10000}
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/config/join$").
		Reply(200).
		JSON(`{"data":"OK"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config/nodes$").
		Reply(200).
		JSON(`{"data":[{"node":"node1"},{"node":"node2"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config/qdevice$").
		Reply(200).
		JSON(`{"data":{"state":"running","mode":"sync"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/config/totem$").
		Reply(200).
		JSON(`{"data":{"transport":"knet","token":10000,"version":2}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/config$").
		Reply(200).
		JSON(`{"data":"OK"}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/config/nodes/node2$").
		Reply(200).
		JSON(`{"data":{"corosync_authkey":"key-bytes","corosync_conf":"conf-bytes","warnings":[]}}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/config/nodes/node2$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/qemu/cpu-flags + custom-cpu-models --------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/qemu/cpu-flags$").
		Reply(200).
		JSON(`{"data":[
			{"name":"aes","description":"AES-NI","supported-on":["node1","node2"]},
			{"name":"avx2","description":"AVX2","supported-on":["node1"]}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/qemu/custom-cpu-models$").
		Reply(200).
		JSON(`{"data":[
			{"cputype":"custom-epyc","reported-model":"EPYC","flags":"+aes","hidden":0}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/qemu/custom-cpu-models/custom-epyc$").
		Reply(200).
		JSON(`{"data":{"cputype":"custom-epyc","reported-model":"EPYC","flags":"+aes","hidden":0,"phys-bits":"host"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/qemu/custom-cpu-models$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/qemu/custom-cpu-models/custom-epyc$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/cluster/qemu/custom-cpu-models/custom-epyc$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/bulk-action -------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/bulk-action$").
		Reply(200).
		JSON(`{"data":[{"subdir":"guest"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/bulk-action/guest$").
		Reply(200).
		JSON(`{"data":[{"subdir":"start"},{"subdir":"shutdown"},{"subdir":"suspend"},{"subdir":"migrate"}]}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/bulk-action/guest/start$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:bulk_action:cluster:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/bulk-action/guest/shutdown$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:bulk_action:cluster:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/bulk-action/guest/suspend$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:bulk_action:cluster:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/bulk-action/guest/migrate$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:bulk_action:cluster:root@pam:"}`)

	// --- /cluster/ceph/flags + /cluster/ceph/metadata -------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ceph/flags$").
		Reply(200).
		JSON(`{"data":[
			{"name":"noout","description":"OSDs will not be marked out","value":true},
			{"name":"noscrub","description":"Scrubbing disabled","value":false}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ceph/flags/noout$").
		Reply(200).
		JSON(`{"data":"set"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/ceph/metadata$").
		Reply(200).
		JSON(`{"data":{
			"version":{"version":"18.2.0","buildcommit":"abc"},
			"osd":[{"id":0,"ceph_version":"18.2.0","hostname":"node1"}],
			"mon":[],
			"mgr":[],
			"mds":[]
		}}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ceph/flags$").
		Reply(200).
		JSON(`{"data":"UPID:node1:0000ABCD:00ABCDEF:00000000:cephflags:cluster:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/ceph/flags/noout$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/backup-info -------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/backup-info$").
		Reply(200).
		JSON(`{"data":[{"subdir":"not-backed-up"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/backup-info/not-backed-up$").
		Reply(200).
		JSON(`{"data":[
			{"vmid":100,"type":"qemu","name":"server1"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/backup/backup-1/included_volumes$").
		Reply(200).
		JSON(`{"data":{
			"children":[
				{"id":100,"type":"qemu","name":"server1","children":[
					{"id":"scsi0","name":"local-lvm:vm-100-disk-0","included":true,"reason":"in-backup"}
				]}
			]
		}}`)

	// --- /cluster/ha/status/{arm,disarm}-ha -----------------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/status/arm-ha$").
		Reply(200).
		JSON(`{"data":null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/cluster/ha/status/disarm-ha$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/metrics (top-level) + /cluster/metrics/export ---------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/metrics$").
		Reply(200).
		JSON(`{"data":[{"subdir":"server"},{"subdir":"export"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/metrics/export").
		Reply(200).
		JSON(`{"data":{"data":[
			{"id":"node/node1","metric":"cpu","timestamp":1700000000,"type":"gauge","value":0.42},
			{"id":"qemu/100","metric":"mem","timestamp":1700000000,"type":"gauge","value":123456789}
		]}}`)

	// --- /cluster/options -----------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/options$").
		Reply(200).
		JSON(`{"data":{
			"console":"html5",
			"language":"en",
			"max_workers":4,
			"mac_prefix":"BC:24:11",
			"migration":"secure",
			"ha":"shutdown_policy=conditional",
			"next-id":"lower=100,upper=1000000",
			"crs":"ha=basic"
		}}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/cluster/options$").
		Reply(200).
		JSON(`{"data":null}`)

	// --- /cluster/log + /cluster/notifications/endpoints + firewall group rule

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/log").
		Reply(200).
		JSON(`{"data":[
			{"node":"node1","time":1700000000,"user":"root@pam","pri":6,"tag":"task","msg":"start"},
			{"node":"node2","time":1700000050,"user":"root@pam","pri":6,"tag":"task","msg":"done"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/notifications/endpoints$").
		Reply(200).
		JSON(`{"data":[{"subdir":"sendmail"},{"subdir":"gotify"},{"subdir":"smtp"},{"subdir":"webhook"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/firewall/groups/test-group/0$").
		Reply(200).
		JSON(`{"data":{"pos":0,"type":"in","action":"ACCEPT","enable":1,"source":"10.0.0.0/24"}}`)

	// --- /cluster/tasks ------------------------------------------------------
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/tasks$").
		Reply(200).
		JSON(`{"data":[
			{"upid":"UPID:node1:00000010:00000010:00000010:vzdump:100:root@pam:","node":"node1","type":"vzdump","id":"100","user":"root@pam","status":"OK","starttime":1700000000,"endtime":1700000060},
			{"upid":"UPID:node2:00000011:00000011:00000011:qmstart:101:root@pam:","node":"node2","type":"qmstart","id":"101","user":"root@pam","status":"running","starttime":1700000100}
		]}`)
}
