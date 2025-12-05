package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func pool() {
	gock.New(config.C.URI).
		Get("^/pools$").
		Reply(200).
		JSON(`{"data": [
    {
      "poolid": "test-pool",
      "comment": "Test pool"
    }
  ]
}`)

	// Mock for the new /pools/?poolid=test-pool endpoint (returns array)
	gock.New(config.C.URI).
		Get("^/pools/$").
		MatchParam("poolid", "test-pool").
		Reply(200).
		JSON(`{"data": [
    {
      "poolid": "test-pool",
      "comment": "Test pool",
      "members": [
        {
          "disk": 0,
          "uptime": 88341,
          "diskwrite": 7389189632,
          "netout": 73088105,
          "maxmem": 17179869184,
          "maxdisk": 10737418240,
          "type": "qemu",
          "vmid": 100,
          "template": 0,
          "cpu": 0.0378721079577321,
          "name": "test-vm",
          "node": "pve1",
          "id": "qemu/100",
          "diskread": 500085248,
          "netin": 485707842,
          "maxcpu": 4,
          "mem": 3650678784,
          "status": "running"
        },
        {
          "diskread": 486947840,
          "id": "qemu/106",
          "cpu": 0.02889417191544,
          "template": 0,
          "name": "test-vm2",
          "node": "pve2",
          "maxcpu": 1,
          "netin": 337525157,
          "status": "running",
          "mem": 1744027648,
          "disk": 0,
          "diskwrite": 1382547968,
          "uptime": 88204,
          "netout": 21769689,
          "type": "qemu",
          "vmid": 106,
          "maxdisk": 10737418240,
          "maxmem": 2147483648
        },
        {
          "node": "node1",
          "maxdisk": 948340654080,
          "type": "storage",
          "id": "storage/node1/local",
          "status": "available",
          "storage": "local",
          "content": "backup,vztmpl,iso",
          "plugintype": "dir",
          "disk": 10486939648,
          "shared": 0
        }
      ]
    }
  ]
}`)

	// Mock for the deprecated /pools/test-pool endpoint (kept for backwards compatibility)
	gock.New(config.C.URI).
		Get("^/pools/test-pool$").
		Reply(200).
		JSON(`{"data": {
    "comment": "Test pool",
    "members": [
      {
        "disk": 0,
        "uptime": 88341,
        "diskwrite": 7389189632,
        "netout": 73088105,
        "maxmem": 17179869184,
        "maxdisk": 10737418240,
        "type": "qemu",
        "vmid": 100,
        "template": 0,
        "cpu": 0.0378721079577321,
        "name": "test-vm",
        "node": "pve1",
        "id": "qemu/100",
        "diskread": 500085248,
        "netin": 485707842,
        "maxcpu": 4,
        "mem": 3650678784,
        "status": "running"
      },
      {
        "diskread": 486947840,
        "id": "qemu/106",
        "cpu": 0.02889417191544,
        "template": 0,
        "name": "test-vm2",
        "node": "pve2",
        "maxcpu": 1,
        "netin": 337525157,
        "status": "running",
        "mem": 1744027648,
        "disk": 0,
        "diskwrite": 1382547968,
        "uptime": 88204,
        "netout": 21769689,
        "type": "qemu",
        "vmid": 106,
        "maxdisk": 10737418240,
        "maxmem": 2147483648
      },
	  {
		"node": "node1",
		"maxdisk": 948340654080,
		"type": "storage",
		"id": "storage/node1/local",
		"status": "available",
		"storage": "local",
		"content": "backup,vztmpl,iso",
		"plugintype": "dir",
		"disk": 10486939648,
		"shared": 0
	  }
    ]
  }
}`)

	gock.New(config.C.URI).
		Post("^/pools$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Put("^/pools/test-pool$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Delete("^/pools/test-pool$").
		Reply(200).
		JSON(`{"data": null}`)
}
