package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/capture"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func storage() {
	// GET /storage - List all cluster storages
	gock.New(config.C.URI).
		Persist().
		Get("^/storage$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "content": "vztmpl,iso,backup",
            "digest": "1234567890abcdef1234567890abcdef12345678",
            "storage": "local",
            "type": "dir",
            "shared": 0,
            "path": "/var/lib/vz"
        },
        {
            "content": "images,rootdir",
            "digest": "abcdef1234567890abcdef1234567890abcdef12",
            "storage": "local-lvm",
            "type": "lvmthin",
            "shared": 0,
            "thinpool": "data",
            "vgname": "pve"
        },
        {
            "content": "images,rootdir",
            "digest": "fedcba0987654321fedcba0987654321fedcba09",
            "storage": "nfs-storage",
            "type": "nfs",
            "shared": 1,
            "path": "/mnt/pve/nfs-storage",
            "nodes": "node1,node2"
        }
    ]
}`)

	// GET /storage/{storage} - Get specific storage
	gock.New(config.C.URI).
		Persist().
		Get("^/storage/local$").
		Reply(200).
		JSON(`{
    "data": {
        "content": "vztmpl,iso,backup",
        "digest": "1234567890abcdef1234567890abcdef12345678",
        "storage": "local",
        "type": "dir",
        "shared": 0,
        "path": "/var/lib/vz"
    }
}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/storage/local-lvm$").
		Reply(200).
		JSON(`{
    "data": {
        "content": "images,rootdir",
        "digest": "abcdef1234567890abcdef1234567890abcdef12",
        "storage": "local-lvm",
        "type": "lvmthin",
        "shared": 0,
        "thinpool": "data",
        "vgname": "pve"
    }
}`)

	// POST /storage - Create new storage
	gock.New(config.C.URI).
		Post("^/storage$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// PUT /storage/{storage} - Update storage
	gock.New(config.C.URI).
		Put("^/storage/local$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// DELETE /storage/{storage} - Delete storage
	gock.New(config.C.URI).
		Delete("^/storage/test-storage$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000001:00000001:00000001:storage:delete:root@pam:"
}`)

	// GET /nodes/{node}/storage/{storage}/content - Get storage content
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "volid": "local:iso/debian-12.0.0-amd64-netinst.iso",
            "format": "iso",
            "size": 654311424,
            "ctime": 1693252591
        },
        {
            "volid": "local:vztmpl/debian-12-standard_12.0-1_amd64.tar.zst",
            "format": "tar.zst",
            "size": 128974848,
            "ctime": 1693252600
        },
        {
            "volid": "local:backup/vzdump-qemu-100-2023_08_28-12_00_00.vma.zst",
            "format": "vma.zst",
            "size": 2147483648,
            "ctime": 1693252800,
            "vmid": 100
        }
    ]
}`)

	// DELETE /nodes/{node}/storage/{storage}/content/{volume} - Delete storage content
	gock.New(config.C.URI).
		Delete("^/nodes/node1/storage/local/content/local:iso/test.iso$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000002:00000002:00000002:imgdel:delete:root@pam:"
}`)

	// POST /nodes/{node}/storage/{storage}/download-url - Download from URL
	gock.New(config.C.URI).
		Post("^/nodes/node1/storage/local/download-url$").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000003:00000003:00000003:download:iso:root@pam:"
}`)

	// POST /nodes/{node}/storage/{storage}/upload - Upload content (iso, vztmpl, snippets, ...)
	// Multipart bodies are recorded by capture.UploadMatcher so tests can
	// assert on the content/filename/body fields.
	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/storage/local/upload$").
		AddMatcher(capture.UploadMatcher()).
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000004:00000004:00000004:imgcopy:upload:root@pam:"
}`)

	// GET /nodes/{node}/storage/{storage}/prunebackups - Dryrun prune preview.
	// Matches with or without filter query params.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/prunebackups").
		Reply(200).
		JSON(`{
    "data": [
        {
            "volid": "local:backup/vzdump-qemu-100-2024_01_15-03_00_00.vma.zst",
            "ctime": 1705287600,
            "mark": "keep",
            "type": "qemu",
            "vmid": 100
        },
        {
            "volid": "local:backup/vzdump-qemu-100-2024_01_08-03_00_00.vma.zst",
            "ctime": 1704682800,
            "mark": "remove",
            "type": "qemu",
            "vmid": 100
        },
        {
            "volid": "local:backup/vzdump-qemu-100-2023_12_25-03_00_00.vma.zst",
            "ctime": 1703473200,
            "mark": "protected",
            "type": "qemu",
            "vmid": 100
        },
        {
            "volid": "local:backup/manual-snapshot-before-upgrade.vma.zst",
            "ctime": 1703300000,
            "mark": "renamed",
            "type": "qemu",
            "vmid": 100
        }
    ]
}`)

	// DELETE /nodes/{node}/storage/{storage}/prunebackups - Execute prune.
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/storage/local/prunebackups").
		Reply(200).
		JSON(`{
    "data": "UPID:node1:00000005:00000005:00000005:prunebackups:local:root@pam:"
}`)

	// GET /nodes/{node}/storage/{storage}/import-metadata - ESXi disk import metadata.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/esxi/import-metadata").
		Reply(200).
		JSON(`{
    "data": {
        "type": "vm",
        "source": "esxi",
        "create-args": {
            "name": "imported-vm",
            "memory": 4096,
            "cores": 2,
            "ostype": "l26"
        },
        "disks": {
            "scsi0": "esxi:ha-datacenter/MyVM/MyVM.vmdk",
            "scsi1": "esxi:ha-datacenter/MyVM/MyVM_1.vmdk"
        },
        "net": {
            "net0": {
                "model": "vmxnet3",
                "bridge": "vmbr0"
            }
        },
        "warnings": [
            {
                "type": "guest-is-running",
                "key": "power",
                "value": "poweredOn"
            }
        ]
    }
}`)

	// --- /nodes/{node}/storage/{storage}/content extras ---------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/storage/local-lvm/content$").
		Reply(200).
		JSON(`{"data": "local-lvm:vm-100-disk-1"}`)

	gock.New(config.C.URI).
		Persist().
		Put("^/nodes/node1/storage/local/content/local:backup/").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/storage/local-lvm/content/local-lvm:vm-100-disk-0$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000006:00000006:00000006:imgcopy:vm-100-disk-0:root@pam:"}`)

	// --- OCI registry pull --------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Post("^/nodes/node1/storage/local/oci-registry-pull$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000007:00000007:00000007:ocipull:alpine:root@pam:"}`)

	// --- file-restore -------------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/pbs/file-restore/list").
		Reply(200).
		JSON(`{"data": [
			{"filepath": "/etc/hostname", "type": "f", "size": 12, "mtime": 1715000000},
			{"filepath": "/etc/network", "type": "d", "leaf": 0}
		]}`)

	// --- rrddata ------------------------------------------------------------

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/rrddata").
		Reply(200).
		JSON(`{"data": [
			{"time": 1715000000, "used": 1000000, "total": 2000000},
			{"time": 1715000060, "used": 1100000, "total": 2000000}
		]}`)

	// --- per-volume GETs for ISO/VzTmpl/Import/Backup ----------------------

	// ISO volume detail.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content/local:iso/debian-12\\.iso$").
		Reply(200).
		JSON(`{"data": {
			"volid": "local:iso/debian-12.iso",
			"format": "iso",
			"size": 654311424,
			"ctime": 1693252591,
			"path": "/var/lib/vz/template/iso/debian-12.iso"
		}}`)

	// vztmpl volume detail.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content/local:vztmpl/debian-12-standard\\.tar\\.zst$").
		Reply(200).
		JSON(`{"data": {
			"volid": "local:vztmpl/debian-12-standard.tar.zst",
			"format": "tar.zst",
			"size": 128974848,
			"ctime": 1693252600,
			"path": "/var/lib/vz/template/cache/debian-12-standard.tar.zst"
		}}`)

	// import volume detail (lives on the "esxi" storage in fixtures).
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/esxi/content/esxi:import/MyVM\\.vmx$").
		Reply(200).
		JSON(`{"data": {
			"volid": "esxi:import/MyVM.vmx",
			"format": "vmx",
			"size": 4096,
			"ctime": 1700000000,
			"path": "/mnt/esxi/MyVM/MyVM.vmx"
		}}`)

	// backup volume detail.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content/local:backup/vzdump-qemu-100\\.vma\\.zst$").
		Reply(200).
		JSON(`{"data": {
			"volid": "local:backup/vzdump-qemu-100.vma.zst",
			"format": "vma.zst",
			"size": 2147483648,
			"ctime": 1693252800,
			"path": "/var/lib/vz/dump/vzdump-qemu-100.vma.zst"
		}}`)

	// DELETE per-volume endpoints used by (*ISO|*VzTmpl|*Backup).Delete.
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/storage/local/content/local:iso/debian-12\\.iso$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000010:00000010:00000010:imgdel:iso:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/storage/local/content/local:vztmpl/debian-12-standard\\.tar\\.zst$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000011:00000011:00000011:imgdel:vztmpl:root@pam:"}`)

	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/storage/local/content/local:backup/vzdump-qemu-100\\.vma\\.zst$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000012:00000012:00000012:imgdel:backup:root@pam:"}`)

	// deleteVolume regenerates volid from path when volid is empty. Register
	// the rebuilt path so the path-only branch is exercisable.
	gock.New(config.C.URI).
		Persist().
		Delete("^/nodes/node1/storage/local/content/local:backup/from-path\\.vma\\.zst$").
		Reply(200).
		JSON(`{"data": "UPID:node1:00000013:00000013:00000013:imgdel:backup:root@pam:"}`)

	// Per-volume GETs that intentionally omit volid in the body so the
	// ISO/VzTmpl/Import "regenerate VolID" branch is exercised.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content/local:iso/no-volid\\.iso$").
		Reply(200).
		JSON(`{"data": {"format": "iso", "size": 1, "path": "/var/lib/vz/template/iso/no-volid.iso"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/local/content/local:vztmpl/no-volid\\.tar\\.zst$").
		Reply(200).
		JSON(`{"data": {"format": "tar.zst", "size": 1, "path": "/var/lib/vz/template/cache/no-volid.tar.zst"}}`)

	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/storage/esxi/content/esxi:import/no-volid\\.vmx$").
		Reply(200).
		JSON(`{"data": {"format": "vmx", "size": 1, "path": "/mnt/esxi/no-volid.vmx"}}`)
}
