package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func containers() {
	// GET /nodes/{node}/lxc - List containers
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "vmid": 100,
            "status": "running",
            "name": "ct-test-1",
            "cpus": 2,
            "maxmem": 2147483648,
            "maxdisk": 10737418240,
            "maxswap": 536870912,
            "uptime": 12345,
            "tags": "prod;web"
        },
        {
            "vmid": 101,
            "status": "stopped",
            "name": "ct-test-2",
            "cpus": 1,
            "maxmem": 1073741824,
            "maxdisk": 8589934592,
            "maxswap": 268435456,
            "uptime": 0,
            "tags": "tag1;tag2"
        },
        {
            "vmid": 102,
            "status": "running",
            "name": "ct-test-3",
            "cpus": 4,
            "maxmem": 4294967296,
            "maxdisk": 21474836480,
            "maxswap": 1073741824,
            "uptime": 54321,
            "tags": ""
        }
    ]
}`)

	// GET /nodes/{node}/lxc/{vmid}/status/current - Get container current status
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/status/current$").
		Reply(200).
		JSON(`{
    "data": {
        "vmid": 101,
        "status": "stopped",
        "name": "ct-test-2",
        "cpus": 1,
        "maxmem": 1073741824,
        "maxdisk": 8589934592,
        "maxswap": 268435456,
        "uptime": 0,
        "tags": "tag1;tag2"
    }
}`)

	// GET /nodes/{node}/lxc/{vmid}/config - Get container configuration
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/config$").
		Reply(200).
		JSON(`{
    "data": {
        "arch": "amd64",
        "cores": 1,
        "hostname": "ct-test-2",
        "memory": 1024,
        "net0": "name=eth0,bridge=vmbr0,firewall=1,hwaddr=BC:24:11:2E:C5:7E,ip=dhcp,type=veth",
        "ostype": "ubuntu",
        "rootfs": "local-lvm:vm-101-disk-0,size=8G",
        "swap": 256,
        "tags": "tag1;tag2",
        "digest": "abcdef1234567890"
    }
}`)

	// POST /nodes/{node}/lxc/{vmid}/clone - Clone container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/clone$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzmigrate:101:root@pam:"}`)

	// DELETE /nodes/{node}/lxc/{vmid} - Delete container
	gock.New(config.C.URI).
		Delete("^/nodes/node1/lxc/101$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdestroy:101:root@pam:"}`)

	// PUT /nodes/{node}/lxc/{vmid}/config - Update container config
	gock.New(config.C.URI).
		Put("^/nodes/node1/lxc/101/config$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzconfig:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/start - Start container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/start$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzstart:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/stop - Stop container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/stop$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzstop:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/suspend - Suspend container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/suspend$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzsuspend:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/reboot - Reboot container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/reboot$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzreboot:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/resume - Resume container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/resume$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzresume:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/status/shutdown - Shutdown container
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/shutdown$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzshutdown:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/template - Convert to template
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/template$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /nodes/{node}/lxc/{vmid}/snapshot - List snapshots
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/snapshot$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "name": "snapshot1",
            "description": "First snapshot",
            "snaptime": 1609459200
        },
        {
            "name": "snapshot2",
            "description": "Second snapshot",
            "snaptime": 1609545600
        },
        {
            "name": "snapshot3",
            "description": "Third snapshot",
            "snaptime": 1609632000
        }
    ]
}`)

	// POST /nodes/{node}/lxc/{vmid}/snapshot - Create snapshot
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/snapshot$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzsnapshot:101:root@pam:"}`)

	// GET /nodes/{node}/lxc/{vmid}/snapshot/{snapname} - Get snapshot
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/snapshot/snapshot1$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "name": "snapshot1",
            "description": "First snapshot",
            "snaptime": 1609459200
        }
    ]
}`)

	// DELETE /nodes/{node}/lxc/{vmid}/snapshot/{snapname} - Delete snapshot
	gock.New(config.C.URI).
		Delete("^/nodes/node1/lxc/101/snapshot/snapshot1$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzdelsnapshot:101:root@pam:"}`)

	// POST /nodes/{node}/lxc/{vmid}/snapshot/{snapname}/rollback - Rollback snapshot
	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/snapshot/snapshot1/rollback$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:5A3B7C8D:vzrollback:101:root@pam:"}`)

	// GET /nodes/{node}/lxc/{vmid}/interfaces - Get network interfaces
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/interfaces$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "hwaddr": "00:00:00:00:00:00",
            "inet": "127.0.0.1/8",
            "inet6": "::1/128",
            "name": "lo"
        },
        {
            "hwaddr": "bc:24:11:89:67:07",
            "inet": "192.168.3.95/22",
            "inet6": "fe80::be24:11ff:fe89:6707/64",
            "name": "eth0"
        }
    ]
}`)
}
