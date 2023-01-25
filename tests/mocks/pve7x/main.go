package pve7x

import (
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

func init() {
	version()
	index()
}

var routes types.Routes

func Routes() types.Routes {
	return routes
}

func r(r ...types.Route) {
	routes = append(routes, r...)
}

func version() {
	r(types.Route{
		Method: "GET",
		Path:   "/version",
		Reply:  200,
		JSON: `
{
    "data": {
        "repoid": "777777",
        "release": "7.7",
        "version": "7.7-7"
    }
}`,
	})
}

func index() {
	r(types.Route{
		Method: "GET",
		Path:   "/",
		Reply:  200,
		JSON: `
{
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
    ]
}`,
	})
}
