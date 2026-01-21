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

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/status$").
		Reply(200).
		JSON(`{
    "data": {
        "idle": 0,
        "cpu": 0.00260552371026576,
        "ksm": {
            "shared": 0
        },
        "swap": {
            "total": 0,
            "free": 0,
            "used": 0
        },
        "pveversion": "pve-manager/7.4-16/0f39f621",
        "wait": 0,
        "uptime": 2501631,
        "kversion": "Linux 5.15.108-1-pve #1 SMP PVE 5.15.108-2 (2023-07-20T10:06Z)",
        "cpuinfo": {
            "mhz": "3400.000",
            "model": "Intel(R) Core(TM) i7-6700 CPU @ 3.40GHz",
            "sockets": 1,
            "flags": "fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx smx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault invpcid_single pti tpr_shadow vnmi flexpriority ept vpid ept_ad fsgsbase tsc_adjust bmi1 hle avx2 smep bmi2 erms invpcid rtm mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp",
            "hvm": "1",
            "cores": 4,
            "user_hz": 100,
            "cpus": 8
        },
        "loadavg": [
            "0.00",
            "0.00",
            "0.00"
        ],
        "memory": {
            "total": 65919459328,
            "free": 57824059392,
            "used": 8095399936
        },
        "rootfs": {
            "total": 948338819072,
            "free": 937851224064,
            "used": 10487595008,
            "avail": 937851224064
        }
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/doesntexist/status$").
		Reply(500).
		JSON(`{
    "data": null
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/version$").
		Reply(200).
		JSON(`{
    "data": {
        "release": "7.4",
        "version": "7.4-16",
        "repoid": "0f39f621"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node2/network$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "bridge_vlan_aware": 1,
            "method6": "static",
            "bridge_vids": "2-4094",
            "address": "192.168.1.10",
            "bridge_stp": "off",
            "families": [
                "inet",
                "inet6"
            ],
            "mtu": "1500",
            "bridge_ports": "enp8s0",
            "iface": "vmbr2",
            "netmask": "24",
            "method": "static",
            "type": "bridge",
            "bridge_fd": "0",
            "cidr6": "fd66:5ac3:eeaf:3200::10/64",
            "active": 1,
            "cidr": "192.168.1.10/24",
            "address6": "fd66:5ac3:eeaf:3200::10",
            "comments": "comment\n",
            "autostart": 1,
            "priority": 9,
            "netmask6": "64"
        },
        {
            "active": 1,
            "cidr6": "fd66:5ac3:eeaf::10/64",
            "address6": "fd66:5ac3:eeaf::10",
            "comments": "comment\n",
            "autostart": 1,
            "priority": 11,
            "cidr": "192.168.0.10/24",
            "netmask6": "64",
            "vlan-raw-device": "vmbr0",
            "exists": null,
            "method6": "static",
            "address": "192.168.0.10",
            "families": [
                "inet",
                "inet6"
            ],
            "iface": "vmbr0.2",
            "netmask": "24",
            "method": "static",
            "vlan-id": "2",
            "type": "vlan"
        }
    ]
}`)

	// LXC
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/interfaces").
		Reply(200).
		JSON(`{
		"data": [
				{
						"inet":"127.0.0.1/8",
						"hwaddr":"00:00:00:00:00:00",
						"name":"lo",
						"inet6":"::1/128"
				},
				{
						"inet6":"fe80::be24:11ff:fe89:6707/64",
						"name":"eth0",
						"hwaddr":"bc:24:11:89:67:07",
						"inet":"192.168.3.95/22"
				}
		]
}`)

	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc$").
		Reply(200).
		JSON(`{
"data": [{"cpu":0,"cpus":1,"disk":640397312,"diskread":273694720,"diskwrite":200982528,"maxdisk":8350298112,"maxmem":536870912,"maxswap":536870912,"mem":34304000,"name":"test","netin":94558593,"netout":1618542,"pid":248173,"status":"running","swap":0,"type":"lxc","uptime":919760,"vmid":"106"},{"cpu":0,"cpus":1,"disk":639303680,"diskread":283123712,"diskwrite":201687040,"maxdisk":8350298112,"maxmem":536870912,"maxswap":536870912,"mem":34508800,"name":"zort","netin":94560801,"netout":1619838,"pid":248045,"status":"running","swap":0,"type":"lxc","uptime":919761,"vmid":"105"},{"cpu":0,"cpus":1,"disk":0,"diskread":0,"diskwrite":0,"maxdisk":8589934592,"maxmem":536870912,"maxswap":536870912,"mem":0,"name":"test-container","netin":0,"netout":0,"status":"stopped","swap":0,"template":1,"type":"lxc","uptime":0,"vmid":"101"}]
}`)

	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/status/current").
		Reply(200).
		JSON(`{
    "data": {
        "cpu":0,
        "cpus":2,
        "disk":0,
        "diskread":0,
        "diskwrite":0,
        "ha":{"managed":0},
        "maxdisk":8589934592,
        "maxmem":536870912,
        "maxswap":536870912,
        "mem":0,
        "name":"test-container",
        "netin":0,
        "netout":0,
        "status":"stopped",
        "swap":0,
        "template":1,
        "tags":"tag1;tag2",
        "type":"lxc",
        "uptime":0,
        "vmid":101
    }
}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/clone").
		Reply(200).
		JSON(`{
    "data": null
}`)

	gock.New(config.C.URI).
		Delete("^/nodes/node1/lxc/101").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzdestroy:101:root@pam:"}`)

	gock.New(config.C.URI).
		Put("^/nodes/node1/lxc/101/config").
		Reply(200).
		JSON(`{"data": "null"}`)

		// Used for ContainerConfig
	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/config").
		Reply(200).
		JSON(`{"data":
{
   "arch" : "amd64",
   "cores" : 2,
   "digest" : "5911baea29f4c0073fb063fd9ab29e75892832bf",
   "features" : "fuse=1,mknod=1,nesting=1",
   "hostname" : "test-container",
   "lxc" : [
      [
         "lxc.cgroup2.devices.allow",
         "c 226:0 rwm"
      ],
      [
         "lxc.cgroup2.devices.allow",
         "c 226:128 rwm"
      ],
      [
         "lxc.cgroup2.devices.allow",
         "c 29:0 rwm"
      ],
      [
         "lxc.mount.entry",
         "/dev/dri dev/dri none bind,optional,create=dir"
      ],
      [
         "lxc.mount.entry",
         "/dev/fb0 dev/fb0 none bind,optional,create=file"
      ]
   ],
   "memory" : 4096,
   "mp0" : "/mnt/foo/bar,mp=/storage",
   "net0" : "name=eth0,bridge=vmbr0,firewall=1,hwaddr=5D:CF:BD:B2:C5:39,ip=dhcp,type=veth",
   "onboot" : 1,
   "ostype" : "debian",
   "rootfs" : "vmstore:subvol-101-disk-0,size=30G",
   "swap" : 512,
   "tags" : "tag1;tag2"
}}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/start").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzstart:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/stop").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzstop:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/suspend").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzsuspend:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/reboot").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzreboot:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/resume").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzresume:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/status/shutdown").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzshutdown:101:root"}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/template").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/snapshot").
		Reply(200).
		JSON(`{
    "data": [
        {
            "description": "description1",
            "name":"snapshot1",
            "snaptime":1709753281
        },
        {
            "description":"description2",
            "name":"snapshot2",
            "parent":"parent1",
            "snaptime":1709753290
        },
        {
            "description":"You are here!",
            "digest":"e2f5f35c85b2ca35e5f9ab789436b25c1d71cbad",
            "name":"current",
            "parent":"parent2",
            "running":1
        }
    ]
}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/snapshot").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0031B740:0645340C:23E5BA99:vzsnapshot:101:root"
}`)

	gock.New(config.C.URI).
		Delete("^/nodes/node1/lxc/101/snapshot/snapshot1").
		Reply(200).
		JSON(`{"data": "UPID:node1:0031B740:0645340C:23E5BA99:vzrmsnapshot:101:root"}`)

	gock.New(config.C.URI).
		Get("^/nodes/node1/lxc/101/snapshot/snapshot1").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0031B740:0645340C:23E5BA99:vzsnapshot:101:root"
}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/lxc/101/snapshot/snapshot1/rollback").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:0031B740:0645340C:23E5BA99:vzrollback:101:root"
}`)

	gock.New(config.C.URI).
		Post("^/nodes/node1/qemu/101/clone").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// GET /nodes/{node}/report - Get node report
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/report$").
		Reply(200).
		JSON(`{
    "data": "pve-manager: 7.7-1\nkernel: 6.5.0-1-pve\nproxmox-ve: 7.7-1\nqemu-server: 7.7-1\nlxc-pve: 5.0.0-1"
}`)

	// POST /nodes/{node}/termproxy - Create terminal proxy
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/termproxy$").
		Reply(200).
		JSON(`{
    "data": {
        "port": 5900,
        "ticket": "PVE:termproxy:12345678",
        "upid": "UPID:node1:00001234:00005678:12345678:termproxy:root@pam:",
        "user": "root@pam"
    }
}`)

	// GET /nodes/{node}/aplinfo - List appliances
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/aplinfo$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "template": "ubuntu-22.04-standard",
            "type": "lxc",
            "package": "ubuntu-22.04-standard_22.04-1_amd64.tar.zst",
            "os": "ubuntu",
            "version": "22.04",
            "headline": "Ubuntu 22.04 LTS",
            "infopage": "https://pve.proxmox.com/wiki/Linux_Container",
            "description": "Ubuntu 22.04 LTS (Jammy Jellyfish) standard system",
            "section": "system"
        },
        {
            "template": "debian-12-standard",
            "type": "lxc",
            "package": "debian-12-standard_12.0-1_amd64.tar.zst",
            "os": "debian",
            "version": "12.0",
            "headline": "Debian 12 (Bookworm)",
            "infopage": "https://pve.proxmox.com/wiki/Linux_Container",
            "description": "Debian 12 (Bookworm) standard system",
            "section": "system"
        }
    ]
}`)

	// POST /nodes/{node}/aplinfo - Download appliance
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/aplinfo$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:download:root@pam:"
}`)

	// GET /nodes/{node}/storage - List storages
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "storage": "local",
            "content": "images,rootdir,vztmpl,backup,iso,snippets",
            "type": "dir",
            "active": 1,
            "avail": 50000000000,
            "used": 10000000000,
            "total": 60000000000,
            "enabled": 1,
            "shared": 0
        },
        {
            "storage": "local-lvm",
            "content": "images,rootdir",
            "type": "lvmthin",
            "active": 1,
            "avail": 100000000000,
            "used": 20000000000,
            "total": 120000000000,
            "enabled": 1,
            "shared": 0
        }
    ]
}`)

	// GET /nodes/{node}/storage/{storage}/status - Get storage status
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/status$").
		Reply(200).
		JSON(`{
    "data": {
        "storage": "local",
        "content": "images,rootdir,vztmpl,backup,iso,snippets",
        "type": "dir",
        "active": 1,
        "avail": 50000000000,
        "used": 10000000000,
        "total": 60000000000,
        "enabled": 1,
        "shared": 0
    }
}`)

	// GET /nodes/{node}/storage/{storage}/content - List storage content
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content").
		MatchParam("content", "vztmpl").
		Reply(200).
		JSON(`{
    "data": [
        {
            "volid": "local:vztmpl/ubuntu-22.04-standard_22.04-1_amd64.tar.zst",
            "content": "vztmpl",
            "format": "tgz",
            "size": 123456789,
            "ctime": 1234567890
        },
        {
            "volid": "local:vztmpl/debian-12-standard_12.0-1_amd64.tar.zst",
            "content": "vztmpl",
            "format": "tgz",
            "size": 98765432,
            "ctime": 1234567890
        }
    ]
}`)

	// POST /nodes/{node}/storage/{storage}/download-url - Download from URL
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/storage/local/download-url$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:download:root@pam:"
}`)

	// GET /nodes/{node}/firewall/options - Get firewall options
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/firewall/options$").
		Reply(200).
		JSON(`{
    "data": {
        "enable": true,
        "log_level_in": "info",
        "log_level_out": "info",
        "ndp": true,
        "nf_conntrack_allow_invalid": false,
        "nf_conntrack_max": 262144,
        "nf_conntrack_tcp_timeout_established": 432000,
        "nosmurfs": true,
        "protection_synflood": false,
        "protection_synflood_burst": 1000,
        "protection_synflood_rate": 200,
        "smurf_log_level": "info",
        "tcp_flags_log_level": "nolog",
        "tcpflags": false
    }
}`)

	// PUT /nodes/{node}/firewall/options - Update firewall options
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/firewall/options$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// GET /nodes/{node}/firewall/rules - Get firewall rules
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/firewall/rules$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "pos": 0,
            "type": "in",
            "action": "ACCEPT",
            "enable": 1,
            "iface": "vmbr0",
            "source": "192.168.1.0/24",
            "dest": "192.168.1.100",
            "proto": "tcp",
            "dport": "22",
            "comment": "Allow SSH from LAN"
        },
        {
            "pos": 1,
            "type": "in",
            "action": "DROP",
            "enable": 1,
            "proto": "tcp",
            "dport": "22",
            "comment": "Block all other SSH"
        }
    ]
}`)

	// POST /nodes/{node}/firewall/rules - Create firewall rule
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/firewall/rules$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// PUT /nodes/{node}/firewall/rules/{pos} - Update firewall rule
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/firewall/rules/[0-9]+$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// DELETE /nodes/{node}/firewall/rules/{pos} - Delete firewall rule
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/firewall/rules/[0-9]+$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// GET /nodes/{node}/certificates/info - Get certificates
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/certificates/info$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "filename": "/etc/pve/nodes/node1/pve-ssl.pem",
            "fingerprint": "80:D4:F2:DF:64:95:CD:8D:A0:82:82:AC:48:BA:C0:7A:1B:6B:87:8B:FE:B9:83:1C:95:4E:79:58:77:99:69:F5",
            "issuer": "Proxmox Virtual Environment",
            "notafter": 1735689600,
            "notbefore": 1704153600,
            "subject": "node1.example.com",
            "san": [
                "DNS:node1",
                "DNS:node1.example.com",
                "IP:192.168.1.100"
            ],
            "pem": "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKZx...\n-----END CERTIFICATE-----"
        }
    ]
}`)

	// POST /nodes/{node}/certificates/custom - Upload custom certificate
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/certificates/custom$").
		Reply(200).
		JSON(`{
    "data": {
        "filename": "/etc/pve/nodes/node1/pve-ssl.pem",
        "fingerprint": "AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90",
        "issuer": "Custom CA",
        "notafter": 1767225600,
        "notbefore": 1735689600,
        "subject": "node1.example.com"
    }
}`)

	// DELETE /nodes/{node}/certificates/custom - Delete custom certificate
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/certificates/custom$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// POST /nodes/{node}/vzdump - Backup VMs
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/vzdump$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:vzdump:root@pam:"
}`)

	// GET /nodes/{node}/vzdump/extractconfig - Extract backup config
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/vzdump/extractconfig").
		Reply(200).
		JSON(`{
    "data": "cores: 2\nmemory: 2048\nostype: debian\nrootfs: local-lvm:vm-100-disk-0,size=8G\nnet0: name=eth0,bridge=vmbr0,ip=dhcp"
}`)
}
