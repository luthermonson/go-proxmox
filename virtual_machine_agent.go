package proxmox

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
)

// This file wraps /nodes/{node}/qemu/{vmid}/agent/* — the QEMU guest-agent
// surface. Pre-existing agent helpers (Ping, AgentExec, AgentExecStatus,
// AgentGetHostName, AgentGetNetworkIFaces, AgentOsInfo, AgentSetUserPassword,
// WaitForAgent, WaitForAgentExecExit) live in virtual_machine.go and are not
// duplicated here.
//
// PVE wraps most QGA responses in a single {"result": ...} envelope on top of
// the standard {"data": ...} envelope the client already strips. The helpers
// below unwrap that inner "result" before returning a typed value.

// AgentCommandIndex lists the QEMU guest-agent sub-commands the PVE node
// exposes for this VM. This is the index served at GET /agent — it surfaces
// the routing table the proxy itself knows about and is independent of what
// the in-guest agent actually advertises (see AgentGetInfo for that).
func (v *VirtualMachine) AgentCommandIndex(ctx context.Context) ([]*AgentCommandIndexEntry, error) {
	var out []*AgentCommandIndexEntry
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent", v.Node, v.VMID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AgentCommand is the low-level POST /agent helper that runs any QGA command
// by name. The specific helpers (AgentPing, AgentGetTime, …) are preferable
// for known commands because they return typed values; this exists for the
// rare command not otherwise wrapped (or for forward-compat with new QGA
// verbs PVE adds). PVE constrains the accepted names — see the enum on the
// API endpoint — but we don't gate that here so callers stay forward-compat.
// The raw {"result": ...} envelope payload is returned as a JSON map.
func (v *VirtualMachine) AgentCommand(ctx context.Context, command string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"command": command,
	}
	var resp struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent", v.Node, v.VMID), body, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentPing is a cheap probe that the guest-agent socket is reachable. PVE
// returns an empty result on success; any non-nil error means the agent is
// unreachable or the VM is off.
func (v *VirtualMachine) AgentPing(ctx context.Context) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/ping", v.Node, v.VMID), nil, nil)
}

// AgentGetTime returns the guest's wall-clock time in nanoseconds since the
// Unix epoch.
func (v *VirtualMachine) AgentGetTime(ctx context.Context) (AgentTime, error) {
	var resp struct {
		Result AgentTime `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-time", v.Node, v.VMID), &resp); err != nil {
		return 0, err
	}
	return resp.Result, nil
}

// AgentGetTimezone returns the guest IANA timezone string (e.g. "UTC").
// QGA reports both a string zone and a UTC offset; only the zone is exposed
// here — callers needing the offset can use AgentInfo to fall back.
func (v *VirtualMachine) AgentGetTimezone(ctx context.Context) (string, error) {
	var resp struct {
		Result struct {
			Zone   string `json:"zone"`
			Offset int    `json:"offset"`
		} `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-timezone", v.Node, v.VMID), &resp); err != nil {
		return "", err
	}
	return resp.Result.Zone, nil
}

