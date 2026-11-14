// failover shows a client that survives the loss of the node it was pointed at.
//
// Any node in a quorate Proxmox cluster serves the whole cluster API, and
// tickets and API tokens are cluster-wide, so a request that failed because one
// node is down can be answered by another. WithAdditionalEndpoints gives the
// client those other nodes; the NewClient URL stays the primary.
//
// The program polls the cluster every few seconds and prints which call
// succeeded. While it runs, power off or firewall the node you pointed
// PROXMOX_URL at: the polls keep succeeding, answered by a fallback node. The
// client logs each failover at warn level.
//
// Failover triggers on transport failures (DNS, dial, connection refused, TLS
// handshake, timeout or reset) and on the gateway statuses 502, 503, 504, 595
// and 596. It never triggers on any other HTTP response. Writes are replayed
// on another node only when the request provably never left the client; a
// write that may have been delivered is returned as-is, so nothing runs twice.
//
// Circuit breaker: an endpoint that fails at the transport layer
// WithEndpointFailureThreshold times in a row is benched for a cooldown and
// skipped without being dialed while another endpoint is healthy. The cooldown
// starts at the base given to WithEndpointCooldown and doubles per further
// failure up to the max. A success clears the endpoint's state.
//
// Usage:
//
//	go run .
//
// Expects PROXMOX_URL (primary), PROXMOX_TOKENID and PROXMOX_SECRET in the
// environment. PROXMOX_FALLBACK_URLS is a comma-separated list of the other
// nodes' API URLs. PROXMOX_CA_FILE is an optional PEM file with the cluster CA;
// without it the example skips TLS verification, which is for labs only.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	primary := envOr("PROXMOX_URL", "https://pve1:8006/api2/json")

	var fallbacks []string
	for _, u := range strings.Split(os.Getenv("PROXMOX_FALLBACK_URLS"), ",") {
		if u = strings.TrimSpace(u); u != "" {
			fallbacks = append(fallbacks, u)
		}
	}
	if len(fallbacks) == 0 {
		log.Fatal("set PROXMOX_FALLBACK_URLS to a comma-separated list of the other nodes' API URLs")
	}

	opts := []proxmox.Option{
		proxmox.WithAPIToken(mustEnv("PROXMOX_TOKENID"), mustEnv("PROXMOX_SECRET")),

		// The primary is the NewClient URL above. These are tried, in order,
		// when the current endpoint is unreachable.
		proxmox.WithAdditionalEndpoints(fallbacks...),

		// Bench an endpoint after 2 consecutive transport failures (default 3).
		proxmox.WithEndpointFailureThreshold(2),

		// First bench lasts 5s, doubling per further failure up to 1m
		// (defaults: 10s and 180s). The endpoint is re-probed once it expires.
		proxmox.WithEndpointCooldown(5*time.Second, time.Minute),

		// Failovers are logged at warn level; the default level is error.
		proxmox.WithLogger(&proxmox.LeveledLogger{Level: proxmox.LevelWarn}),

		// A short timeout keeps a black-holed node from stalling each attempt.
		proxmox.WithTimeout(5 * time.Second),
	}
	if ca := os.Getenv("PROXMOX_CA_FILE"); ca != "" {
		// One CA verifies every node; a single TLS config applies to all endpoints.
		opts = append(opts, proxmox.WithRootCAFile(ca))
	} else {
		opts = append(opts, proxmox.WithInsecureSkipVerify())
	}

	client := proxmox.NewClient(primary, opts...)

	for {
		poll(client)
		time.Sleep(3 * time.Second)
	}
}

// poll makes two cheap reads. Each one is a separate request, so a failover
// that happens on the first leaves the client on the new endpoint for the
// second (the last endpoint that answered becomes the current one).
func poll(client *proxmox.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	v, err := client.Version(ctx)
	if err != nil {
		// Every endpoint was tried and none answered.
		log.Printf("version: all endpoints failed: %v", err)
		return
	}

	nodes, err := client.Nodes(ctx)
	if err != nil {
		log.Printf("nodes: %v", err)
		return
	}

	online := 0
	for _, n := range nodes {
		if n.Status == "online" {
			online++
		}
	}
	fmt.Printf("%s  PVE %s, %d/%d nodes online\n", time.Now().Format(time.TimeOnly), v.Version, online, len(nodes))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
