// Package record provides Mage targets for recording the cassettes consumed
// by tests under tests/recorder/. The pipeline self-hosts: it uses
// go-proxmox itself to drive an outer Proxmox host, spin up a nested PVE
// via the unattended autoinstall flow, run the recorder against it, and
// destroy the nested VM.
//
// Targets:
//
//	mage record:pve9     Record cassettes against a fresh PVE 9 install.
//	mage record:pve8     Record cassettes against a fresh PVE 8 install.
//	mage record:all      Record both pve9 and pve8 in series.
//	mage record:verify   Replay every checked-in cassette without touching the
//	                     network. Run in CI to guard against drift.
//	mage record:plan     Print what mage record:all would do; touches nothing.
//	mage record:smoke    Run only the recorder loop's hand-crafted smoke
//	                     cassette (TestRecorder_Smoke). No outer host needed.
//
// See tests/recorder/SEED.md for the deterministic state contract the nested
// PVE produces, and the README of this package for env var configuration.
package record

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Pve9 records the PVE 9 cassettes against a freshly installed nested host.
// Requires the full PROXMOX_URL/_TOKENID/_SECRET/_NODE_NAME/_NODE_STORAGE
// envset; see config.go for the optional knobs.
func Pve9() error { return recordOne("pve9") }

// Pve8 records the PVE 8 cassettes. Same requirements as Pve9.
func Pve8() error { return recordOne("pve8") }

// All runs Pve9 then Pve8 in series. Halts on first failure.
func All() error {
	for _, m := range []string{"pve9", "pve8"} {
		if err := recordOne(m); err != nil {
			return err
		}
	}
	return nil
}

// Verify replays every checked-in cassette under tests/recorder/testdata
// without recording. Hits no network. CI default.
func Verify() error {
	cmd := exec.Command("go", "test", "./tests/recorder/...", "-count=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Explicitly drop PROXMOX_URL so the recorder helper picks ModeReplayOnly.
	env := []string{}
	for _, kv := range os.Environ() {
		if !startsWith(kv, "PROXMOX_URL=") {
			env = append(env, kv)
		}
	}
	cmd.Env = env
	return cmd.Run()
}

// Plan prints the sequence of operations record:all would perform without
// touching the outer host. Useful for first-time configuration: shows
// exactly which env vars are read and which steps would run.
func Plan() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	fmt.Printf(`mage record:all would:

  1. Connect to outer Proxmox API
       URL:      %s
       Node:     %s
       Storage:  %s
       Token:    %s
       Insecure: %v

  2. SSH into the outer host (for proxmox-auto-install-assistant)
       Host:     %s
       User:     %s
       Key:      %s

  3. For each release in (pve9, pve8):
       a. Skip download if the upstream ISO is already in storage
       b. Otherwise download from upstream
       c. Re-master the ISO with answer.toml + first-boot seed
       d. Create nested VM %d (%s)
            CPU=host  cores=%d  ram=%dMB  disk=%dGB
            bridge=%s  ip=%s  gw=%s
       e. Wait up to 25 minutes for install to complete
       f. SSH-poll /var/lib/go-proxmox-seeded for seed completion
       g. Run go test ./tests/recorder/... in record mode
       h. Destroy the nested VM

  4. Cassettes land in:
       tests/recorder/testdata/pve9/
       tests/recorder/testdata/pve8/

`,
		cfg.OuterURL,
		cfg.OuterNode,
		cfg.OuterStorage,
		cfg.OuterTokenID,
		cfg.OuterInsecure,
		cfg.SSHHost,
		cfg.SSHUser,
		cfg.SSHKey,
		cfg.NestedVMID,
		cfg.NestedName,
		cfg.NestedCPU,
		cfg.NestedRAM,
		cfg.NestedDiskGB,
		cfg.NestedBridge,
		cfg.NestedIP,
		cfg.NestedGateway,
	)
	return nil
}

// Smoke runs just the recorder's hand-crafted smoke cassette
// (TestRecorder_Smoke). No outer host or env vars needed; useful as a
// sanity check that the recorder package itself is wired correctly.
func Smoke() error {
	cmd := exec.Command("go", "test", "./tests/recorder/...", "-run", "TestRecorder_Smoke", "-v", "-count=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func recordOne(major string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	rel, ok := Releases[major]
	if !ok {
		return fmt.Errorf("unknown release %q (have: %v)", major, releaseKeys())
	}
	p, err := NewPipeline(context.Background(), cfg, rel)
	if err != nil {
		return err
	}
	return p.Run(context.Background())
}

func releaseKeys() []string {
	keys := make([]string, 0, len(Releases))
	for k := range Releases {
		keys = append(keys, k)
	}
	return keys
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
