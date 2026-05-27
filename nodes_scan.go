package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// /nodes/{node}/scan — storage discovery probes. All are GET-with-query, all
// require Datastore.Allocate on /storage (except the diridx). Used during
// "add storage" UI flows: pick a server, list what's there.

// ScanIndex enumerates the children of /nodes/{node}/scan ("zfs", "lvm",
// "nfs", ...). PVE schema returns [{"method":...}]; collapsed to []string.
func (n *Node) ScanIndex(ctx context.Context) ([]string, error) {
	var items []struct {
		Method string `json:"method"`
	}
	if err := n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan", n.Name), &items); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Method)
	}
	return out, nil
}

// ScanZFS lists ZFS pools on this node — local probe, no parameters.
func (n *Node) ScanZFS(ctx context.Context) (pools []*ScanZFSPool, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/zfs", n.Name), &pools)
	return
}

// ScanLVM lists LVM volume groups on this node — local probe, no parameters.
func (n *Node) ScanLVM(ctx context.Context) (vgs []*ScanLVMVG, err error) {
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/lvm", n.Name), &vgs)
	return
}

// ScanLVMThin lists thin pools inside the given LVM volume group.
func (n *Node) ScanLVMThin(ctx context.Context, vg string) (pools []*ScanLVMThinPool, err error) {
	if vg == "" {
		return nil, errors.New("vg is required")
	}
	q := url.Values{}
	q.Set("vg", vg)
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/lvmthin?%s", n.Name, q.Encode()), &pools)
	return
}

// ScanNFS lists NFS exports on the given server. Server is name or IP.
func (n *Node) ScanNFS(ctx context.Context, server string) (exports []*ScanNFSExport, err error) {
	if server == "" {
		return nil, errors.New("server is required")
	}
	q := url.Values{}
	q.Set("server", server)
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/nfs?%s", n.Name, q.Encode()), &exports)
	return
}

// ScanCIFSOptions is the query payload for ScanCIFS. Server is required;
// username/password/domain are needed only for non-anonymous shares.
type ScanCIFSOptions struct {
	Server   string `url:"server"`
	Username string `url:"username,omitempty"`
	Password string `url:"password,omitempty"`
	Domain   string `url:"domain,omitempty"`
}

// ScanCIFS lists shares on the given SMB/CIFS server.
func (n *Node) ScanCIFS(ctx context.Context, opts *ScanCIFSOptions) (shares []*ScanCIFSShare, err error) {
	if opts == nil || opts.Server == "" {
		return nil, errors.New("server is required")
	}
	q := url.Values{}
	q.Set("server", opts.Server)
	if opts.Username != "" {
		q.Set("username", opts.Username)
	}
	if opts.Password != "" {
		q.Set("password", opts.Password)
	}
	if opts.Domain != "" {
		q.Set("domain", opts.Domain)
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/cifs?%s", n.Name, q.Encode()), &shares)
	return
}

// ScanPBSOptions is the query payload for ScanPBS. Server, Username, and
// Password are required; Fingerprint is needed when PBS uses a self-signed
// cert; Port defaults to 8007 server-side.
type ScanPBSOptions struct {
	Server      string
	Username    string
	Password    string
	Fingerprint string
	Port        int
}

// ScanPBS lists datastores on the given Proxmox Backup Server.
func (n *Node) ScanPBS(ctx context.Context, opts *ScanPBSOptions) (stores []*ScanPBSStore, err error) {
	if opts == nil || opts.Server == "" || opts.Username == "" || opts.Password == "" {
		return nil, errors.New("server, username, and password are required")
	}
	q := url.Values{}
	q.Set("server", opts.Server)
	q.Set("username", opts.Username)
	q.Set("password", opts.Password)
	if opts.Fingerprint != "" {
		q.Set("fingerprint", opts.Fingerprint)
	}
	if opts.Port > 0 {
		q.Set("port", fmt.Sprintf("%d", opts.Port))
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/pbs?%s", n.Name, q.Encode()), &stores)
	return
}

// ScanISCSI lists iSCSI targets reachable through the portal (IP or DNS,
// optionally with :port).
func (n *Node) ScanISCSI(ctx context.Context, portal string) (targets []*ScanISCSITarget, err error) {
	if portal == "" {
		return nil, errors.New("portal is required")
	}
	q := url.Values{}
	q.Set("portal", portal)
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/scan/iscsi?%s", n.Name, q.Encode()), &targets)
	return
}
