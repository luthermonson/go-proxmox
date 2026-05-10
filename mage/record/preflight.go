package record

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	proxmox "github.com/luthermonson/go-proxmox"
)

// Preflight validates the outer-host configuration and reachability before
// any provisioning. Catches the failure modes that would otherwise show up
// as a 15-minute install timeout: missing tools, wrong VMID, unreachable
// nested-VM subnet, empty ISO digests.
//
// All checks run; the function returns an error only if at least one check
// FAILed (warnings are advisory). A clean run prints a green table and
// exits 0 — that's the precondition for `mage record:pve9` / `record:all`.
func Preflight() error {
	results := []checkResult{}

	cfg, err := LoadConfig()
	if err != nil {
		results = append(results, fail("config load", err.Error()))
		return reportResults(results)
	}
	results = append(results, ok("config load", "all required env vars set"))

	ctx := context.Background()

	// API + cluster-level checks.
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.OuterInsecure}, //nolint:gosec
		},
	}
	parts := splitTokenID(cfg.OuterTokenID)
	if parts == nil {
		results = append(results, fail("token id", `expected "user@realm!tokenname"`))
		return reportResults(results)
	}
	client := proxmox.NewClient(cfg.OuterURL,
		proxmox.WithHTTPClient(httpClient),
		proxmox.WithAPIToken(parts[0], cfg.OuterSecret),
	)

	results = append(results, checkAPI(ctx, client, cfg))
	node, nodeRes := checkNode(ctx, client, cfg)
	results = append(results, nodeRes)
	if node != nil {
		results = append(results, checkStorage(ctx, node, cfg))
		results = append(results, checkVMIDFree(ctx, node, cfg))
	}

	// SSH-based host-level checks.
	ssh, sshErr := NewSSHClient(cfg)
	if sshErr != nil {
		results = append(results, fail("ssh client", sshErr.Error()))
	} else {
		results = append(results, checkSSH(ssh, cfg))
		results = append(results, checkBridge(ssh, cfg))
		results = append(results, checkAutoInstallAssistant(ssh))
		results = append(results, checkNestedVirt(ssh))
	}

	// Local sanity checks (no remote calls).
	results = append(results, checkNestedNetwork(cfg))
	results = append(results, checkReleases()...)

	return reportResults(results)
}

// ── Individual checks ─────────────────────────────────────────────────────

func checkAPI(ctx context.Context, c *proxmox.Client, cfg *Config) checkResult {
	v, err := c.Version(ctx)
	if err != nil {
		return fail("outer Proxmox API", fmt.Sprintf("%s: %v", cfg.OuterURL, err))
	}
	return ok("outer Proxmox API", fmt.Sprintf("version=%s", v.Version))
}

func checkNode(ctx context.Context, c *proxmox.Client, cfg *Config) (*proxmox.Node, checkResult) {
	node, err := c.Node(ctx, cfg.OuterNode)
	if err != nil {
		return nil, fail(fmt.Sprintf("node %q", cfg.OuterNode), err.Error())
	}
	return node, ok(fmt.Sprintf("node %q", cfg.OuterNode), "found")
}

func checkStorage(ctx context.Context, node *proxmox.Node, cfg *Config) checkResult {
	st, err := node.Storage(ctx, cfg.OuterStorage)
	if err != nil {
		return fail(fmt.Sprintf("storage %q", cfg.OuterStorage), err.Error())
	}
	if !strings.Contains(st.Content, "iso") {
		return fail(fmt.Sprintf("storage %q", cfg.OuterStorage),
			fmt.Sprintf("does not advertise iso content (has %q)", st.Content))
	}
	const minBytes = 10 * 1024 * 1024 * 1024 // 10 GiB headroom for ISO + VM disk
	if st.Avail > 0 && st.Avail < minBytes {
		return warnf(fmt.Sprintf("storage %q", cfg.OuterStorage),
			fmt.Sprintf("avail=%dMB, recommend >10GB free", st.Avail/(1024*1024)))
	}
	availGB := "unknown"
	if st.Avail > 0 {
		availGB = fmt.Sprintf("%dGB", st.Avail/(1024*1024*1024))
	}
	return ok(fmt.Sprintf("storage %q", cfg.OuterStorage),
		fmt.Sprintf("avail=%s, supports iso", availGB))
}

