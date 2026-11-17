package apidrift

import (
	"reflect"

	proxmox "github.com/luthermonson/go-proxmox"
)

// Registry is the curated endpoint→Go-type map apidrift diffs against the
// schema. Paths and methods are written exactly as apidoc.js spells them
// (`{node}`, `{vmid}`, …). For array responses the mapped type is the element
// type (e.g. GET .../content → StorageContent, decoded into []StorageContent).
//
// This is deliberately a seed, not the full API surface: the library does not
// annotate which struct decodes which endpoint, so a complete automatic
// mapping is impossible. Endpoints absent here surface in the report as
// "unmapped" so they can be wired up over time (that is what the weekly
// workflow tracks). Add a line here whenever you point a Go type at a new
// endpoint.
func Registry() []Mapping {
	t := reflect.TypeOf
	return []Mapping{
		{Method: "GET", Path: "/version", Type: t(proxmox.Version{})},
		{Method: "GET", Path: "/nodes/{node}/status", Type: t(proxmox.NodeStatus{})},
		{Method: "GET", Path: "/nodes/{node}/dns", Type: t(proxmox.NodeDNS{})},
		{Method: "GET", Path: "/nodes/{node}/certificates/info", Type: t(proxmox.NodeCertificate{})},
		{Method: "GET", Path: "/nodes/{node}/disks/list", Type: t(proxmox.Disk{})},
		{Method: "GET", Path: "/nodes/{node}/storage/{storage}/status", Type: t(proxmox.StorageStatus{})},
		{Method: "GET", Path: "/nodes/{node}/storage/{storage}/content", Type: t(proxmox.StorageContent{})},
		{Method: "GET", Path: "/nodes/{node}/lxc/{vmid}/migrate", Type: t(proxmox.ContainerMigratePreconditions{})},
		{Method: "GET", Path: "/cluster/ceph/flags", Type: t(proxmox.CephFlag{})},
		{Method: "GET", Path: "/access/users", Type: t(proxmox.User{})},
		{Method: "GET", Path: "/access/users/{userid}", Type: t(proxmox.User{})},
		{Method: "GET", Path: "/cluster/sdn/zones", Type: t(proxmox.SDNZone{})},
		{Method: "GET", Path: "/cluster/sdn/zones/{zone}", Type: t(proxmox.SDNZone{})},
	}
}
