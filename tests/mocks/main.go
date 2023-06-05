package mocks

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

func On(config types.Config) {
	if config.Version != nil {
		config.Version(config)
		return
	}

	ProxmoxVE7x(config) // default latest
}

func Off() {
	gock.Off()
}
