package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func nodes() {
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

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes$").
		Reply(200).
		JSON(`{
  "data": [
    {
      "uptime": 2236708,
      "level": "",
      "maxmem": 33568288768,
      "disk": 2310930432,
      "node": "node1",
      "maxdisk": 940743983104,
      "mem": 11508809728,
      "ssl_fingerprint": "80:D4:F2:DF:64:95:CD:8D:A0:82:82:AC:48:BA:C0:7A:1B:6B:87:8B:FE:B9:83:1C:95:4E:79:58:77:99:69:F5",
      "status": "online",
      "type": "node",
      "cpu": 0.00348605577689243,
      "id": "node/node1",
      "maxcpu": 12
    },
    {
      "level": "",
      "uptime": 6256882,
      "node": "node2",
      "maxdisk": 482721529856,
      "maxmem": 16651751424,
      "disk": 2303721472,
      "ssl_fingerprint": "17:F1:B6:52:8B:0C:22:4A:97:1F:B2:F2:90:3D:29:0A:D0:DF:BE:0E:76:5A:B5:EC:F6:2E:6E:8F:60:E6:C5:C0",
      "status": "online",
      "mem": 1838854144,
      "maxcpu": 4,
      "cpu": 0.00722831505483549,
      "type": "node",
      "id": "node/node2"
    },
    {
      "maxdisk": 482717728768,
      "node": "node3",
      "disk": 2315386880,
      "maxmem": 16668868608,
      "level": "",
      "uptime": 6258488,
      "maxcpu": 4,
      "id": "node/node3",
      "cpu": 0.00821557582405153,
      "type": "node",
      "status": "online",
      "ssl_fingerprint": "1D:56:94:B4:75:4B:5C:33:46:DD:14:38:6C:EC:6E:12:A8:F0:66:64:5E:F2:40:F7:60:2A:C0:9F:BF:6C:51:3C",
      "mem": 1858961408
    },
    {
      "maxmem": 65919561728,
      "disk": 9992273920,
      "node": "node4",
      "maxdisk": 951055024128,
      "uptime": 6257222,
      "level": "",
      "cpu": 0.00748876684972541,
      "type": "node",
      "id": "node/node4",
      "maxcpu": 8,
      "mem": 2268295168,
      "ssl_fingerprint": "0D:78:80:CD:64:8E:96:E5:31:87:1C:45:3C:62:93:2F:23:4C:D5:02:42:FE:C8:40:DC:AF:3D:2A:F8:B4:F6:CE",
      "status": "online"
    }
  ]
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/network$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "priority": 3,
            "method6": "manual",
            "exists": 1,
            "method": "manual",
            "active": 1,
            "families": [
                "inet"
            ],
            "type": "eth",
            "iface": "enp0s31f6"
        },
        {
            "priority": 4,
            "cidr": "192.168.5.1/24",
            "active": 1,
            "netmask": "24",
            "bridge_ports": "enp0s31f6",
            "method6": "manual",
            "autostart": 1,
            "bridge_fd": "0",
            "method": "static",
            "gateway": "192.168.1.1",
            "iface": "vmbr0",
            "type": "bridge",
            "families": [
                "inet"
            ],
            "address": "192.168.5.1",
            "bridge_stp": "off"
        }
    ]
}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/network/vmbr0$").
		Reply(200).
		JSON(`{
    "data": {
        "method6": "manual",
        "autostart": 1,
        "method": "static",
        "bridge_fd": "0",
        "gateway": "192.168.1.1",
        "type": "bridge",
        "address": "192.168.5.1",
        "bridge_stp": "off",
        "families": [
            "inet"
        ],
        "cidr": "192.168.5.1/24",
        "priority": 4,
        "netmask": "24",
        "active": 1,
        "bridge_ports": "enp0s31f6"
    }
}`)

}
