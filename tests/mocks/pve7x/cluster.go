package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func cluster() {

	gock.New(config.C.URI).
		Get("/cluster/nextid").
		Reply(200).
		JSON(`{"data": "100"}`)

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
}