func checkVMIDFree(ctx context.Context, node *proxmox.Node, cfg *Config) checkResult {
	name := fmt.Sprintf("VMID %d", cfg.NestedVMID)
	vms, err := node.VirtualMachines(ctx)
	if err != nil {
		return fail(name, fmt.Sprintf("list VMs: %v", err))
	}
	for _, vm := range vms {
		if int(vm.VMID) == cfg.NestedVMID {
			return fail(name, fmt.Sprintf("in use by qemu/%d (%s)", cfg.NestedVMID, vm.Name))
		}
	}
	cts, err := node.Containers(ctx)
	if err != nil {
		return fail(name, fmt.Sprintf("list CTs: %v", err))
	}
	for _, ct := range cts {
		if int(ct.VMID) == cfg.NestedVMID {
			return fail(name, fmt.Sprintf("in use by lxc/%d (%s)", cfg.NestedVMID, ct.Name))
		}
	}
	return ok(name, "available")
}

func checkBridge(ssh *SSHClient, cfg *Config) checkResult {
	// `ip -br link show <bridge>` exits 0 if the bridge exists and is up.
	// Run via SSH because /nodes/{node}/network requires permissions that
	// not every API token grants — the SSH path is the universal one.
	cmd := fmt.Sprintf("ip -br link show %q 2>/dev/null", cfg.NestedBridge)
	out, err := ssh.Run(cmd)
	if err != nil || strings.TrimSpace(out) == "" {
		return fail(fmt.Sprintf("bridge %q", cfg.NestedBridge),
			fmt.Sprintf("not found on outer host (or `ip` not in PATH)"))
	}
	return ok(fmt.Sprintf("bridge %q", cfg.NestedBridge), strings.TrimSpace(out))
}

func checkSSH(ssh *SSHClient, cfg *Config) checkResult {
	out, err := ssh.Run("hostname")
	if err != nil {
		return fail("SSH access", fmt.Sprintf("%s@%s: %v", cfg.SSHUser, cfg.SSHHost, err))
	}
	return ok("SSH access", fmt.Sprintf("%s@%s (%s)", cfg.SSHUser, cfg.SSHHost, strings.TrimSpace(out)))
}

func checkAutoInstallAssistant(ssh *SSHClient) checkResult {
	out, err := ssh.Run("which proxmox-auto-install-assistant 2>/dev/null || true")
	if err != nil {
		return fail("proxmox-auto-install-assistant", err.Error())
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return fail("proxmox-auto-install-assistant",
			"not found; install on outer host: apt install proxmox-auto-install-assistant")
	}
	return ok("proxmox-auto-install-assistant", out)
}

func checkNestedVirt(ssh *SSHClient) checkResult {
	out, err := ssh.Run(`cat /sys/module/kvm_intel/parameters/nested 2>/dev/null || ` +
		`cat /sys/module/kvm_amd/parameters/nested 2>/dev/null || echo missing`)
	if err != nil {
		return fail("nested virtualization", err.Error())
	}
	v := strings.TrimSpace(out)
	switch v {
	case "Y", "1":
		return ok("nested virtualization", "kvm.nested="+v)
	case "missing":
		return fail("nested virtualization",
			"neither kvm_intel nor kvm_amd module appears loaded on outer host")
	default:
		return fail("nested virtualization",
			fmt.Sprintf("kvm.nested=%q; enable with /etc/modprobe.d/kvm.conf and reboot", v))
	}
}

