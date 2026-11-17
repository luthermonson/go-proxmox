# Changelog

Per-release upgrade instructions live in [`migration/`](./migration/). Releases before v0.9.0 are listed on the [GitHub releases page](https://github.com/luthermonson/go-proxmox/releases).

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
