package pve6x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func Load() {
	version()
}

func version() {
	gock.New(config.C.URI).
		Get("/version").
		Reply(200).
		JSON(`{"data":{"repoid":"666666","release":"6.6","version":"6.6-6"}}`)
}
