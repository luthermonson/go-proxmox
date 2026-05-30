// sdn-walkthrough demonstrates the /cluster/sdn surface end-to-end:
// create a VXLAN zone, a vnet inside it, a subnet, and an EVPN controller;
// preview the pending config via dry-run; apply atomically under a lock with
// rollback safety; then clean up.
//
// Run it against a lab cluster only — it mutates SDN state and rolls
// changes back on the slightest hiccup. Expects PROXMOX_URL,
// PROXMOX_TOKENID, PROXMOX_SECRET in the environment.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/luthermonson/go-proxmox"
)

const (
	zoneName       = "example-vxlan"
	vnetName       = "example-vnet1"
	subnetCIDR     = "10.42.0.0/24"
	controllerName = "example-evpn"
	asn            = 65000
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := proxmox.NewClient(envOr("PROXMOX_URL", "https://localhost:8006/api2/json"),
		proxmox.WithAPIToken(mustEnv("PROXMOX_TOKENID"), mustEnv("PROXMOX_SECRET")),
		proxmox.WithHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}),
	)

	cluster, err := client.Cluster(ctx)
	must(err, "client.Cluster")

	// Capability probe: SDNSubdirs returns only the sub-areas this token can
	// read. Useful sanity check that the credential is actually scoped for SDN.
	subdirs, err := cluster.SDNSubdirs(ctx)
	must(err, "cluster.SDNSubdirs")
	fmt.Printf("SDN areas visible to this token: %v\n\n", subdirs)

	// --- Create resources ---------------------------------------------------

	fmt.Printf("Creating zone %q (vxlan)\n", zoneName)
	must(cluster.NewSDNZone(ctx, &proxmox.SDNZoneOptions{
		Name: zoneName,
		Type: "vxlan",
		// On the POST/PUT side PVE wants the comma-joined string form.
		// (On the read side the same fields surface as proxmox.CSV — see
		// the SDNZone read-back below.)
		Nodes: "pve1,pve2,pve3",
		Peers: "10.0.0.1,10.0.0.2,10.0.0.3",
	}), "NewSDNZone")

	// Read the zone back to demonstrate CSV unmarshaling: PVE returns
	// `"nodes": "pve1,pve2,pve3"` on the wire, and proxmox.CSV (a typed
	// []string) presents it as a slice without forcing the caller to split.
	zone, err := cluster.SDNZone(ctx, zoneName)
	must(err, "cluster.SDNZone")
	fmt.Printf("  zone nodes (CSV → []string): %v\n", []string(zone.Nodes))

	fmt.Printf("Creating vnet %q in zone %q\n", vnetName, zoneName)
	must(cluster.NewSDNVNet(ctx, &proxmox.VNetOptions{
		Name: vnetName,
		Zone: zoneName,
		Tag:  100,
		Type: "vnet",
	}), "NewSDNVNet")

	// The vnet instance handle carries `client` + parent identifying fields,
	// so subnet operations don't need to re-thread anything.
	vnet, err := cluster.SDNVNet(ctx, vnetName)
	must(err, "cluster.SDNVNet")

	fmt.Printf("Creating subnet %s in vnet %q\n", subnetCIDR, vnetName)
	must(vnet.NewSubnet(ctx, &proxmox.SDNSubnetOptions{
		Subnet:  subnetCIDR,
		Type:    "subnet",
		Gateway: "10.42.0.1",
	}), "vnet.NewSubnet")

	fmt.Printf("Creating EVPN controller %q (ASN %d)\n", controllerName, asn)
	must(cluster.NewSDNController(ctx, &proxmox.SDNControllerOptions{
		Controller: controllerName,
		Type:       "evpn",
		ASN:        asn,
		Peers:      "10.0.0.1,10.0.0.2",
	}), "NewSDNController")

	// --- Preview, then apply atomically under a lock ------------------------

	// DryRun returns the FRR + interfaces diff between the running config and
	// what Apply would push. Use it to gate the actual apply in a CI flow.
	dry, err := cluster.SDNDryRun(ctx, "")
	must(err, "SDNDryRun")
	fmt.Println("\nDry-run preview:")
	fmt.Println("  frr-diff:", firstLine(dry.FRRDiff))
	fmt.Println("  ifaces-diff:", firstLine(dry.InterfacesDiff))

	// SDNLock returns a token that scopes the subsequent apply. allowPending=false
	// fails if anyone else has unapplied edits — the safe default.
	token, err := cluster.SDNLock(ctx, false)
	must(err, "SDNLock")
	fmt.Printf("\nAcquired SDN lock: %s\n", token)

	applyTask, err := cluster.SDNApply(ctx)
	if err != nil {
		// Roll back pending state and surrender the lock so we don't leave
		// the cluster mid-edit. SDNRollback(_, _, releaseLock=true) does both.
		_ = cluster.SDNRollback(ctx, token, true)
		log.Fatalf("apply failed (rolled back): %v", err)
	}
	if err := applyTask.WaitFor(ctx, 60); err != nil {
		_ = cluster.SDNRollback(ctx, token, true)
		log.Fatalf("apply task timed out (rolled back): %v", err)
	}
	must(cluster.SDNReleaseLock(ctx, token, false), "SDNReleaseLock")
	fmt.Println("Apply succeeded.")

	// --- Cleanup (instance-pattern delete) ----------------------------------

	fmt.Println("\nCleaning up:")
	// Subnet → vnet → zone (must delete dependents first), then the controller.
	subnet := vnet.Subnet(subnetCIDR)
	must(subnet.Delete(ctx), "subnet.Delete")
	must(cluster.DeleteSDNVNet(ctx, vnetName), "DeleteSDNVNet")
	must(cluster.DeleteSDNZone(ctx, zoneName), "DeleteSDNZone")
	must(cluster.SDNController(controllerName).Delete(ctx), "SDNController.Delete")

	// Re-apply to commit the deletes. Same lock dance as above.
	token2, err := cluster.SDNLock(ctx, false)
	must(err, "SDNLock (cleanup)")
	cleanupTask, err := cluster.SDNApply(ctx)
	if err != nil {
		_ = cluster.SDNRollback(ctx, token2, true)
		log.Fatalf("cleanup apply failed (rolled back): %v", err)
	}
	must(cleanupTask.WaitFor(ctx, 60), "cleanup apply wait")
	must(cluster.SDNReleaseLock(ctx, token2, false), "SDNReleaseLock (cleanup)")
	fmt.Println("Cleanup applied.")
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

func must(err error, label string) {
	if err != nil {
		log.Fatalf("%s: %v", label, err)
	}
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	if s == "" {
		return "(empty)"
	}
	return s
}
