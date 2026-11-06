// failover shows one client spread across several nodes of a Proxmox
// cluster. It polls the cluster every few seconds; reboot or stop the primary
// node while it runs and the calls carry on through the fallbacks, then move
// back to the primary once it is up again. Failovers are logged at warn level
// and recoveries at info.
//
// Read-only: it lists nodes and nothing else. Expects PROXMOX_URL (the
// primary), PROXMOX_FALLBACK_URLS (comma-separated), PROXMOX_TOKENID and
// PROXMOX_SECRET in the environment; see the README.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	primary := mustEnv("PROXMOX_URL")
	fallbacks := strings.Split(mustEnv("PROXMOX_FALLBACK_URLS"), ",")

	opts := []proxmox.Option{
		proxmox.WithAPIToken(mustEnv("PROXMOX_TOKENID"), mustEnv("PROXMOX_SECRET")),
		proxmox.WithFallbackEndpoints(fallbacks...),
		proxmox.WithEndpointCooldown(10*time.Second, time.Minute),
		proxmox.WithLogger(&proxmox.LeveledLogger{Level: proxmox.LevelInfo}),
	}
	if ca := os.Getenv("PROXMOX_CA_FILE"); ca != "" {
		// The cluster CA (/etc/pve/pve-root-ca.pem) signs every node's
		// certificate, so one bundle verifies all the endpoints.
		opts = append(opts, proxmox.WithRootCAFile(ca))
	} else {
		opts = append(opts, proxmox.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // lab example; set PROXMOX_CA_FILE to verify
			},
		}))
	}
	if os.Getenv("PROXMOX_ROUND_ROBIN") != "" {
		opts = append(opts, proxmox.WithEndpointStrategy(proxmox.EndpointRoundRobin))
	}
	client := proxmox.NewClient(primary, opts...)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()
	for {
		poll(ctx, client)
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

// poll makes one bounded read. Each call has its own deadline, which bounds
// every endpoint tried for it.
func poll(ctx context.Context, client *proxmox.Client) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	nodes, err := client.Nodes(ctx)
	if err != nil {
		log.Printf("list nodes: %v", err)
		return
	}
	var states []string
	for _, n := range nodes {
		states = append(states, fmt.Sprintf("%s=%s", n.Node, n.Status))
	}
	log.Printf("nodes: %s", strings.Join(states, " "))
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("%s is not set", name)
	}
	return v
}
