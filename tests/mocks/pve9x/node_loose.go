package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

// nodeLoose registers gock fixtures for the small "loose" /nodes/{node}/*
// endpoints — node config, hosts, power, journal/syslog/netstat, execute,
// console launchers, RRD, URL/OCI/vzdump queries, firewall extras,
// tasks list, network revert, and storage identity/rrd.
func nodeLoose() {
	// ---- /nodes/{node}/config -------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/config").
		Reply(200).
		JSON(`{"data":{
			"description":"primary hypervisor",
			"ballooning-target":80,
			"startall-onboot-delay":30,
			"wakeonlan":"mac=aa:bb:cc:dd:ee:ff,bind-interface=vmbr0",
			"digest":"abc123"
		}}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/config$").
		Reply(200).
		JSON(`{"data":null}`)

	// ---- /nodes/{node}/hosts --------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hosts$").
		Reply(200).
		JSON(`{"data":{
			"data":"127.0.0.1 localhost\n10.0.0.1 node1.example.com node1\n",
			"digest":"hostsdigest123"
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/hosts$").
		Reply(200).
		JSON(`{"data":null}`)

	// ---- /nodes/{node}/status (POST = reboot/shutdown) ------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/status$").
		Reply(200).
		JSON(`{"data":null}`)

	// ---- /nodes/{node}/journal ------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/journal").
		Reply(200).
		JSON(`{"data":[
			"Jan 01 00:00:01 node1 systemd[1]: Started PVE API Daemon.",
			"Jan 01 00:00:02 node1 pveproxy[1234]: starting server"
		]}`)

	// ---- /nodes/{node}/syslog -------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/syslog").
		Reply(200).
		JSON(`{"data":[
			{"n":1,"t":"Jan 01 00:00:01 node1 kernel: Linux version 6.8.0"},
			{"n":2,"t":"Jan 01 00:00:02 node1 systemd[1]: Reached target Multi-User"}
		]}`)

	// ---- /nodes/{node}/netstat ------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/netstat$").
		Reply(200).
		JSON(`{"data":[
			{"dev":"tap100i0","vmid":100,"in":12345,"out":67890},
			{"dev":"veth101i0","vmid":101,"in":111,"out":222}
		]}`)

	// ---- /nodes/{node}/execute ------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/execute$").
		Reply(200).
		JSON(`{"data":[{"status":200,"data":"ok"},{"status":200,"data":"ok"}]}`)

	// ---- /nodes/{node}/vncshell + spiceshell ----------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/vncshell$").
		Reply(200).
		JSON(`{"data":{
			"cert":"-----BEGIN CERTIFICATE-----\nABC\n-----END CERTIFICATE-----",
			"port":5901,
			"ticket":"PVE:root@pam:VNCSHELL",
			"upid":"UPID:node1:00000001:00000001:00000001:vncshell::root@pam:",
			"user":"root@pam",
			"password":"vncpw"
		}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/spiceshell$").
		Reply(200).
		JSON(`{"data":{
			"host":"node1.example.com",
			"password":"spicepw",
			"proxy":"node1.example.com",
			"tls-port":"3128",
			"type":"spice"
		}}`)

	// ---- /nodes/{node}/rrd + rrddata ------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/rrd$").
		Reply(200).
		JSON(`{"data":{"filename":"rrd-node-graph.png"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/rrddata").
		Reply(200).
		JSON(`{"data":[
			{"time":1715000000,"cpu":0.12,"mem":1000000,"maxmem":8000000},
			{"time":1715000060,"cpu":0.18,"mem":1100000,"maxmem":8000000}
		]}`)

	// ---- /nodes/{node}/query-url-metadata -------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/query-url-metadata").
		Reply(200).
		JSON(`{"data":{
			"filename":"debian-12.iso",
			"mimetype":"application/x-iso9660-image",
			"size":654311424
		}}`)

	// ---- /nodes/{node}/query-oci-repo-tags ------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/query-oci-repo-tags").
		Reply(200).
		JSON(`{"data":["latest","3.18","3.19","3.20"]}`)

	// ---- /nodes/{node}/vzdump/defaults ----------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/vzdump/defaults").
		Reply(200).
		JSON(`{"data":{
			"mode":"snapshot",
			"compress":"zstd",
			"storage":"backup",
			"remove":1,
			"prune-backups":"keep-last=3,keep-daily=7"
		}}`)

	// ---- /nodes/{node}/firewall/log + /firewall/rules/{pos} -------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/firewall/log").
		Reply(200).
		JSON(`{"data":[
			{"n":1,"t":"DROP IN=vmbr0 OUT= MAC=aa:bb:cc"},
			{"n":2,"t":"ACCEPT IN=vmbr0 OUT= MAC=dd:ee:ff"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/firewall/rules/0$").
		Reply(200).
		JSON(`{"data":{
			"pos":0,
			"type":"in",
			"action":"ACCEPT",
			"enable":1,
			"proto":"tcp",
			"dport":"22",
			"comment":"ssh"
		}}`)

	// ---- /nodes/{node}/tasks --------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get(`^/nodes/node1/tasks$`).
		Reply(200).
		JSON(`{"data":[
			{
				"upid":"UPID:node1:00000001:00000001:00000001:vzdump:100:root@pam:",
				"node":"node1","pid":1,"pstart":1,"starttime":1715000000,
				"endtime":1715000100,"type":"vzdump","id":"100","user":"root@pam",
				"status":"OK"
			},
			{
				"upid":"UPID:node1:00000002:00000002:00000002:qmstart:100:root@pam:",
				"node":"node1","pid":2,"pstart":2,"starttime":1715000200,
				"type":"qmstart","id":"100","user":"root@pam"
			}
		]}`)

	// ---- /nodes/{node}/network (DELETE = revert) ------------------------

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/network$").
		Reply(200).
		JSON(`{"data":null}`)

	// ---- /nodes/{node}/storage/{storage}/identity + rrd -----------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/identity$").
		Reply(200).
		JSON(`{"data":{"id":"local","type":"dir"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/rrd$").
		Reply(200).
		JSON(`{"data":{"filename":"rrd-storage-graph.png"}}`)
}
