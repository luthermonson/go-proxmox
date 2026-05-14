package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

// virtualMachines registers PVE 7.x VM endpoints. Only the QEMU guest-agent
// surface is mocked here today; full VM CRUD lives in pve9x/pve8x. Add more
// as needed when version-specific behaviour appears.
func virtualMachines() {
	// ----- QEMU guest-agent endpoints (vmid 101) -----
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/ping$").
		Reply(200).
		JSON(`{"data": {"result": {}}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-time$").
		Reply(200).
		JSON(`{"data": {"result": 1715600000000000000}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-timezone$").
		Reply(200).
		JSON(`{"data": {"result": {"zone": "UTC", "offset": 0}}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-users$").
		Reply(200).
		JSON(`{"data": {"result": []}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-vcpus$").
		Reply(200).
		JSON(`{"data": {"result": [{"logical-id": 0, "online": true}]}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-fsinfo$").
		Reply(200).
		JSON(`{"data": {"result": []}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/get-memory-blocks$").
		Reply(200).
		JSON(`{"data": {"result": []}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/info$").
		Reply(200).
		JSON(`{"data": {"result": {"version": "5.2.0"}}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-freeze$").
		Reply(200).
		JSON(`{"data": {"result": 1}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-thaw$").
		Reply(200).
		JSON(`{"data": {"result": 1}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fsfreeze-status$").
		Reply(200).
		JSON(`{"data": {"result": "thawed"}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/fstrim$").
		Reply(200).
		JSON(`{"data": {"result": {}}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/shutdown$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-disk$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-hybrid$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/suspend-ram$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/qemu/101/agent/file-read$").
		Reply(200).
		JSON(`{"data": {"content": "hello world\n", "truncated": 0}}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/qemu/101/agent/file-write$").
		Reply(200).
		JSON(`{"data": null}`)
}
