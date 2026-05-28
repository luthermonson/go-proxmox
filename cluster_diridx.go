package proxmox

import (
	"context"
)

// Directory-index ("diridx") wrappers for /cluster/* endpoints. See the
// header comment in nodes_diridx.go for the rationale — same shape, same
// shared subdir-list decoder.

// Subdirs enumerates the children of /cluster (typically "replication",
// "metrics", "config", "firewall", "backup", "backupinfo", "ha",
// "acme", "ceph", "jobs", "sdn", "log", "resources", "tasks", "options",
// "status", "nextid", "qemu"). ACL-filtered.
func (cl *Cluster) Subdirs(ctx context.Context) ([]string, error) {
	return cl.clusterDiridx(ctx, "/cluster")
}

// ACMESubdirs enumerates the children of /cluster/acme ("plugins",
// "account", "tos", "meta", "directories", "challenge-schema").
func (cl *Cluster) ACMESubdirs(ctx context.Context) ([]string, error) {
	return cl.acmeDiridx(ctx, "/cluster/acme")
}

// FirewallSubdirs enumerates the children of /cluster/firewall ("groups",
// "rules", "ipset", "aliases", "options", "macros", "refs").
func (cl *Cluster) FirewallSubdirs(ctx context.Context) ([]string, error) {
	return cl.firewallDiridx(ctx, "/cluster/firewall")
}

// SDNSubdirs enumerates the children of /cluster/sdn ("vnets", "zones",
// "controllers", "ipams", "dns", "fabrics", "subnets").
func (cl *Cluster) SDNSubdirs(ctx context.Context) ([]string, error) {
	return cl.sdnDiridx(ctx, "/cluster/sdn")
}

// CephSubdirs enumerates the children of /cluster/ceph ("metadata", "status",
// "flags").
func (cl *Cluster) CephSubdirs(ctx context.Context) ([]string, error) {
	return cl.cephDiridx(ctx, "/cluster/ceph")
}

// ConfigSubdirs enumerates the children of /cluster/config ("nodes", "join",
// "totem", "qdevice", "apiversion").
func (cl *Cluster) ConfigSubdirs(ctx context.Context) ([]string, error) {
	return cl.configDiridx(ctx, "/cluster/config")
}

// HASubdirs enumerates the children of /cluster/ha ("groups", "resources",
// "status", "rules").
func (cl *Cluster) HASubdirs(ctx context.Context) ([]string, error) {
	return cl.haDiridx(ctx, "/cluster/ha")
}

// HAStatusSubdirs enumerates the children of /cluster/ha/status ("current",
// "manager_status").
func (cl *Cluster) HAStatusSubdirs(ctx context.Context) ([]string, error) {
	return cl.haDiridx(ctx, "/cluster/ha/status")
}

// QEMUSubdirs enumerates the children of /cluster/qemu. Each entry's "subdir"
// is a numeric VMID — these are the QEMU guests visible to the caller
// cluster-wide. Distinct from (*Node).VirtualMachines, which scopes to one
// node and returns rich VM records.
func (cl *Cluster) QEMUSubdirs(ctx context.Context) ([]string, error) {
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

func (cl *Cluster) bulkActionDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) backupInfoDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) metricsDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}

func (cl *Cluster) notificationsDiridx(ctx context.Context, path string) ([]string, error) {
	return decodeSubdirList(ctx, cl.client, path)
}
