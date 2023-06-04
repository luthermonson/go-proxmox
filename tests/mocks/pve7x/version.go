package pve7x

import (
	"github.com/h2non/gock"
)

func version() {
	gock.New(config.TestURI).
		Get("/version").
		Reply(200).
		JSON(`
{
    "data": {
        "repoid": "777777",
        "release": "7.7",
        "version": "7.7-7"
    }
}`)
}
