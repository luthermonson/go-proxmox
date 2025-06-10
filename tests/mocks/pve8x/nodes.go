package pve8x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func nodes() {
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
            "active": 1,
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
}
