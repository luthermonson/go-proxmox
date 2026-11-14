# Multi-endpoint failover

A client built from one URL is a single point of failure: when that node is down the client is dead, even though the rest of a quorate cluster is healthy. Every node serves the whole cluster API and tickets and tokens are cluster-wide, so the client can be given other nodes to fall back to.

```go
client := proxmox.NewClient("https://pve1:8006/api2/json",
    proxmox.WithAPIToken(tokenID, secret),
    proxmox.WithRootCAFile("/etc/pve/pve-root-ca.pem"),   // one CA verifies every node
    proxmox.WithAdditionalEndpoints("https://pve2:8006/api2/json", "https://pve3:8006/api2/json"),
    proxmox.WithEndpointFailureThreshold(2),              // bench after 2 consecutive failures (default 3)
    proxmox.WithEndpointCooldown(5*time.Second, time.Minute), // bench 5s, doubling, capped at 1m (defaults 10s / 180s)
)
```

- **Triggers:** transport failures (DNS, dial, refused, TLS handshake, timeout or reset) and the gateway statuses `502`, `503`, `504`, `595` and `596`. Any other HTTP response is a real answer and is returned as-is.
- **Idempotency:** a write is replayed on another node only when the request provably never left the client. A write that may have been delivered is not replayed, so a create or delete never runs twice. Reads (`GET`, `HEAD`, `OPTIONS`) fail over in every case.
- **Circuit breaker:** an endpoint that keeps failing is benched and skipped without being dialed while another is healthy. The bench is evaluated on the request path; there is no background goroutine.
- **Sticky:** after a node answers it becomes the current endpoint, and the client does not move back to the primary until the current one fails.

See [Multiple endpoints and failover](../../README.md#multiple-endpoints-and-failover) for the full rules.

## Run it

```shell
export PROXMOX_URL="https://pve1.example.test:8006/api2/json"
export PROXMOX_FALLBACK_URLS="https://pve2.example.test:8006/api2/json,https://pve3.example.test:8006/api2/json"
export PROXMOX_TOKENID="root@pam!failover-example"
export PROXMOX_SECRET="<token-secret>"
# export PROXMOX_CA_FILE=/path/to/cluster-ca.pem   # omit in a lab to skip TLS verification

cd examples/failover
go run .
```

The program prints the PVE version and node count every three seconds. Power off the primary node, or drop its port with a firewall rule, while it runs. The output keeps going, and the client logs a `failing over` warning (the example sets `WithLogger` to warn level; the default level is error).
