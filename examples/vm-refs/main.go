// vm-refs shows the two ways to get a VM handle and when to use each.
//
// Fetch: client.Node and node.VirtualMachine GET the node status, the VM
// status and the VM config before returning. Use it when you want to read
// that data.
//
// Ref: client.NodeRef and node.VMRef build the handle locally with no API
// call. Use it when you already know the node and VMID and only want to act:
// vm.Stop on a ref is a single POST. Call vm.Ping to fill a ref with status
// and config when you do need them.
//
// Usage:
//
//	go run . <node> <vmid> [status|start|stop|shutdown|reboot]
//
// The default action is status, which only reads. The other actions change
// the VM's power state, so point them at a lab VM. Expects PROXMOX_URL,
// PROXMOX_TOKENID and PROXMOX_SECRET in the environment.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <node> <vmid> [status|start|stop|shutdown|reboot]", os.Args[0])
	}
	nodeName := os.Args[1]
	vmid, err := strconv.Atoi(os.Args[2])
	must(err, "vmid")
	action := "status"
	if len(os.Args) > 3 {
		action = os.Args[3]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := proxmox.NewClient(envOr("PROXMOX_URL", "https://localhost:8006/api2/json"),
		proxmox.WithAPIToken(mustEnv("PROXMOX_TOKENID"), mustEnv("PROXMOX_SECRET")),
		proxmox.WithHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}),
	)

	if action == "status" {
		showStatus(ctx, client, nodeName, vmid)
		return
	}

	// --- Ref pattern: act without fetching ---------------------------------
	//
	// No request happens until the action itself. A controller that already
	// knows where its VMs are (from its own state, or from one
	// cluster.Resources call) can drive hundreds of them this way without
	// three GETs per VM, and a config read that fails (for example while
	// another task holds the VM's lock) can't block a stop.
	vm := client.NodeRef(nodeName).VMRef(vmid)

	var task *proxmox.Task
	switch action {
	case "start":
		task, err = vm.Start(ctx)
	case "stop":
		task, err = vm.Stop(ctx)
	case "shutdown":
		task, err = vm.Shutdown(ctx)
	case "reboot":
		task, err = vm.Reboot(ctx)
	default:
		log.Fatalf("unknown action %q", action)
	}
	must(err, action)
	must(task.Wait(ctx, time.Second, 3*time.Minute), "wait for "+action)

	// The ref knows only node and VMID. IsRunning reads vm.Status, which is
	// empty until Ping fills it, so refresh before reading state.
	must(vm.Ping(ctx), "vm.Ping")
	fmt.Printf("%s %d after %s: status=%s running=%v\n", nodeName, vmid, action, vm.Status, vm.IsRunning())
}

// showStatus is the fetch pattern: the accessors GET the data up front, which
// is what you want when you are going to read it.
func showStatus(ctx context.Context, client *proxmox.Client, nodeName string, vmid int) {
	node, err := client.Node(ctx, nodeName)
	must(err, "client.Node")
	vm, err := node.VirtualMachine(ctx, vmid)
	must(err, "node.VirtualMachine")

	fmt.Printf("node %s: uptime %ds, %s\n", node.Name, node.Uptime, node.PVEVersion)
	fmt.Printf("vm %d (%s): status=%s running=%v cpus=%d tags=%q\n",
		vmid, vm.Name, vm.Status, vm.IsRunning(), vm.CPUs, vm.VirtualMachineConfig.Tags)
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
