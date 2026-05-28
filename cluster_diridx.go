package proxmox

import (
	"context"
)

// Directory-index ("diridx") wrappers for /cluster/* endpoints. See the
// header comment in nodes_diridx.go for the rationale — same shape, same
// shared subdir-list decoder.

// ClusterIndex enumerates the children of /cluster (typically "replication",
// "metrics", "config", "firewall", "backup", "backupinfo", "ha",
// "acme", "ceph", "jobs", "sdn", "log", "resources", "tasks", "options",
// "status", "nextid", "qemu"). ACL-filtered.
func (cl *Cluster) ClusterIndex(ctx context.Context) ([]string, error) {
	return cl.clusterDiridx(ctx, "/cluster")
}

// ACMEIndex enumerates the children of /cluster/acme ("plugins",
// "account", "tos", "meta", "directories", "challenge-schema").
func (cl *Cluster) ACMEIndex(ctx context.Context) ([]string, error) {
	return cl.acmeDiridx(ctx, "/cluster/acme")
}

// FirewallIndex enumerates the children of /cluster/firewall ("groups",
// "rules", "ipset", "aliases", "options", "macros", "refs").
func (cl *Cluster) FirewallIndex(ctx context.Context) ([]string, error) {
	return cl.firewallDiridx(ctx, "/cluster/firewall")
}

// SDNIndex enumerates the children of /cluster/sdn ("vnets", "zones",
// "controllers", "ipams", "dns", "fabrics", "subnets").
func (cl *Cluster) SDNIndex(ctx context.Context) ([]string, error) {
	return cl.sdnDiridx(ctx, "/cluster/sdn")
}

// CephIndex enumerates the children of /cluster/ceph ("metadata", "status",
// "flags").
func (cl *Cluster) CephIndex(ctx context.Context) ([]string, error) {
	return cl.cephDiridx(ctx, "/cluster/ceph")
}

// ConfigIndex enumerates the children of /cluster/config ("nodes", "join",
// "totem", "qdevice", "apiversion").
func (cl *Cluster) ConfigIndex(ctx context.Context) ([]string, error) {
	return cl.configDiridx(ctx, "/cluster/config")
}

// HAIndex enumerates the children of /cluster/ha ("groups", "resources",
// "status", "rules").
func (cl *Cluster) HAIndex(ctx context.Context) ([]string, error) {
	return cl.haDiridx(ctx, "/cluster/ha")
}

// HAStatusIndex enumerates the children of /cluster/ha/status ("current",
// "manager_status").
func (cl *Cluster) HAStatusIndex(ctx context.Context) ([]string, error) {
	return cl.haDiridx(ctx, "/cluster/ha/status")
}

// QEMUIndex enumerates the children of /cluster/qemu. Each entry's "subdir"
// is a numeric VMID — these are the QEMU guests visible to the caller
// cluster-wide. Distinct from (*Node).VirtualMachines, which scopes to one
// node and returns rich VM records.
func (cl *Cluster) QEMUIndex(ctx context.Context) ([]string, error) {
	return cl.qemuDiridx(ctx, "/cluster/qemu")
}

// --- diridx helpers ---------------------------------------------------------

func (cl *Cluster) clusterDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) acmeDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) firewallDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) sdnDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) cephDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) configDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) haDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) qemuDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}
