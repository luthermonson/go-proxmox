package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

var config types.Config

func Load(c types.Config) {
	config = c

	version()
	index()
}

func index() {
	indexContent := `{
    "data": [
        {
            "subdir": "version"
        },
        {
            "subdir": "cluster"
        },
        {
            "subdir": "nodes"
        },
        {
            "subdir": "storage"
        },
        {
            "subdir": "access"
        },
        {
            "subdir": "pools"
        }
    ]}`

	gock.New(config.TestURI).
		Get("/").
		Reply(200).
		JSON(indexContent)

	gock.New(config.TestURI).
		Post("/"). // fake to test client Post method
		Reply(200).
		JSON(indexContent)

	gock.New(config.TestURI).
		Put("/"). // fake to test client Put method
		Reply(200).
		JSON(indexContent)

	gock.New(config.TestURI).
		Delete("/"). // fake to test client Delete method
		Reply(200).
		JSON(indexContent)
}
