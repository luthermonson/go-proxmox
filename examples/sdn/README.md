# SDN walkthrough

End-to-end example exercising the `/cluster/sdn` surface:

1. Capability probe via `SDNSubdirs` (ACL-filtered diridx).
2. Create a VXLAN zone using `proxmox.CSV` for the comma-joined `nodes`/`peers` fields.
3. Create a vnet inside the zone.
4. Create a subnet on the vnet via the `*VNet` instance handle.
5. Create an EVPN controller.
6. Preview the pending config via `SDNDryRun`.
7. Acquire a lock with `SDNLock`, run `SDNApply`, and roll back if anything fails.
8. Clean up (subnet → vnet → zone → controller) and re-apply.

## Run it

```shell
export PROXMOX_URL="https://lab.example.test:8006/api2/json"
export PROXMOX_TOKENID="root@pam!sdn-example"
export PROXMOX_SECRET="<token-secret>"

cd examples/sdn
go run .
```

The example talks to a real PVE cluster. It mutates SDN state — point it at a lab, not production. It rolls back on any apply failure and cleans up on the happy path, but a partial failure between steps may leave a `example-vxlan` zone, `example-vnet1` vnet, `example-evpn` controller, or `10.42.0.0/24` subnet behind. Delete those by name if so.

## Resolve dependencies

The example uses a `replace` directive to depend on the parent module from `../..`, so no separate `go get` is needed for the SDK itself. If `go run` complains about indirect deps, run `go mod tidy` once in this directory.
