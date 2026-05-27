package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

// hardware registers gock fixtures for the /nodes/{node}/scan,
// /nodes/{node}/capabilities, and /nodes/{node}/hardware endpoint families.
// All read-only — no body assertions.
func hardware() {
	// ---- /nodes/{node}/scan ------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan$").
		Reply(200).
		JSON(`{"data":[
			{"method":"zfs"},{"method":"lvm"},{"method":"lvmthin"},
			{"method":"nfs"},{"method":"cifs"},{"method":"pbs"},{"method":"iscsi"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/zfs$").
		Reply(200).
		JSON(`{"data":[{"pool":"rpool"},{"pool":"tank"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/lvm$").
		Reply(200).
		JSON(`{"data":[{"vg":"pve"},{"vg":"data"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/lvmthin").
		Reply(200).
		JSON(`{"data":[{"lv":"data"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/nfs").
		Reply(200).
		JSON(`{"data":[
			{"path":"/exports/backup","options":"rw,no_root_squash"},
			{"path":"/exports/iso","options":"ro"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/cifs").
		Reply(200).
		JSON(`{"data":[
			{"share":"backup","description":"Backup share"},
			{"share":"iso"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/pbs").
		Reply(200).
		JSON(`{"data":[
			{"store":"main","comment":"primary datastore"},
			{"store":"archive"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/scan/iscsi").
		Reply(200).
		JSON(`{"data":[
			{"target":"iqn.2024.example:storage.0","portal":"192.0.2.10:3260"},
			{"target":"iqn.2024.example:storage.1","portal":"192.0.2.10:3260"}
		]}`)

	// ---- /nodes/{node}/capabilities ---------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/capabilities$").
		Reply(200).
		JSON(`{"data":[{"name":"qemu"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/capabilities/qemu$").
		Reply(200).
		JSON(`{"data":[
			{"name":"cpu"},{"name":"cpu-flags"},{"name":"machines"},{"name":"migration"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get(`^/nodes/node1/capabilities/qemu/cpu(\?|$)`).
		Reply(200).
		JSON(`{"data":[
			{"name":"host","vendor":"GenuineIntel","custom":false},
			{"name":"x86-64-v3","vendor":"GenuineIntel","custom":false,"abstract":true},
			{"name":"custom-mycpu","vendor":"GenuineIntel","custom":true}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/capabilities/qemu/cpu-flags").
		Reply(200).
		JSON(`{"data":[
			{"name":"aes","description":"AES instruction set","supported-on":["node1","node2"]},
			{"name":"avx2","description":"Advanced Vector Extensions 2"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/capabilities/qemu/machines").
		Reply(200).
		JSON(`{"data":[
			{"id":"pc-q35-9.0","type":"q35","version":"9.0"},
			{"id":"pc-i440fx-9.0","type":"i440fx","version":"9.0","changes":"pveX backport"}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/capabilities/qemu/migration$").
		Reply(200).
		JSON(`{"data":{"has-dbus-vmstate":true}}`)

	// ---- /nodes/{node}/hardware -------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hardware$").
		Reply(200).
		JSON(`{"data":[{"type":"pci"},{"type":"usb"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hardware/pci$").
		Reply(200).
		JSON(`{"data":[
			{
				"id":"0000:01:00.0","class":"0x030000",
				"vendor":"0x10de","vendor_name":"NVIDIA Corporation",
				"device":"0x2204","device_name":"GA102 [GeForce RTX 3090]",
				"iommugroup":15,"mdev":true
			},
			{
				"id":"0000:02:00.0","class":"0x010802",
				"vendor":"0x8086","device":"0xa808",
				"iommugroup":16
			}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hardware/usb$").
		Reply(200).
		JSON(`{"data":[
			{
				"busnum":1,"devnum":2,"port":1,"level":1,"class":9,
				"vendid":"0x1d6b","prodid":"0x0002",
				"speed":"480","manufacturer":"Linux Foundation","product":"USB 2.0 root hub",
				"usbpath":"1-1"
			},
			{
				"busnum":2,"devnum":1,"port":0,"level":0,"class":9,
				"vendid":"0x1d6b","prodid":"0x0003","speed":"5000"
			}
		]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hardware/pci/0000:01:00.0$").
		Reply(200).
		JSON(`{"data":[{"method":"mdev"}]}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/hardware/pci/0000:01:00.0/mdev$").
		Reply(200).
		JSON(`{"data":[
			{"type":"nvidia-256","name":"GRID A40-4Q","description":"4GB profile","available":2},
			{"type":"nvidia-257","name":"GRID A40-8Q","description":"8GB profile","available":1}
		]}`)
}
