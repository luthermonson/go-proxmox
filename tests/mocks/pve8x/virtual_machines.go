package pve8x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func virtualMachines() {
	// GET /nodes/{node}/qemu/{vmid}/status/current - VM status
	gock.New(config.C.URI).
		Get("^/nodes/node1/qemu/101/status/current$").
		Reply(200).
		JSON(`{
    "data": {
        "pid": 1563102,
        "shares": 1000,
        "agent": 1,
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

	// POST /nodes/{node}/qemu/{vmid}/clone - Clone VM
	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/101/clone").
		Reply(200).
		JSON(`{
    "data": null
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
}
