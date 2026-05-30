
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


