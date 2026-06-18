package record

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"
)

// Pipeline owns the end-to-end record loop for a single PVE major:
//
//  1. Build an outer-host API client and an SSH client.
//  2. Ensure the upstream installer ISO is on the outer host (download if
//     missing, validate SHA256 if specified).
//  3. Re-master the ISO with answer.toml + first-boot seed embedded.
//  4. Provision the nested VM with the prepared ISO and wait for the API.
//  5. Run the recorder driver against the nested PVE; cassettes land under
//     tests/recorder/testdata/<release.CassetteDir>/.
//  6. Always destroy the nested VM on exit.
type Pipeline struct {
	cfg     *Config
	release Release

	outer   *proxmox.Client
	ssh     *SSHClient
	iso     *ISOManager
	vm      *VMLifecycle
}

// NewPipeline constructs the dependencies. Errors here are configuration
// problems (missing env, bad SSH key) rather than runtime ones.
func NewPipeline(ctx context.Context, cfg *Config, release Release) (*Pipeline, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.OuterInsecure}, //nolint:gosec
		},
	}
	parts := splitTokenID(cfg.OuterTokenID)
	if parts == nil {
		return nil, fmt.Errorf("PROXMOX_TOKENID %q is not in <user>@<realm>!<token-id> format", cfg.OuterTokenID)
	}
	outer := proxmox.NewClient(cfg.OuterURL,
		proxmox.WithHTTPClient(httpClient),
		proxmox.WithAPIToken(parts[0], cfg.OuterSecret),
	)

	ssh, err := NewSSHClient(cfg)
	if err != nil {
		return nil, err
	}
	iso, err := NewISOManager(ctx, cfg, outer)
	if err != nil {
		return nil, err
	}
	vm, err := NewVMLifecycle(ctx, cfg, outer)
	if err != nil {
		return nil, err
	}
	return &Pipeline{cfg: cfg, release: release, outer: outer, ssh: ssh, iso: iso, vm: vm}, nil
}

// Run executes the pipeline end-to-end. The nested VM is destroyed on
// every exit path, including failure during provisioning or recording, and
// the local HTTP server is shut down on the way out.
func (p *Pipeline) Run(ctx context.Context) (err error) {
	fmt.Printf("[record:%s] downloading upstream ISO if needed\n", p.release.Major)
	upstream, err := p.iso.EnsureUpstream(ctx, p.release)
	if err != nil {
		return fmt.Errorf("ensure upstream ISO: %w", err)
	}

	// Local HTTP server hosts answer.toml + first-boot.sh for the duration
	// of the install. Compute the URLs from config first so the answer
	// file can reference its own first-boot URL before we bind anything.
	urls, err := PlanURLs(p.cfg)
	if err != nil {
		return fmt.Errorf("plan http server URLs: %w", err)
	}
	answer := []byte(AnswerTOML(p.cfg, urls.FirstBootURL))
	firstBoot := []byte(FirstBootScript(p.cfg))

	fmt.Printf("[record:%s] starting local HTTP server on %s\n", p.release.Major, urls.BaseURL)
	httpServer, err := NewFileServer(p.cfg, answer, firstBoot)
	if err != nil {
		return fmt.Errorf("start http server: %w", err)
	}
	defer func() { _ = httpServer.Stop() }()

	fmt.Printf("[record:%s] preparing autoinstall ISO with fetch-from-http URL\n", p.release.Major)
	prepared, err := p.iso.PrepareAutoInstall(ctx, p.release, upstream, urls.AnswerURL)
	if err != nil {
		return fmt.Errorf("prepare auto-install ISO: %w", err)
	}

	fmt.Printf("[record:%s] provisioning nested VM (this typically takes 5-15 minutes)\n", p.release.Major)
	if err := p.vm.Provision(ctx, prepared); err != nil {
		return fmt.Errorf("provision nested VM: %w", err)
	}
	defer func() {
		fmt.Printf("[record:%s] destroying nested VM\n", p.release.Major)
		if dErr := p.vm.Destroy(ctx); dErr != nil && err == nil {
			err = fmt.Errorf("destroy nested VM: %w", dErr)
		}
	}()

	if err := p.waitForSeed(ctx); err != nil {
		return fmt.Errorf("wait for first-boot seed: %w", err)
	}

	fmt.Printf("[record:%s] driving recorder against nested PVE\n", p.release.Major)
	if err := p.recordCassettes(ctx); err != nil {
		return fmt.Errorf("record cassettes: %w", err)
	}

	fmt.Printf("[record:%s] done; cassettes in tests/recorder/testdata/%s/\n", p.release.Major, p.release.CassetteDir)
	return nil
}

// waitForSeed polls /var/lib/go-proxmox-seeded on the nested PVE via SSH
// until the file appears, signalling the first-boot seed script has
// completed. Times out after 10 minutes.
func (p *Pipeline) waitForSeed(ctx context.Context) error {
	nestedSSH := *p.cfg
	nestedSSH.SSHHost = hostOnly(p.vm.NestedURL())
	client, err := NewSSHClient(&nestedSSH)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		if _, err := client.Run("test -f /var/lib/go-proxmox-seeded"); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
		}
	}
	return fmt.Errorf("seed marker /var/lib/go-proxmox-seeded never appeared")
}

// recordCassettes runs `go test ./tests/recorder/...` with the env vars
// the recorder package consumes (PROXMOX_URL points at the nested PVE,
// PROXMOX_TOKENID/PROXMOX_SECRET are the recorder@pve token from seed).
//
// The release's CassetteDir is exposed via PROXMOX_RECORDER_CASSETTE_DIR
// so tests can scope their cassettes per major.
func (p *Pipeline) recordCassettes(ctx context.Context) error {
	dir := filepath.Join("tests", "recorder", "testdata", p.release.CassetteDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Force a fresh capture: ModeRecordOnce only records when the cassette
	// file is missing. Removing existing per-release cassettes is the
	// supported "re-record everything" workflow.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			if rmErr := os.Remove(filepath.Join(dir, e.Name())); rmErr != nil {
				return rmErr
			}
		}
	}

	cmd := exec.CommandContext(ctx, "go", "test", "./tests/recorder/...", "-count=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"PROXMOX_URL="+p.vm.NestedURL(),
		"PROXMOX_TOKENID=recorder@pve!cassettes",
		"PROXMOX_SECRET="+p.cfg.NestedTokenSecret,
		"PROXMOX_RECORDER_CASSETTE_DIR="+p.release.CassetteDir,
	)
	return cmd.Run()
}

// splitTokenID parses "user@realm!tokenname" into ["user@realm", "tokenname"].
// Returns nil on malformed input.
func splitTokenID(t string) []string {
	for i := len(t) - 1; i >= 0; i-- {
		if t[i] == '!' {
			return []string{t[:i], t[i+1:]}
		}
	}
	return nil
}
