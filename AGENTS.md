# AGENTS.md

Guidance for AI coding agents (Claude Code, Codex, etc.) working in this
repository. `CLAUDE.md` defers to this file.

## Build, test, lint

This project uses [Mage](https://magefile.org/) as the task runner. Targets are
defined in `magefile.go` and `mage/`.

- `mage test` — unit tests (alias for `go test` in the root package).
- `mage test:coverage` — unit tests with `-race -coverprofile=coverage.txt -covermode=atomic`. CI runs this.
- `mage test:build` — compile check using `go build -tags test` (the `test`
  tag activates `build.go`'s no-op `main`).
- `mage lint` — runs `golangci-lint` (version pinned in
  `mage/install/install.go`; currently `v2.8.0`). Installs it on demand.
- `mage ci` — install deps, lint, coverage, build (matches the GitHub Actions job).
- `mage test:integration` — runs `go test ./tests/integration -tags "nodes containers vms"` against a real PVE cluster.
- `mage env` — print the env vars the integration suite reads, masking secrets.
- `mage endpoints:sync` / `mage endpoints:coverage` — snapshot the upstream PVE
  API surface (`mage/endpoints/`) and diff the package's wrapper coverage
  against it. `coverage` writes `.cache/pve-api/coverage.txt` (uncovered
  endpoints) and `coverage_by_area.txt` (per-area breakdown). The repo is at
  100% coverage of considered endpoints — see "Endpoint coverage tooling"
  below for how the by-design exclusion list works.

Run a single unit test directly with `go test -run TestName` from the repo root
(unit tests live in the root package). To run only one integration suite, use the
matching build tag, e.g. `go test ./tests/integration -tags nodes -run TestNode`.

### Integration test env vars

Required before `mage test:integration`:

```
PROXMOX_URL, PROXMOX_USERNAME, PROXMOX_PASSWORD,
PROXMOX_TOKENID, PROXMOX_SECRET
```

Optional: `PROXMOX_OTP`, `PROXMOX_NODE_NAME`, `PROXMOX_NODE_STORAGE`,
`PROXMOX_APPLIANCE_PREFIX`, `PROXMOX_ISO_URL`. Integration tests are expected to
clean up after themselves — don't leave artifacts on the target cluster.

## Architecture

The repo is a single Go package, `github.com/luthermonson/go-proxmox`, that
wraps the Proxmox VE `/api2/json` REST API. Files at the root are organized by
Proxmox resource (`nodes.go`, `virtual_machine.go`, `containers.go`,
`cluster.go`, `storage.go`, `access.go`, `tasks.go`, …) and one large
`types.go` (~75 KB) holds every JSON-bound struct.

### Client and request layer (`proxmox.go`)

`Client` is the entry point. It carries the HTTP client, base URL, auth
material (token *or* credentials/session), logger, and a session mutex. All
resource methods route through `Client.Req`, which:

1. Prefixes relative paths with `baseURL`.
2. Adds auth headers via `authHeaders` — `PVEAPIToken=` when a token is set,
   else `Cookie: PVEAuthCookie=` + `CSRFPreventionToken` from the session.
3. On 401/403 with credentials configured, calls `CreateSession` once and
   retries; otherwise returns `ErrNotAuthorized`.
4. Hands the response to `handleResponse`, which unwraps the standard Proxmox
   `{"data": ...}` envelope before unmarshalling into the caller's `v`.

Public helpers `Get`, `GetWithParams`, `Post`, `Put`, `Delete`, `Upload`, and
`UploadReader` all sit on top of `Req`. Sentinel errors (`ErrNotFound`,
`ErrNotAuthorized`, `ErrTimeout`, `ErrNoop`, `ErrSessionExists`) have matching
`Is*` predicates — use those rather than string matching.

Two websocket helpers (`TermWebSocket`, `VNCWebSocket`) upgrade `https://` →
`wss://` and return `(send, recv, errs, closer, err)` channels for the
xterm.js / VNC protocols Proxmox exposes.

### Configuration via functional options (`options.go`)

`NewClient(baseURL, opts...)` accepts `Option` funcs. Prefer the non-deprecated
forms: `WithHTTPClient`, `WithCredentials`, `WithAPIToken`, `WithSession`,
`WithUserAgent`, `WithLogger`, plus the auth/transport options below.
`WithClient` and `WithLogins` are kept for backwards compatibility — don't
introduce new uses.

**TLS and transport options** (`WithInsecureSkipVerify`, `WithRootCAs`,
`WithRootCAFile`, `WithClientCertificate`, `WithTimeout`, plus the Tier 3
proxy/retry/interceptor options when they land) defer their side effects
until `finalizeOptions` runs after all option funcs. This makes option order
irrelevant — `WithHTTPClient(custom) + WithInsecureSkipVerify()` and the
reverse both end up with the option applied to `custom.Transport`. The
shared private helpers `ensureOwnHTTPClient`, `ensureTransport`, and
`ensureTLSConfig` handle the lazy promotion of `http.DefaultClient` /
`http.DefaultTransport` to writable clones so transport mutations never
bleed into the package-global singletons. When you add a new
transport-touching option, route through those helpers rather than mutating
`c.httpClient` directly.

**Auth ergonomics.** `WithOTP` stashes a single-use TOTP/2FA code on the
client and threads it into `Credentials.Otp` inside `sessionCredentials`
(called by `CreateSession`). The OTP is zeroed after first use so a
re-authentication after session loss doesn't resend a stale code.
`WithDefaultRealm` is applied at the same merge point — it sets
`Credentials.Realm` only when both the field is empty and the username has
no `@realm` suffix.

**Eager auth.** `WithEagerAuth` exists specifically because pveproxy
enforces a hardcoded 3-second delay on every 401 from `auth_handler` (PVE
upstream: `# always delay unauthorized calls by 3 seconds` in
`PVE::APIServer::AnyEvent`). The library's credential-auth path is
intentionally reactive — the first request goes out unauthenticated and
triggers the retry-after-/access/ticket loop — which means the first
user-facing call eats that 3 seconds. `WithEagerAuth` calls `CreateSession`
synchronously inside `NewClient` to pay that cost at construction time.
Errors are swallowed (option funcs can't return errors) and logged at debug
level; the next user request retries via the existing lazy path. Callers
who need auth errors surfaced at startup should call `CreateSession`
directly.

### Resource model

Resources are plain structs with an unexported `client *Client` field
populated by their parent's accessor. The traversal pattern is:

```
client.Cluster(ctx) → *Cluster
client.Node(ctx, name) → *Node
node.VirtualMachine(ctx, vmid) → *VirtualMachine
vm.Config(ctx, opts...) → *Task
```

Mutating endpoints typically return a `*Task` (built from a Proxmox UPID
string). `Task.Wait`, `Task.WaitFor`, `Task.WaitForCompleteStatus`, and
`Task.Watch` poll `/nodes/<node>/tasks/<upid>/status` and `…/log`. UPID
parsing happens in `NewTask` (`tasks.go`).

`VirtualMachineOption` (`virtual_machine_config.go`) and similar option types
are name/value pairs flattened into the request body — this is how the
package handles Proxmox's free-form, version-dependent config keys.

### Required: pick the right shape for new endpoints (singleton vs. instance type)

Every new endpoint wrapper must decide where to hang. The rule is mechanical:
**look at the upstream PVE path** and apply one of two shapes.

**Singleton sub-resource (no identifier in the path).** Path looks like
`/parent/<sub>` or `/parent/<sub>/<fixed-action>` with no `/{id}/` segment.
Wrap as a **method on the parent** type. Examples: `/nodes/{node}/apt/update`,
`/nodes/{node}/dns`, `/nodes/{node}/ceph/status`. There's exactly one of these
per parent — no identifier to hold.

```go
func (n *Node) APTUpdate(ctx context.Context, notify, quiet bool) (*Task, error) { … }
func (n *Node) DNS(ctx context.Context) (*NodeDNS, error) { … }
```

**Multi-instance sub-resource (identifier in the path).** Path has
`/{name}/` or `/{id}/` selecting one of many. Wrap as a **getter on the parent
that returns an instance handle**, with operations as methods on the handle.
Examples: `/nodes/{node}/storage/{storage}/...`, `/nodes/{node}/ceph/osd/{id}/...`,
`/nodes/{node}/services/{service}/...`. The handle carries identity in its
receiver so callers don't re-thread it on every call.

```go
// Getter: no API call, just constructs the handle.
func (n *Node) CephOSD(id int) *CephOSD {
    return &CephOSD{client: n.client, Node: n.Name, ID: id}
}

// Instance type: unexported client + exported identifying fields.
type CephOSD struct {
    client *Client
    Node   string `json:"-"`
    ID     int    `json:"-"`
    // ...plus any data fields the list endpoint populates...
}

// Operations live on the instance, not the parent.
func (o *CephOSD) Scrub(ctx context.Context, deep bool) error { … }
func (o *CephOSD) Delete(ctx context.Context, cleanup bool) (*Task, error) { … }
```

**Anti-pattern to avoid:** `node.CephOSDScrub(ctx, id, ...)` or
`node.DeleteCephOSD(ctx, id, ...)`. Threading the id through every call when
the schema clearly identifies the resource in the path duplicates the
identity argument across N methods and makes the godoc surface of the parent
explode. `*Storage`, `*VirtualMachine`, `*Container`, `*NodeNetwork`,
`*NodeService`, `*NodeReplicationJob`, `*CephOSD`, `*CephPool`,
`*CephMon`/`Mgr`/`MDS`, `*CephFS`, `*FirewallRule`, `*VirtualMachineSnapshot`,
and `*ContainerSnapshot` are all instance types — mirror their shape.

**Construction:** the getter never makes an API call; it just builds the
handle. The corresponding `List`/`Plural` accessor on the parent (e.g.
`node.CephOSDs(ctx)`) populates `client` + identifying fields on each
returned `*CephOSD` so list results are immediately chainable.

**Creation stays on the parent.** `node.CreateCephOSD(ctx, opts)` — there's
no instance yet, so the parent owns the constructor call. Same for `New*`
patterns.

**Singleton subsystems with many endpoints (e.g. ceph itself, disks,
firewall-options) stay on the parent.** Don't introduce a top-level handle
just because the area is big. Threshold: does the schema path namespace by
identifier? If yes → handle; if no → parent method.

**Inventory of existing instance types** (mirror their shape for new ones):
`*Storage`, `*VirtualMachine`, `*Container`, `*VirtualMachineSnapshot`,
`*ContainerSnapshot`, `*NodeNetwork`, `*NodeService`, `*NodeReplicationJob`,
`*Task`, `*CephOSD`, `*CephPool`, `*CephMon`, `*CephMgr`, `*CephMDS`,
`*CephFS`, `*FirewallRule`, `*FirewallSecurityGroup`, `*PCIDevice`,
`*SDNController`, `*SDNDNS`, `*SDNIPAM`, `*SDNFabric`, `*SDNFabricNode`,
`*SDNPrefixList`, `*SDNPrefixListEntry`, `*SDNRouteMapEntry`, `*VNet`,
`*VNetSubnet`, `*ACMEAccount`, `*ACMEPlugin`, `*CustomCPUModel`,
`*HAResource`, `*HAGroup`, `*HARule`, `*ReplicationJob`, `*RealmSyncJob`,
`*MetricServer`, `*NotificationGotifyEndpoint`,
`*NotificationSMTPEndpoint`, `*NotificationSendmailEndpoint`,
`*NotificationWebhookEndpoint`, `*NotificationMatcher`, `*PCIMapping`,
`*USBMapping`, `*DirMapping`, `*ClusterBackup`.

### Required: directory-index ("diridx") endpoints use the `Subdirs` naming

PVE exposes many GET endpoints whose body is a static `[{"subdir":"x"}, ...]`
catalogue of the available sub-resources, ACL-filtered server-side. They're a
supported permission/capability probe (see how the web UI itself uses them).

**Naming convention:**

- Public method: `Subdirs()` (when there's no scoping prefix) or
  `<Area>Subdirs()` when several diridx GETs live on the same receiver and
  the bare name would clash with sibling action methods. Examples:
  `(*Cluster).Subdirs`, `(*Cluster).FirewallSubdirs`, `(*Node).Subdirs`,
  `(*Node).DisksSubdirs`, `(*NodeReplicationJob).Subdirs`.
- Unexported per-area helper: `<area>Diridx(ctx, path) ([]string, error)`,
  e.g. `sdnDiridx`, `capabilitiesDiridx`, `clusterDiridx`. The endpoint
  scanner (`mage/endpoints`) keys off the literal `Diridx` suffix to attribute
  the call site as a GET against the path argument — keep the suffix.
- Shared decode: every per-area helper one-lines into `decodeSubdirList(ctx,
  c, path)` (in `nodes_diridx.go`). Don't duplicate the
  `[{"subdir":...}]` unmarshalling.

The handful of paths that *look* like diridx but actually return a typed
payload (e.g. `GET /nodes/{node}/storage/{storage}` returns the storage
status object, not subdirs) are named after the payload they produce:
`(*Storage).Status() *StorageStatus`. Don't shoehorn them through the
`Subdirs` shape.

### Required: comma-joined PVE responses use the `CSV` type

Several PVE fields serialize as a comma-joined string on the wire — most
notoriously `SDNZone.Nodes` and `.Peers`, but also user `Groups`, the various
firewall ipset references, and a few realm config fields. Declaring these as
`[]string` is broken: PVE rejects the JSON-array form on PUT and refuses to
unmarshal the comma-string form on GET, so the round-trip fails.

`CSV` (`types.go`, defined as `type CSV []string`) is the canonical wrapper.
Its `UnmarshalJSON` accepts both `"a,b,c"` and `["a","b","c"]`; `MarshalJSON`
always emits `"a,b,c"`. Use `CSV` for any field where the upstream
description mentions "comma-separated list" or where the schema type is
`string` but the data is plural-by-convention. When in doubt, check whether
the live PVE API returns a string for that field — if yes and the value
contains commas, it's `CSV`.

### Required: comment every pointer field tied to a PVE default

When you add a `*T` field per the rule above, ship a brief comment on the
same line block explaining the specific PVE default and the failure mode the
pointer prevents. Example shape:

```go
// CPUUnits — PVE default 1024 (cgroup v1) / 100 (cgroup v2). Plain int
// would default to 0 and override the server's CPU weight on edit. See #199.
CPUUnits *int `json:"cpuunits,omitempty"`
```

Future readers shouldn't need `git blame` to understand why a field is a
pointer. Don't ship `FIXME(issue-199)` comments — the FIXME-era pointerize
work is done; new fields are either pointers-with-rationale or plain values.

### Required: don't clobber PVE-side defaults on config structs

Background: see issue #199. Several config structs (e.g. `ContainerConfig`,
`VirtualMachineConfig`) historically declared fields as plain values with
`,omitempty`. That works only when Go's zero value matches Proxmox's
documented default. When it doesn't — for example `Console IntOrBool` with
a `json:"console,omitempty"` tag, where PVE defaults to `1` (true) but Go's
zero is `0`/false — `omitempty` *cannot* tell "user left it unset" apart
from "user wanted false", so the marshaller drops the field on the unset
case and the API call silently flips a server-side default.

**Rule for every new or modified config struct field:**

1. **Look up the upstream default.** Cross-reference the field against the
   [PVE API viewer](https://pve.proxmox.com/pve-docs/api-viewer/index.html)
   for that endpoint. The default is in the parameter description, e.g.
   `console: <boolean> (default = 1)`.
2. **If the upstream default differs from Go's zero value, use a pointer.**
   That means:
   - `*bool` when the PVE default is `true` (Go zero: `false`).
   - `*int` / `*StringOrInt` / `*IntOrBool` when the PVE default is any
     non-zero number, or when `0` is itself a meaningful value the user
     might want to send explicitly.
   - `*string` when the PVE default is a non-empty string, or when the empty
     string is a meaningful value.
   - Slices, maps, and pointer types are already nil-able and don't need
     wrapping.
3. **If Go's zero value matches the upstream default, leave it unboxed** with
   `,omitempty`. Don't pointer-ify defensively — it just adds caller
   ergonomics noise for no correctness gain.
4. **Mirror the change on any sibling option/builder type.** If you add
   `*bool` to `ContainerConfig.Console`, the matching
   `ContainerOption`/`VirtualMachineOption` helpers and any test fixtures
   that round-trip through the struct must agree.
5. **Add a regression test.** Marshal the struct with the field unset and
   assert the JSON does *not* contain the key — that's the failure mode this
   rule exists to prevent. Cover the explicit-false / explicit-zero case
   too: set the pointer and assert the key *is* present with the right
   value.

When in doubt: the question to ask is "if the user never touches this field,
will an unintended value reach Proxmox?" If yes, the field needs to be a
pointer.

The `audit/` directory contains a tool that diffs the package's config
structs against the live PVE API schema and writes `audit/report.md`. Run
it (see the header comment in `audit/main.go` for the regen command) when
adding fields or after a PVE release to catch new mismatches; existing
known mismatches are tagged with `FIXME` comments in `types.go`.

### Mock-based unit tests (`tests/mocks/`)

Unit tests use [`h2non/gock`](https://github.com/h2non/gock) to intercept HTTP
calls. The `mocks` package exposes `On(config)` (default = PVE 9.x) plus
`ProxmoxVE6x`, `ProxmoxVE7x`, `ProxmoxVE8x`, `ProxmoxVE9x` to load
version-specific fixtures from `tests/mocks/pve{6,7,8,9}x/`. The standard test
shape:

```go
mocks.On(mockConfig)        // or ProxmoxVE7x(mockConfig) for version-specific
defer mocks.Off()
client := mockClient()
// ...exercise client...
```

`tests/mocks/capture/` records calls so you can assert on them.

### Required: add gock mocks for every new/changed endpoint in a PR

Any PR that adds or changes a method on `*Client` (or any resource type) which
calls `c.Get`/`Post`/`Put`/`Delete`/`Upload` **must** ship matching unit tests
backed by gock fixtures. Do not rely on integration tests alone — CI only runs
unit tests, and integration coverage is opt-in.

Workflow when you touch an endpoint:

1. **Find the registration file.** Mock fixtures live in
   `tests/mocks/pve9x/<resource>.go` (e.g., `nodes.go`, `virtual_machines.go`,
   `storage.go`, `tasks.go`). The file is selected by the resource group, not
   the HTTP path. Each file exports a single lowercase function (`nodes()`,
   `virtualMachines()`, …) wired into `pve9x/proxmox.go`'s `Load()`. If you are
   adding a brand-new resource group, add a new file *and* register its
   loader in `proxmox.go`.
2. **Register the route.** Use `gock.New(config.C.URI)` with an anchored regex
   path (`^/nodes/node1/qemu/101/config$`) and a `Reply(200).JSON(...)` body
   that matches the real Proxmox response — including the outer `"data": { … }`
   envelope. Use `.Persist()` only for routes called more than once per test
   run (e.g., list endpoints reused across tests); one-shot routes should not
   persist.
3. **Match the test data conventions.** Existing fixtures use node `node1`,
   VMID `101`, container `100`, etc. Reuse those identifiers so new tests
   compose with existing ones. If you need a new identifier, add it
   consistently across every fixture file that references it.
4. **Backport to older versions only when behavior differs.** Default to
   `pve9x/`. Only add `pve6x/`, `pve7x/`, or `pve8x/` fixtures when the test
   explicitly calls `mocks.ProxmoxVE{6,7,8}x(mockConfig)` to exercise
   version-specific behavior (see `TestClient_Version6/7/9` in
   `proxmox_test.go`).
5. **Write the unit test.** Add it to the matching `<resource>_test.go` in the
   repo root using the standard shape above. Cover the happy path and at
   least one error path (e.g., `gock.New(...).Reply(404)` or
   `Reply(401)` to exercise `ErrNotFound` / `ErrNotAuthorized`).
6. **Verify.** Run `mage test` (or `go test -run NewTestName`) and `mage lint`
   before opening the PR. An unmocked endpoint will hit `gock`'s
   "no match" failure, surfaced as a real HTTP error in the test output.

Example: a new `Client.Foo` calling `GET /foo/bar` needs a new entry in (or new
section added to) the appropriate `pve9x/*.go` file:

```go
gock.New(config.C.URI).
    Get("^/foo/bar$").
    Reply(200).
    JSON(`{"data": {"field": "value"}}`)
```

…plus a `TestClient_Foo` in the matching `*_test.go` that calls `mocks.On` /
`mocks.Off` and asserts on the parsed response.

### Endpoint coverage tooling (`mage/endpoints`)

`mage endpoints:sync` fetches the upstream `apidoc.js` schema and flattens it
to `(method, path)` tuples in `.cache/pve-api/`. `mage endpoints:coverage`
scans the package source for call sites, normalizes their path templates, and
diffs against the schema.

The scanner is deliberately conservative — it only counts call sites it can
statically prove correspond to a schema endpoint. To keep it accurate it
recognises a few non-obvious shapes:

- `<x>.Get|Post|Put|Delete(WithParams)?(ctx, "/path"|fmt.Sprintf("/path", …), …)`
  — the canonical form.
- `<x>.VNCWebSocket|TermWebSocket(...)` — counts as GET against the path arg.
- `<x>.<Name>[Dd]iridx(ctx, path, …)` — counts as GET. **This is why diridx
  helpers must keep the `Diridx`/`diridx` suffix.**
- `<x>.client.Upload[\w]*(path, …)` — counts as POST. Catches both
  `Upload` and `UploadReader` (and multi-line forms where `(` is at EOL).
- `<var> := url.URL{Path: …}` followed by `<var>.String()` and
  `<var> := fmt.Sprintf("/…")` followed by a bare-var call argument — both
  resolve via a per-file tracked path map.

**By-design exclusions** live in `intentionallyExcluded` (top of
`endpoints.go`). The map is reconciled against the live schema on every run,
so a stale entry (PVE renamed/removed the endpoint) fails loudly rather than
silently masking a regression. Adding a new entry requires an explanatory
rationale (URL-builder helpers, streaming binary downloads, etc.). Diridx
GETs do **not** belong here — we wrap them.

**Coverage as a quality gate.** The repo currently sits at 100% of considered
endpoints. New PRs that add a method should bump the coverage in the right
direction. If a PR introduces a wrapper the scanner can't detect, prefer
teaching the scanner over adding to the exclusion list.

### Integration tests (`tests/integration/`)

Gated by build tags so they don't run by default: `nodes`, `containers`,
`vms`. Tests share a `TestingData` struct populated in `init()` from the
`PROXMOX_*` env vars and verify behavior against a live cluster. Treat them as
destructive — they create and delete VMs/containers — and always pair a
"create" step with a deferred cleanup.

## Release migration guides

Every release that ships at least one apidiff `incompatible` change must
include a `migration/v<VERSION>.md` file describing the upgrade path FROM
the previous release. The file is created as part of the release-cut PR,
not retrofitted later, and is linked from the GitHub release body.

Confirm what's incompatible with:

```shell
go install golang.org/x/exp/cmd/apidiff@latest
git worktree add /tmp/old-version v<previous-version>
cd /tmp/old-version && apidiff -w /tmp/old.api .
cd <repo>          && apidiff -w /tmp/new.api .
apidiff -incompatible /tmp/old.api /tmp/new.api
```

The migration doc should group breaks into themed sections (e.g. "snapshots
refactored to instance pattern", "indexed device fields dropped",
"bool → IntOrBool") rather than listing them as a flat dump. Each section
shows a before/after code example. `v0.7.0.md` is the template; copy its
structure.

Releases with zero incompatible changes don't need a migration file — the
GitHub release notes are sufficient. The directory is a record of the
breaks we've shipped, not a per-release ritual.

## Repo-specific gotchas

- **Local replace in `go.mod`.** `go.mod` contains
  `replace github.com/diskfs/go-diskfs => C:/Users/luthe/go-diskfs`. This is a
  developer-local override on the maintainer's machine; do not commit changes
  that depend on it, and do not propagate that path elsewhere. If you touch
  `go.mod`/`go.sum`, leave the replace alone unless explicitly asked.
- **Upload size limit.** `Client.Upload` is bound by Proxmox's ~16 KB POST
  cap (see the comment in `proxmox.go`). For large ISOs use `DownloadURL` on
  the node or store the file out-of-band.
- **Session retry.** A 401/403 triggers exactly one `CreateSession` retry;
  don't add a second retry layer in callers.
- **Go version.** `go.mod` is pinned to `go 1.25`; CI uses Go 1.25.
- **`types.go` is the dumping ground.** New JSON shapes go there alongside
  the existing types rather than in per-resource files.
