# VM template builder

Builds a reusable qemu template from the Ubuntu Noble cloud image, end to end:

1. Download the cloud image onto an import-capable storage via `Storage.DownloadURL` (skipped if already present).
2. Get a free VMID from `Cluster.NextID` — no invented numbers.
3. Create the VM with `Node.NewVirtualMachine` and `VirtualMachineOption` key/value pairs — including `import-from=` on `scsi0` so the disk is cloned from the image, and an `ide2` cloud-init drive.
4. Recover the VMID from the returned task (`task.ID` is the UPID's id field, which for a `qmcreate` task is the VMID).
5. Set cloud-init defaults (`ciuser`, `ipconfig0`, optionally `sshkeys` via `proxmox.EncodeSSHKeys`) with `VirtualMachine.ConfigSync`.
6. Grow the base disk with `VirtualMachine.ResizeDisk`.
7. Convert to a template with `VirtualMachine.ConvertToTemplate`.

This is also the worked answer to a common point of confusion: **`VirtualMachineConfig` is the read side** — it's what `GET /config` unmarshals into, and the library never sends it to the API. Creating or editing a VM always goes through `VirtualMachineOption` pairs whose names are the raw parameter names from the [PVE API viewer](https://pve.proxmox.com/pve-docs/api-viewer/#/nodes/{node}/qemu).

## Run it

```shell
export PROXMOX_URL="https://lab.example.test:8006/api2/json"
export PROXMOX_TOKENID="root@pam!template-example"
export PROXMOX_SECRET="<token-secret>"

# optional
export PROXMOX_NODE_NAME="pve1"        # default: first node in the cluster
export PROXMOX_IMPORT_STORAGE="local"  # needs the "import" content type enabled
export PROXMOX_DISK_STORAGE="local-lvm"
export PROXMOX_CI_USER="ubuntu"
export PROXMOX_SSH_PUBKEY="$(cat ~/.ssh/id_ed25519.pub)"

cd examples/vm-template
go run .
```

The example talks to a real PVE node. It downloads ~600 MB to the import storage and creates a VM at the cluster's next free VMID (printed as it runs) — point it at a lab, not production. It does not clean up on the happy path (the template *is* the artifact); delete the created VM and the downloaded image to reset.

Requirements: PVE 8.2+ (for `import-from` with qcow2 on an import storage), and the import storage must have the **Import** content type enabled (Datacenter → Storage → Edit → Content).

## Clone the template

```go
template, err := node.VirtualMachine(ctx, vmid) // the VMID the run printed
if err != nil {
    panic(err)
}
newid, task, err := template.Clone(ctx, &proxmox.VirtualMachineCloneOptions{
    Name: "instance-1", // NewID 0 = next free VMID from the cluster
})
if err != nil {
    panic(err)
}
if err := task.Wait(ctx, time.Second, 5*time.Minute); err != nil {
    panic(err)
}
fmt.Println("cloned to", newid)
```

## Resolve dependencies

The example uses a `replace` directive to depend on the parent module from `../..`, so no separate `go get` is needed for the SDK itself. If `go run` complains about indirect deps, run `go mod tidy` once in this directory.
