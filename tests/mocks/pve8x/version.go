package pve8x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func version() {
	versionJSON := `
{
    "data": {
        "repoid": "761609f7",
        "release": "8.4",
        "version": "8.4-1"
    }
}`
	gock.New(config.C.URI).
		Get("^/version$").
		Reply(200).
		JSON(versionJSON)

	gock.New(config.C.URI).
		Post("^/version$"). // fake to test client Post method
		Reply(200).
		JSON(versionJSON)

	gock.New(config.C.URI).
		Put("^/version$"). // fake to test client Put method
		Reply(200).
		JSON(versionJSON)

	gock.New(config.C.URI).
		Delete("^/version$"). // fake to test client Delete method
		Reply(200).
		JSON(versionJSON)
}
