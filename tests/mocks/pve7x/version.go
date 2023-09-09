package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func version() {
	versionJSON := `
{
    "data": {
        "repoid": "777777",
        "release": "7.7",
        "version": "7.7-7"
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
