package record

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"
)

// ISOManager handles the upstream-ISO download cache and the answer-file
// remastering step. All operations target the outer PVE host: downloads use
// /storage/{name}/download-url, remastering shells out via SSH because the
// proxmox-auto-install-assistant CLI lives on the outer host.
type ISOManager struct {
	cfg     *Config
	client  *proxmox.Client
	storage *proxmox.Storage
	ssh     *SSHClient
}

// NewISOManager wires up the proxmox storage handle and the SSH client.
func NewISOManager(ctx context.Context, cfg *Config, client *proxmox.Client) (*ISOManager, error) {
	node, err := client.Node(ctx, cfg.OuterNode)
	if err != nil {
		return nil, fmt.Errorf("look up outer node %q: %w", cfg.OuterNode, err)
	}
	storage, err := node.Storage(ctx, cfg.OuterStorage)
	if err != nil {
		return nil, fmt.Errorf("look up outer storage %q: %w", cfg.OuterStorage, err)
	}
	ssh, err := NewSSHClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ISOManager{cfg: cfg, client: client, storage: storage, ssh: ssh}, nil
}

// EnsureUpstream downloads the upstream installer ISO into the outer host's
// storage if it isn't already present. Returns the storage volume id that
// can be attached to a VM (e.g. "local:iso/proxmox-ve_9.0-1.iso").
func (m *ISOManager) EnsureUpstream(ctx context.Context, r Release) (string, error) {
	volid := m.volID(r.ISOFilename)
	have, err := m.hasISO(ctx, volid)
	if err != nil {
		return "", err
	}
	if have {
		return volid, nil
	}

	if r.ISOSHA256 == "" {
		return "", fmt.Errorf("recorder: %s release has no SHA256 in versions.go; "+
			"populate it from the Proxmox downloads page before recording", r.Major)
	}

	task, err := m.storage.DownloadURLWithHash(ctx, "iso", r.ISOFilename, r.ISOURL,
		r.ISOSHA256, "sha256")
	if err != nil {
		return "", fmt.Errorf("start download of %s: %w", r.ISOURL, err)
	}
	if err := task.Wait(ctx, 5*time.Second, 30*time.Minute); err != nil {
		return "", fmt.Errorf("wait for ISO download: %w", err)
	}
	return volid, nil
}

// PrepareAutoInstall generates an answer.toml + first-boot script and runs
// proxmox-auto-install-assistant on the outer host to bake them into a new
// ISO. Returns the volume id of the prepared ISO.
//
// The prepared ISO has a deterministic filename (one per major) so re-runs
// overwrite the previous artifact rather than accumulating.
func (m *ISOManager) PrepareAutoInstall(ctx context.Context, r Release, upstreamVolid string) (string, error) {
	out, err := m.ssh.Run("which proxmox-auto-install-assistant")
	if err != nil || strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("proxmox-auto-install-assistant not found on %s; "+
			"install it with: apt install proxmox-auto-install-assistant", m.cfg.SSHHost)
	}

	// Resolve the on-disk path of the upstream ISO. For a "local" Proxmox
	// directory storage that's /var/lib/vz/template/iso/<filename>; for
	// other storages we ask the API. To stay generic, ask via `pvesm path`.
	upstreamPath, err := m.ssh.Run(fmt.Sprintf("pvesm path %q", upstreamVolid))
	if err != nil {
		return "", fmt.Errorf("resolve upstream ISO path: %w", err)
	}
	upstreamPath = strings.TrimSpace(upstreamPath)
	if upstreamPath == "" {
		return "", fmt.Errorf("pvesm path %q returned empty", upstreamVolid)
	}

	// Embed the first-boot script as a base64 data URL inside answer.toml so
	// the autoinstall process pulls it on first boot without us needing an
	// HTTP server reachable from the nested VM.
	script := FirstBootScript(m.cfg)
	scriptB64 := base64.StdEncoding.EncodeToString([]byte(script))
	answer := strings.Replace(AnswerTOML(m.cfg), "REPLACE_ME", scriptB64, 1)

	// Write answer.toml to a scratch path on the outer host.
	answerPath := TempPath("answer") + ".toml"
	if err := m.ssh.WriteFile(answerPath, []byte(answer)); err != nil {
		return "", fmt.Errorf("write answer.toml: %w", err)
	}
	defer func() { _, _ = m.ssh.Run(fmt.Sprintf("rm -f %q", answerPath)) }()

	preparedFilename := fmt.Sprintf("auto-%s.iso", r.Major)
	preparedPath := strings.Replace(upstreamPath, r.ISOFilename, preparedFilename, 1)

	cmd := fmt.Sprintf(
		"proxmox-auto-install-assistant prepare-iso %q --fetch-from iso "+
			"--answer-file %q --output %q",
		upstreamPath, answerPath, preparedPath)
	if _, err := m.ssh.Run(cmd); err != nil {
		return "", fmt.Errorf("prepare-iso: %w", err)
	}

	return m.volID(preparedFilename), nil
}

// hasISO returns true if the storage already contains the named volume id.
func (m *ISOManager) hasISO(ctx context.Context, volid string) (bool, error) {
	contents, err := m.storage.GetContent(ctx)
	if err != nil {
		return false, fmt.Errorf("list storage contents: %w", err)
	}
	for _, e := range contents {
		if e.Volid == volid {
			return true, nil
		}
	}
	return false, nil
}

func (m *ISOManager) volID(filename string) string {
	return fmt.Sprintf("%s:iso/%s", m.cfg.OuterStorage, filename)
}
