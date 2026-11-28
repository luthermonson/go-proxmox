package proxmox

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GuestMeta is the parsed form of a guest's read-only "meta" config property,
// a comma-joined key=value string such as
// "creation-qemu=9.2.0,ctime=1700000000,rtime=1764150000000,uid=abcdefghij234567klmn".
//
// PVE 9.3 added uid and rtime. A guest gets them the first time a 9.3 node
// writes its configuration, so on older PVE, and for guests no 9.3 node has
// written yet, both are empty.
type GuestMeta struct {
	// UID is the guest's random identity ([a-z2-7]{20}). It is kept across
	// configuration changes, migrations, snapshot rollbacks, backups and
	// restores, also under another VMID or in another cluster. Pair it with
	// Cluster.ClusterID to tell apart equal UIDs in different clusters.
	UID string
	// RTime is when the guest was registered in this cluster, in
	// milliseconds since the UNIX epoch. See RegisteredAt.
	RTime int64
	// CTime is the guest's creation time, in seconds since the UNIX epoch.
	// See CreatedAt.
	CTime int64
	// CreationQEMU is the QEMU machine version the VM was created with.
	// Containers do not have it.
	CreationQEMU string
	// Extra holds any key this package does not know yet, so a newer PVE
	// does not make parsing fail.
	Extra map[string]string
}

// RegisteredAt returns RTime as a time.Time, or the zero time when unset.
func (m *GuestMeta) RegisteredAt() time.Time {
	if m == nil || m.RTime == 0 {
		return time.Time{}
	}
	return time.UnixMilli(m.RTime)
}

// CreatedAt returns CTime as a time.Time, or the zero time when unset.
func (m *GuestMeta) CreatedAt() time.Time {
	if m == nil || m.CTime == 0 {
		return time.Time{}
	}
	return time.Unix(m.CTime, 0)
}

// ParseGuestMeta parses a guest's "meta" property string. An empty string
// yields an empty GuestMeta and no error. A malformed entry or a non-numeric
// ctime/rtime is an error.
func ParseGuestMeta(s string) (*GuestMeta, error) {
	m := &GuestMeta{}
	if strings.TrimSpace(s) == "" {
		return m, nil
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("parse guest meta %q: entry %q is not key=value", s, part)
		}
		switch k {
		case "uid":
			m.UID = v
		case "creation-qemu":
			m.CreationQEMU = v
		case "ctime", "rtime":
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse guest meta %q: %s: %w", s, k, err)
			}
			if k == "ctime" {
				m.CTime = n
			} else {
				m.RTime = n
			}
		default:
			if m.Extra == nil {
				m.Extra = make(map[string]string)
			}
			m.Extra[k] = v
		}
	}
	return m, nil
}

// GuestMeta parses the VM's meta property. See ParseGuestMeta.
func (vmc *VirtualMachineConfig) GuestMeta() (*GuestMeta, error) {
	return ParseGuestMeta(vmc.Meta)
}

// GuestMeta parses the container's meta property. See ParseGuestMeta.
func (cc *ContainerConfig) GuestMeta() (*GuestMeta, error) {
	return ParseGuestMeta(cc.Meta)
}
