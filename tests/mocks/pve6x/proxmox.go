package pve6x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

var config types.Config

func Load(c types.Config) {
	config = c

	version()
}

func version() {
	gock.New(config.TestURI).
		Get("/version").
		Reply(200).
		JSON(`{"data":{"repoid":"666666","release":"6.6","version":"6.6-6"}}`)
}
