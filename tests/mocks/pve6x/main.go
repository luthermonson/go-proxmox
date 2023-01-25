package pve6x

import (
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

var routes types.Routes

func init() {
	// access
	// cluster
	// nodes
	// pools
	// storage
	version()
}

func Routes() types.Routes {
	return routes
}

func version() {
	routes = append(routes, types.Route{
		Method: "GET",
		Path:   "/version",
		Reply:  200,
		JSON:   `{"data":{"repoid":"666666","release":"6.6","version":"6.6-6"}}`,
	})
}
