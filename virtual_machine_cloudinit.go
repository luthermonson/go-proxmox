package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// VirtualMachineCloudInitPending is one row from GET /qemu/{vmid}/cloudinit —
// a single config key with its currently-applied value and a pending value if
// a regenerate is needed. Delete=1 means the key is pending removal.
type VirtualMachineCloudInitPending struct {
	Delete  IntOrBool `json:"delete,omitempty"`
	Key     string    `json:"key,omitempty"`
	Pending string    `json:"pending,omitempty"`
	Value   string    `json:"value,omitempty"`
}

// CloudInitPending lists per-VM cloud-init config differences between the
// applied image and the current VM config. Empty list means the image is in
// sync with the config.
func (v *VirtualMachine) CloudInitPending(ctx context.Context) (pending []*VirtualMachineCloudInitPending, err error) {
	return pending, v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/cloudinit", v.Node, v.VMID), &pending)
}

// CloudInitRegenerate rewrites the cloud-init ISO/cd-rom so the next guest
// boot picks up pending changes. Synchronous — PVE returns null on success.
func (v *VirtualMachine) CloudInitRegenerate(ctx context.Context) error {
	return v.client.Put(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/cloudinit", v.Node, v.VMID), nil, nil)
}

// CloudInitDump returns the rendered cloud-init configuration document of a
// given kind. kind must be one of "user", "meta", "network".
func (v *VirtualMachine) CloudInitDump(ctx context.Context, kind string) (string, error) {
	if kind == "" {
		return "", errors.New("cloudinit dump type is required (user|meta|network)")
	}
	q := url.Values{}
	q.Set("type", kind)
	var out string
	err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/cloudinit/dump?%s", v.Node, v.VMID, q.Encode()), &out)
	return out, err
}
