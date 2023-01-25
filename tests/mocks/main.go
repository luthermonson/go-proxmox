package mocks

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/types"
)

const TestURI = "http://test.localhost"

func data(d map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": d,
	}
}

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

func LoadRoutes(config types.Config, routes types.Routes) {
	for _, route := range routes {
		req := gock.New(config.TestURI)
		if route.Path != "" && route.Method != "" {
			switch route.Method {
			case "GET":
				req.Get(route.Path)
			case "POST":
				req.Post(route.Path)
			case "PUT":
				req.Put(route.Path)
			case "DELETE":
				req.Delete(route.Path)
			}
		}
		if route.MatchType != "" {
			req.MatchType(route.MatchType)
		}
		if route.Reply > 0 {
			_ = req.Reply(route.Reply)
		}
		if route.JSON != nil {
			req.Response.JSON(route.JSON)
		}
		if route.BodyString != "" {
			req.Response.BodyString(route.BodyString)
		}
	}
}
