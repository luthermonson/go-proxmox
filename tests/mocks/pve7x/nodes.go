package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func nodes() {
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/rrddata").
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
