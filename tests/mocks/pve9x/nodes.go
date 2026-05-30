package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func nodes() {
	// GET /nodes - List all nodes
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

	// GET /nodes/node1/version - Node version
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/version$").
		Reply(200).
		JSON(`{
    "data": {
        "release": "9.1",
        "version": "9.1-1",
        "repoid": "9a1b2c3d"
    }
}`)

	// GET /nodes/doesntexist/status - Error case for non-existent node
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/doesntexist/status$").
		Reply(500).
		JSON(`{
    "data": null
}`)

	// GET /nodes/node1/status - Node status
	gock.New(config.C.URI).
		Get("^/nodes/node1/status$").
		Reply(200).
		JSON(`{
    "data": {
        "cpu": 0.0123456789,
        "memory": {
            "used": 2147483648,
            "total": 8589934592,
            "free": 6442450944
        },
        "uptime": 86400,
        "loadavg": ["0.15", "0.25", "0.30"],
        "kversion": "Linux 6.11.0-1-pve",
        "pveversion": "pve-manager/9.1-1/9a1b2c3d",
        "cpuinfo": {
            "model": "Intel(R) Xeon(R) CPU E5-2680 v4",
            "cores": 4,
            "cpus": 8,
            "sockets": 1
        }
    }
}`)

	// GET /nodes/node2/status - Node status for node2
	gock.New(config.C.URI).
		Get("^/nodes/node2/status$").
		Reply(200).
		JSON(`{
    "data": {
        "cpu": 0.0234567890,
        "memory": {
            "used": 4294967296,
            "total": 17179869184,
            "free": 12884901888
        },
        "uptime": 172800,
        "loadavg": ["0.25", "0.35", "0.40"],
        "kversion": "Linux 6.11.0-1-pve",
        "pveversion": "pve-manager/9.1-1/9a1b2c3d",
        "cpuinfo": {
            "model": "Intel(R) Xeon(R) CPU E5-2680 v4",
            "cores": 8,
            "cpus": 16,
            "sockets": 2
        }
    }
}`)

	// GET /nodes/{node}/network/{iface} - Get specific network interface
	gock.New(config.C.URI).
		Get("^/nodes/node1/network/vmbr0$").
		Reply(200).
		JSON(`{
    "data": {
        "iface": "vmbr0",
        "type": "bridge",
        "method": "static",
        "method6": "manual",
        "address": "192.168.1.100",
        "netmask": "24",
        "cidr": "192.168.1.100/24",
        "gateway": "192.168.1.1",
        "bridge_ports": "eno1",
        "bridge_stp": "off",
        "bridge_fd": "0",
        "autostart": 1,
        "active": 1,
        "priority": 10,
        "families": ["inet"]
    }
}`)

	// POST /nodes/{node}/network — create new interface, echoes a body back.
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/network$").
		Reply(200).
		JSON(`{
    "data": {
        "iface": "vmbr99",
        "type": "bridge",
        "method": "static",
        "method6": "manual",
        "address": "10.0.0.1",
        "netmask": "24",
        "cidr": "10.0.0.1/24",
        "autostart": 1
    }
}`)

	// PUT /nodes/{node}/network — reload pending changes; returns UPID.
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/network$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00009999:00009999:00009999:srvreload:networking:root@pam:"}`)

	// PUT /nodes/{node}/network/{iface} — update interface.
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/network/vmbr99$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /nodes/{node}/network/{iface} — drop interface.
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/network/vmbr99$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00009998:00009998:00009998:srvreload:networking:root@pam:"}`)

	// POST /nodes/{node}/qemu — create new VM. Returns a UPID. Lives on the
	// nodes group because the (*Node).NewVirtualMachine creates against the
	// node, not a specific VMID.
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00007777:00007777:00007777:qmcreate:300:root@pam:"}`)

	// POST /nodes/{node}/lxc — create new container.
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/lxc$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00007778:00007778:00007778:vzcreate:300:root@pam:"}`)

	// GET /cluster/sdn/ipams/{node}/status — IPAM listing scoped to node.
	gock.New(config.C.URI).
		Persist().
		Get("^/cluster/sdn/ipams/node1/status$").
		Reply(200).
		JSON(`{"data": [
			{"hostname": "vm100", "ip": "10.0.0.10", "mac": "aa:bb:cc:dd:ee:01", "subnet": "10.0.0.0/24", "vmid": "100", "vnet": "vnet1", "zone": "zone1"},
			{"hostname": "vm101", "ip": "10.0.0.11", "mac": "aa:bb:cc:dd:ee:02", "subnet": "10.0.0.0/24", "vmid": "101", "vnet": "vnet1", "zone": "zone1"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/network").
		ParamPresent("type").
		Reply(200).
		JSON(`{
    "data": [
        {
            "iface": "vmbr1",
            "bridge_fd": "0",
            "autostart": 1,
            "bridge_ports": "eno1.2 vmbr2.10",
            "priority": 31,
            "families": [
                "inet"
            ],
            "bridge_vids": "2-4094",
            "active": "1",
            "bridge_stp": "off",
            "bridge_vlan_aware": 1,
            "type": "bridge",
            "comments": "some comment\n",
            "method6": "manual",
            "method": "manual"
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
            "method": "manual",
            "method6": "manual",
            "priority": 20,
            "families": [
                "inet"
            ],
            "type": "eth",
            "exists": 1,
            "iface": "eno1"
        },
        {
            "netmask": "32",
            "priority": 16,
            "families": [
                "inet"
            ],
            "address": "192.168.2.50",
            "cidr": "192.168.2.50/32",
            "vlan-raw-device": "eno1",
            "vlan-id": "2",
            "iface": "eno1.2",
            "autostart": 1,
            "exists": 1,
            "options": [
                "metric 200"
            ],
            "method6": "manual",
            "method": "static",
            "type": "vlan",
            "comments": "Some Comment\n",
            "active": 1
        },
        {
            "iface": "vmbr1",
            "bridge_fd": "0",
            "autostart": 1,
            "bridge_ports": "eno1.2 vmbr2.10",
            "priority": 31,
            "families": [
                "inet"
            ],
            "bridge_vids": "2-4094",
            "active": 1,
            "bridge_stp": "off",
            "bridge_vlan_aware": 1,
            "type": "bridge",
            "comments": "some comment\n",
            "method6": "manual",
            "method": "manual"
        },
        {
            "options": [
                "metric 100"
            ],
            "exists": null,
            "iface": "vmbr2.2",
            "autostart": 1,
            "vlan-id": "2",
            "vlan-raw-device": "vmbr2",
            "cidr": "192.168.22.31/24",
            "address": "192.168.22.31",
            "priority": 33,
            "families": [
                "inet"
            ],
            "netmask": "24",
            "active": 1,
            "type": "vlan",
            "method": "static",
            "method6": "manual"
        },
        {
            "families": [
                "inet"
            ],
            "priority": 35,
            "netmask": "24",
            "cidr": "172.16.20.1/24",
            "vlan-raw-device": "vmbr2",
            "address": "172.20.0.1",
            "iface": "vmbr2.8",
            "exists": null,
            "autostart": 1,
            "vlan-id": "8",
            "method": "static",
            "method6": "manual",
            "comments": "Some Network\n",
            "type": "vlan",
            "active": 1
        }
    ]
}`)

	// GET /nodes/node2/network - Node2 network interfaces (should return 2 per test)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node2/network$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "iface": "eno1",
            "type": "eth",
            "method": "manual",
            "method6": "manual",
            "priority": 20,
            "families": ["inet"],
            "exists": 1
        },
        {
            "iface": "vmbr0",
            "type": "bridge",
            "method": "static",
            "method6": "manual",
            "address": "192.168.1.101",
            "netmask": "24",
            "cidr": "192.168.1.101/24",
            "gateway": "192.168.1.1",
            "bridge_ports": "eno1",
            "bridge_stp": "off",
            "bridge_fd": "0",
            "autostart": 1,
            "active": 1,
            "priority": 10,
            "families": ["inet"]
        }
    ]
}`)

	// GET /nodes/{node}/report - Get node report
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/report$").
		Reply(200).
		JSON(`{
    "data": "pve-manager: 9.1-1\nkernel: 6.11.0-1-pve\nproxmox-ve: 9.1-1\nqemu-server: 9.1-1\nlxc-pve: 6.0.0-1"
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

	// GET /nodes/{node}/storage/local-lvm/status — same listing entry as
	// above's /storage but called by name. local-lvm only stores images and
	// rootdir; used by tests that verify cloud-init refuses non-iso storages.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local-lvm/status$").
		Reply(200).
		JSON(`{
    "data": {
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

	// GET /nodes/{node}/certificates - Directory index
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/certificates$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "info"},
        {"name": "custom"},
        {"name": "acme"}
    ]
}`)

	// GET /nodes/{node}/certificates/acme - ACME directory index
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/certificates/acme$").
		Reply(200).
		JSON(`{
    "data": [
        {"name": "certificate"}
    ]
}`)

	// POST /nodes/{node}/certificates/acme/certificate - Order ACME cert
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/certificates/acme/certificate$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:acme-new-cert:pveproxy:root@pam:"
}`)

	// PUT /nodes/{node}/certificates/acme/certificate - Renew ACME cert
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/certificates/acme/certificate$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:acme-renew-cert:pveproxy:root@pam:"
}`)

	// DELETE /nodes/{node}/certificates/acme/certificate - Revoke ACME cert
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/certificates/acme/certificate$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00001234:00005678:12345678:acme-revoke-cert:pveproxy:root@pam:"
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

	// GET /nodes/{node}/dns - Read resolver config.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/dns$").
		Reply(200).
		JSON(`{
    "data": {
        "search": "example.com",
        "dns1": "1.1.1.1",
        "dns2": "8.8.8.8",
        "dns3": "9.9.9.9"
    }
}`)

	// PUT /nodes/{node}/dns - Replace resolver config. Returns null on success.
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/dns$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /nodes/{node}/services
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/services$").
		Reply(200).
		JSON(`{"data": [
			{"service": "pveproxy", "name": "pveproxy", "desc": "PVE API Proxy", "state": "running", "active-state": "active", "unit-state": "enabled"},
			{"service": "sshd",     "name": "sshd",     "desc": "OpenSSH server",  "state": "running", "active-state": "active", "unit-state": "enabled"}
		]}`)

	// GET /nodes/{node}/services/{service}/state
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/services/pveproxy/state$").
		Reply(200).
		JSON(`{"data": {"service": "pveproxy", "name": "pveproxy", "desc": "PVE API Proxy", "state": "running", "active-state": "active", "unit-state": "enabled"}}`)

	// POST /nodes/{node}/services/{service}/{action}
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/services/pveproxy/(start|stop|restart|reload)$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00012345:67890123:srvstart:pveproxy:root@pam:"}`)

	// GET /nodes/{node}/time - Read time + timezone
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/time$").
		Reply(200).
		JSON(`{"data": {"time": 1715500000, "localtime": 1715500000, "timezone": "UTC"}}`)

	// PUT /nodes/{node}/time - Set timezone
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/time$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /nodes/{node}/subscription
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/subscription$").
		Reply(200).
		JSON(`{"data": {"status": "active", "key": "pve8c-1234567890", "level": "c", "productname": "PVE Community 1 CPU/year", "sockets": 1, "nextduedate": "2026-12-31"}}`)

	// PUT /nodes/{node}/subscription
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/subscription$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /nodes/{node}/subscription
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/subscription$").
		Reply(200).
		JSON(`{"data": null}`)

	// DELETE /nodes/{node}/subscription
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/subscription$").
		Reply(200).
		JSON(`{"data": null}`)

	// POST /nodes/{node}/startall — mass start
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/startall$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001111:00002222:5A3B7C8D:startall::root@pam:"}`)

	// POST /nodes/{node}/stopall — mass stop
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/stopall$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001111:00002222:5A3B7C8D:stopall::root@pam:"}`)

	// POST /nodes/{node}/suspendall — mass suspend
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/suspendall$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001111:00002222:5A3B7C8D:suspendall::root@pam:"}`)

	// POST /nodes/{node}/migrateall — mass migrate
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/migrateall$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001111:00002222:5A3B7C8D:migrateall::root@pam:"}`)

	// POST /nodes/{node}/wakeonlan — returns the MAC string
	gock.New(config.C.URI).
		Post("^/nodes/node1/wakeonlan$").
		Reply(200).
		JSON(`{"data": "AA:BB:CC:DD:EE:FF"}`)

	// GET /nodes/{node}/replication
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/replication$").
		Reply(200).
		JSON(`{"data": [
			{"id": "100-0", "type": "local", "source": "node1", "target": "node2", "guest": 100, "jobnum": 0, "last_sync": 1715500000, "next_sync": 1715501000, "state": "idle"}
		]}`)

	// GET /nodes/{node}/replication/{id}/status
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/replication/100-0/status$").
		Reply(200).
		JSON(`{"data": {"id": "100-0", "type": "local", "target": "node2", "guest": 100, "jobnum": 0, "last_sync": 1715500000, "state": "idle", "fail_count": 0}}`)

	// GET /nodes/{node}/replication/{id}/log
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/replication/100-0/log$").
		Reply(200).
		JSON(`{"data": [
			{"n": 1, "t": "2025-05-12 10:00:00 100-0: start replication job"},
			{"n": 2, "t": "2025-05-12 10:00:05 100-0: end replication job"}
		]}`)

	// POST /nodes/{node}/replication/{id}/schedule_now
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/replication/100-0/schedule_now$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00012345:67890124:replicate:100-0:root@pam:"}`)

	// GET /nodes/{node}/apt
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/apt$").
		Reply(200).
		JSON(`{"data": [
			{"id": "changelog"},
			{"id": "repositories"},
			{"id": "update"},
			{"id": "versions"}
		]}`)

	// GET /nodes/{node}/apt/update — pending upgrades
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/apt/update$").
		Reply(200).
		JSON(`{"data": [
			{"Package": "pve-manager", "Title": "Proxmox VE Management", "Description": "PVE manager", "Version": "9.1-2", "OldVersion": "9.1-1", "Origin": "Proxmox", "Arch": "amd64", "Section": "admin", "Priority": "optional"}
		]}`)

	// POST /nodes/{node}/apt/update — refresh index, returns UPID
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/apt/update$").
		Reply(200).
		JSON(`{"data": "UPID:node1:0000ABCD:0001ABCD:67890125:aptupdate::root@pam:"}`)

	// GET /nodes/{node}/apt/changelog
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/apt/changelog").
		Reply(200).
		JSON(`{"data": "pve-manager (9.1-2) bookworm; urgency=medium\n  * fix things\n"}`)

	// GET /nodes/{node}/apt/repositories
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/apt/repositories$").
		Reply(200).
		JSON(`{"data": {
			"digest": "abcdef0123456789",
			"files": [
				{
					"path": "/etc/apt/sources.list",
					"file-type": "list",
					"digest": [1,2,3,4],
					"repositories": [
						{"Enabled": true, "FileType": "list", "Types": ["deb"], "URIs": ["http://deb.debian.org/debian"], "Suites": ["bookworm"], "Components": ["main", "contrib"]}
					]
				}
			],
			"errors": [],
			"infos": [
				{"path": "/etc/apt/sources.list.d/pve-enterprise.list", "index": "0", "kind": "warning", "message": "Enterprise repo configured without subscription", "property": "Enabled"}
			],
			"standard-repos": [
				{"handle": "enterprise", "name": "Proxmox VE Enterprise", "status": false},
				{"handle": "no-subscription", "name": "Proxmox VE No-Subscription"}
			]
		}}`)

	// POST /nodes/{node}/apt/repositories — change (enable/disable) existing repo
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/apt/repositories$").
		Reply(200).
		JSON(`{"data": null}`)

	// PUT /nodes/{node}/apt/repositories — add standard repository
	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/apt/repositories$").
		Reply(200).
		JSON(`{"data": null}`)

	// GET /nodes/{node}/apt/versions
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/apt/versions$").
		Reply(200).
		JSON(`{"data": [
			{"Package": "pve-manager", "Title": "Proxmox VE Management", "Description": "PVE manager", "Version": "9.1-1", "Origin": "Proxmox", "Arch": "amd64", "Section": "admin", "Priority": "optional", "CurrentState": "Installed", "ManagerVersion": "9.1-1"},
			{"Package": "proxmox-ve", "Title": "Proxmox VE", "Description": "Proxmox VE metapackage", "Version": "9.1-1", "Origin": "Proxmox", "Arch": "all", "Section": "admin", "Priority": "optional", "CurrentState": "Installed", "RunningKernel": "6.8.12-1-pve"}
		]}`)

	// --- /nodes/{node}/disks -------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/list").
		Reply(200).
		JSON(`{"data": [
			{"devpath": "/dev/sda", "size": 512000000000, "model": "Samsung SSD 870 EVO", "type": "ssd", "health": "PASSED", "gpt": 1, "used": "LVM"},
			{"devpath": "/dev/sdb", "size": 2000000000000, "model": "Seagate ST2000DM008", "type": "hdd", "health": "PASSED", "gpt": 0}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/smart").
		Reply(200).
		JSON(`{"data": {"health": "PASSED", "type": "ata", "attributes": []}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/disks/initgpt$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:diskinit:/dev/sda:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/disks/wipedisk$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:wipedisk:/dev/sda:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/directory$").
		Reply(200).
		JSON(`{"data": [
			{"path": "/mnt/pve/iso", "device": "/dev/sda1", "type": "ext4", "options": "defaults"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/disks/directory$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:mkdir:iso:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/disks/directory/iso").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:rmdir:iso:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/lvm$").
		Reply(200).
		JSON(`{"data": {
			"leaf": 0,
			"children": [
				{"name": "pve", "size": 500000000000, "free": 100000000000, "leaf": 0, "children": [
					{"name": "/dev/sda", "size": 500000000000, "free": 100000000000, "leaf": 1}
				]}
			]
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/disks/lvm$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:mklvm:pve:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/disks/lvm/pve").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:rmlvm:pve:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/lvmthin$").
		Reply(200).
		JSON(`{"data": [
			{"lv": "data", "lv_size": 400000000000, "metadata_size": 100000000, "metadata_used": 1000000, "used": 50000000000}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/disks/lvmthin$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:mklvmthin:data:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/disks/lvmthin/data").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:rmlvmthin:data:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/zfs$").
		Reply(200).
		JSON(`{"data": [
			{"name": "rpool", "health": "ONLINE", "size": 1000000000000, "alloc": 200000000000, "free": 800000000000, "frag": 5, "dedup": 1.0}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks/zfs/rpool$").
		Reply(200).
		JSON(`{"data": {
			"name": "rpool",
			"state": "ONLINE",
			"status": "healthy",
			"scan": "scrub repaired 0B in 00:12:34 with 0 errors on Sun Apr 14 00:24:35 2024",
			"errors": "No known data errors",
			"children": [
				{"name": "raidz2-0", "state": "ONLINE", "leaf": 0, "children": [
					{"name": "/dev/sda", "state": "ONLINE", "leaf": 1, "read": 0, "write": 0, "cksum": 0},
					{"name": "/dev/sdb", "state": "ONLINE", "leaf": 1, "read": 0, "write": 0, "cksum": 0}
				]}
			]
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/disks/zfs$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:mkzfs:rpool:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/disks/zfs/rpool").
		Reply(200).
		JSON(`{"data": "UPID:node1:00001234:00005678:12345678:rmzfs:rpool:root@pam:"}`)

	// --- diridx endpoints (see nodes_diridx.go) ----------------------------

	// GET /nodes/{node} — node root diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"qemu"},
			{"subdir":"lxc"},
			{"subdir":"storage"},
			{"subdir":"network"},
			{"subdir":"tasks"},
			{"subdir":"services"},
			{"subdir":"subscription"},
			{"subdir":"firewall"},
			{"subdir":"replication"}
		]}`)

	// GET /nodes/{node}/firewall — node firewall diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/firewall$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"rules"},
			{"subdir":"options"},
			{"subdir":"log"}
		]}`)

	// GET /nodes/{node}/disks — node disks diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/disks$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"list"},
			{"subdir":"smart"},
			{"subdir":"initgpt"},
			{"subdir":"wipedisk"},
			{"subdir":"directory"},
			{"subdir":"lvm"},
			{"subdir":"lvmthin"},
			{"subdir":"zfs"}
		]}`)

	// GET /nodes/{node}/services/{service} — per-service diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/services/pveproxy$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"state"},
			{"subdir":"start"},
			{"subdir":"stop"},
			{"subdir":"restart"},
			{"subdir":"reload"}
		]}`)

	// GET /nodes/{node}/replication/{id} — per-replication-job diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/replication/101-0$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"status"},
			{"subdir":"log"},
			{"subdir":"schedule_now"}
		]}`)

	// GET /nodes/{node}/storage/{storage} — diridx path that returns the
	// storage's status/capability object directly (not a subdir list).
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local$").
		Reply(200).
		JSON(`{"data":{
			"type":"dir",
			"content":"images,rootdir,vztmpl,backup,iso,snippets",
			"active":1,
			"enabled":1,
			"shared":0,
			"total":60000000000,
			"used":10000000000,
			"avail":50000000000,
			"used_fraction":0.16667
		}}`)
}
