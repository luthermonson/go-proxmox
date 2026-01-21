package pve9x

import (
	"github.com/h2non/gock"
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
}
