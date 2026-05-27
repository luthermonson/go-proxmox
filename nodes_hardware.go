package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// /nodes/{node}/hardware — physical device inventory. PCI and USB only; the
// /hardware/pci/{id}/ branch is multi-instance per AGENTS.md and gets a
// *PCIDevice handle.

// HardwareIndex enumerates the hardware subtypes ("pci", "usb").
func (n *Node) HardwareIndex(ctx context.Context) ([]string, error) {
	var items []struct {
		Type string `json:"type"`
	}
	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/hardware", n.Name), &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Type)
	}
	return out, nil
}

// HardwarePCIOptions is the query payload for ListPCIDevices. ClassBlacklist
// is a list of PCI class hex codes to filter out (default server-side:
// "05;06;0b" — memory, bridge, processor). Verbose unset means the verbose
// default (1) — set Terse=true to get only PCI IDs.
type HardwarePCIOptions struct {
	ClassBlacklist []string
	Terse          bool
}

// ListPCIDevices returns the PCI devices on this node. Returned *PCIDevice
// handles are pre-populated with client/Node so callers can chain
// Index()/Mdev() directly.
func (n *Node) ListPCIDevices(ctx context.Context, opts *HardwarePCIOptions) (devices []*PCIDevice, err error) {
	path := fmt.Sprintf("/nodes/%s/hardware/pci", n.Name)
	if opts != nil {
		q := url.Values{}
		if len(opts.ClassBlacklist) > 0 {
			q.Set("pci-class-blacklist", strings.Join(opts.ClassBlacklist, ";"))
		}
		if opts.Terse {
			q.Set("verbose", "0")
		}
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}
	if err = n.client.Get(ctx, path, &devices); err != nil {
		return
	}
	for _, d := range devices {
		d.client = n.client
		d.Node = n.Name
	}
	return
}

// ListUSBDevices returns the USB devices on this node.
func (n *Node) ListUSBDevices(ctx context.Context) (devices []*USBDevice, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/hardware/usb", n.Name), &devices)
	return
}

// PCIDevice returns a handle for a single PCI device by id ("0000:01:00.0")
// or by cluster-mapping name. No API call.
func (n *Node) PCIDevice(id string) *PCIDevice {
	return &PCIDevice{
		client: n.client,
		Node:   n.Name,
		ID:     id,
	}
}

// Index enumerates the subresources of the PCI device — currently just
// ["mdev"].
func (d *PCIDevice) Index(ctx context.Context) ([]string, error) {
	if d.ID == "" {
		return nil, errors.New("pci id is required")
	}
	var items []struct {
		Method string `json:"method"`
	}
	if err := d.client.Get(ctx, fmt.Sprintf("/nodes/%s/hardware/pci/%s", d.Node, d.ID), &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Method)
	}
	return out, nil
}

// Mdev lists the mediated-device types this PCI device supports. Empty
// result for cards without SR-IOV/mdev capability.
func (d *PCIDevice) Mdev(ctx context.Context) (types []*PCIMdevType, err error) {
	if d.ID == "" {
		return nil, errors.New("pci id is required")
	}
	err = d.client.Get(ctx, fmt.Sprintf("/nodes/%s/hardware/pci/%s/mdev", d.Node, d.ID), &types)
	return
}
