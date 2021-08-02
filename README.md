# Proxmox API Client Go Package
A Go package to consume the Proxmox VE api2/json. Inspiration drawn from the existing
[Telmate](https://github.com/Telmate/proxmox-api-go/tree/master/proxmox) package but looking to improve
in the following ways.
* Treated as a package instead of a cli with an http client in a sub-directory
* Proper types and JSON marshal/unmarshalling for all end points
* Testing, unit testing and integration tests against an API endpoint
* Options to configure the client at creation

## Usage
Create a client and use the public methods at access Proxmox resources.

### Basic usage with login credentials
```go
package main

import (
	"fmt"
	"github.com/luthermonson/go-proxmox"
)
func main() {
    client := proxmox.NewClient("https://localhost:8006/api2/json")
    if _, err := client.Login(proxmox.Credentials{
    	Username: "root@pam",
    	Password: "password",
    }); err != nil {
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
	
    client := proxmox.NewClient("https://localhost:8006",
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
vars in your shell before running `make`. The integration tests will test both logging in and using an API token  
credentials so make sure you set all five env vars before running tests for them to pass.

//TODO make the 

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
will likely use some Proxmox homelab and consuming extra resources via tests will lead to frustration.
