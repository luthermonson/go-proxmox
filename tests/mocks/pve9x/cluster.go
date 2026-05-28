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
}
