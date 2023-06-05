package mocks

import (
	"github.com/luthermonson/go-proxmox/tests/mocks/pve6x"
	"github.com/luthermonson/go-proxmox/tests/mocks/pve7x"
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

func ProxmoxVE7x(config types.Config) {
	pve7x.Load(config)
}

func ProxmoxVE6x(config types.Config) {
	pve6x.Load(config)
}
