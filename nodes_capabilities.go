package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// /nodes/{node}/capabilities — node-level "what can this host run".
// Currently only QEMU lives under here; arch parameter defaults to the host
// architecture server-side.

// CapabilitiesIndex enumerates the top-level capability subsystems ("qemu").
func (n *Node) CapabilitiesIndex(ctx context.Context) ([]string, error) {
	return n.capabilitiesDiridx(ctx, fmt.Sprintf("/nodes/%s/capabilities", n.Name))
}

// QEMUCapabilitiesIndex enumerates the QEMU capability subresources
// ("cpu", "cpu-flags", "machines", "migration").
func (n *Node) QEMUCapabilitiesIndex(ctx context.Context) ([]string, error) {
	return n.capabilitiesDiridx(ctx, fmt.Sprintf("/nodes/%s/capabilities/qemu", n.Name))
}

// QEMUCPUModels lists all available CPU models, both built-in QEMU types and
// any custom models defined on the cluster. arch is "" (host default),
// "x86_64", or "aarch64".
func (n *Node) QEMUCPUModels(ctx context.Context, arch string) (models []*QEMUCPUModel, err error) {
	path := fmt.Sprintf("/nodes/%s/capabilities/qemu/cpu", n.Name)
	if arch != "" {
		q := url.Values{}
		q.Set("arch", arch)
		path = path + "?" + q.Encode()
	}
	err = n.client.Get(ctx, path, &models)
	return
}

// QEMUCPUFlags lists VM-visible CPU flags supported on this node. accel is
// "kvm" (default) or "tcg"; arch is "" / "x86_64" / "aarch64". aarch64
// returns an empty list per PVE — no VM-specific flags are defined yet.
func (n *Node) QEMUCPUFlags(ctx context.Context, arch, accel string) (flags []*QEMUCPUFlag, err error) {
	q := url.Values{}
	if arch != "" {
		q.Set("arch", arch)
	}
	if accel != "" {
		q.Set("accel", accel)
	}
	path := fmt.Sprintf("/nodes/%s/capabilities/qemu/cpu-flags", n.Name)
	if len(q) > 0 {
		path = path + "?" + q.Encode()
	}
	err = n.client.Get(ctx, path, &flags)
	return
}

// QEMUMachineTypes returns the q35 / i440fx versions available on this host.
// arch is "" (host default), "x86_64", or "aarch64".
func (n *Node) QEMUMachineTypes(ctx context.Context, arch string) (types []*QEMUMachineType, err error) {
	path := fmt.Sprintf("/nodes/%s/capabilities/qemu/machines", n.Name)
	if arch != "" {
		q := url.Values{}
		q.Set("arch", arch)
		path = path + "?" + q.Encode()
	}
	err = n.client.Get(ctx, path, &types)
	return
}

// QEMUMigrationCapabilities returns node-specific live migration features —
// currently just whether dbus-vmstate is available for live-migrating
// additional VM state.
func (n *Node) QEMUMigrationCapabilities(ctx context.Context) (caps *QEMUMigrationCapabilities, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/capabilities/qemu/migration", n.Name), &caps)
	return
}

func (n *Node) capabilitiesDiridx(ctx context.Context, path string) ([]string, error) {
	// PVE schema declares the items as `{}` (no properties) but in practice
	// emits {"name": ...} for both indexes. Decode loosely.
	var items []map[string]string
	if err := n.client.Get(ctx, path, &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		if v, ok := it["name"]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}
