# Proxmox API Client Go Package
A Go package to consume the Proxmox VE api2/json. Inspiration drawn from the existing
[Telmate](https://github.com/Telmate/proxmox-api-go/tree/master/proxmox) package but looking to improve
in the following ways.
* Treated as a proper standalone go package
* Types and JSON marshal/unmarshalling for all end points
* Full Testing, unit testing and integration tests against an API endpoint
* Configuration options when creating a client for flexible usage
* Client logging for debugging within your code
* Added functionality for better go tooling built on this library, some things we'd like
  * Boot VM from qcow URL, inspiration: [Proxmox Linux Templates](https://www.phillipsj.net/posts/proxmox-linux-templates/)
  * Dynamic host targeting for VM, Proxmox lacks a scheduler when given VM params it will try and locate a host with resources to put it
  * cloud-init support via no-cloud ISOs uploaded to node data stores and auto-mounted before boot, inspiration [quiso](https://github.com/luthermonson/quiso)
  * Unattend XML Support via ISOs similar to cloud-init ideas
  * node/vm/container shell command support via KVM proxy already built into proxmox

Core developers are home lab enthusiasts working in the virtualization and kubernetes space. The common use case we have for
Proxmox is dev stress testing and validation of functionality in the products we work on and we plan to build the following tooling 
around this library to make that easier.
* [Docker Machine Driver](https://github.com/luthermonson/docker-machine-driver-proxmox) for consumption by (Rancher)[https://rancher.com/docs/rancher/v1.5/en/configuration/machine-drivers/]
* [Terminal UI](https://github.com/luthermonson/p9s) inspired by [k9s](https://github.com/derailed/k9s) for quick management of PVE Clusters
* [Terraform Provider](https://github.com/luthermonson/terraform-provider-proxmox) with better local-exec and cloud-init/unattend xml support

## Usage
Create a client and use the public methods to access Proxmox resources.

### Basic usage with login credentials
```go
package main

import (
	"fmt"
	"github.com/luthermonson/go-proxmox"
)

func main() {
    client := proxmox.NewClient("https://localhost:8006/api2/json")
    if err := client.Login("root@pam", "password"); err != nil {
        panic(err)
    }
    version, err := client.Version()
    if err != nil {
        panic(err)
    }
    fmt.Println(version.Release) // 6.3
}
```

### Usage with Client Options
```go
package main

import (
	"fmt"
	"github.com/luthermonson/go-proxmox"
)

func main() {
    insecureHTTPClient := http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
    }
    tokenID := "root@pam!mytoken"
    secret := "somegeneratedapitokenguidefromtheproxmoxui"
    
    client := proxmox.NewClient("https://localhost:8006/api2/json",
        proxmox.WithClient(&insecureHTTPClient),
        proxmox.WithAPIToken(tokenID, secret),
    )
    
    version, err := client.Version()
    if err != nil {
        panic(err)
    }
    fmt.Println(version.Release) // 6.3
}
```
## Testing
When developing this package you can run the testing suite against an existing Proxmox API. To do this set some env
vars in your shell before running `mage ci`. The integration tests will test both logging in and using an API token  
credentials so make sure you set all five env vars before running tests for them to pass.

### Bash
```shell
export PROXMOX_URL="https://192.168.1.6:8006/api2/json"
export PROXMOX_USERNAME="root@pam"
export PROXMOX_PASSWORD="password"
export PROXMOX_TOKENID="root@pam!mytoken"
export PROXMOX_SECRET="somegeneratedapitokenguidefromtheproxmoxui"

make
```

### Powershell
```powershell
$Env:PROXMOX_URL = "https://192.168.1.6:8006/api2/json"
$Env:PROXMOX_USERNAME = "root@pam"
$Env:PROXMOX_PASSWORD = "password"
$Env:PROXMOX_TOKENID = "root@pam!mytoken"
$Env:PROXMOX_SECRET = "somegeneratedapitokenguidefromtheproxmoxui"

./make
```

Please leave no trace when developing integration tests. All tests should create and remove all testing data they 
are generating so they can be repeatably run against the same proxmox environment. Most people working on this package
will likely use their personal Proxmox homelab and consuming extra resources via tests will lead to frustration.
