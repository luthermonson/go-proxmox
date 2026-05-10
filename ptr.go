package proxmox

// Ptr returns a pointer to v. It exists so callers can construct optional
// fields on config/options structs inline without a temporary local variable:
//
//	cfg := &proxmox.ContainerConfig{
//		Arch:    proxmox.Ptr("amd64"),
//		Console: proxmox.Ptr(proxmox.IntOrBool(true)),
//	}
//
// Pointer-typed fields on these structs distinguish "unset, use the PVE
// server-side default" (nil) from "explicitly send this value" (non-nil) —
// without that distinction, a Go zero value silently overrides the server
// default at marshal time. See issue #199.
func Ptr[T any](v T) *T { return &v }
