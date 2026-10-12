# VM refs vs. fetched VMs

There are two ways to get a handle on a VM, and they cost different amounts:

| | Fetch | Ref |
|---|---|---|
| Build the handle | `client.Node(ctx, name)` then `node.VirtualMachine(ctx, vmid)` | `client.NodeRef(name).VMRef(vmid)` |
| Requests to build it | 3 GETs (node status, VM status, VM config) | none |
| `vm.Stop(ctx)` afterwards | 1 POST | 1 POST |
| `vm.Status`, `vm.Name`, `vm.VirtualMachineConfig` | filled | empty until `vm.Ping(ctx)` |

Use **fetch** when you are going to read the VM: show its status, compare its config, branch on `IsRunning`.

Use a **ref** when you already know the node and VMID and only want to act: start, stop, migrate, change the config, delete. This is the controller case. If you reconcile hundreds of VMs whose placement you already know, a ref skips three GETs per VM. A config read that fails (for example while another task holds the VM's lock) also can't block a stop.

Containers work the same way: `client.NodeRef(name).ContainerRef(vmid)` and `ct.Ping(ctx)`.

What works on a ref without fetching first:

- **Actions** (`Start`, `Stop`, `Shutdown`, `Reboot`, `Migrate`, `Config`, `Clone`, snapshots, firewall, agent, ...) need only the node and VMID. They send exactly the one request they always did.
- **Config read-modify-write** (`AddTag`, `RemoveTag`, `CloudInit`, `UnmountCloudInitISO`, and the cloud-init ISO cleanup inside `Delete`) fetch the config themselves when the handle doesn't have it. That is one extra GET, and it keeps the existing tags instead of overwriting them.
- **Accessors** (`IsRunning`, `IsStopped`, `IsPaused`, `IsHibernated`, `HasTag`) take no context, so they can't fetch. On a ref they report `false` until you call `Ping`.

## Run it

```shell
export PROXMOX_URL="https://lab.example.test:8006/api2/json"
export PROXMOX_TOKENID="root@pam!refs-example"
export PROXMOX_SECRET="<token-secret>"

cd examples/vm-refs
go run . pve1 100            # fetch pattern: read-only status
go run . pve1 100 reboot     # ref pattern: one POST, then Ping to read the result
```

`status` only reads. `start`, `stop`, `shutdown` and `reboot` change the VM's power state, so point them at a lab VM.
