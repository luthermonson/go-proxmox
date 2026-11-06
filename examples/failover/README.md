# Endpoint failover

One client spread across several nodes of a Proxmox cluster with `WithFallbackEndpoints`. It lists the cluster's nodes every 3 seconds. Stop or reboot the primary while it runs: the next call fails over to a fallback (logged at warn), calls keep succeeding, and once the primary is back and its cooldown has passed, traffic returns to it (logged at info).

The example is read-only. It only calls `GET /nodes`.

## Run it

```shell
export PROXMOX_URL="https://pve1.example.test:8006/api2/json"
export PROXMOX_FALLBACK_URLS="https://pve2.example.test:8006/api2/json,https://pve3.example.test:8006/api2/json"
export PROXMOX_TOKENID="root@pam!failover-example"
export PROXMOX_SECRET="<token-secret>"

# optional
export PROXMOX_CA_FILE="./pve-root-ca.pem"  # the cluster CA from /etc/pve; without it TLS isn't verified
export PROXMOX_ROUND_ROBIN=1                # spread calls across healthy nodes instead of preferring the primary

cd examples/failover
go run .
```

Sample output while `pve1` reboots:

```text
nodes: pve1=online pve2=online pve3=online
[WARN] proxmox endpoint https://pve1.example.test:8006/api2/json failed (dial tcp 10.0.0.11:8006: connect: connection refused); failing over to https://pve2.example.test:8006/api2/json
nodes: pve1=offline pve2=online pve3=online
nodes: pve1=offline pve2=online pve3=online
[INFO] proxmox endpoint https://pve1.example.test:8006/api2/json is answering again
nodes: pve1=online pve2=online pve3=online
```

See "Multiple endpoints and failover" in the top-level README for exactly what fails over, why writes are never replayed, and which calls stay pinned.
