package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Wrappers for /cluster/qemu/cpu-flags and /cluster/qemu/custom-cpu-models —
// the QEMU CPU flag catalog and CRUD for custom CPU model definitions used by
// `cpu: custom-<name>` in VM configs.

// QEMUCPUFlags returns the catalog of CPU flags available across the cluster
// (currently x86_64 only — aarch64 returns an empty list). arch and accel are
// optional; pass "" to use PVE defaults (host arch + kvm acceleration).
//
// GET /cluster/qemu/cpu-flags
func (cl *Cluster) QEMUCPUFlags(ctx context.Context, arch, accel string) (flags []*QEMUCPUFlag, err error) {
	path := "/cluster/qemu/cpu-flags"
	q := url.Values{}
	if arch != "" {
		q.Set("arch", arch)
	}
	if accel != "" {
		q.Set("accel", accel)
	}
	if enc := q.Encode(); enc != "" {
		path = path + "?" + enc
	}
	err = cl.client.Get(ctx, path, &flags)
	return
}

// CustomCPUModels lists configured custom CPU model definitions. Only entries
// the caller has Mapping.{Audit,Use,Modify} on are returned.
//
// GET /cluster/qemu/custom-cpu-models
func (cl *Cluster) CustomCPUModels(ctx context.Context) (models []*CustomCPUModel, err error) {
	if err = cl.client.Get(ctx, "/cluster/qemu/custom-cpu-models", &models); err != nil {
		return nil, err
	}
	for _, m := range models {
		m.client = cl.client
	}
	return
}

// CustomCPUModel returns a handle for a single custom CPU model. No API call.
// The "custom-" prefix on cputype is optional per PVE.
//
// GET /cluster/qemu/custom-cpu-models/{cputype}
func (cl *Cluster) CustomCPUModel(cputype string) *CustomCPUModel {
	return &CustomCPUModel{client: cl.client, CPUType: cputype}
}

// NewCustomCPUModel creates a new custom CPU model. opts.CPUType is required;
// opts.ReportedModel is required by PVE per the schema (optional=0).
//
// POST /cluster/qemu/custom-cpu-models
func (cl *Cluster) NewCustomCPUModel(ctx context.Context, opts *CustomCPUModelOptions) error {
	if opts == nil || opts.CPUType == "" {
		return errors.New("custom cpu model: cputype is required")
	}
	if opts.ReportedModel == "" {
		return errors.New("custom cpu model: reported-model is required")
	}
	return cl.client.Post(ctx, "/cluster/qemu/custom-cpu-models", opts, nil)
}

// Read populates the receiver with the current definition.
//
// GET /cluster/qemu/custom-cpu-models/{cputype}
func (m *CustomCPUModel) Read(ctx context.Context) error {
	if m.CPUType == "" {
		return errors.New("custom cpu model: cputype is required")
	}
	return m.client.Get(ctx, fmt.Sprintf("/cluster/qemu/custom-cpu-models/%s", m.CPUType), m)
}

// Update mutates an existing custom CPU model.
//
// PUT /cluster/qemu/custom-cpu-models/{cputype}
func (m *CustomCPUModel) Update(ctx context.Context, opts *CustomCPUModelOptions) error {
	if m.CPUType == "" {
		return errors.New("custom cpu model: cputype is required")
	}
	if opts == nil {
		opts = &CustomCPUModelOptions{}
	}
	return m.client.Put(ctx, fmt.Sprintf("/cluster/qemu/custom-cpu-models/%s", m.CPUType), opts, nil)
}

// Delete removes a custom CPU model definition.
//
// DELETE /cluster/qemu/custom-cpu-models/{cputype}
func (m *CustomCPUModel) Delete(ctx context.Context) error {
	if m.CPUType == "" {
		return errors.New("custom cpu model: cputype is required")
	}
	return m.client.Delete(ctx, fmt.Sprintf("/cluster/qemu/custom-cpu-models/%s", m.CPUType), nil)
}
