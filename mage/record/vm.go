package record

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	proxmox "github.com/luthermonson/go-proxmox"
)

// installTimeout caps how long we wait for a fresh PVE install to complete
// and the API to come up. The PVE installer typically reboots once and is
// reachable within ~5 minutes; 25 minutes leaves headroom for slow disks
// and post-install package updates triggered by the answer file.
const installTimeout = 25 * time.Minute

// VMLifecycle drives the nested PVE VM through its full life: create,
// attach the prepared ISO, start, wait for the API to come up, then on
// teardown stop and delete unconditionally.
type VMLifecycle struct {
	cfg    *Config
	client *proxmox.Client
	node   *proxmox.Node
}

// NewVMLifecycle wires up the outer-host node handle.
func NewVMLifecycle(ctx context.Context, cfg *Config, client *proxmox.Client) (*VMLifecycle, error) {
	node, err := client.Node(ctx, cfg.OuterNode)
	if err != nil {
		return nil, fmt.Errorf("look up outer node %q: %w", cfg.OuterNode, err)
	}
	return &VMLifecycle{cfg: cfg, client: client, node: node}, nil
}

// Provision creates the nested VM with the prepared installer ISO attached
// as the boot disk, starts it, and waits for the PVE API to come up on the
// configured static IP. Returns when the API responds to /version, or
// errors out after installTimeout.
func (l *VMLifecycle) Provision(ctx context.Context, preparedISO string) error {
	if err := l.Destroy(ctx); err != nil {
		return fmt.Errorf("preflight destroy: %w", err)
	}

	create, err := l.node.NewVirtualMachine(ctx, l.cfg.NestedVMID,
		proxmox.VirtualMachineOption{Name: "name", Value: l.cfg.NestedName},
		proxmox.VirtualMachineOption{Name: "ostype", Value: "l26"},
		proxmox.VirtualMachineOption{Name: "cpu", Value: "host"},
		proxmox.VirtualMachineOption{Name: "cores", Value: l.cfg.NestedCPU},
		proxmox.VirtualMachineOption{Name: "memory", Value: l.cfg.NestedRAM},
		proxmox.VirtualMachineOption{Name: "scsihw", Value: "virtio-scsi-pci"},
		proxmox.VirtualMachineOption{Name: "scsi0", Value: fmt.Sprintf("%s:%d,format=qcow2",
			l.cfg.OuterStorage, l.cfg.NestedDiskGB)},
		proxmox.VirtualMachineOption{Name: "ide2", Value: fmt.Sprintf("%s,media=cdrom", preparedISO)},
		proxmox.VirtualMachineOption{Name: "boot", Value: "order=ide2;scsi0"},
		proxmox.VirtualMachineOption{Name: "net0", Value: fmt.Sprintf("virtio,bridge=%s",
			l.cfg.NestedBridge)},
		proxmox.VirtualMachineOption{Name: "agent", Value: "1"},
	)
	if err != nil {
		return fmt.Errorf("create nested VM %d: %w", l.cfg.NestedVMID, err)
	}
	if err := create.Wait(ctx, 2*time.Second, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for nested VM create: %w", err)
	}

	vm, err := l.node.VirtualMachine(ctx, l.cfg.NestedVMID)
	if err != nil {
		return fmt.Errorf("look up nested VM %d: %w", l.cfg.NestedVMID, err)
	}
	start, err := vm.Start(ctx)
	if err != nil {
		return fmt.Errorf("start nested VM: %w", err)
	}
	if err := start.Wait(ctx, 2*time.Second, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for nested VM start: %w", err)
	}

	return l.waitForAPI(ctx, installTimeout)
}

// Destroy stops and removes the nested VM. Idempotent: missing VM is not
// an error, since the caller is invoking it both as a preflight cleanup
// and as a post-record teardown.
func (l *VMLifecycle) Destroy(ctx context.Context) error {
	vm, err := l.node.VirtualMachine(ctx, l.cfg.NestedVMID)
	if err != nil {
		// Treat any not-found as already-gone. go-proxmox surfaces these
		// as a status 500 with a "does not exist" message; rather than
		// brittle string-matching, we list and check.
		exists, listErr := l.exists(ctx)
		if listErr != nil {
			return fmt.Errorf("look up nested VM %d: %w", l.cfg.NestedVMID, err)
		}
		if !exists {
			return nil
		}
		return fmt.Errorf("look up nested VM %d: %w", l.cfg.NestedVMID, err)
	}

	if vm.IsRunning() {
		stop, sErr := vm.Stop(ctx)
		if sErr != nil {
			return fmt.Errorf("stop nested VM: %w", sErr)
		}
		if wErr := stop.Wait(ctx, 2*time.Second, 2*time.Minute); wErr != nil {
			return fmt.Errorf("wait for nested VM stop: %w", wErr)
		}
	}

	del, err := vm.Delete(ctx)
	if err != nil {
		return fmt.Errorf("delete nested VM: %w", err)
	}
	if err := del.Wait(ctx, 2*time.Second, 2*time.Minute); err != nil {
		return fmt.Errorf("wait for nested VM delete: %w", err)
	}
	return nil
}

// exists checks whether the nested VMID currently has a VM record on the
// outer node. Used to make Destroy idempotent without string-matching API
// errors.
func (l *VMLifecycle) exists(ctx context.Context) (bool, error) {
	vms, err := l.node.VirtualMachines(ctx)
	if err != nil {
		return false, err
	}
	for _, v := range vms {
		if int(v.VMID) == l.cfg.NestedVMID {
			return true, nil
		}
	}
	return false, nil
}

// waitForAPI polls the nested PVE's /version endpoint until it responds OK
// or the timeout elapses. Uses an InsecureSkipVerify HTTP client because
// the nested PVE serves a self-signed cert we don't bother trusting.
func (l *VMLifecycle) waitForAPI(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("%s/version", l.NestedURL())
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				// 401 is fine — it means the API is up but our (intentionally
				// missing here) auth didn't pass. Either way the VM is reachable.
				return nil
			}
			lastErr = fmt.Errorf("nested PVE returned %s", resp.Status)
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	if lastErr == nil {
		lastErr = errors.New("nested PVE never came up")
	}
	return fmt.Errorf("waitForAPI: %w", lastErr)
}

// NestedURL returns the API base URL for the nested PVE. Strips any /N
// prefix from the static-IP CIDR.
func (l *VMLifecycle) NestedURL() string {
	ip := l.cfg.NestedIP
	if i := strings.Index(ip, "/"); i > 0 {
		ip = ip[:i]
	}
	return fmt.Sprintf("https://%s:8006/api2/json", ip)
}
