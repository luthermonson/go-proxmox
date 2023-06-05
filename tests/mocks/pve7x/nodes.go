package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func nodes() {
	gock.New(config.C.URI).
		Persist().
		Get("/nodes").
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

}
