package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func sdn() {
	// GET /nodes/{node}/sdn - top-level diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"vnets"},
			{"subdir":"zones"},
			{"subdir":"fabrics"}
		]}`)

	// GET /nodes/{node}/sdn/fabrics/{fabric} - fabric diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/fabrics/fab1$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"routes"},
			{"subdir":"neighbors"},
			{"subdir":"interfaces"}
		]}`)

	// GET /nodes/{node}/sdn/fabrics/{fabric}/interfaces
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/fabrics/fab1/interfaces$").
		Reply(200).
		JSON(`{"data":[
			{"name":"eth0","state":"up","type":"Point-to-Point"},
			{"name":"eth1","state":"down","type":"Broadcast"}
		]}`)

	// GET /nodes/{node}/sdn/fabrics/{fabric}/neighbors
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/fabrics/fab1/neighbors$").
		Reply(200).
		JSON(`{"data":[
			{"neighbor":"10.0.0.2","status":"Established","uptime":"8h24m12s"}
		]}`)

	// GET /nodes/{node}/sdn/fabrics/{fabric}/routes
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/fabrics/fab1/routes$").
		Reply(200).
		JSON(`{"data":[
			{"route":"10.0.0.0/24","via":["10.0.0.1","10.0.0.2"]},
			{"route":"10.0.1.0/24","via":["10.0.0.3"]}
		]}`)

	// GET /nodes/{node}/sdn/vnets/{vnet} - vnet diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/vnets/vnet1$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"mac-vrf"}
		]}`)

	// GET /nodes/{node}/sdn/vnets/{vnet}/mac-vrf
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/vnets/vnet1/mac-vrf$").
		Reply(200).
		JSON(`{"data":[
			{"ip":"10.0.0.10","mac":"aa:bb:cc:dd:ee:ff","nexthop":"10.0.0.1"},
			{"ip":"10.0.0.11","mac":"aa:bb:cc:dd:ee:00","nexthop":"10.0.0.2"}
		]}`)

	// GET /nodes/{node}/sdn/zones
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/zones$").
		Reply(200).
		JSON(`{"data":[
			{"zone":"zone1","status":"available"},
			{"zone":"zone2","status":"pending"}
		]}`)

	// GET /nodes/{node}/sdn/zones/{zone} - zone diridx
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/zones/zone1$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"content"},
			{"subdir":"bridges"},
			{"subdir":"ip-vrf"}
		]}`)

	// GET /nodes/{node}/sdn/zones/{zone}/bridges
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/zones/zone1/bridges$").
		Reply(200).
		JSON(`{"data":[
			{
				"name":"vnet1",
				"vlan_filtering":"1",
				"ports":[
					{"name":"tap100i0","index":"0","primary_vlan":100,"vlans":["200","300-310"],"vmid":100},
					{"name":"fwln100i0"}
				]
			}
		]}`)

	// GET /nodes/{node}/sdn/zones/{zone}/content
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/zones/zone1/content$").
		Reply(200).
		JSON(`{"data":[
			{"vnet":"vnet1","status":"available"},
			{"vnet":"vnet2","status":"pending","statusmsg":"awaiting reload"}
		]}`)

	// GET /nodes/{node}/sdn/zones/{zone}/ip-vrf
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/sdn/zones/zone1/ip-vrf$").
		Reply(200).
		JSON(`{"data":[
			{"ip":"10.0.0.0/24","metric":20,"nexthops":["10.0.0.1"],"protocol":"bgp"},
			{"ip":"10.0.1.0/24","metric":0,"protocol":"connected"}
		]}`)
}