// AgentGetUsers lists users currently logged into the guest.
func (v *VirtualMachine) AgentGetUsers(ctx context.Context) ([]*AgentUser, error) {
	var resp struct {
		Result []*AgentUser `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-users", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentGetVCPUs lists the guest's logical CPUs and their online state.
func (v *VirtualMachine) AgentGetVCPUs(ctx context.Context) ([]*AgentVCPU, error) {
	var resp struct {
		Result []*AgentVCPU `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-vcpus", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentGetFsInfo returns one entry per mounted filesystem in the guest, with
// disk usage and the underlying block device topology.
func (v *VirtualMachine) AgentGetFsInfo(ctx context.Context) ([]*AgentFsInfo, error) {
	var resp struct {
		Result []*AgentFsInfo `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-fsinfo", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentGetMemoryBlocks lists hot-pluggable memory blocks visible to the
// guest. Older kernels / non-Linux guests return an empty slice.
func (v *VirtualMachine) AgentGetMemoryBlocks(ctx context.Context) ([]*AgentMemoryBlock, error) {
	var resp struct {
		Result []*AgentMemoryBlock `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-memory-blocks", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentGetMemoryBlockInfo returns hot-pluggable memory-block sizing for the
// guest — the per-block byte size that AgentGetMemoryBlocks's phys-index
// references. Returns zero on guests without memory-hotplug support.
func (v *VirtualMachine) AgentGetMemoryBlockInfo(ctx context.Context) (*AgentMemoryBlockInfo, error) {
	var resp struct {
		Result *AgentMemoryBlockInfo `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/get-memory-block-info", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentGetInfo returns the guest-agent's own version and the list of QGA
// commands it advertises. Useful to feature-detect before issuing an op the
// agent may not support.
func (v *VirtualMachine) AgentGetInfo(ctx context.Context) (*AgentInfo, error) {
	var resp struct {
		Result *AgentInfo `json:"result"`
	}
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/info", v.Node, v.VMID), &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentFsfreezeFreeze freezes all guest filesystems; returns the number of
// filesystems frozen. Pair with AgentFsfreezeThaw — leaving a guest frozen
// will hang every write inside it.
func (v *VirtualMachine) AgentFsfreezeFreeze(ctx context.Context) (int, error) {
	var resp struct {
		Result int `json:"result"`
	}
	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/fsfreeze-freeze", v.Node, v.VMID), nil, &resp); err != nil {
		return 0, err
	}
	return resp.Result, nil
}

// AgentFsfreezeThaw thaws all previously-frozen guest filesystems and
// returns the count thawed.
func (v *VirtualMachine) AgentFsfreezeThaw(ctx context.Context) (int, error) {
	var resp struct {
		Result int `json:"result"`
	}
	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/fsfreeze-thaw", v.Node, v.VMID), nil, &resp); err != nil {
		return 0, err
	}
	return resp.Result, nil
}

// AgentFsfreezeStatus reports whether the guest filesystems are currently
// frozen ("thawed" or "frozen").
func (v *VirtualMachine) AgentFsfreezeStatus(ctx context.Context) (AgentFsfreezeStatus, error) {
	var resp struct {
		Result AgentFsfreezeStatus `json:"result"`
	}
	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/fsfreeze-status", v.Node, v.VMID), nil, &resp); err != nil {
		return "", err
	}
	return resp.Result, nil
}

// AgentFstrim issues fstrim across all mounted guest filesystems. PVE
// returns a per-mountpoint trim report which we expose as the raw JSON map.
func (v *VirtualMachine) AgentFstrim(ctx context.Context) (map[string]interface{}, error) {
	var resp struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/fstrim", v.Node, v.VMID), nil, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// AgentShutdown asks the guest-agent to shut the guest down cleanly. Unlike
// VirtualMachine.Shutdown (which goes through QEMU and returns a Task), this
// is a synchronous QGA call — the guest may take many seconds to actually
// halt after this returns.
func (v *VirtualMachine) AgentShutdown(ctx context.Context) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/shutdown", v.Node, v.VMID), nil, nil)
}

// AgentSuspendDisk suspends the guest to disk (S4 / hibernation).
func (v *VirtualMachine) AgentSuspendDisk(ctx context.Context) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/suspend-disk", v.Node, v.VMID), nil, nil)
}

// AgentSuspendHybrid suspends the guest to a hybrid sleep state (RAM + disk).
func (v *VirtualMachine) AgentSuspendHybrid(ctx context.Context) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/suspend-hybrid", v.Node, v.VMID), nil, nil)
}

// AgentSuspendRAM suspends the guest to RAM (S3).
func (v *VirtualMachine) AgentSuspendRAM(ctx context.Context) error {
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/suspend-ram", v.Node, v.VMID), nil, nil)
}

// AgentFileRead reads a file from the guest via the guest-agent. The PVE
// endpoint caps the response at 16 MiB and exposes a Truncated flag so
// callers can detect partial reads. Note: unlike most agent endpoints, the
// response is NOT wrapped in a "result" envelope here — content and
// truncated sit at the top level under "data".
func (v *VirtualMachine) AgentFileRead(ctx context.Context, file string) (*AgentFileRead, error) {
	// Path is a free-form filesystem path, so url-encode it.
	q := url.Values{}
	q.Set("file", file)
	var out AgentFileRead
	if err := v.client.Get(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-read?%s", v.Node, v.VMID, q.Encode()), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AgentFileWrite writes content to a file inside the guest. QGA's
// guest-file-write requires a base64 payload, so we encode here (binary-safe)
// and pass encode=0 to tell PVE the content is already base64 and to forward
// it to QGA verbatim. With encode=1 PVE would base64-encode our payload a
// second time, and QGA's single decode would leave the base64 text on disk.
// The per-call body size cap is ~60 KiB. Pass the raw bytes; this helper
// handles the encoding.
func (v *VirtualMachine) AgentFileWrite(ctx context.Context, file string, content []byte) error {
	body := map[string]interface{}{
		"file":    file,
		"content": base64.StdEncoding.EncodeToString(content),
		"encode":  0,
	}
	return v.client.Post(ctx, fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-write", v.Node, v.VMID), body, nil)
}
