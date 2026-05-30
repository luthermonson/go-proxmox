package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func virtualMachines() {
	// GET /nodes/{node}/qemu/{vmid}/config - Get VM config (vmid 101, all merged device types for tests)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/config$").
		Reply(200).
		JSON(`{
    "data": {
        "digest": "abc123def456",
        "name": "matt",
        "vmid": 101,
        "cores": 2,
        "memory": 2048,
        "sockets": 1,
        "ostype": "l26",
        "boot": "order=scsi0;ide2;net0",
        "scsihw": "virtio-scsi-pci",
        "tags": "production;webserver",
        "ide0": "local:100/vm-101-disk-0.qcow2",
        "ide2": "local:iso/debian-12.iso,media=cdrom",
        "scsi0": "local-lvm:vm-101-disk-0,size=32G",
        "sata0": "local-lvm:vm-101-disk-1,size=32G",
        "virtio0": "local-lvm:vm-101-disk-2,size=64G",
        "unused0": "local-lvm:vm-101-unused",
        "net0": "virtio=BC:24:11:2E:C5:4A,bridge=vmbr0",
        "numa0": "cpus=0-1,memory=2048",
        "hostpci0": "0000:01:00.0",
        "serial0": "socket",
        "usb0": "host=1234:5678",
        "parallel0": "/dev/parport0",
        "ipconfig0": "ip=192.168.1.10/24,gw=192.168.1.1"
    }
}`)

	// GET /nodes/{node}/qemu/102/config — high-index device entries plus
	// prefix-collision fields (scsihw, bare numa). Issue #211 regression coverage:
	// proves the UnmarshalJSON router routes net15..net31, scsi30, unused255,
	// hostpci15, ipconfig20 into the maps and does NOT route scsihw or numa.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/102/config$").
		Reply(200).
		JSON(`{
    "data": {
        "digest": "abc123def456beyondten",
        "name": "wide",
        "vmid": 102,
        "cores": 4,
        "memory": 4096,
        "ostype": "l26",
        "scsihw": "virtio-scsi-pci",
        "numa": 1,
        "scsi0": "local-lvm:vm-102-disk-0,size=32G",
        "scsi30": "local-lvm:vm-102-disk-30,size=32G",
        "net0": "virtio=00:00:00:00:00:00,bridge=vmbr0",
        "net15": "virtio=00:00:00:00:00:15,bridge=vmbr15",
        "net31": "virtio=00:00:00:00:00:31,bridge=vmbr31",
        "unused15": "local-lvm:vm-102-unused-15",
        "unused255": "local-lvm:vm-102-unused-255",
        "hostpci15": "0000:0f:00.0",
        "numa0": "cpus=0-1,memory=2048",
        "ipconfig20": "ip=10.0.0.20/24,gw=10.0.0.1"
    }
}`)

	// GET /nodes/node1/qemu/102/status/current — minimal status payload so
	// node.VirtualMachine(ctx, 102) succeeds before fetching /config.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/102/status/current$").
		Reply(200).
		JSON(`{
    "data": {
        "vmid": 102,
        "name": "wide",
        "status": "running",
        "uptime": 1234,
        "cpus": 4,
        "maxmem": 4294967296,
        "mem": 0,
        "maxdisk": 34359738368,
        "disk": 0
    }
}`)

	// GET /nodes/{node}/qemu/{vmid}/status/current - VM status
	gock.New(config.C.URI).
		Get("^/nodes/node1/qemu/101/status/current$").
		Reply(200).
		JSON(`{
    "data": {
        "pid": 1563102,
        "shares": 1000,
        "agent": 1,
        "spice": 1,
        "diskwrite": 1515457024,
        "cpus": 8,
        "ha": {
            "managed": 0
        },
        "maxmem": 2097152000,
        "blockstat": {
            "scsi0": {
                "rd_total_time_ns": 7089432813,
                "flush_total_time_ns": 7442045713,
                "wr_total_time_ns": 65889619830,
                "failed_rd_operations": 0,
                "rd_bytes": 649448960,
                "wr_bytes": 1515457024,
                "unmap_operations": 469,
                "failed_unmap_operations": 0,
                "failed_wr_operations": 0,
                "account_failed": true,
                "invalid_unmap_operations": 0,
                "wr_operations": 157514,
                "rd_operations": 15582,
                "failed_flush_operations": 0,
                "invalid_wr_operations": 0,
                "account_invalid": true,
                "unmap_total_time_ns": 9514953,
                "unmap_merged": 0,
                "timed_stats": [],
                "unmap_bytes": 15973687808,
                "invalid_flush_operations": 0,
                "idle_time_ns": 4427685914,
                "flush_operations": 15494,
                "invalid_rd_operations": 0,
                "wr_highest_offset": 2808696832,
                "rd_merged": 0,
                "wr_merged": 0
            },
            "ide2": {
                "unmap_merged": 0,
                "timed_stats": [],
                "unmap_bytes": 0,
                "invalid_flush_operations": 0,
                "idle_time_ns": 170803536780303,
                "flush_operations": 0,
                "invalid_rd_operations": 0,
                "wr_highest_offset": 0,
                "rd_merged": 0,
                "wr_merged": 0,
                "failed_flush_operations": 0,
                "invalid_wr_operations": 0,
                "account_invalid": true,
                "unmap_total_time_ns": 0,
                "unmap_operations": 0,
                "failed_unmap_operations": 0,
                "failed_wr_operations": 0,
                "account_failed": true,
                "invalid_unmap_operations": 0,
                "wr_operations": 0,
                "rd_operations": 98,
                "rd_total_time_ns": 10689186,
                "flush_total_time_ns": 0,
                "wr_total_time_ns": 0,
                "failed_rd_operations": 0,
                "rd_bytes": 344348,
                "wr_bytes": 0
            }
        },
        "uptime": 170815,
        "cpu": 0.0112815646165076,
        "running-machine": "pc-i440fx-8.0+pve0",
        "balloon": 2097152000,
        "qmpstatus": "running",
        "status": "running",
        "maxdisk": 18467520512,
        "diskread": 649793308,
        "freemem": 887222272,
        "ballooninfo": {
            "actual": 2097152000,
            "max_mem": 2097152000,
            "free_mem": 887222272,
            "major_page_faults": 1811,
            "minor_page_faults": 3803793,
            "mem_swapped_out": 0,
            "mem_swapped_in": 0,
            "total_mem": 2015014912,
            "last_update": 1693252591
        },
        "vmid": 101,
        "balloon_min": 2097152000,
        "mem": 1127792640,
        "proxmox-support": {
            "pbs-dirty-bitmap-savevm": true,
            "pbs-dirty-bitmap": true,
            "query-bitmap-info": true,
            "pbs-masterkey": true,
            "backup-max-workers": true,
            "pbs-dirty-bitmap-migration": true,
            "pbs-library-version": "1.4.0 (UNKNOWN)"
        },
        "running-qemu": "8.0.2",
        "name": "matt",
        "netout": 14139344,
        "netin": 547369168,
        "nics": {
            "tap1001i0": {
                "netout": 14139344,
                "netin": 547369168
            }
        },
        "disk": 0
    }
}`)

	// GET /nodes/{node}/qemu/{vmid}/rrddata - VM RRD data
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/rrddata$").
		MatchParams(map[string]string{
			"timeframe": "hour",
		}).
		Reply(200).
		JSON(`{
    "data": [
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693110660,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693110720,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693110780,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "time": 1693110840,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693110900
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693110960
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693111020,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "time": 1693111080,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693111140,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "time": 1693111200,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693111260,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693111320,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693111380,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693111440
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693111500,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693111560,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693111620,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693111680
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693111740,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693111800,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693111860,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693111920,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693111980,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693112040,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "time": 1693112100,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693112160,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693112220,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "time": 1693112280,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "time": 1693112340,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693112400,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693112460
        },
        {
            "time": 1693112520,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693112580,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693112640
        },
        {
            "time": 1693112700,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693112760,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693112820,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693112880,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693112940
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693113000,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113060,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693113120,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113180
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113240,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "time": 1693113300,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113360
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113420,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "time": 1693113480,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113540,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693113600,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693113660,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693113720,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113780
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693113840
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693113900
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693113960,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693114020,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693114080,
            "maxcpu": 4,
            "maxmem": 16777216000
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693114140,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693114200
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "disk": 0,
            "maxdisk": 68719476736,
            "time": 1693114260
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693114320,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "time": 1693114380,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "time": 1693114440,
            "disk": 0,
            "maxdisk": 68719476736,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693114500,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693114560,
            "maxdisk": 68719476736,
            "disk": 0
        },
        {
            "maxmem": 16777216000,
            "maxcpu": 4,
            "time": 1693114620,
            "disk": 0,
            "maxdisk": 68719476736
        },
        {
            "time": 1693114680,
            "maxdisk": 68719476736,
            "disk": 0,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxdisk": 68719476736,
            "disk": 0,
            "time": 1693114740,
            "maxmem": 16777216000,
            "maxcpu": 4
        },
        {
            "maxcpu": 4,
            "maxmem": 16777216000,
            "time": 1693114800,
            "disk": 0,
            "maxdisk": 68719476736
        }
    ]
}`)

	// GET /nodes/{node}/qemu/{vmid}/rrd - render single-DS PNG, returns filename
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/rrd$").
		Reply(200).
		JSON(`{"data": {"filename": "/var/lib/rrdcached/db/pve2-vm/101.png"}}`)

	// GET /nodes/{node}/qemu/{vmid}/migrate - migration preconditions
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/migrate$").
		Reply(200).
		JSON(`{
    "data": {
        "running": true,
        "has-dbus-vmstate": true,
        "allowed_nodes": ["node2", "node3"],
        "not_allowed_nodes": {
            "node4": {
                "unavailable_storages": ["local-lvm"],
                "blocking-ha-resources": [
                    {"sid": "vm:101", "cause": "node-affinity"}
                ]
            }
        },
        "local_disks": [
            {"volid": "local-lvm:vm-101-disk-0", "size": 34359738368, "cdrom": false, "is_unused": false}
        ],
        "local_resources": [],
        "mapped-resources": [],
        "mapped-resource-info": {},
        "dependent-ha-resources": []
    }
}`)

	// POST /nodes/{node}/qemu/{vmid}/remote_migrate - cross-cluster migration
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/remote_migrate$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00009ABC:0000DEAD:5A3B7C8D:qmremote-migrate:101:root@pam:"}`)

	// POST /nodes/{node}/qemu/{vmid}/clone - Clone VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/101/clone").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// POST /nodes/{node}/qemu/{vmid}/monitor - Access the Virtual Machine monitor
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/101/monitor").
		Reply(200).
		JSON(`{
    "data": "help text"
}`)

	// GET /nodes/{node}/qemu/{vmid}/config - Get VM config
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/config$").
		Reply(200).
		JSON(`{
    "data": {
        "digest": "abc123def456",
        "name": "test-vm",
        "vmid": 100,
        "cores": 2,
        "memory": 2048,
        "sockets": 1,
        "ostype": "l26",
        "boot": "order=scsi0;ide2;net0",
        "scsi0": "local-lvm:vm-100-disk-0,size=32G",
        "ide2": "local:iso/debian-12.iso,media=cdrom",
        "net0": "virtio=BC:24:11:2E:C5:4A,bridge=vmbr0",
        "scsihw": "virtio-scsi-pci",
        "tags": "production;webserver"
    }
}`)

	// POST /nodes/{node}/qemu/{vmid}/config - Update VM config (for tag management)
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/config$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000004:00000004:00000004:qmconfig:100:root@pam:"
}`)

	// PUT /nodes/{node}/qemu/{vmid}/config - synchronous VM config update
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/100/config$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /nodes/{node}/qemu/{vmid}/feature - feature availability check
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/feature$").
		MatchParam("feature", "[a-z]+").
		Reply(200).
		JSON(`{
    "data": {
        "hasFeature": true,
        "nodes": ["node1", "node2"]
    }
}`)

	// POST /nodes/{node}/qemu/{vmid}/dbus-vmstate - control dbus-vmstate helper
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/100/dbus-vmstate$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /nodes/{node}/qemu/{vmid}/status/start - Start VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/start$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000005:00000005:00000005:qmstart:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/stop - Stop VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/stop$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000006:00000006:00000006:qmstop:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/shutdown - Shutdown VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/shutdown$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000007:00000007:00000007:qmshutdown:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/reboot - Reboot VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/reboot$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000008:00000008:00000008:qmreboot:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/reset - Reset VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/reset$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000009:00000009:00000009:qmreset:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/suspend - Pause VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/suspend$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000A:0000000A:0000000A:qmsuspend:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/status/resume - Resume VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/status/resume$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000B:0000000B:0000000B:qmresume:100:root@pam:"
}`)

	// DELETE /nodes/{node}/qemu/{vmid} - Delete VM
	gock.New(config.C.URI).
		Delete("^/nodes/node1/qemu/999$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000C:0000000C:0000000C:qmdestroy:999:root@pam:"
}`)

	// GET /nodes/{node}/qemu/{vmid}/snapshot - List VM snapshots
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/snapshot$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "name": "current",
            "description": "You are here!",
            "snaptime": 0
        },
        {
            "name": "snap1",
            "description": "Before upgrade",
            "snaptime": 1693252591,
            "vmstate": 1,
            "parent": "current"
        },
        {
            "name": "snap2",
            "description": "After upgrade",
            "snaptime": 1693252600,
            "parent": "snap1"
        }
    ]
}`)

	// POST /nodes/{node}/qemu/{vmid}/snapshot - Create VM snapshot
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/snapshot$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000D:0000000D:0000000D:qmsnapshot:100:root@pam:"
}`)

	// POST /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/rollback - Rollback snapshot
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/100/snapshot/snap1/rollback$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000E:0000000E:0000000E:qmrollback:100:root@pam:"
}`)

	// DELETE /nodes/{node}/qemu/{vmid}/snapshot/{snapname} - Delete snapshot
	gock.New(config.C.URI).
		Delete("^/nodes/node1/qemu/100/snapshot/snap2$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0000000F:0000000F:0000000F:qmdelsnapshot:100:root@pam:"
}`)

	// GET /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/config
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/snapshot/snap1/config$").
		Reply(200).
		JSON(`{
    "data": {
        "description": "Before upgrade",
        "parent": "snap0",
        "cores": 4,
        "memory": 8192,
        "name": "snap1"
    }
}`)

	// PUT /nodes/{node}/qemu/{vmid}/snapshot/{snapname}/config
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/100/snapshot/snap1/config$").
		Reply(200).
		JSON(`{"data": null}`)

	// ----- Per-VM firewall (vmid 100) -----

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall$").
		Reply(200).
		JSON(`{
    "data": {
        "rules": [{"pos": 0, "action": "ACCEPT", "type": "in"}],
        "aliases": [{"name": "internal", "cidr": "10.0.0.0/8"}],
        "ipset": [{"name": "blocked", "comment": "blocked clients"}]
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall/rules/0$").
		Reply(200).
		JSON(`{
    "data": {"pos": 0, "action": "ACCEPT", "type": "in", "enable": 1, "comment": "allow http"}
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall/log").
		Reply(200).
		JSON(`{
    "data": [
        [0, "block: IN=eth0 SRC=1.2.3.4"],
        [1, "block: IN=eth0 SRC=5.6.7.8"]
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall/refs").
		Reply(200).
		JSON(`{
    "data": [
        {"type": "alias", "name": "internal", "comment": "vm-local"},
        {"type": "ipset", "name": "blocked"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall/aliases$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "internal", "cidr": "10.0.0.0/8", "comment": "RFC1918"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/100/firewall/aliases$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/firewall/aliases/internal$").
		Reply(200).
		JSON(`{
    "data": {"name": "internal", "cidr": "10.0.0.0/8", "comment": "RFC1918"}
}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/100/firewall/aliases/internal$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/nodes/node1/qemu/100/firewall/aliases/internal$").
		Reply(200).
		JSON(`{"data": null}`)

	// ----- Cloud-init (vmid 100) -----

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/cloudinit$").
		Reply(200).
		JSON(`{
    "data": [
        {"key": "ipconfig0", "value": "ip=10.0.0.10/24,gw=10.0.0.1", "pending": "ip=10.0.0.11/24,gw=10.0.0.1"},
        {"key": "sshkeys", "value": "ssh-rsa AAAA...old", "pending": "ssh-rsa AAAA...new"}
    ]
}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/100/cloudinit$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/cloudinit/dump").
		Reply(200).
		JSON(`{"data": "#cloud-config\nhostname: node1\n"}`)

	// ----- QEMU guest-agent endpoints (vmid 101) -----
	// All synchronous QGA wrappers return {"data": {"result": ...}} except
	// file-read (top-level data) and file-write (null).

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent$").
		Reply(200).
		JSON(`{"data": [
			{"name": "exec"},
			{"name": "ping"},
			{"name": "fsfreeze-status"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent$").
		Reply(200).
		JSON(`{"data": {"result": {"echoed": "ping"}}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-memory-block-info$").
		Reply(200).
		JSON(`{"data": {"result": {"size": 134217728}}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/ping$").
		Reply(200).
		JSON(`{"data": {"result": {}}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-time$").
		Reply(200).
		JSON(`{"data": {"result": 1715600000000000000}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-timezone$").
		Reply(200).
		JSON(`{"data": {"result": {"zone": "UTC", "offset": 0}}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-users$").
		Reply(200).
		JSON(`{"data": {"result": [
			{"user": "root", "login-time": 1715500000.123},
			{"user": "luther", "domain": "WORKGROUP", "login-time": 1715500050.5}
		]}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-vcpus$").
		Reply(200).
		JSON(`{"data": {"result": [
			{"logical-id": 0, "online": true, "can-offline": false},
			{"logical-id": 1, "online": true, "can-offline": true}
		]}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-fsinfo$").
		Reply(200).
		JSON(`{"data": {"result": [
			{"name": "sda1", "mountpoint": "/", "type": "ext4", "used-bytes": 1234567890, "total-bytes": 53687091200, "disk": [{"serial": "drive-scsi0", "bus-type": "scsi", "bus": 0, "unit": 0, "target": 0, "dev": "/dev/sda1", "pci-controller": {"domain": 0, "bus": 0, "slot": 5, "function": 0}}]}
		]}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-memory-blocks$").
		Reply(200).
		JSON(`{"data": {"result": [
			{"phys-index": 0, "online": true, "can-offline": false},
			{"phys-index": 1, "online": true, "can-offline": true}
		]}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/info$").
		Reply(200).
		JSON(`{"data": {"result": {"version": "7.2.0", "supported_commands": [
			{"name": "guest-ping", "enabled": true, "success-response": true},
			{"name": "guest-exec", "enabled": true, "success-response": true}
		]}}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-freeze$").
		Reply(200).
		JSON(`{"data": {"result": 3}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-thaw$").
		Reply(200).
		JSON(`{"data": {"result": 3}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-status$").
		Reply(200).
		JSON(`{"data": {"result": "thawed"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fstrim$").
		Reply(200).
		JSON(`{"data": {"result": {"/": {"trimmed": 1048576, "minimum": 0}}}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/shutdown$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-disk$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-hybrid$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-ram$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/file-read$").
		Reply(200).
		JSON(`{"data": {"content": "hello world\n", "truncated": 0}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/file-write$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /nodes/{node}/qemu/{vmid}/spiceproxy
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/spiceproxy$").
		Reply(200).
		JSON(`{
    "data": {
        "type": "spice",
        "host": "node1.example.com",
        "port": "61024",
        "tls-port": "61025",
        "password": "secret-ticket",
        "proxy": "http://proxy.example.com",
        "title": "VM 101",
        "host-subject": "OU=PVE Cluster Node,O=Proxmox VE,CN=node1",
        "ca": "-----BEGIN CERTIFICATE-----\nMIIB...==\n-----END CERTIFICATE-----",
        "delete-this-file": "1",
        "secure-attention": "Ctrl+Alt+Ins",
        "release-cursor": "Ctrl+Alt+R",
        "toggle-fullscreen": "Shift+F11"
    }
}`)

	// GET /nodes/{node}/qemu/{vmid} — per-VM directory index (vmdiridx)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100$").
		Reply(200).
		JSON(`{
    "data": [
        {"subdir": "config"},
        {"subdir": "status"},
        {"subdir": "snapshot"},
        {"subdir": "firewall"},
        {"subdir": "agent"},
        {"subdir": "rrd"},
        {"subdir": "rrddata"},
        {"subdir": "monitor"},
        {"subdir": "termproxy"},
        {"subdir": "vncproxy"},
        {"subdir": "vncwebsocket"},
        {"subdir": "spiceproxy"},
        {"subdir": "feature"},
        {"subdir": "clone"},
        {"subdir": "move_disk"},
        {"subdir": "migrate"},
        {"subdir": "resize"},
        {"subdir": "sendkey"},
        {"subdir": "unlink"},
        {"subdir": "template"},
        {"subdir": "cloudinit"},
        {"subdir": "pending"},
        {"subdir": "mtunnel"},
        {"subdir": "mtunnelwebsocket"}
    ]
}`)

	// GET /nodes/{node}/qemu/{vmid}/status — status directory index (vmcmdidx)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/status$").
		Reply(200).
		JSON(`{
    "data": [
        {"subdir": "current"},
        {"subdir": "start"},
        {"subdir": "stop"},
        {"subdir": "reset"},
        {"subdir": "shutdown"},
        {"subdir": "suspend"},
        {"subdir": "resume"},
        {"subdir": "reboot"}
    ]
}`)

	// GET /nodes/{node}/qemu/{vmid}/snapshot/{snapname} — snapshot directory index (snapshot_cmd_idx)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/100/snapshot/snap1$").
		Reply(200).
		JSON(`{
    "data": [
        {"subdir": "config"},
        {"subdir": "rollback"}
    ]
}`)

	// POST /nodes/{node}/qemu/{vmid}/mtunnel — open migration tunnel
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/100/mtunnel$").
		Reply(200).
		JSON(`{
    "data": {
        "socket": "/run/qemu-server/100.mtunnel",
        "ticket": "PVEMTUNNELTICKET:abc123",
        "upid": "UPID:node1:00001234:00005678:00009ABC:qmtunnel:100:root@pam:"
    }
}`)

	// ===== Additional fixtures for coverage tests =====

	// POST /nodes/{node}/qemu/{vmid}/status/suspend with todisk for Hibernate
	// (uses vmid 101 to avoid colliding with Pause's 100 mock above).
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/status/suspend$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000A001:0000A001:0000A001:qmsuspend:101:root@pam:"}`)

	// POST /nodes/{node}/qemu/{vmid}/migrate - Migrate VM
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/migrate$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000B001:0000B001:0000B001:qmigrate:101:root@pam:"}`)

	// PUT /nodes/{node}/qemu/{vmid}/resize - ResizeDisk
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/101/resize$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000B002:0000B002:0000B002:qmresize:101:root@pam:"}`)

	// PUT /nodes/{node}/qemu/{vmid}/unlink - UnlinkDisk
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/101/unlink$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000B003:0000B003:0000B003:qmunlink:101:root@pam:"}`)

	// POST /nodes/{node}/qemu/{vmid}/move_disk - MoveDisk
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/move_disk$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000B004:0000B004:0000B004:qmmove:101:root@pam:"}`)

	// POST /nodes/{node}/qemu/{vmid}/template - ConvertToTemplate
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/template$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000B005:0000B005:0000B005:qmtemplate:101:root@pam:"}`)

	// GET /nodes/{node}/qemu/{vmid}/pending - Pending config
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/pending$").
		Reply(200).
		JSON(`{"data": [
			{"key": "cores", "value": 2, "pending": 4},
			{"key": "memory", "value": 2048}
		]}`)

	// PUT /nodes/{node}/qemu/{vmid}/sendkey - SendKey
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/101/sendkey$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /nodes/{node}/qemu/{vmid}/termproxy - TermProxy
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/termproxy$").
		Reply(200).
		JSON(`{"data": {"user": "root@pam", "ticket": "PVEVNC:ABC123", "upid": "UPID:node1:0000C001:0000C001:0000C001:vncproxy:101:root@pam:", "port": "5901"}}`)

	// POST /nodes/{node}/qemu/{vmid}/vncproxy - VNCProxy
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/vncproxy").
		Reply(200).
		JSON(`{"data": {"user": "root@pam", "cert": "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----", "ticket": "PVEVNC:DEF456", "upid": "UPID:node1:0000C002:0000C002:0000C002:vncproxy:101:root@pam:", "port": "5902"}}`)

	// ----- Per-VM legacy firewall helpers (vmid 101) -----

	// GET /firewall/ipset - list IPSets
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/firewall/ipset$").
		Reply(200).
		JSON(`{"data": [{"name": "blocked", "comment": "blocked clients", "digest": "abc"}]}`)

	// POST /firewall/ipset - create IPSet
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/firewall/ipset$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /firewall/ipset/{name} - list IPSet entries
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/firewall/ipset/blocked$").
		Reply(200).
		JSON(`{"data": [{"cidr": "10.1.2.3", "comment": "client", "digest": "abc"}]}`)

	// POST /firewall/ipset/{name} - add entry
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/firewall/ipset/blocked$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /firewall/ipset/{name}
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/qemu/101/firewall/ipset/blocked$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET single IPSet entry
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/firewall/ipset/blocked/10\\.1\\.2\\.3$").
		Reply(200).
		JSON(`{"data": {"cidr": "10.1.2.3", "comment": "client", "digest": "abc"}}`)

	// PUT update IPSet entry
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/101/firewall/ipset/blocked/10\\.1\\.2\\.3$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE single IPSet entry
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/qemu/101/firewall/ipset/blocked/10\\.1\\.2\\.3$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /firewall/options - FirewallOptionGet
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/firewall/options$").
		Reply(200).
		JSON(`{"data": {"enable": 1, "policy_in": "ACCEPT", "policy_out": "ACCEPT"}}`)

	// PUT /firewall/options - FirewallOptionSet
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/qemu/101/firewall/options$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /firewall/rules - FirewallRules (list)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/firewall/rules$").
		Reply(200).
		JSON(`{"data": [
			{"pos": 0, "type": "in", "action": "ACCEPT", "enable": 1, "comment": "allow http"},
			{"pos": 1, "type": "out", "action": "DROP", "enable": 0}
		]}`)

	// POST /firewall/rules - NewFirewallRule
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/firewall/rules$").
		Reply(200).
		JSON(`{"data": null}`)

	// ----- QGA endpoints for vmid 101 not yet covered -----

	// GET /agent/get-host-name - AgentGetHostName
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-host-name$").
		Reply(200).
		JSON(`{"data": {"result": {"host-name": "vm-101.example.com"}}}`)

	// GET /agent/network-get-interfaces - AgentGetNetworkIFaces (includes lo to test filtering)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/network-get-interfaces$").
		Reply(200).
		JSON(`{"data": {"result": [
			{"name": "lo", "hardware-address": "00:00:00:00:00:00", "ip-addresses": []},
			{"name": "eth0", "hardware-address": "BC:24:11:2E:C5:4A", "ip-addresses": [{"ip-address": "10.0.0.10", "ip-address-type": "ipv4", "prefix": 24}]}
		]}}`)

	// POST /agent/exec - AgentExec
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/exec$").
		Reply(200).
		JSON(`{"data": {"pid": 1234}}`)

	// GET /agent/exec-status?pid=... - AgentExecStatus (exited=1)
	// Unlike most agent endpoints AgentExecStatus does NOT wrap in a "result"
	// envelope — the struct sits directly under "data".
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/exec-status$").
		Reply(200).
		JSON(`{"data": {"exited": 1, "exitcode": 0, "out-data": "hello\n", "out-truncated": false}}`)

	// GET /agent/get-osinfo - AgentOsInfo
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-osinfo$").
		Reply(200).
		JSON(`{"data": {"result": {"id": "debian", "name": "Debian GNU/Linux", "pretty-name": "Debian GNU/Linux 12 (bookworm)", "version": "12 (bookworm)", "version-id": "12", "machine": "x86_64", "kernel-release": "6.1.0", "kernel-version": "#1 SMP"}}}`)

	// POST /agent/set-user-password - AgentSetUserPassword
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/set-user-password$").
		Reply(200).
		JSON(`{"data": null}`)

	// vmid 102: POST /config for tag mutations (Add/RemoveTag).
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/102/config$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000A002:0000A002:0000A002:qmconfig:102:root@pam:"}`)

	// ----- deleteCloudInitISO happy-path scaffolding (cinode + vmid 503) -----
	// Provide a self-contained node ("cinode") with a single iso-capable
	// storage that already contains user-data-503.iso. Using a dedicated
	// node sidesteps registration-order collisions with nodes.go's persisted
	// /nodes/node1/storage list.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/status$").
		Reply(200).
		JSON(`{"data": {"name": "cinode", "status": "online"}}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/storage$").
		Reply(200).
		JSON(`{"data": [
			{"storage": "cidata", "type": "dir", "enabled": 1, "content": "iso", "active": 1}
		]}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/storage/cidata$").
		Reply(200).
		JSON(`{"data": {"name": "cidata", "type": "dir", "enabled": 1, "content": "iso", "active": 1}}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/storage/cidata/status$").
		Reply(200).
		JSON(`{"data": {"name": "cidata", "type": "dir", "enabled": 1, "content": "iso", "active": 1}}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/storage/cidata/content$").
		Reply(200).
		JSON(`{"data": [
			{"volid": "cidata:iso/user-data-503.iso", "format": "iso", "size": 374784}
		]}`)
	// Storage.ISO(name) issues a GET on the individual content endpoint.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/storage/cidata/content/cidata:iso/user-data-503\\.iso$").
		Reply(200).
		JSON(`{"data": {"path": "/var/lib/vz/template/iso/user-data-503.iso", "volid": "cidata:iso/user-data-503.iso", "format": "iso", "size": 374784}}`)
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/cinode/storage/cidata/content/cidata:iso/user-data-503\\.iso$").
		Reply(200).
		JSON(`{"data": "UPID:cinode:0000D001:0000D001:0000D001:imgdel:cidata:root@pam:"}`)
	// Task status mock for the delete worker — return "stopped" immediately
	// so task.WaitFor returns nil on the first poll.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/cinode/tasks/UPID:cinode:0000D001:0000D001:0000D001:imgdel:cidata:root@pam:/status$").
		Reply(200).
		JSON(`{"data": {"status": "stopped", "exitstatus": "OK", "node": "cinode", "type": "imgdel", "id": "cidata", "user": "root@pam", "upid": "UPID:cinode:0000D001:0000D001:0000D001:imgdel:cidata:root@pam:"}}`)
}
