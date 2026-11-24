# Changelog

Per-release upgrade instructions live in [`migration/`](./migration/). Releases before v0.9.0 are listed on the [GitHub releases page](https://github.com/luthermonson/go-proxmox/releases).

## v0.12.0

Upgrade guide: [`migration/v0.12.0.md`](./migration/v0.12.0.md). Decode fixes for fields PVE returns as `0`/`1` integers that were typed `bool` (so they failed to decode or dropped data), plus one more option field made a pointer so `0`/`false` can be sent. Seven source-level changes; if your code compiles, you've migrated.

### Fixes

- `Cluster.CephFlag` decodes PVE's `0`/`1` flag state correctly (it errored on the integer before); the `(bool, error)` signature is unchanged. `CephFlag.Value` is now `IntOrBool` (#383)
- `ContainerMigratePreconditions.NotAllowedNodes` is keyed by node name (`map[string]*ContainerMigrateNotAllowedNodes`) like the qemu endpoint; v0.11.0's flat struct silently dropped the per-node blocking-HA data (#383)
- `ContainerMigratePreconditions.Running`, `VirtualMachineMigratePreconditions.Running` and `.HasDBusVMState` are `IntOrBool` — PVE sends them as `0`/`1`, which a plain `bool` failed to decode (#383)
- `FirewallNodeOption.NDP` and `FirewallVirtualMachineOption.NDP` are `*IntOrBool`, so `Ptr(IntOrBool(false))` sends `ndp=0` instead of being dropped and PVE keeping its default of 1 (#380, rebased)
- `Cluster.CheckID` recognises PVE 9.2's reworded "guest ID N is already in use" error (older PVE: "VM N already exists"), so a taken VMID is reported as taken instead of surfacing the 400; matches on the typed `*APIError`. Behaviour-only, no API change (#378)

### Breaking changes

Seven incompatible field changes (`apidiff v0.11.0..v0.12.0` reports exactly these); see the [migration guide](./migration/v0.12.0.md). All are "the field couldn't decode what PVE sends" fixes: `CephFlag.Value`, the three migrate-precondition `Running`/`HasDBusVMState` fields, `ContainerMigratePreconditions.NotAllowedNodes`, and the two firewall `NDP` fields. The `CephFlag(...)` method signature is unchanged. (#383, #380)

Thanks to @lbarrosby (#380) and @themoriarti (#378).

## v0.11.0

Upgrade guide: [`migration/v0.11.0.md`](./migration/v0.11.0.md). Mostly decode fixes — fields and methods whose Go type or JSON tag did not match what PVE returns, so they never decoded. Every change is source-incompatible, but in each case the old shape could not carry the real value, so fixing your code gets you data you weren't getting before.

### Features

- Schema return-type drift tooling: `cmd/apidrift` diffs the library's mapped return types against the upstream PVE API schema and reports drift; a weekly GitHub Actions job runs it and opens a tracking issue when the schema moves. Maintainer tooling, not part of the client API (#375)

### Fixes

- `Cluster.CephFlag` returns `bool`, matching the API (was `string`) (#375)
- `Container.Firewall` / `VirtualMachine.Firewall` replaced by `FirewallSubdirs() ([]string, error)`; the endpoint is a directory index, not a firewall object. Added on `Cluster` and `Node` too (#375)
- `User.Tokens` is a `UserTokens` with a custom unmarshaller accepting both the array (`GET /access/users`) and object (`GET /access/users/{userid}`) forms (#375)
- `NodeCertificate.NotAfter` / `.NotBefore` are `int64` Unix timestamps, with the correct tags (were `string`) (#375)
- `SDNZone.Pending` is `map[string]any` (PVE returns a pending-changes object, not a bool); `SDNZoneOptions.ExitNodesLocalRouting` is `IntOrBool` (#375)
- `Disk.MountPoint string` replaced by `Disk.Mounted IntOrBool`; added `Disk.OSDIDList`; `Storage.Status` targets `/storage/{storage}/status` so its fields decode (#375)
- `ContainerMigratePreconditions` reshaped to the LXC schema (`*ContainerMigrateNotAllowedNodes`, `DependentHAResources`); `StorageContent.ApproximateSize` added (#375)
- `ClusterBackupOptions.IoNice`, `.LockWait`, `.StopWait` and `.Zstd` are `*uint`, so `0` (a valid non-default value) reaches PVE instead of being dropped — matching `VirtualMachineBackupOptions` (#379)

### Breaking changes

- 17 source-level signature changes, all of the "field/method couldn't decode what PVE sends" kind. See the [migration guide](./migration/v0.11.0.md) for the full list and before/after. The notable ones: `CephFlag` return type, `Firewall` methods removed, `User.Tokens` type, `NodeCertificate` time fields, `SDNZone.Pending`, `Disk.MountPoint`→`Mounted` (and `Disk` is no longer comparable), `ContainerMigratePreconditions` shape, and the four `ClusterBackupOptions` `*uint` fields (#375, #379)

## v0.10.0

Upgrade guide: [`migration/v0.10.0.md`](./migration/v0.10.0.md). Two source-level breaks, both the same shape: an option field that defaults to true on the PVE side became a pointer so `false` can actually be sent. Wrap the value you were passing in `Ptr(...)`; unset behaves as before. If your code compiles, you have migrated.

### Fixes

- `NodeStopAllOptions.ForceStop` is `*IntOrBool` (was `IntOrBool`): `force-stop=0` now reaches PVE, so `StopAll` can let a graceful shutdown time out instead of hard-stopping guests. Unset still omits the field (PVE default 1) (#374)
- `ClusterBackupOptions.Enabled`, `.Remove` and `.StdExcludes` are `*bool` (was `bool`): a backup job can now be created or updated disabled, and `remove=0`/`stdexcludes=0` reach PVE instead of being dropped (#373)

### Breaking changes

- `NodeStopAllOptions.ForceStop` is now `*IntOrBool`; set it with `Ptr(IntOrBool(false))` (#374)
- `ClusterBackupOptions.Enabled`, `.Remove` and `.StdExcludes` are now `*bool`; set them with `Ptr(false)` / `Ptr(true)` (#373)

## v0.9.0

Upgrade guide: [`migration/v0.9.0.md`](./migration/v0.9.0.md). One source-level break (`AgentExecStatus` field types); the rest changes behaviour without changing a signature, so read the guide even if your code builds.

### Features

- Multi-endpoint failover: `WithAdditionalEndpoints`, `WithEndpointFailureThreshold`, `WithEndpointCooldown`. Fails over on transport errors and on `502`/`503`/`504`/`595`/`596`, replays writes only when the request never left the client, benches repeatedly failing endpoints with a circuit breaker. See `examples/failover` (#369)
- Zero-request handles: `Client.NodeRef`, `Node.VirtualMachineRef`, `Node.ContainerRef`. Act on a guest without three GETs first (#348)
- `Node.StorageRef` and `Storage.Ping`, the storage counterpart of the refs (#370)
- `Client.FindVirtualMachine` and `Client.FindContainer` resolve a VMID to a node-bound handle in one `GET /cluster/resources` (#370)
- Typed `*APIError` for every non-2xx response, with the sentinels `ErrForbidden`, `ErrBadRequest`, `ErrRateLimited`, `ErrNodeUnreachable`, the matching `Is*` helpers and `AsAPIError`. `IsNotFound` also matches PVE's `500`/`400` "does not exist" messages (#350)
- `*TaskFailedError`, `IsTaskFailed` and `Task.Warnings`; `WARNINGS: n` is now a successful exit status (#347)
- `Container.Ping` and `VirtualMachine.Ping` refresh a handle in place from status and config (#348)
- Websocket terminal and VNC connections dial the current failover endpoint (#369)

### Fixes

- `Task.Wait`, `WaitFor` and `WaitForCompleteStatus` honour context cancellation, no longer double-ping, and recover a final status from the node's task history when PVE answers `no such task` (#347)
- `Task.Watch` streams log lines in order and runs to the end of the task, including the final `TASK OK` line (#364)
- `VirtualMachine.AgentExec` returns the API error instead of a misleading "no pid returned" error (#363)
- `AgentExecStatus` decodes `err-truncated` and `signal` the way PVE sends them (#355)
- `VirtualMachine.WaitForAgent` matches the typed error, so it works behind an HTTP/2 proxy where the status line has no reason phrase (#350)
- `TermWebSocket` no longer panics on close or dereferences a nil error on a bad handshake reply, and waits for its workers before closing channels (#371)
- `Node.VzTmpls` results carry the client, node and storage, so `Delete` works on them (#368)
- `Storage.Backup` fills `VolID` from the requested name (#366)
- Integer request parameters are no longer sent in exponent notation (`1.7e+09`), and null fields are no longer sent as `<nil>` (#365)
- `VirtualMachineSnapshotUpdateOptions` and `ContainerSnapshotUpdateOptions` always send `description`, so an empty value clears it (#367)
- Uploads send `content`, `checksum-algorithm` and `checksum` ahead of the file part, the order pveproxy requires (#352)
- Guest firewall IPSet delete sends `force` and `digest` on the query string, where PVE reads them (#356)
- `VirtualMachine.CloudInit` waits for the cloud-init tag task before attaching the ISO (PVE bugzilla 8124)

### Deprecations

- `(*Node).New`: use `Client.NodeRef` (#348)
- `(*VirtualMachine).New`: use `Node.VirtualMachineRef` (#348)
- `Node.Storage`: use `Node.StorageRef`, and `Storage.Ping` for the status fields (#370)

Still working, still compiling.

### Breaking changes / behavior changes

- `AgentExecStatus.ErrTruncated` is now `IntOrBool` (was `bool`) and `AgentExecStatus.Signal` is now `int` (was `bool`). Source-incompatible if you assign or compare them as `bool` (#355)
- `Task.Wait` and `WaitFor` return `*TaskFailedError` when the task fails; they used to return `nil` (#347)
- `WaitForCompleteStatus` returns `*TaskFailedError` as `err` for a failed task (#347)
- `Task.Ping` returns the error and leaves the task untouched when the status request fails (#347)
- Non-2xx statuses other than 400/401/403/500/501 (`404`, `502`, `595`, ...) now return an error; they used to fall through to JSON decoding and usually returned `nil` (#350)
- `401`/`403` return an `*APIError` instead of the bare `ErrNotAuthorized` value, so `err == ErrNotAuthorized` no longer matches; use `errors.Is` (#350)
- `VirtualMachine.Delete` reads the config first when the handle has none, one extra GET (#348)
- `Container.Ping` makes two requests (status and config) and replaces the handle's fields (#348)
- `VirtualMachine.Ping` replaces the handle's status fields, clearing ones PVE no longer sends (#348)
- Snapshot description updates with an empty `Description` now clear it rather than being a no-op (#367)