func checkNestedNetwork(cfg *Config) checkResult {
	name := fmt.Sprintf("IP %s", cfg.NestedIP)
	ip, ipnet, err := net.ParseCIDR(cfg.NestedIP)
	if err != nil {
		return fail(name, "must be in CIDR form (e.g. 10.0.10.250/24)")
	}
	gw := net.ParseIP(cfg.NestedGateway)
	if gw == nil {
		return fail(name, fmt.Sprintf("gateway %q is not a valid IP", cfg.NestedGateway))
	}
	if !ipnet.Contains(gw) {
		return fail(name,
			fmt.Sprintf("gateway %s not in subnet %s", gw, ipnet))
	}
	if ip.Equal(gw) {
		return fail(name, "IP and gateway are the same address")
	}
	if ip.IsLinkLocalUnicast() || ip.IsLoopback() {
		return fail(name, "IP is link-local or loopback")
	}
	// TEST-NET-1 (192.0.2.0/24) is the documentation default. If the user
	// hasn't overridden it, the recorder probably can't reach the nested
	// VM — TEST-NET-1 is non-routable on real LANs. Warn rather than fail
	// in case someone really did configure it as a host-only bridge.
	testNet := &net.IPNet{IP: net.IPv4(192, 0, 2, 0), Mask: net.CIDRMask(24, 32)}
	if testNet.Contains(ip) {
		return warnf(name, "in TEST-NET-1 (192.0.2.0/24); set PROXMOX_RECORDER_IP "+
			"to a real LAN address unless you've explicitly bridged it")
	}
	return ok(name, fmt.Sprintf("in subnet %s, gateway %s", ipnet, gw))
}

func checkReleases() []checkResult {
	out := []checkResult{}
	for _, major := range []string{"pve9", "pve8"} {
		r, has := Releases[major]
		name := fmt.Sprintf("%s ISO SHA256", major)
		if !has {
			out = append(out, fail(name, "missing from Releases map"))
			continue
		}
		if r.ISOSHA256 == "" {
			out = append(out, fail(name,
				fmt.Sprintf("empty in versions.go; populate from "+
					"https://www.proxmox.com/en/downloads (filename %s)", r.ISOFilename)))
			continue
		}
		out = append(out, ok(name, "populated"))
	}
	return out
}

// ── Result + reporting plumbing ───────────────────────────────────────────

type checkStatus int

const (
	statusOK checkStatus = iota
	statusWarn
	statusFail
)

type checkResult struct {
	name    string
	status  checkStatus
	message string
}

func ok(name, msg string) checkResult     { return checkResult{name, statusOK, msg} }
func warnf(name, msg string) checkResult  { return checkResult{name, statusWarn, msg} }
func fail(name, msg string) checkResult   { return checkResult{name, statusFail, msg} }

func (r checkResult) tag() string {
	switch r.status {
	case statusOK:
		return "[ OK ]"
	case statusWarn:
		return "[WARN]"
	default:
		return "[FAIL]"
	}
}

func reportResults(results []checkResult) error {
	// Compute the longest name so columns align across rows.
	maxName := 0
	for _, r := range results {
		if len(r.name) > maxName {
			maxName = len(r.name)
		}
	}

	failed, warned := 0, 0
	for _, r := range results {
		fmt.Printf("%s  %-*s   %s\n", r.tag(), maxName, r.name, r.message)
		switch r.status {
		case statusFail:
			failed++
		case statusWarn:
			warned++
		}
	}
	fmt.Println()
	switch {
	case failed > 0 && warned > 0:
		fmt.Printf("%s, %s. Fix failures and re-run mage record:preflight.\n",
			plural(failed, "failure", "failures"), plural(warned, "warning", "warnings"))
	case failed > 0:
		fmt.Printf("%s. Fix and re-run mage record:preflight.\n",
			plural(failed, "failure", "failures"))
	case warned > 0:
		fmt.Printf("%s. Recording will likely work, but review them above.\n",
			plural(warned, "warning", "warnings"))
	default:
		fmt.Println("All checks passed. Ready for mage record:pve9 / record:all.")
	}
	if failed > 0 {
		return errors.New("preflight failed")
	}
	return nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
