// vm-template builds a reusable qemu template from an Ubuntu cloud image:
// download the image into an import-capable storage, create a VM whose disk
// is imported from it, set cloud-init defaults, and convert the VM to a
// template ready for cloning.
//
// This is also the canonical demonstration of how VM creation works in this
// package: options, not structs. Node.NewVirtualMachine and
// VirtualMachine.Config take VirtualMachineOption key/value pairs whose names
// are the raw qemu parameter names from the PVE API viewer
// (https://pve.proxmox.com/pve-docs/api-viewer/#/nodes/{node}/qemu).
// VirtualMachineConfig is the *read* side — GET /config unmarshals into it —
// and is never sent to the API.
//
// Run it against a lab cluster only — it downloads ~600 MB onto the target
// storage and creates a VM at the cluster's next free VMID. Expects
// PROXMOX_URL, PROXMOX_TOKENID, PROXMOX_SECRET in the environment; see the
// README for the optional knobs.
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

const (
	templateName = "ubuntu-noble-template"

	imageURL  = "https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img"
	imageName = "noble-server-cloudimg-amd64.qcow2" // name on the PVE storage; must end .qcow2 for import
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	client := proxmox.NewClient(envOr("PROXMOX_URL", "https://localhost:8006/api2/json"),
		proxmox.WithAPIToken(mustEnv("PROXMOX_TOKENID"), mustEnv("PROXMOX_SECRET")),
		proxmox.WithHTTPClient(&http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}),
	)

	// --- Resolve the node ---------------------------------------------------

	nodeName := os.Getenv("PROXMOX_NODE_NAME")
	if nodeName == "" {
		nodes, err := client.Nodes(ctx)
		must(err, "client.Nodes")
		nodeName = nodes[0].Node
	}
	node, err := client.Node(ctx, nodeName)
	must(err, "client.Node")
	fmt.Printf("Using node %q\n", nodeName)

	// --- Fetch the cloud image onto an import-capable storage ---------------

	// The upload endpoint is capped at ~16 KB POSTs, so large images go
	// through the download-url endpoint instead: PVE pulls the file onto the
	// storage itself. Content type "import" requires a storage with the
	// import content type enabled (Datacenter → Storage → Content).
	importStorageName := envOr("PROXMOX_IMPORT_STORAGE", "local")
	importStorage, err := node.Storage(ctx, importStorageName)
	must(err, "node.Storage (import)")

	imageVolID := fmt.Sprintf("%s:import/%s", importStorageName, imageName)
	if existing, err := importStorage.GetContent(ctx); err == nil && hasVolume(existing, imageVolID) {
		fmt.Printf("Image %s already present, skipping download\n", imageVolID)
	} else {
		fmt.Printf("Downloading %s → %s\n", imageURL, imageVolID)
		task, err := importStorage.DownloadURL(ctx, "import", imageName, imageURL)
		must(err, "storage.DownloadURL")
		must(task.Wait(ctx, 5*time.Second, 20*time.Minute), "download task")
	}

	// --- Create the VM ------------------------------------------------------

	// Ask the cluster for the next free VMID rather than inventing one. If
	// you prefer templates at a fixed, well-known ID (9000 is a common
	// convention), pass that instead — creation fails if the ID is taken.
	cluster, err := client.Cluster(ctx)
	must(err, "client.Cluster")
	vmid, err := cluster.NextID(ctx)
	must(err, "cluster.NextID")

	// Creation is key/value options matching the PVE API parameters — the
	// same names you would pass to `qm create`. This is the write-side
	// counterpart to the VirtualMachineConfig struct you get back on reads.
	diskStorageName := envOr("PROXMOX_DISK_STORAGE", "local-lvm")
	fmt.Printf("Creating VM %d (%s)\n", vmid, templateName)
	task, err := node.NewVirtualMachine(ctx, vmid,
		proxmox.VirtualMachineOption{Name: "name", Value: templateName},
		proxmox.VirtualMachineOption{Name: "ostype", Value: "l26"},
		proxmox.VirtualMachineOption{Name: "memory", Value: 2048},
		proxmox.VirtualMachineOption{Name: "cores", Value: 2},
		proxmox.VirtualMachineOption{Name: "cpu", Value: "x86-64-v2-AES"},
		proxmox.VirtualMachineOption{Name: "scsihw", Value: "virtio-scsi-single"},
		// "storage:0" means "allocate on this storage, size taken from the
		// import source"; import-from clones the cloud image into the new disk.
		proxmox.VirtualMachineOption{Name: "scsi0", Value: fmt.Sprintf("%s:0,import-from=%s,discard=on", diskStorageName, imageVolID)},
		// The cloud-init state disk. PVE generates its content at boot from
		// the ciuser/sshkeys/ipconfig0 values set below.
		proxmox.VirtualMachineOption{Name: "ide2", Value: fmt.Sprintf("%s:cloudinit", diskStorageName)},
		proxmox.VirtualMachineOption{Name: "net0", Value: "virtio,bridge=vmbr0"},
		// Cloud images ship a serial console; point the display at it.
		proxmox.VirtualMachineOption{Name: "serial0", Value: "socket"},
		proxmox.VirtualMachineOption{Name: "vga", Value: "serial0"},
		proxmox.VirtualMachineOption{Name: "agent", Value: "enabled=1"},
		proxmox.VirtualMachineOption{Name: "boot", Value: "order=scsi0"},
	)
	must(err, "node.NewVirtualMachine")
	must(task.Wait(ctx, time.Second, 5*time.Minute), "create task")

	// The create task itself carries the new VMID: task.ID is the UPID's id
	// field, which for a qmcreate task is the VMID. We already know it here,
	// but a caller holding only the *Task can recover it like this.
	createdID, err := strconv.Atoi(task.ID)
	must(err, "parse task.ID")
	if createdID != vmid {
		log.Fatalf("task.ID %d does not match requested vmid %d", createdID, vmid)
	}

	vm, err := node.VirtualMachine(ctx, createdID)
	must(err, "node.VirtualMachine")

	// --- Cloud-init defaults ------------------------------------------------

	// Post-create edits use the same option pattern: Config (async POST,
	// returns a task) or ConfigSync (synchronous PUT).
	ciOptions := []proxmox.VirtualMachineOption{
		{Name: "ciuser", Value: envOr("PROXMOX_CI_USER", "ubuntu")},
		{Name: "ipconfig0", Value: "ip=dhcp"},
	}
	if pubkey := os.Getenv("PROXMOX_SSH_PUBKEY"); pubkey != "" {
		// PVE rejects Go's default query escaping for this field —
		// EncodeSSHKeys produces the exact urlencoded form it validates.
		ciOptions = append(ciOptions, proxmox.VirtualMachineOption{
			Name: "sshkeys", Value: proxmox.EncodeSSHKeys(pubkey),
		})
	}
	must(vm.ConfigSync(ctx, ciOptions...), "vm.ConfigSync (cloud-init)")

	// Cloud images ship a ~3.5 GB filesystem; grow the base disk so clones
	// start from something usable. cloud-init expands the fs on first boot.
	resizeTask, err := vm.ResizeDisk(ctx, "scsi0", "+8G")
	must(err, "vm.ResizeDisk")
	must(resizeTask.Wait(ctx, time.Second, 2*time.Minute), "resize task")

	// --- Convert to template ------------------------------------------------

	tmplTask, err := vm.ConvertToTemplate(ctx)
	must(err, "vm.ConvertToTemplate")
	must(tmplTask.Wait(ctx, time.Second, 2*time.Minute), "template task")

	fmt.Printf("\nTemplate %d (%s) ready. Clone it with:\n\n", createdID, templateName)
	fmt.Printf("    template, err := node.VirtualMachine(ctx, %d)\n", createdID)
	fmt.Printf("    if err != nil { ... }\n")
	fmt.Printf("    newid, task, err := template.Clone(ctx, &proxmox.VirtualMachineCloneOptions{\n")
	fmt.Printf("        Name: \"instance-1\", // NewID 0 = next free VMID from the cluster\n")
	fmt.Printf("    })\n")
	fmt.Printf("    if err != nil { ... }\n")
	fmt.Printf("    if err := task.Wait(ctx, time.Second, 5*time.Minute); err != nil { ... }\n")
}

func hasVolume(content []*proxmox.StorageContent, volid string) bool {
	for _, c := range content {
		if c.Volid == volid {
			return true
		}
	}
	return false
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
