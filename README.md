
# Proxmox API Client Go Package
[![Continuous Integration](https://github.com/luthermonson/go-proxmox/actions/workflows/ci.yaml/badge.svg)](https://github.com/luthermonson/go-proxmox/actions/workflows/ci.yaml) [![GitHub license](https://img.shields.io/github/license/luthermonson/go-proxmox)](https://github.com/luthermonson/go-proxmox/blob/main/LICENSE)
[![GitHub issues](https://img.shields.io/github/issues/luthermonson/go-proxmox)](https://github.com/luthermonson/go-proxmox/issues)
[![GitHub release](https://img.shields.io/github/release/luthermonson/go-proxmox.svg)](https://GitHub.com/luthermonson/go-proxmox/releases/) [![codecov](https://codecov.io/gh/luthermonson/go-proxmox/graph/badge.svg?token=GQSSZ0ZHZ4)](https://codecov.io/gh/luthermonson/go-proxmox) [![Go Report Card](https://goreportcard.com/badge/github.com/luthermonson/go-proxmox)](https://goreportcard.com/report/github.com/luthermonson/go-proxmox) [![Go Reference](https://pkg.go.dev/badge/github.com/luthermonson/go-proxmox.svg)](https://pkg.go.dev/github.com/luthermonson/go-proxmox)

Join the community to discuss ongoing client development usage, the proxmox API or tooling in the [#go-proxmox](https://gophers.slack.com/archives/C05920LDDD3) channel on the Gophers Slack and see the [self generated docs](https://pkg.go.dev/github.com/luthermonson/go-proxmox) for more usage details.

[![Slack](https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white)](https://gophers.slack.com/archives/C05920LDDD3)


A go client for [Proxmox VE](https://www.proxmox.com/). The client implements [/api2/json](https://pve.proxmox.com/pve-docs/api-viewer/index.html) and inspiration was drawn from the existing [Telmate](https://github.com/Telmate/proxmox-api-go/tree/master/proxmox) package but looking to improve in the following ways...
* Treated as a proper standalone go package
* Types and JSON marshal/unmarshalling for all end points
* Full Testing, unit testing with mocks and integration tests against an API endpoint
* Configuration options when creating a client for flexible usage
* Client logging for debugging within your code
* Context support
* Added functionality for better go tooling built on this library, some things we'd like
  * Boot VM from qcow URL, inspiration: [Proxmox Linux Templates](https://www.phillipsj.net/posts/proxmox-linux-templates/)
  * Dynamic host targeting for VM, Proxmox lacks a scheduler when given VM params it will try and locate a host with resources to put it
  * cloud-init support via no-cloud ISOs uploaded to node data stores and auto-mounted before boot, inspiration [quiso](https://github.com/luthermonson/quiso)
  * Unattended XML Support via ISOs similar to cloud-init ideas
  * node/vm/container shell command support via KVM proxy already built into proxmox

## API coverage

`go-proxmox` wraps 100% of the upstream PVE `/api2/json` surface for PVE 8.x and 9.x — every endpoint in [the API viewer](https://pve.proxmox.com/pve-docs/api-viewer/index.html) has a typed Go wrapper, with three intentional exceptions documented in `mage/endpoints/endpoints.go`: the two `mtunnelwebsocket` URL builders (the library returns the signed URL via `MigrationTunnelWebSocketPath`; the caller plumbs into their own websocket dialer) and the `file-restore/download` streaming binary endpoint.

Coverage is tracked in CI via `mage endpoints:coverage`, which diffs the live PVE schema against the package's call sites. Run it locally to confirm a fresh schema bump didn't add anything new:

```shell
mage endpoints:sync       # refresh .cache/pve-api/endpoints.json from upstream
mage endpoints:coverage   # print per-area coverage; lists any missing endpoints
```

Core developers are home lab enthusiasts working in the virtualization and kubernetes space. The common use case we have for
Proxmox is dev stress testing and validation of functionality in the products we work on, we plan to build the following tooling 
around this library to make that easier.
* [Docker Machine Driver](https://github.com/luthermonson/docker-machine-driver-proxmox) for consumption by [Rancher](https://rancher.com/docs/rancher/v1.5/en/configuration/machine-drivers/)
* [Terminal UI](https://github.com/luthermonson/p9s) inspired by [k9s](https://github.com/derailed/k9s) for quick management of PVE Clusters
* [Terraform Provider](https://github.com/luthermonson/terraform-provider-proxmox) with better local-exec and cloud-init/unattend xml support
* [Cluster API Provider Proxmox](https://github.com/luthermonson/cluster-api-provider-proxmox) to create kubernetes clusters

## Usage
Create a client and use the public methods to access Proxmox resources.

### Basic usage with login with a username and password credential
```go
package main

import (
	"context"
	"fmt"
	
	"github.com/luthermonson/go-proxmox"
)

func main() {
    credentials := proxmox.Credentials{
		Username: "root@pam", 
		Password: "12345",
    }
    client := proxmox.NewClient("https://localhost:8006/api2/json",
		proxmox.WithCredentials(&credentials),
    )
	
    version, err := client.Version(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println(version.Release) // 7.4
}
```

### Usage with Client Options

Lab setup (self-signed PVE, short timeout, API token):

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/luthermonson/go-proxmox"
)

func main() {
	client := proxmox.NewClient("https://localhost:8006/api2/json",
		proxmox.WithAPIToken("root@pam!mytoken", "somegeneratedapitokenguidefromtheproxmoxui"),
		proxmox.WithInsecureSkipVerify(),     // lab only
		proxmox.WithTimeout(30*time.Second),  // http.DefaultClient has no timeout
	)

	version, err := client.Version(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(version.Release) // 6.3
}
```

Production setup (pinned CA, mTLS optional):

```go
caBundle, _ := os.ReadFile("/etc/ssl/certs/pve-cluster.pem")
pool := x509.NewCertPool()
pool.AppendCertsFromPEM(caBundle)

client := proxmox.NewClient("https://pve.example.com:8006/api2/json",
	proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
	proxmox.WithRootCAs(pool),
	proxmox.WithTimeout(30*time.Second),
	// optional mTLS:
	// proxmox.WithClientCertificate(clientCert),
)
```

### Credential authentication and 2FA

API tokens are the right choice for daemons and CI. For interactive tools — anything with a human typing a password and a TOTP code — use credentials with `WithOTP` and `WithEagerAuth`:

```go
client := proxmox.NewClient(url,
	proxmox.WithCredentials(&proxmox.Credentials{
		Username: "admin",
		Password: password,
	}),
	proxmox.WithDefaultRealm("pam"),  // "admin" gets Realm "pam" merged in
	proxmox.WithOTP("123456"),        // one-shot; consumed on first /access/ticket
	proxmox.WithEagerAuth(),          // see note below
)
```

`WithEagerAuth` is worth calling out. PVE's pveproxy enforces a **hardcoded 3-second delay on every 401 response** as a brute-force mitigation (see `PVE::APIServer::AnyEvent`'s `# always delay unauthorized calls by 3 seconds` block). With credential auth the library's first request goes out unauthenticated by design — the ticket isn't issued until `/access/ticket` succeeds — so the first user-facing call eats the full 3 seconds. `WithEagerAuth` runs `CreateSession` inside `NewClient` so that cost is paid once at startup and every subsequent request is a normal ticket-authenticated call. Token auth doesn't trigger the 401 path at all and doesn't need this.

The OTP is consumed exactly once; subsequent `RefreshTicket` calls renew the session via the ticket itself and don't need a new code. If the session is fully lost later (PVE restart invalidates tickets), construct a fresh client with a fresh OTP — TOTP codes can't be cached.

### Resource traversal: instance handles

Most resources identified by an id/name follow a getter-returns-handle pattern. The handle carries the parent's client and identifying fields, so callers don't re-thread `(node, id)` on every call.

```go
cluster, _ := client.Cluster(ctx)

// SDN controllers — getter on the parent, operations on the instance.
ctrl := cluster.SDNController("evpn-1")
if err := ctrl.Update(ctx, &proxmox.SDNControllerOptions{ASN: proxmox.IntOrBool(65000)}); err != nil {
    panic(err)
}
_, _ = ctrl.Delete(ctx)

// Same pattern for VM/container snapshots, firewall rules, ceph OSDs/pools/mons,
// HA resources, ACME accounts, custom CPU models, fabrics, IPAMs, prefix-lists, etc.
```

See `AGENTS.md` (Required: pick the right shape for new endpoints) for the full inventory of instance types.

### Permission / capability discovery via `Subdirs`

PVE's directory-index GETs are ACL-filtered: the response only lists the sub-resources the calling token is permitted to read. Use them to probe what an API token can do without try-and-403 against every endpoint:

```go
node, _ := client.Node(ctx, "pve1")
subdirs, _ := node.Subdirs(ctx)         // ["qemu", "lxc", "storage", ...] filtered by ACL
fw, _ := node.FirewallSubdirs(ctx)      // ["rules", "options", "log"] if reachable

cluster, _ := client.Cluster(ctx)
sdnAreas, _ := cluster.SDNSubdirs(ctx)  // ["vnets", "zones", "controllers", ...]
```

### `CSV` type for comma-joined PVE fields

A few PVE fields serialize as comma-joined strings on the wire (most notably `SDNZone.Nodes` and `.Peers`, plus various group/realm fields). `proxmox.CSV` is a typed `[]string` whose `UnmarshalJSON` accepts both `"a,b,c"` and `["a","b","c"]`, and whose `MarshalJSON` always emits the comma-joined form PVE expects:

```go
// Read side: PVE returns `"nodes": "pve1,pve2,pve3"`. CSV presents it as a slice.
zone, _ := cluster.SDNZone(ctx, "vxlan-1")
for _, n := range zone.Nodes {                  // proxmox.CSV — iterable like []string
    fmt.Println(n)
}
allNodes := []string(zone.Nodes)                // explicit cast when you need []string

// Write side: the matching *Options types take a plain comma-joined string,
// because PVE only accepts that form on POST/PUT.
_ = cluster.NewSDNZone(ctx, &proxmox.SDNZoneOptions{
    Name:  "vxlan-1",
    Type:  "vxlan",
    Nodes: "pve1,pve2,pve3",
})
```

### Proxies

Route every request through a forward proxy. Works for `http://`, `https://`, and `socks5://` URLs:

```go
proxyURL, _ := url.Parse("http://proxy.corp.example.com:3128")
client := proxmox.NewClient("https://pve.example.com:8006/api2/json",
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithProxy(proxyURL),
)
```

Or use the standard `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` env-var convention:

```go
client := proxmox.NewClient("https://pve.example.com:8006/api2/json",
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithProxyFromEnvironment(),
)
```

`WithProxyFromEnvironment` reads env vars per-request (via Go's standard `http.ProxyFromEnvironment`), so changes after `NewClient` take effect on the next call. Composes with `WithHTTPClient`, the TLS options, retry, and request interceptors — option order doesn't matter.

### Retries on transient failures

PVE returns `502` / `503` during cluster transitions and `429` when rate-limited. `WithRetry` installs a `RoundTripper` wrapper that retries with full-jitter exponential backoff, honors `Retry-After` on `429` / `503`, and respects request-context cancellation:

```go
client := proxmox.NewClient(url,
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithRetry(),  // defaults: 3 attempts, 200ms–5s backoff
)
```

Tune the defaults for flakier upstreams:

```go
client := proxmox.NewClient(url,
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithRetry(
        proxmox.WithRetryMax(5),
        proxmox.WithRetryBackoff(500*time.Millisecond, 30*time.Second),
    ),
)
```

Or replace the predicate that decides what to retry — for example to also retry `423 Locked` while a cluster transition is in flight:

```go
retryOn423 := func(res *http.Response, err error) bool {
    if err != nil {
        return true
    }
    switch res.StatusCode {
    case http.StatusLocked, http.StatusBadGateway,
        http.StatusServiceUnavailable, http.StatusGatewayTimeout,
        http.StatusTooManyRequests:
        return true
    }
    return false
}
client := proxmox.NewClient(url,
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithRetry(proxmox.WithRetryCondition(retryOn423)),
)
```

The default predicate retries network errors plus `502` / `503` / `504` / `429`. Only idempotent verbs (`GET`, `PUT`, `DELETE`) and `POST` with a fully-buffered body are eligible — this client always buffers request bodies as `[]byte`, so `POST` is rewindable in practice.

### Request interceptors

Run a function on every outgoing request after the auth headers are populated and before the request is sent. Useful for tracing, correlation IDs, custom audit headers, request logging:

```go
addCorrelationID := func(req *http.Request) error {
    req.Header.Set("X-Correlation-Id", "build-1234")
    return nil
}
client := proxmox.NewClient(url,
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithRequestInterceptor(addCorrelationID),
)
```

Multiple interceptors compose — each `WithRequestInterceptor` call appends to the chain, and they run in registration order. The first non-nil error short-circuits the request (with a `request interceptor:` prefix so callers can `errors.Is` against their own sentinels):

```go
tracing := func(req *http.Request) error {
    // pull the active span from req.Context() and inject W3C traceparent.
    req.Header.Set("Traceparent", traceparent.From(req.Context()))
    return nil
}
audit := func(req *http.Request) error {
    log.Info().Str("method", req.Method).Str("path", req.URL.Path).Msg("pve")
    return nil
}
client := proxmox.NewClient(url,
    proxmox.WithAPIToken("automation@pve!ci", "<secret>"),
    proxmox.WithRequestInterceptor(tracing),
    proxmox.WithRequestInterceptor(audit),
)
```

The chain fires from `Req`, `Upload`, and `UploadReader`. Websocket upgrades (`TermWebSocket`, `VNCWebSocket`) are exempt — the dialer doesn't surface a `*http.Request` the chain could mutate.

### More examples

- [`examples/sdn`](./examples/sdn/) — full SDN walkthrough: create a zone, vnet, subnet, controller, dry-run / apply / rollback.
- [`examples/term-and-vnc`](./examples/term-and-vnc/) — websocket terminal and VNC proxy via a small Gin server.

### Upgrading between releases

Per-release migration guides live in [`migration/`](./migration/). Each file describes the source-level changes when upgrading FROM the named release.

- [`migration/v0.6.0.md`](./migration/v0.6.0.md) — upgrading from v0.5.x
- [`migration/v0.7.0.md`](./migration/v0.7.0.md) — upgrading from v0.6.0 (the major cleanup release)

# Developing
This project relies on [Mage](https://magefile.org/) for cross os/arch compatibility, please see their installation guide. 

## Unit Testing
Run `mage test` to run the unit tests in the root directory.

## Integration Testing
To run the integration testing suite against an existing Proxmox API set some env vars in your shell before running `mage testIntegration`. The integration tests will test logging in and using an API token credentials so make sure you set all five env vars before running tests for them to pass.

Please leave no trace when developing integration tests. All tests should create and remove all testing data they generate then they can be repeatably run against the same proxmox environment. Most people working on this package will likely use their personal Proxmox VE home lab and consuming extra resources via tests will lead to frustration.

### Bash
```shell
export PROXMOX_URL="https://192.168.1.6:8006/api2/json"
export PROXMOX_USERNAME="root@pam"
export PROXMOX_PASSWORD="password"
export PROXMOX_TOKENID="root@pam!mytoken"
export PROXMOX_SECRET="somegeneratedapitokenguidefromtheproxmoxui"

mage test:integration
```

### Powershell
```powershell
$Env:PROXMOX_URL = "https://192.168.1.6:8006/api2/json"
$Env:PROXMOX_USERNAME = "root@pam"
$Env:PROXMOX_PASSWORD = "password"
$Env:PROXMOX_TOKENID = "root@pam!mytoken"
$Env:PROXMOX_SECRET = "somegeneratedapitokenguidefromtheproxmoxui"

mage test:integration
```


