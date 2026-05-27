package proxmox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

var (
	isFloat = regexp.MustCompile(`^[0-9.]*$`)
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Otp      string `json:"otp,omitempty"` // One-time password for Two-factor authentication.
	Path     string `json:"path,omitempty"`
	Privs    string `json:"privs,omitempty"`
	Realm    string `json:"realm,omitempty"`
}

type Permission map[string]IntOrBool
type Permissions map[string]Permission

type PermissionsOptions struct {
	Path   string // path to limit the return e.g. / or /nodes
	UserID string // username e.g. root@pam or token
}

type Session struct {
	Username            string `json:"username"`
	CSRFPreventionToken string `json:"CSRFPreventionToken,omitempty"`

	// Cap is being returned but not documented in the API docs, likely will get rewritten later with better types
	Cap         map[string]map[string]int `json:"cap,omitempty"`
	ClusterName string                    `json:"clustername,omitempty"`
	Ticket      string                    `json:"ticket,omitempty"`
}

type Version struct {
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
	Version string `json:"version"`
}

// NodeService is one row of the services list and the response shape of
// /nodes/{node}/services/{service}/state. The same struct fits both because
// the list returns the same per-service info, just batched.
type NodeService struct {
	Service     string `json:"service"`
	Name        string `json:"name,omitempty"`
	Desc        string `json:"desc,omitempty"`
	State       string `json:"state,omitempty"`        // running / stopped / unknown
	ActiveState string `json:"active-state,omitempty"` // active / inactive / failed
	UnitState   string `json:"unit-state,omitempty"`   // enabled / disabled / masked
}

// NodeTime is the response from GET /nodes/{node}/time. Time and Localtime
// are unix epoch seconds (UTC and local-tz-adjusted respectively); Timezone
// is the IANA name.
type NodeTime struct {
	Time      int64  `json:"time"`
	Localtime int64  `json:"localtime"`
	Timezone  string `json:"timezone"`
}

// Subscription mirrors GET /nodes/{node}/subscription. Fields are loosely
// typed because PVE's response varies between license levels (community,
// basic, standard, premium) and between current-vs-expired states.
type Subscription struct {
	Status         string `json:"status,omitempty"` // active / inactive / notfound / expired / suspended / new
	Key            string `json:"key,omitempty"`
	Level          string `json:"level,omitempty"` // c=community, b=basic, s=standard, p=premium
	ProductName    string `json:"productname,omitempty"`
	RegDate        string `json:"regdate,omitempty"`     // YYYY-MM-DD HH:MM:SS
	NextDueDate    string `json:"nextduedate,omitempty"` // YYYY-MM-DD
	Validdirectory string `json:"validdirectory,omitempty"`
	Sockets        int    `json:"sockets,omitempty"`
	Checktime      int64  `json:"checktime,omitempty"` // epoch
	ServerID       string `json:"serverid,omitempty"`
	URL            string `json:"url,omitempty"`
	Message        string `json:"message,omitempty"`
	Signature      string `json:"signature,omitempty"`
}

// LogEntry is a single line from PVE log endpoints (task log, replication
// log, etc.) — N is the 1-based line number, T the line text.
type LogEntry struct {
	N int    `json:"n"`
	T string `json:"t"`
}

// NodeReplicationStatus is runtime state for a replication job on a node:
// what was last synced, fail count, next-sync time. The cluster-wide
// configuration of the job lives at /cluster/replication; this is the
// per-node view of how that job is *running*.
type NodeReplicationStatus struct {
	ID        string  `json:"id"`
	Type      string  `json:"type,omitempty"`
	Source    string  `json:"source,omitempty"`
	Target    string  `json:"target,omitempty"`
	Guest     int     `json:"guest,omitempty"`
	JobNum    int     `json:"jobnum,omitempty"`
	Schedule  string  `json:"schedule,omitempty"`
	LastSync  int64   `json:"last_sync,omitempty"` // epoch
	LastTry   int64   `json:"last_try,omitempty"`  // epoch
	NextSync  int64   `json:"next_sync,omitempty"` // epoch
	Duration  float64 `json:"duration,omitempty"`  // seconds
	FailCount int     `json:"fail_count,omitempty"`
	Error     string  `json:"error,omitempty"`
	PID       int     `json:"pid,omitempty"`
	State     string  `json:"state,omitempty"`
}

// APTIndexEntry is one row of the /nodes/{node}/apt directory index. PVE
// returns objects with a single "id" field naming each child resource
// (changelog, repositories, update, versions).
type APTIndexEntry struct {
	ID string `json:"id"`
}

// APTUpdate is one pending package upgrade as reported by /apt/update. Fields
// use PVE's upper-case names verbatim — these come straight from the apt
// metadata.
type APTUpdate struct {
	Package      string `json:"Package"`
	Title        string `json:"Title"`
	Description  string `json:"Description"`
	Version      string `json:"Version"`                // new version
	OldVersion   string `json:"OldVersion,omitempty"`   // installed version
	Origin       string `json:"Origin"`                 // "Proxmox", "Debian", ...
	Arch         string `json:"Arch"`
	Section      string `json:"Section"`
	Priority     string `json:"Priority"`
	NotifyStatus string `json:"NotifyStatus,omitempty"` // version PVE has already notified about
}

// APTPackageVersion is one row of /apt/versions — package info for the
// Proxmox-relevant subset of installed packages, including current install
// state. Used by the GUI's "Updates → Package Versions" panel.
type APTPackageVersion struct {
	Package        string `json:"Package"`
	Title          string `json:"Title"`
	Description    string `json:"Description"`
	Version        string `json:"Version"`
	OldVersion     string `json:"OldVersion,omitempty"`
	Origin         string `json:"Origin"`
	Arch           string `json:"Arch"`
	Section        string `json:"Section"`
	Priority       string `json:"Priority"`
	CurrentState   string `json:"CurrentState"`             // Installed / NotInstalled / ...
	ManagerVersion string `json:"ManagerVersion,omitempty"` // only on pve-manager
	RunningKernel  string `json:"RunningKernel,omitempty"`  // only on proxmox-ve
	NotifyStatus   string `json:"NotifyStatus,omitempty"`
}

// APTRepositories is the parsed view of /etc/apt/sources.list(.d) plus a
// global Digest used as an etag for concurrent edits. StandardRepos is PVE's
// catalog of repositories the GUI knows how to add; the per-handle Status is
// nil when the repo isn't configured on the node.
type APTRepositories struct {
	Digest        string                  `json:"digest"`
	Files         []*APTRepositoryFile    `json:"files,omitempty"`
	Errors        []*APTRepositoryError   `json:"errors,omitempty"`
	Infos         []*APTRepositoryInfo    `json:"infos,omitempty"`
	StandardRepos []*APTStandardRepo      `json:"standard-repos,omitempty"`
}

// APTRepositoryFile is one apt sources file on disk. FileType is "list"
// (one-line) or "sources" (deb822). Digest is the per-file digest as a byte
// array (PVE returns it as a JSON array of integers).
type APTRepositoryFile struct {
	Path         string            `json:"path"`
	FileType     string            `json:"file-type"`
	Digest       []int             `json:"digest,omitempty"`
	Repositories []*APTRepository  `json:"repositories,omitempty"`
}

// APTRepository is a single repository entry within a file. Components,
// Options and Comment are only populated where the underlying file format
// supports them.
type APTRepository struct {
	Enabled    bool                   `json:"Enabled"`
	FileType   string                 `json:"FileType"`
	Types      []string               `json:"Types"`
	URIs       []string               `json:"URIs"`
	Suites     []string               `json:"Suites"`
	Components []string               `json:"Components,omitempty"`
	Options    []*APTRepositoryOption `json:"Options,omitempty"`
	Comment    string                 `json:"Comment,omitempty"`
}

type APTRepositoryOption struct {
	Key    string   `json:"Key"`
	Values []string `json:"Values"`
}

type APTRepositoryError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// APTRepositoryInfo is a warning/info note PVE attaches to a specific entry
// within a specific file (e.g. "use of subscription repo on free system").
// Index is a string in the schema even though it names a numeric position.
type APTRepositoryInfo struct {
	Path     string `json:"path"`
	Index    string `json:"index"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
	Property string `json:"property,omitempty"`
}

// APTStandardRepo is one entry from PVE's catalog of well-known repos.
// Status is *bool because tri-state: true = configured+enabled,
// false = configured+disabled, nil = not present on the node.
type APTStandardRepo struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Status *bool  `json:"status,omitempty"`
}

type Term struct {
	Port   StringOrInt
	Ticket string
	UPID   string
	User   string
}

type VNCConfig struct {
	GeneratePassword bool `json:"generate-password,omitempty"`
	Websocket        bool `json:"websocket,omitempty"`
}

type VNC struct {
	Cert     string
	Port     StringOrInt
	Ticket   string
	UPID     string
	User     string
	Password string `json:",omitempty"`
}

type Cluster struct {
	client  *Client
	Version int
	Quorate int
	Nodes   NodeStatuses
	Name    string
	ID      string
}

func (cl *Cluster) UnmarshalJSON(b []byte) error {
	var tmp []map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	for _, d := range tmp {
		t, ok := d["type"]
		if !ok {
			break
		}

		switch t.(string) {
		case "cluster":
			if v, ok := d["id"]; ok {
				cl.ID = v.(string)
			}
			if v, ok := d["name"]; ok {
				cl.Name = v.(string)
			}
			if v, ok := d["version"]; ok {
				cl.Version = int(v.(float64))
			}
			if v, ok := d["quorate"]; ok {
				cl.Quorate = int(v.(float64))
			}
		case "node":
			ns := NodeStatus{
				Status: "offline",
				Type:   "node",
			}
			if v, ok := d["name"]; ok {
				ns.Name = v.(string)
			}
			if v, ok := d["level"]; ok {
				ns.Level = v.(string)
			}
			if v, ok := d["online"]; ok {
				ns.Online = int(v.(float64))
				if ns.Online == 1 {
					ns.Status = "online"
				}
			}
			if v, ok := d["id"]; ok {
				ns.ID = v.(string)
			}
			if v, ok := d["ip"]; ok {
				ns.IP = v.(string)
			}
			if v, ok := d["local"]; ok {
				ns.Local = int(v.(float64))
			}

			cl.Nodes = append(cl.Nodes, &ns)
		}
	}

	return nil
}

type ClusterResources []*ClusterResource

type ClusterResource struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	CGroupMode uint64  `json:"cgroup-mode,omitempty"`
	Content    string  `json:",omitempty"`
	CPU        float64 `json:",omitempty"`
	Disk       uint64  `json:",omitempty"` // documented as string but this is an int
	DiskRead   uint64  `json:",omitempty"`
	DiskWrite  uint64  `json:",omitempty"`
	HAstate    string  `json:",omitempty"`
	Level      string  `json:",omitempty"`
	MaxCPU     uint64  `json:",omitempty"`
	MaxDisk    uint64  `json:",omitempty"`
	MaxMem     uint64  `json:",omitempty"`
	Mem        uint64  `json:",omitempty"` // documented as string but this is an int
	Name       string  `json:",omitempty"`
	NetIn      uint64  `json:",omitempty"`
	NetOut     uint64  `json:",omitempty"`
	Node       string  `json:",omitempty"`
	PluginType string  `json:",omitempty"`
	Pool       string  `json:",omitempty"`
	Shared     uint64  `json:",omitempty"`
	Status     string  `json:",omitempty"`
	Storage    string  `json:",omitempty"`
	Tags       string  `json:",omitempty"`
	Template   uint64  `json:",omitempty"`
	Uptime     uint64  `json:",omitempty"`
	VMID       uint64  `json:",omitempty"`
}

type Ceph struct {
	client *Client
}

type ClusterCephStatus struct {
	ElectionEpoch  int            `json:"election_epoch"`
	Fsid           string         `json:"fsid"`
	Fsmap          CephFsMap      `json:"fsmap"`
	Health         CephHealth     `json:"health"`
	Mgrmap         CephMgrMap     `json:"mgrmap"`
	Monmap         CephMonMap     `json:"monmap"`
	Osdmap         CephOsdMap     `json:"osdmap"`
	Pgmap          CephPgMap      `json:"pgmap"`
	ProgressEvents struct{}       `json:"progress_events"`
	Quorum         []int          `json:"quorum"`
	QuorumAge      int            `json:"quorum_age"`
	QuorumNames    []string       `json:"quorum_names"`
	Servicemap     CephServiceMap `json:"servicemap"`
}

type CephHealthCheckName string
type CephHealthCheckDetail struct {
	Message string `json:"message"`
}
type CephHealthCheckSummary struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}
type CephHealthCheck struct {
	Detail   []CephHealthCheckDetail `json:"detail"`
	Muted    bool                    `json:"muted"`
	Severity string                  `json:"severity"`
	Summary  CephHealthCheckSummary  `json:"summary"`
}

type CephHealth struct {
	Checks map[CephHealthCheckName]CephHealthCheck `json:"checks"`
	Mutes  []interface{}                           `json:"mutes"`
	Status string                                  `json:"status"`
}

type CephOsdMap struct {
	Epoch          int `json:"epoch"`
	NumInOsds      int `json:"num_in_osds"`
	NumOsds        int `json:"num_osds"`
	NumRemappedPgs int `json:"num_remapped_pgs"`
	NumUpOsds      int `json:"num_up_osds"`
	OsdInSince     int `json:"osd_in_since"`
	OsdUpSince     int `json:"osd_up_since"`
}

type CephPgMap struct {
	BytesAvail int64 `json:"bytes_avail"`
	BytesTotal int64 `json:"bytes_total"`
	BytesUsed  int64 `json:"bytes_used"`
	DataBytes  int64 `json:"data_bytes"`
	NumObjects int   `json:"num_objects"`
	NumPgs     int   `json:"num_pgs"`
	NumPools   int   `json:"num_pools"`
	PgsByState []struct {
		Count     int    `json:"count"`
		StateName string `json:"state_name"`
	} `json:"pgs_by_state"`
	ReadBytesSec  int `json:"read_bytes_sec"`
	ReadOpPerSec  int `json:"read_op_per_sec"`
	WriteBytesSec int `json:"write_bytes_sec"`
	WriteOpPerSec int `json:"write_op_per_sec"`
}

type CephMonMap struct {
	Created           time.Time       `json:"created"`
	DisallowedLeaders string          `json:"disallowed_leaders: "`
	ElectionStrategy  int             `json:"election_strategy"`
	Epoch             int             `json:"epoch"`
	Features          CephMonFeatures `json:"features"`
	Fsid              string          `json:"fsid"`
	MinMonRelease     int             `json:"min_mon_release"`
	MinMonReleaseName string          `json:"min_mon_release_name"`
	Modified          time.Time       `json:"modified"`
	Mons              []ClusterCephMon `json:"mons"`
	Quorum            []int           `json:"quorum"`
	RemovedRanks      string          `json:"removed_ranks: "`
	StretchMode       bool            `json:"stretch_mode"`
	TiebreakerMon     string          `json:"tiebreaker_mon"`
}

// ClusterCephMon is the cluster-status snapshot of a monitor (used inside
// ClusterCephStatus.Monmap.Mons). It's distinct from *CephMon, the per-node
// monitor handle returned by Node.CephMon(name) that carries operations.
type ClusterCephMon struct {
	Addr          string `json:"addr"`
	CrushLocation string `json:"crush_location"`
	Name          string `json:"name"`
	Priority      int    `json:"priority"`
	Rank          int    `json:"rank"`
	Weight        int    `json:"weight"`
	PublicAddr    string `json:"public_addr"`
	PublicAddrs   struct {
		Addrvec []CephMgrAddrVector `json:"addrvec"`
	} `json:"public_addrs"`
}

type CephMonFeatures struct {
	Optional   []interface{} `json:"optional"`
	Persistent []string      `json:"persistent"`
}

type CephFsMap struct {
	ByRank []struct {
		FilesystemID int    `json:"filesystem_id"`
		Gid          int    `json:"gid"`
		Name         string `json:"name"`
		Rank         int    `json:"rank"`
		Status       string `json:"status"`
	} `json:"by_rank"`
	Epoch     int `json:"epoch"`
	ID        int `json:"id"`
	In        int `json:"in"`
	Max       int `json:"max"`
	Up        int `json:"up"`
	UpStandby int `json:"up:standby"`
}

type CephServiceMap struct {
	Epoch    int      `json:"epoch"`
	Modified string   `json:"modified"`
	Services struct{} `json:"services"`
}

type CephMgrMap struct {
	ActiveAddr          string                   `json:"active_addr"`
	ActiveAddrs         CephMgrActiveAddresses   `json:"active_addrs"`
	ActiveChange        string                   `json:"active_change"`
	ActiveClients       []CephMgrActiveClient    `json:"active_clients"`
	ActiveGid           int                      `json:"active_gid"`
	ActiveMgrFeatures   int64                    `json:"active_mgr_features"`
	ActiveName          string                   `json:"active_name"`
	AlwaysOnModules     CephMgrAlwaysOnModules   `json:"always_on_modules"`
	Available           bool                     `json:"available"`
	AvailableModules    []CephMgrAvailableModule `json:"available_modules"`
	Epoch               int                      `json:"epoch"`
	LastFailureOsdEpoch int                      `json:"last_failure_osd_epoch"`
	Modules             []string                 `json:"modules"`
	Services            CephMgrServices          `json:"services"`
	Standbys            []CephMgrStandby         `json:"standbys"`
}

type CephMgrAvailableModule struct {
	CanRun        bool                          `json:"can_run"`
	ErrorString   string                        `json:"error_string"`
	ModuleOptions CephMgrAvailableModuleOptions `json:"module_options"`
	Name          string                        `json:"name"`
}

type CephMgrAvailableModuleOptions struct {
	Interval          CephMgrAvailableModuleOption `json:"interval"`
	LogLevel          CephMgrAvailableModuleOption `json:"log_level"`
	LogToCluster      CephMgrAvailableModuleOption `json:"log_to_cluster"`
	LogToClusterLevel CephMgrAvailableModuleOption `json:"log_to_cluster_level"`
	LogToFile         CephMgrAvailableModuleOption `json:"log_to_file"`
	SMTPDestination   CephMgrAvailableModuleOption `json:"smtp_destination"`
	SMTPFromName      CephMgrAvailableModuleOption `json:"smtp_from_name"`
	SMTPHost          CephMgrAvailableModuleOption `json:"smtp_host"`
	SMTPPassword      CephMgrAvailableModuleOption `json:"smtp_password"`
	SMTPPort          CephMgrAvailableModuleOption `json:"smtp_port"`
	SMTPSender        CephMgrAvailableModuleOption `json:"smtp_sender"`
	SMTPSsl           CephMgrAvailableModuleOption `json:"smtp_ssl"`
	SMTPUser          CephMgrAvailableModuleOption `json:"smtp_user"`
}

type CephMgrAvailableModuleOption struct {
	DefaultValue string        `json:"default_value"`
	Desc         string        `json:"desc"`
	EnumAllowed  []string      `json:"enum_allowed"`
	Flags        int           `json:"flags"`
	Level        string        `json:"level"`
	LongDesc     string        `json:"long_desc"`
	Max          string        `json:"max"`
	Min          string        `json:"min"`
	Name         string        `json:"name"`
	SeeAlso      []interface{} `json:"see_also"`
	Tags         []interface{} `json:"tags"`
	Type         string        `json:"type"`
}

type CephMgrServices struct {
	Dashboard  string `json:"dashboard"`
	Prometheus string `json:"prometheus"`
}

type CephMgrStandby struct {
	AvailableModules []CephMgrAvailableModule `json:"available_modules"`
	Gid              int                      `json:"gid"`
	MgrFeatures      int64                    `json:"mgr_features"`
	Name             string                   `json:"name"`
}

type CephMgrActiveAddresses struct {
	Addrvec []CephMgrAddrVector `json:"addrvec"`
}

type CephMgrAddrVector struct {
	Addr  string `json:"addr"`
	Nonce int    `json:"nonce"`
	Type  string `json:"type"`
}

type CephMgrActiveClient struct {
	Addrvec []CephMgrAddrVector `json:"addrvec"`
	Name    string              `json:"name"`
}

type CephMgrAlwaysOnModules struct {
	Octopus []string `json:"octopus"`
	Pacific []string `json:"pacific"`
	Quincy  []string `json:"quincy"`
	Reef    []string `json:"reef"`
	Squid   []string `json:"squid"`
}

// CephFS is a single entry from the list at GET /nodes/{node}/ceph/fs AND the
// operations handle returned by Node.CephFS(name). A CephFS may have multiple
// data pools — DataPool/MetadataPool are the legacy scalar fields (kept for
// backwards compatibility) and DataPools/DataPoolIDs expose the full set.
type CephFS struct {
	client         *Client
	Node           string   `json:"-"`
	Name           string   `json:"name"`
	MetadataPool   string   `json:"metadata_pool"`
	MetadataPoolID int      `json:"metadata_pool_id,omitempty"`
	DataPool       string   `json:"data_pool"`
	DataPools      []string `json:"data_pools,omitempty"`
	DataPoolIDs    []int    `json:"data_pool_ids,omitempty"`
}

// CephFSOptions is the body for POST /nodes/{node}/ceph/fs/{name}. All
// fields are optional: PVE defaults Name to "cephfs", PgNum to 128, and
// AddStorage to false.
type CephFSOptions struct {
	PgNum      int       `json:"pg_num,omitempty"`
	AddStorage IntOrBool `json:"add-storage,omitempty"`
}

// CephCfgDBEntry is a single row from the Ceph mon config DB
// (GET /nodes/{node}/ceph/cfg/db). Value is always a string — Ceph stores
// every option as a string regardless of its underlying type.
type CephCfgDBEntry struct {
	Section            string    `json:"section"`
	Name               string    `json:"name"`
	Value              string    `json:"value"`
	Level              string    `json:"level,omitempty"`
	Mask               string    `json:"mask,omitempty"`
	CanUpdateAtRuntime IntOrBool `json:"can_update_at_runtime,omitempty"`
}

// CephCfgValue is the response to GET /nodes/{node}/ceph/cfg/value: a
// two-level map of section → key → value. Underscores in both section and
// key names are normalised to hyphens by PVE.
type CephCfgValue map[string]map[string]string

// CephIndexEntry is one row of the /nodes/{node}/ceph directory index — each
// entry names a child resource (osd, mon, mgr, pool, fs, status, log, …).
type CephIndexEntry struct {
	Subdir string `json:"subdir,omitempty"`
}

// CephInitOptions are the parameters for POST /nodes/{node}/ceph/init — the
// one-time bootstrap that seeds /etc/ceph/ceph.conf with cluster fsid,
// default pool sizing, and network settings. All fields are optional; PVE
// applies sensible defaults (size=3, min_size=2, etc.). Re-calling init on
// a node that already has a [global] section is a no-op for most fields.
type CephInitOptions struct {
	// Network restricts all Ceph traffic to the given CIDR. Required when
	// you want to pin Ceph to a non-default subnet.
	Network string `json:"network,omitempty"`
	// ClusterNetwork is the optional separate CIDR for OSD heartbeat /
	// replication / recovery traffic. PVE rejects this without Network.
	ClusterNetwork string `json:"cluster-network,omitempty"`
	// Size is the target number of replicas per object (1-7). PVE default 3.
	Size int `json:"size,omitempty"`
	// MinSize is the minimum replicas required to accept I/O (1-7). PVE
	// default 2. Must be <= Size.
	MinSize int `json:"min_size,omitempty"`
	// PGBits is the legacy default placement-group bit count (6-14, default
	// 6). Deprecated upstream in recent Ceph releases; usually leave unset.
	PGBits int `json:"pg_bits,omitempty"`
	// DisableCephx turns off cephx authentication. Dangerous on untrusted
	// networks — only set true when the cluster network is fully private.
	DisableCephx bool `json:"disable_cephx,omitempty"`
}

// CephLogEntry is one line of the cluster log as returned by
// /nodes/{node}/ceph/log. PVE uses single-letter field names ("n", "t") for
// the line-number and text — matching the dump_logfile wire format.
type CephLogEntry struct {
	N int    `json:"n"` // 1-based log-file line number
	T string `json:"t"` // log line text
}

// CephRule is one entry of the CRUSH rules list. PVE returns only the rule
// name here; the rule body lives in the CRUSH map dumped by CephCrush.
type CephRule struct {
	Name string `json:"name"`
}

// CephCmdSafety is the response from the cmd-safety probe — true means Ceph
// considers the requested action (stop/destroy of a service) safe to perform
// right now without losing data redundancy. Status carries the
// human-readable reason when Safe is false.
type CephCmdSafety struct {
	Safe   bool   `json:"safe"`
	Status string `json:"status,omitempty"`
}

// --- Ceph pool (per-node /nodes/{node}/ceph/pool/*) ------------------------

// CephPool is one row returned by GET /nodes/{node}/ceph/pool AND the
// operations handle returned by Node.CephPool(name). Optional fields
// (statistics-bearing, autoscaler-derived) may be absent depending on Ceph
// release and whether the pool reports usage.
type CephPool struct {
	client              *Client
	Node                string         `json:"-"`
	ApplicationMetadata map[string]any `json:"application_metadata,omitempty"`
	AutoscaleStatus     map[string]any `json:"autoscale_status,omitempty"`
	BytesUsed           uint64         `json:"bytes_used,omitempty"`
	CrushRule           int            `json:"crush_rule"`
	CrushRuleName       string         `json:"crush_rule_name,omitempty"`
	MinSize             int            `json:"min_size"`
	PercentUsed         float64        `json:"percent_used,omitempty"`
	PgAutoscaleMode     string         `json:"pg_autoscale_mode,omitempty"`
	PgNum               int            `json:"pg_num"`
	PgNumFinal          int            `json:"pg_num_final,omitempty"`
	PgNumMin            int            `json:"pg_num_min,omitempty"`
	Pool                int            `json:"pool"`
	PoolName            string         `json:"pool_name"`
	Size                int            `json:"size"`
	TargetSize          uint64         `json:"target_size,omitempty"`
	TargetSizeRatio     float64        `json:"target_size_ratio,omitempty"`
	Type                string         `json:"type"`
}

// CephPoolSubdir is one row from GET /nodes/{node}/ceph/pool/{name} — the
// sub-resource directory index. Currently the only entry is "status".
type CephPoolSubdir struct {
	Subdir string `json:"subdir,omitempty"`
}

// CephPoolStatus is the response body of GET /nodes/{node}/ceph/pool/{name}/status.
// Statistics is only populated when the request was made with verbose=1.
type CephPoolStatus struct {
	Application          string         `json:"application,omitempty"`
	ApplicationList      []string       `json:"application_list,omitempty"`
	AutoscaleStatus      map[string]any `json:"autoscale_status,omitempty"`
	CrushRule            string         `json:"crush_rule,omitempty"`
	FastRead             bool           `json:"fast_read"`
	HashPSPool           bool           `json:"hashpspool"`
	ID                   int            `json:"id"`
	MinSize              int            `json:"min_size,omitempty"`
	Name                 string         `json:"name"`
	NoDeepScrub          bool           `json:"nodeep-scrub"`
	NoDelete             bool           `json:"nodelete"`
	NoPGChange           bool           `json:"nopgchange"`
	NoScrub              bool           `json:"noscrub"`
	NoSizeChange         bool           `json:"nosizechange"`
	PgAutoscaleMode      string         `json:"pg_autoscale_mode,omitempty"`
	PgNum                int            `json:"pg_num,omitempty"`
	PgNumMin             int            `json:"pg_num_min,omitempty"`
	PgpNum               int            `json:"pgp_num"`
	Size                 int            `json:"size,omitempty"`
	Statistics           map[string]any `json:"statistics,omitempty"`
	TargetSize           string         `json:"target_size,omitempty"`
	TargetSizeRatio      float64        `json:"target_size_ratio,omitempty"`
	UseGMTHitset         bool           `json:"use_gmt_hitset"`
	WriteFadviseDontneed bool           `json:"write_fadvise_dontneed"`
}

// CephPoolErasureCoding is the inline "erasure-coding" parameter accepted by
// POST /nodes/{node}/ceph/pool. PVE serializes it as a single comma-separated
// string of key=value pairs (e.g. "k=4,m=2,failure-domain=host"). K and M are
// required; the rest are optional. Build the string with String().
type CephPoolErasureCoding struct {
	K             int    // required: number of data chunks
	M             int    // required: number of coding chunks
	DeviceClass   string // optional: CRUSH device class
	FailureDomain string // optional: CRUSH failure domain (default "host")
	Profile       string // optional: override EC profile name
}

// String serializes the EC config to the PVE wire format
// "k=<int>,m=<int>[,device-class=<class>][,failure-domain=<domain>][,profile=<name>]".
func (ec *CephPoolErasureCoding) String() string {
	if ec == nil {
		return ""
	}
	parts := []string{
		fmt.Sprintf("k=%d", ec.K),
		fmt.Sprintf("m=%d", ec.M),
	}
	if ec.DeviceClass != "" {
		parts = append(parts, "device-class="+ec.DeviceClass)
	}
	if ec.FailureDomain != "" {
		parts = append(parts, "failure-domain="+ec.FailureDomain)
	}
	if ec.Profile != "" {
		parts = append(parts, "profile="+ec.Profile)
	}
	return strings.Join(parts, ",")
}

// CephPoolOptions is the POST body for /nodes/{node}/ceph/pool (create) and
// the PUT body for /nodes/{node}/ceph/pool/{name} (update). Name is required
// on create and immutable on update — the URL path supplies it for PUT.
//
// Pointer fields (*int, *bool) are used wherever PVE has a server-side default
// that should be preserved when the caller leaves the field unset; this avoids
// silently clobbering Ceph defaults (size=3, min_size=2, pg_num=128, etc.).
type CephPoolOptions struct {
	Name             string                 `json:"name,omitempty"`
	AddStorages      *bool                  `json:"add_storages,omitempty"`
	Application      string                 `json:"application,omitempty"`
	CrushRule        string                 `json:"crush_rule,omitempty"`
	ErasureCoding    *CephPoolErasureCoding `json:"-"` // serialized by helper, see CreateCephPool
	MinSize          *int                   `json:"min_size,omitempty"`
	PgAutoscaleMode  string                 `json:"pg_autoscale_mode,omitempty"`
	PgNum            *int                   `json:"pg_num,omitempty"`
	PgNumMin         *int                   `json:"pg_num_min,omitempty"`
	Size             *int                   `json:"size,omitempty"`
	TargetSize       string                 `json:"target_size,omitempty"`
	TargetSizeRatio  *float64               `json:"target_size_ratio,omitempty"`
}

type NodeStatuses []*NodeStatus
type NodeStatus struct {
	// shared
	Status string `json:",omitempty"`
	Level  string `json:",omitempty"`
	ID     string `json:",omitempty"` // format "node/<name>"

	// from /nodes endpoint
	Node           string  `json:",omitempty"`
	Type           string  `json:",omitempty"`
	MaxCPU         int     `json:",omitempty"`
	MaxMem         uint64  `json:",omitempty"`
	Disk           uint64  `json:",omitempty"`
	SSLFingerprint string  `json:"ssl_fingerprint,omitempty"`
	MaxDisk        uint64  `json:",omitempty"`
	Mem            uint64  `json:",omitempty"`
	CPU            float64 `json:",omitempty"`
	Uptime         uint64  `json:",omitempty"`

	// from /cluster endpoint
	NodeID int    `json:",omitempty"` // the internal id of the node
	Name   string `json:",omitempty"`
	IP     string `json:",omitempty"`
	Online int    `json:",omitempty"`
	Local  int    `json:",omitempty"`
}

// NodeStartAllOptions is the optional body for POST /nodes/{node}/startall.
type NodeStartAllOptions struct {
	Force IntOrBool `json:"force,omitempty"` // bypass configured startup order
	VMs   string    `json:"vms,omitempty"`   // comma-separated VMID list to limit which guests are started
}

// NodeStopAllOptions is the optional body for POST /nodes/{node}/stopall.
type NodeStopAllOptions struct {
	ForceStop IntOrBool `json:"force-stop,omitempty"` // PVE default 1; pass IntOrBool(false) to allow graceful shutdown to time out
	Timeout   uint64    `json:"timeout,omitempty"`    // per-guest shutdown timeout in seconds (PVE default 180)
	VMs       string    `json:"vms,omitempty"`        // comma-separated VMID list to limit
}

// NodeSuspendAllOptions is the optional body for POST /nodes/{node}/suspendall.
type NodeSuspendAllOptions struct {
	VMs string `json:"vms,omitempty"` // comma-separated VMID list to limit
}

// NodeMigrateAllOptions is the body for POST /nodes/{node}/migrateall.
// Target is required — the destination node name.
type NodeMigrateAllOptions struct {
	Target         string    `json:"target"`
	MaxWorkers     uint64    `json:"maxworkers,omitempty"`       // parallel migration workers
	VMs            string    `json:"vms,omitempty"`              // comma-separated VMID list to limit
	WithLocalDisks IntOrBool `json:"with-local-disks,omitempty"` // include local disks via storage migration
}

type Node struct {
	Name       string
	client     *Client
	Kversion   string
	LoadAvg    []string
	CPU        float64
	RootFS     RootFS
	PVEVersion string
	CPUInfo    CPUInfo
	Swap       Memory
	Idle       int
	Memory     Memory
	Ksm        Ksm
	Uptime     uint64
	Wait       float64
}

type VirtualMachines []*VirtualMachine
type VirtualMachine struct {
	client               *Client
	VirtualMachineConfig *VirtualMachineConfig

	Name string
	Node string

	Agent          IntOrBool
	Spice          IntOrBool
	NetIn          uint64
	CPUs           int
	DiskWrite      uint64
	Status         string
	Lock           string `json:",omitempty"`
	VMID           StringOrUint64
	PID            StringOrUint64
	Netout         uint64
	Disk           uint64
	Mem            uint64
	CPU            float64
	MaxMem         uint64
	MaxDisk        uint64
	DiskRead       uint64
	QMPStatus      string `json:"qmpstatus,omitempty"`
	RunningMachine string `json:"running-machine,omitempty"`
	RunningQemu    string `json:"running-qemu,omitempty"`
	Tags           string `json:"tags,omitempty"`
	Uptime         uint64
	Template       IsTemplate // empty str if a vm, int 1 if a template
	HA             HA         `json:",omitempty"`
}

type HA struct {
	Managed int
}

type RootFS struct {
	Avail uint64
	Total uint64
	Free  uint64
	Used  uint64
}

type CPUInfo struct {
	UserHz  int `json:"user_hz"`
	MHZ     StringOrInt
	Model   string
	Cores   int
	Sockets int
	Flags   string
	CPUs    int
	HVM     string
}

type Memory struct {
	Used  uint64
	Free  uint64
	Total uint64
}

type Ksm struct {
	Shared int64
}

type Time struct {
	Timezone  string
	Time      uint64
	Localtime uint64
}

type Timeframe string

const (
	TimeframeHour  = Timeframe("hour")
	TimeframeDay   = Timeframe("day")
	TimeframeWeek  = Timeframe("week")
	TimeframeMonth = Timeframe("month")
	TimeframeYear  = Timeframe("year")
)

type ConsolidationFunction string

const (
	AVERAGE = ConsolidationFunction("AVERAGE")
	MAX     = ConsolidationFunction("MAX")
)

type RRDData struct {
	Time      uint64
	CPU       float64
	MaxCPU    int
	Mem       float64
	MaxMem    uint64
	Disk      int
	MaxDisk   uint64
	DiskRead  float64
	DiskWrite float64
	NetIn     float64
	NetOut    float64
}

// VirtualMachineOptions A key/value pair used to modify a virtual machine config
// Refer to https://pve.proxmox.com/pve-docs/api-viewer/#/nodes/{node}/qemu/{vmid}/config for a list of valid values
type VirtualMachineOptions []*VirtualMachineOption
type VirtualMachineOption struct {
	Name  string
	Value interface{}
}

type VirtualMachineConfig struct {
	// PVE Metadata
	Digest      string `json:"digest"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Meta        string `json:"meta,omitempty"`
	VMGenID     string `json:"vmgenid,omitempty"` // FIXME(issue-199): PVE default "1 (autogenerated)"; use *string so unset doesn't override.
	Hookscript  string `json:"hookscript,omitempty"`
	Hotplug     string `json:"hotplug,omitempty"` // FIXME(issue-199): PVE default "network,disk,usb"; use *string so unset doesn't override.
	Template    int    `json:"template,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Agent       string `json:"agent,omitempty"`
	Autostart   int    `json:"autostart,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Tablet      int    `json:"tablet,omitempty"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch (int vs boolean) + default differs.
	KVM         int    `json:"kvm,omitempty"`    // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would disable hardware virt).

	Tags      string   `json:"tags,omitempty"`
	TagsSlice []string `json:"-"` // internal helper to manage tags easier

	Protection int    `json:"protection,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Lock       string `json:"lock,omitempty"`

	// Boot configuration
	Boot   string `json:"boot,omitempty"`
	OnBoot int    `json:"onboot,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed). ContainerConfig already uses IntOrBool here.

	// Qemu general specs
	OSType  string `json:"ostype,omitempty"` // FIXME(issue-199): PVE default "other"; use *string so unset doesn't override.
	Machine string `json:"machine,omitempty"`
	Args    string `json:"args,omitempty"`

	// Qemu firmware specs
	Bios     string `json:"bios,omitempty"` // FIXME(issue-199): PVE default "seabios"; use *string so unset doesn't override.
	EFIDisk0 string `json:"efidisk0,omitempty"`
	SMBios1  string `json:"smbios1,omitempty"`
	Acpi     int    `json:"acpi,omitempty"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would disable ACPI).

	// Qemu CPU specs
	Sockets  int             `json:"sockets,omitempty"`  // FIXME(issue-199): PVE default 1; use *int so unset doesn't override.
	Cores    int             `json:"cores,omitempty"`    // FIXME(issue-199): PVE default 1; use *int so unset doesn't override.
	CPU      string          `json:"cpu,omitempty"`
	CPULimit StringOrFloat64 `json:"cpulimit,omitempty"` // FIXME(issue-199): PVE default 0 ("no limit"); empty-string Go zero is silently dropped — fine here, but explicit `0` from the user is also dropped. Consider *StringOrFloat64.
	CPUUnits int             `json:"cpuunits,omitempty"` // FIXME(issue-199): PVE default 1024 (cgroup v1) / 100 (cgroup v2); use *int so unset doesn't override.
	Vcpus    int             `json:"vcpus,omitempty"`
	Affinity string          `json:"affinity,omitempty"`

	// Qemu memory specs
	Numa      int         `json:"numa,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Memory    StringOrInt `json:"memory,omitempty"` // See commit 7f8c808772979f274cdfac1dc7264771a3b7a7ae on qemu-server
	Hugepages string      `json:"hugepages,omitempty"`
	Balloon   int         `json:"balloon,omitempty"`

	// Other Qemu devices
	VGA       string `json:"vga,omitempty"`
	SCSIHW    string `json:"scsihw,omitempty"` // FIXME(issue-199): PVE default "lsi"; use *string so unset doesn't override.
	TPMState0 string `json:"tpmstate0,omitempty"`
	Rng0      string `json:"rng0,omitempty"`
	Audio0    string `json:"audio0,omitempty"`

	// Disk devices
	IDEs map[string]string `json:"-"`
	IDE0 string            `json:"ide0,omitempty"`
	IDE1 string            `json:"ide1,omitempty"`
	IDE2 string            `json:"ide2,omitempty"`
	IDE3 string            `json:"ide3,omitempty"`

	SCSIs  map[string]string `json:"-"`
	SCSI0  string            `json:"scsi0,omitempty"`
	SCSI1  string            `json:"scsi1,omitempty"`
	SCSI2  string            `json:"scsi2,omitempty"`
	SCSI3  string            `json:"scsi3,omitempty"`
	SCSI4  string            `json:"scsi4,omitempty"`
	SCSI5  string            `json:"scsi5,omitempty"`
	SCSI6  string            `json:"scsi6,omitempty"`
	SCSI7  string            `json:"scsi7,omitempty"`
	SCSI8  string            `json:"scsi8,omitempty"`
	SCSI9  string            `json:"scsi9,omitempty"`
	SCSI10 string            `json:"scsi10,omitempty"`
	SCSI11 string            `json:"scsi11,omitempty"`
	SCSI12 string            `json:"scsi12,omitempty"`
	SCSI13 string            `json:"scsi13,omitempty"`
	SCSI14 string            `json:"scsi14,omitempty"`
	SCSI15 string            `json:"scsi15,omitempty"`
	SCSI16 string            `json:"scsi16,omitempty"`
	SCSI17 string            `json:"scsi17,omitempty"`
	SCSI18 string            `json:"scsi18,omitempty"`
	SCSI19 string            `json:"scsi19,omitempty"`
	SCSI20 string            `json:"scsi20,omitempty"`
	SCSI21 string            `json:"scsi21,omitempty"`
	SCSI22 string            `json:"scsi22,omitempty"`
	SCSI23 string            `json:"scsi23,omitempty"`
	SCSI24 string            `json:"scsi24,omitempty"`
	SCSI25 string            `json:"scsi25,omitempty"`
	SCSI26 string            `json:"scsi26,omitempty"`
	SCSI27 string            `json:"scsi27,omitempty"`
	SCSI28 string            `json:"scsi28,omitempty"`
	SCSI29 string            `json:"scsi29,omitempty"`
	SCSI30 string            `json:"scsi30,omitempty"`

	SATAs map[string]string `json:"-"`
	SATA0 string            `json:"sata0,omitempty"`
	SATA1 string            `json:"sata1,omitempty"`
	SATA2 string            `json:"sata2,omitempty"`
	SATA3 string            `json:"sata3,omitempty"`
	SATA4 string            `json:"sata4,omitempty"`
	SATA5 string            `json:"sata5,omitempty"`

	VirtIOs  map[string]string `json:"-"`
	VirtIO0  string            `json:"virtio0,omitempty"`
	VirtIO1  string            `json:"virtio1,omitempty"`
	VirtIO2  string            `json:"virtio2,omitempty"`
	VirtIO3  string            `json:"virtio3,omitempty"`
	VirtIO4  string            `json:"virtio4,omitempty"`
	VirtIO5  string            `json:"virtio5,omitempty"`
	VirtIO6  string            `json:"virtio6,omitempty"`
	VirtIO7  string            `json:"virtio7,omitempty"`
	VirtIO8  string            `json:"virtio8,omitempty"`
	VirtIO9  string            `json:"virtio9,omitempty"`
	VirtIO10 string            `json:"virtio10,omitempty"`
	VirtIO11 string            `json:"virtio11,omitempty"`
	VirtIO12 string            `json:"virtio12,omitempty"`
	VirtIO13 string            `json:"virtio13,omitempty"`
	VirtIO14 string            `json:"virtio14,omitempty"`
	VirtIO15 string            `json:"virtio15,omitempty"`

	Unuseds map[string]string `json:"-"`
	Unused0 string            `json:"unused0,omitempty"`
	Unused1 string            `json:"unused1,omitempty"`
	Unused2 string            `json:"unused2,omitempty"`
	Unused3 string            `json:"unused3,omitempty"`
	Unused4 string            `json:"unused4,omitempty"`
	Unused5 string            `json:"unused5,omitempty"`
	Unused6 string            `json:"unused6,omitempty"`
	Unused7 string            `json:"unused7,omitempty"`
	Unused8 string            `json:"unused8,omitempty"`
	Unused9 string            `json:"unused9,omitempty"`

	// Network devices
	Nets map[string]string `json:"-"`
	Net0 string            `json:"net0,omitempty"`
	Net1 string            `json:"net1,omitempty"`
	Net2 string            `json:"net2,omitempty"`
	Net3 string            `json:"net3,omitempty"`
	Net4 string            `json:"net4,omitempty"`
	Net5 string            `json:"net5,omitempty"`
	Net6 string            `json:"net6,omitempty"`
	Net7 string            `json:"net7,omitempty"`
	Net8 string            `json:"net8,omitempty"`
	Net9 string            `json:"net9,omitempty"`

	// NUMA topology
	Numas map[string]string `json:"-"`
	Numa0 string            `json:"numa0,omitempty"`
	Numa1 string            `json:"numa1,omitempty"`
	Numa2 string            `json:"numa2,omitempty"`
	Numa3 string            `json:"numa3,omitempty"`
	Numa4 string            `json:"numa4,omitempty"`
	Numa5 string            `json:"numa5,omitempty"`
	Numa6 string            `json:"numa6,omitempty"`
	Numa7 string            `json:"numa7,omitempty"`
	Numa8 string            `json:"numa8,omitempty"`
	Numa9 string            `json:"numa9,omitempty"`

	// Host PCI devices
	HostPCIs map[string]string `json:"-"`
	HostPCI0 string            `json:"hostpci0,omitempty"`
	HostPCI1 string            `json:"hostpci1,omitempty"`
	HostPCI2 string            `json:"hostpci2,omitempty"`
	HostPCI3 string            `json:"hostpci3,omitempty"`
	HostPCI4 string            `json:"hostpci4,omitempty"`
	HostPCI5 string            `json:"hostpci5,omitempty"`
	HostPCI6 string            `json:"hostpci6,omitempty"`
	HostPCI7 string            `json:"hostpci7,omitempty"`
	HostPCI8 string            `json:"hostpci8,omitempty"`
	HostPCI9 string            `json:"hostpci9,omitempty"`

	// Serial devices
	Serials map[string]string `json:"-"`
	Serial0 string            `json:"serial0,omitempty"`
	Serial1 string            `json:"serial1,omitempty"`
	Serial2 string            `json:"serial2,omitempty"`
	Serial3 string            `json:"serial3,omitempty"`

	// USB devices
	USBs  map[string]string `json:"-"`
	USB0  string            `json:"usb0,omitempty"`
	USB1  string            `json:"usb1,omitempty"`
	USB2  string            `json:"usb2,omitempty"`
	USB3  string            `json:"usb3,omitempty"`
	USB4  string            `json:"usb4,omitempty"`
	USB5  string            `json:"usb5,omitempty"`
	USB6  string            `json:"usb6,omitempty"`
	USB7  string            `json:"usb7,omitempty"`
	USB8  string            `json:"usb8,omitempty"`
	USB9  string            `json:"usb9,omitempty"`
	USB10 string            `json:"usb10,omitempty"`
	USB11 string            `json:"usb11,omitempty"`
	USB12 string            `json:"usb12,omitempty"`
	USB13 string            `json:"usb13,omitempty"`
	USB14 string            `json:"usb14,omitempty"`

	// Parallel devices
	Parallels map[string]string `json:"-"`
	Parallel0 string            `json:"parallel0,omitempty"`
	Parallel1 string            `json:"parallel1,omitempty"`
	Parallel2 string            `json:"parallel2,omitempty"`

	// Cloud-init
	CIType       string `json:"citype,omitempty"`
	CIUser       string `json:"ciuser,omitempty"`
	CIPassword   string `json:"cipassword,omitempty"`
	Nameserver   string `json:"nameserver,omitempty"`
	Searchdomain string `json:"searchdomain,omitempty"`
	// SSHKeys must be encoded with EncodeSSHKeys — PVE's API validator
	// rejects loose url-encoding (e.g. '+' for spaces). See issue #144.
	SSHKeys  string `json:"sshkeys,omitempty"`
	CICustom string `json:"cicustom,omitempty"`
	CIUpgrade    int    `json:"ciupgrade,omitempty"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would skip package upgrade).

	// Cloud-init interfaces
	IPConfigs map[string]string `json:"-"`
	IPConfig0 string            `json:"ipconfig0,omitempty"`
	IPConfig1 string            `json:"ipconfig1,omitempty"`
	IPConfig2 string            `json:"ipconfig2,omitempty"`
	IPConfig3 string            `json:"ipconfig3,omitempty"`
	IPConfig4 string            `json:"ipconfig4,omitempty"`
	IPConfig5 string            `json:"ipconfig5,omitempty"`
	IPConfig6 string            `json:"ipconfig6,omitempty"`
	IPConfig7 string            `json:"ipconfig7,omitempty"`
	IPConfig8 string            `json:"ipconfig8,omitempty"`
	IPConfig9 string            `json:"ipconfig9,omitempty"`
}

// indexedDeviceMaps lists the JSON-key prefixes that get routed into a
// per-prefix helper map on VirtualMachineConfig. Proxmox accepts more indices
// for each device type than the explicit IDE0..VirtIO15 fields cover (e.g.
// net0..net31, scsi0..scsi30, unused0..unused255), so the maps are the
// authoritative source of truth and the explicit fields stay as a
// compatibility mirror for indices 0..9. Plan: mark the explicit fields
// // Deprecated in v0.7.x and drop them in a later release.
//
// Order does not matter; lookups are by exact prefix followed by a pure-digit
// suffix, which means "scsihw" and the bare "numa" scalar do not collide.
func (vmc *VirtualMachineConfig) indexedDeviceMaps() map[string]*map[string]string {
	return map[string]*map[string]string{
		"ide":      &vmc.IDEs,
		"scsi":     &vmc.SCSIs,
		"sata":     &vmc.SATAs,
		"virtio":   &vmc.VirtIOs,
		"unused":   &vmc.Unuseds,
		"net":      &vmc.Nets,
		"numa":     &vmc.Numas,
		"hostpci":  &vmc.HostPCIs,
		"serial":   &vmc.Serials,
		"usb":      &vmc.USBs,
		"parallel": &vmc.Parallels,
		"ipconfig": &vmc.IPConfigs,
	}
}

func (vmc *VirtualMachineConfig) UnmarshalJSON(data []byte) error {
	type tmpVirtualMachineConfig VirtualMachineConfig

	// create a struct and embed temporary alias of VirtualMachineConfig to avoid recursion
	// this will also populate the rest of the fields using the built in unmarshal function
	tmp := &struct {
		*tmpVirtualMachineConfig
	}{
		tmpVirtualMachineConfig: (*tmpVirtualMachineConfig)(vmc),
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Split the tags on TagSeparator and populate TagsSlice
	vmc.TagsSlice = strings.Split(vmc.Tags, TagSeperator)

	// Walk the raw JSON object once and route every "<prefix><digits>" key
	// into its target map. This captures indices Proxmox returns beyond the
	// explicit struct fields (e.g. net10..net31, scsi10..scsi30).
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	routes := vmc.indexedDeviceMaps()
	for k, v := range raw {
		prefix, ok := indexedDeviceKey(k)
		if !ok {
			continue
		}
		target, ok := routes[prefix]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			// non-string indexed value — skip rather than fail the whole config
			continue
		}
		if *target == nil {
			*target = make(map[string]string)
		}
		(*target)[k] = s
	}

	return nil
}

// indexedDeviceKey returns ("net", true) for "net10" and ("", false) for keys
// like "scsihw" or "numa" that share a prefix but lack a pure-digit suffix.
func indexedDeviceKey(k string) (prefix string, ok bool) {
	for i := 0; i < len(k); i++ {
		c := k[i]
		if c >= '0' && c <= '9' {
			if i == 0 {
				return "", false
			}
			// rest of the string must be all digits
			for j := i + 1; j < len(k); j++ {
				if k[j] < '0' || k[j] > '9' {
					return "", false
				}
			}
			return k[:i], true
		}
	}
	return "", false
}

// VirtualMachineFeature is the response payload of
// GET /nodes/{node}/qemu/{vmid}/feature. HasFeature reports whether the
// requested feature is available for the VM (and optional snapshot); Nodes
// lists the cluster nodes on which the feature is available.
type VirtualMachineFeature struct {
	HasFeature bool     `json:"hasFeature"`
	Nodes      []string `json:"nodes,omitempty"`
}

type VirtualMachineMigrateOptions struct {
	Target           string    `json:"target"`
	BWLimit          uint64    `json:"bwlimit,omitempty"` // FIXME(issue-199): PVE default = datacenter/storage migrate limit; use *uint64 so unset doesn't impose 0.
	Force            IntOrBool `json:"force,omitempty"`
	MigrationNetwork string    `json:"migration_network,omitempty"`
	MigrationType    string    `json:"migration_type,omitempty"`
	Online           IntOrBool `json:"online,omitempty"`
	TargetStorage    string    `json:"targetstorage,omitempty"`
	WithLocalDisks   IntOrBool `json:"with-local-disks,omitempty"`
}

type ContainerMigrateOptions struct {
	Target  string    `json:"target"`
	BWLimit uint64    `json:"bwlimit,omitempty"` // FIXME(issue-199): PVE default = datacenter/storage migrate limit; use *uint64 so unset doesn't impose 0.
	Online  IntOrBool `json:"online,omitempty"`
	Restart IntOrBool `json:"restart,omitempty"`
}

// ContainerDeleteOptions maps to the optional query parameters that
// DELETE /nodes/{node}/lxc/{vmid} accepts. A nil *ContainerDeleteOptions
// passed to Container.Delete is equivalent to all defaults.
type ContainerDeleteOptions struct {
	// Force destroys the container even if it is currently running.
	Force IntOrBool `json:"force,omitempty"`
	// Purge also removes the container from all related configurations
	// (backup jobs, replication jobs, HA), in addition to deleting it.
	Purge IntOrBool `json:"purge,omitempty"`
	// DestroyUnreferencedDisks also destroys disks across enabled storages
	// that match the VMID but are not referenced by the container's config.
	DestroyUnreferencedDisks IntOrBool `json:"destroy-unreferenced-disks,omitempty"`
}

// ContainerRemoteMigrateOptions configures POST /nodes/{node}/lxc/{vmid}/remote_migrate
// (cross-cluster migration). TargetEndpoint is the API-token bundle string PVE
// accepts ("apitoken=PVEAPIToken=... host=... fingerprint=..."); see the
// pvesh docs for the exact shape.
type ContainerRemoteMigrateOptions struct {
	TargetEndpoint string    `json:"target-endpoint"`
	TargetBridge   string    `json:"target-bridge"`            // "src=tgt,src2=tgt2" map
	TargetStorage  string    `json:"target-storage"`           // "src=tgt,src2=tgt2" map
	TargetVMID     int       `json:"target-vmid,omitempty"`
	BWLimit        uint64    `json:"bwlimit,omitempty"`
	Delete         IntOrBool `json:"delete,omitempty"`
	Online         IntOrBool `json:"online,omitempty"`
	Restart        IntOrBool `json:"restart,omitempty"`
	Timeout        uint64    `json:"timeout,omitempty"`
}

// ContainerPending describes a single staged config change returned by
// GET /nodes/{node}/lxc/{vmid}/pending. Value is the currently active value;
// Pending is the value queued for the next start. Delete is set when the key
// is queued for removal.
type ContainerPending struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value,omitempty"`
	Pending interface{} `json:"pending,omitempty"`
	Delete  int         `json:"delete,omitempty"`
}

// ContainerRRD is the response from GET /nodes/{node}/lxc/{vmid}/rrd. PVE
// renders a single PNG on the server and returns its on-disk filename;
// callers typically want RRDData instead for usable numeric series.
type ContainerRRD struct {
	Filename string `json:"filename"`
}

// VirtualMachineRRD is the response from GET /nodes/{node}/qemu/{vmid}/rrd.
// PVE renders a single PNG on the server (one datasource per call) and
// returns its on-disk filename; callers typically want RRDData instead for
// usable numeric series.
type VirtualMachineRRD struct {
	Filename string `json:"filename"`
}

// VirtualMachineRemoteMigrateOptions configures POST
// /nodes/{node}/qemu/{vmid}/remote_migrate (cross-cluster VM migration —
// flagged EXPERIMENTAL upstream). TargetEndpoint is the API-token bundle
// string PVE accepts ("apitoken=PVEAPIToken=... host=... fingerprint=...");
// see the pvesh docs for the exact shape. TargetBridge and TargetStorage are
// "src=tgt,src2=tgt2" pair-list maps; the special value "1" maps each source
// to itself.
type VirtualMachineRemoteMigrateOptions struct {
	TargetEndpoint string    `json:"target-endpoint"`
	TargetBridge   string    `json:"target-bridge"`
	TargetStorage  string    `json:"target-storage"`
	TargetVMID     int       `json:"target-vmid,omitempty"`
	BWLimit        uint64    `json:"bwlimit,omitempty"`
	Delete         IntOrBool `json:"delete,omitempty"`
	Online         IntOrBool `json:"online,omitempty"`
}

// VirtualMachineMigratePreconditionsLocalDisk describes one local disk
// surfaced by GET /nodes/{node}/qemu/{vmid}/migrate. Local disks block
// live migration unless WithLocalDisks is enabled on Migrate.
type VirtualMachineMigratePreconditionsLocalDisk struct {
	VolID    string `json:"volid"`
	Size     uint64 `json:"size,omitempty"`
	CDROM    bool   `json:"cdrom,omitempty"`
	IsUnused bool   `json:"is_unused,omitempty"`
}

// VirtualMachineMigratePreconditionsBlockingHAResource identifies a HA
// resource preventing the VM from migrating to a particular node.
type VirtualMachineMigratePreconditionsBlockingHAResource struct {
	SID   string `json:"sid"`
	Cause string `json:"cause"` // "node-affinity" or "resource-affinity"
}

// VirtualMachineMigratePreconditionsNotAllowedNodes carries the per-node
// reasons a target node is rejected for migration.
type VirtualMachineMigratePreconditionsNotAllowedNodes struct {
	UnavailableStorages  []string                                                `json:"unavailable_storages,omitempty"`
	BlockingHAResources  []*VirtualMachineMigratePreconditionsBlockingHAResource `json:"blocking-ha-resources,omitempty"`
}

// VirtualMachineMigratePreconditions is the response from
// GET /nodes/{node}/qemu/{vmid}/migrate — a dry-run summary of whether the
// VM can be migrated, which nodes accept it, and what local state would
// have to be moved along with it. Pre-flight only; no task is created.
type VirtualMachineMigratePreconditions struct {
	Running             bool                                                          `json:"running"`
	HasDBusVMState      bool                                                          `json:"has-dbus-vmstate"`
	AllowedNodes        []string                                                      `json:"allowed_nodes,omitempty"`
	NotAllowedNodes     map[string]*VirtualMachineMigratePreconditionsNotAllowedNodes `json:"not_allowed_nodes,omitempty"`
	LocalDisks          []*VirtualMachineMigratePreconditionsLocalDisk                `json:"local_disks,omitempty"`
	LocalResources      []string                                                      `json:"local_resources,omitempty"`
	MappedResources     []string                                                      `json:"mapped-resources,omitempty"`
	MappedResourceInfo  map[string]interface{}                                        `json:"mapped-resource-info,omitempty"`
	DependentHAResources []string                                                     `json:"dependent-ha-resources,omitempty"`
}

// SpiceProxy carries the SPICE connection info returned by /spiceproxy.
// The field names match the keys remote-viewer expects in its .vv config.
type SpiceProxy struct {
	Type             string `json:"type"`
	Host             string `json:"host"`
	Port             string `json:"port,omitempty"`
	Password         string `json:"password,omitempty"`
	Proxy            string `json:"proxy,omitempty"`
	Title            string `json:"title,omitempty"`
	TLSPort          string `json:"tls-port,omitempty"`
	CA               string `json:"ca,omitempty"`
	HostSubject      string `json:"host-subject,omitempty"`
	DeleteThisFile   string `json:"delete-this-file,omitempty"`
	SecureAttention  string `json:"secure-attention,omitempty"`
	ReleaseCursor    string `json:"release-cursor,omitempty"`
	ToggleFullscreen string `json:"toggle-fullscreen,omitempty"`
}

// FirewallLogEntry is one line from GET /firewall/log. PVE returns each entry
// as a [line-number, text] JSON tuple — the custom UnmarshalJSON below
// flattens that into named fields.
type FirewallLogEntry struct {
	LineNum int    `json:"n"`
	Text    string `json:"t"`
}

func (f *FirewallLogEntry) UnmarshalJSON(b []byte) error {
	// Tuple form (current PVE): [n, "text"]
	var tuple []interface{}
	if err := json.Unmarshal(b, &tuple); err == nil && len(tuple) == 2 {
		if n, ok := tuple[0].(float64); ok {
			f.LineNum = int(n)
		}
		if t, ok := tuple[1].(string); ok {
			f.Text = t
		}
		return nil
	}
	// Object fallback in case PVE ever switches shape.
	aux := struct {
		N int    `json:"n"`
		T string `json:"t"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	f.LineNum = aux.N
	f.Text = aux.T
	return nil
}

// FirewallRef is one entry from GET /firewall/refs — a referencable IPSet or
// alias visible at this container/VM's scope.
type FirewallRef struct {
	Type    string `json:"type"` // "alias" or "ipset"
	Name    string `json:"name"`
	Comment string `json:"comment,omitempty"`
}

// ContainerSnapshotUpdateOptions is the body for PUT /snapshot/{name}/config.
// PVE accepts only the description field on this endpoint.
type ContainerSnapshotUpdateOptions struct {
	Description string `json:"description,omitempty"`
}

type VirtualMachineCloneOptions struct {
	NewID       int    `json:"newid"`
	BWLimit     uint64 `json:"bwlimit,omitempty"` // FIXME(issue-199): PVE default = datacenter/storage clone limit; use *uint64 so unset doesn't impose 0.
	Description string `json:"description,omitempty"`
	Format      string `json:"format,omitempty"`
	Full        uint8  `json:"full,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default unset; 0/1 are the only meaningful values).
	Name        string `json:"name,omitempty"`
	Pool        string `json:"pool,omitempty"`
	SnapName    string `json:"snapname,omitempty"`
	Storage     string `json:"storage,omitempty"`
	Target      string `json:"target,omitempty"`
}

type VirtualMachineMoveDiskOptions struct {
	Disk         string `json:"disk"`
	BWLimit      uint64 `json:"bwlimit,omitempty"` // FIXME(issue-199): PVE default = datacenter/storage move limit; use *uint64 so unset doesn't impose 0.
	Delete       uint8  `json:"delete,omitempty"`  // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Digest       string `json:"digest,omitempty"`
	Format       string `json:"format,omitempty"`
	Storage      string `json:"storage,omitempty"`
	TargetDigest string `json:"target-digest,omitempty"`
	TargetDisk   string `json:"target-disk,omitempty"`
	TargetVMID   int    `json:"target-vmid,omitempty"`
}

type UPID string

type Tasks []*Task
type Task struct {
	client       *Client
	UPID         UPID
	ID           string
	Type         string
	User         string
	Status       string
	Node         string
	PID          uint64 `json:",omitempty"`
	PStart       uint64 `json:",omitempty"`
	Saved        string `json:",omitempty"`
	ExitStatus   string `json:",omitempty"`
	IsCompleted  bool
	IsRunning    bool
	IsFailed     bool
	IsSuccessful bool
	StartTime    time.Time     `json:"-"`
	EndTime      time.Time     `json:"-"`
	Duration     time.Duration `json:"-"`
}

func (t *Task) UnmarshalJSON(b []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}

	type TempTask Task
	var task TempTask
	if err := json.Unmarshal(b, &task); err != nil {
		return err
	}

	if starttime, ok := tmp["starttime"]; ok {
		task.StartTime = time.Unix(int64(starttime.(float64)), 0)
	}

	if endtime, ok := tmp["endtime"]; ok {
		task.EndTime = time.Unix(int64(endtime.(float64)), 0)
	}

	if !task.StartTime.IsZero() && !task.EndTime.IsZero() {
		task.Duration = task.EndTime.Sub(task.StartTime)
	}

	c := Task(task)
	return copier.Copy(t, &c)
}

type Log map[int]string

// line numbers in the response start a 1  but the start param indexes from 0 so converting to that
func (l *Log) UnmarshalJSON(b []byte) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	log := make(map[int]string, len(data))
	for _, row := range data {
		if n, ok := row["n"]; ok {
			if t, ok := row["t"]; ok {
				log[int(n.(float64))-1] = t.(string)
			}
		}
	}

	return copier.Copy(l, Log(log))
}

type Containers []*Container
type Container struct {
	client          *Client
	ContainerConfig *ContainerConfig

	CPUs    int
	MaxDisk uint64
	MaxMem  uint64
	MaxSwap uint64
	Name    string
	Node    string
	Status  string
	Tags    string
	Uptime  uint64
	VMID    StringOrUint64
}

type ContainerInterfaces []*ContainerInterface

type ContainerInterface struct {
	HWAddr string `json:"hwaddr,omitempty"`
	Name   string `json:"name,omitempty"`
	Inet   string `json:"inet,omitempty"`
	Inet6  string `json:"inet6,omitempty"`
}

type ContainerCloneOptions struct {
	NewID       int    `json:"newid"`
	BWLimit     uint64 `json:"bwlimit,omitempty"` // FIXME(issue-199): PVE default = datacenter/storage clone limit; use *uint64 so unset doesn't impose 0.
	Description string `json:"description,omitempty"`
	Full        uint8  `json:"full,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default unset; 0/1 are the only meaningful values).
	Hostname    string `json:"hostname,omitempty"`
	Pool        string `json:"pool,omitempty"`
	SnapName    string `json:"snapname,omitempty"`
	Storage     string `json:"storage,omitempty"`
	Target      string `json:"target,omitempty"`
}

type ContainerConfig struct {
	Arch         string            `json:"arch,omitempty"`     // FIXME(issue-199): PVE default "amd64"; use *string so unset doesn't override.
	CMode        string            `json:"cmode,omitempty"`    // FIXME(issue-199): PVE default "tty"; use *string so unset doesn't override.
	Console      IntOrBool         `json:"console,omitempty"`  // FIXME(issue-199): PVE default 1 (enabled); use *IntOrBool so unset doesn't disable.
	Cores        int               `json:"cores,omitempty"`
	CPULimit     int               `json:"cpulimit,omitempty"`
	CPUUnits     int               `json:"cpuunits,omitempty"` // FIXME(issue-199): PVE default 1024 (cgroup v1) / 100 (cgroup v2); use *int so unset doesn't override.
	Debug        IntOrBool         `json:"debug,omitempty"`
	Description  string            `json:"description,omitempty"`
	Devs         map[string]string `json:"-"` // internal helper for Dev0..9
	Dev0         string            `json:"dev0,omitempty"`
	Dev1         string            `json:"dev1,omitempty"`
	Dev2         string            `json:"dev2,omitempty"`
	Dev3         string            `json:"dev3,omitempty"`
	Dev4         string            `json:"dev4,omitempty"`
	Dev5         string            `json:"dev5,omitempty"`
	Dev6         string            `json:"dev6,omitempty"`
	Dev7         string            `json:"dev7,omitempty"`
	Dev8         string            `json:"dev8,omitempty"`
	Dev9         string            `json:"dev9,omitempty"`
	Digest       string            `json:"digest"`
	Features     string            `json:"features,omitempty"`
	HookScript   string            `json:"hookscript,omitempty"`
	LXC          [][]string        `json:"lxc,omitempty"`
	Hostname     string            `json:"hostname,omitempty"`
	Lock         string            `json:"lock,omitempty"`
	Memory       int               `json:"memory,omitempty"` // FIXME(issue-199): PVE default 512; use *int so unset doesn't override.
	Mps          map[string]string `json:"-"`                // internal helper for Mp0..9
	Mp0          string            `json:"mp0,omitempty"`
	Mp1          string            `json:"mp1,omitempty"`
	Mp2          string            `json:"mp2,omitempty"`
	Mp3          string            `json:"mp3,omitempty"`
	Mp4          string            `json:"mp4,omitempty"`
	Mp5          string            `json:"mp5,omitempty"`
	Mp6          string            `json:"mp6,omitempty"`
	Mp7          string            `json:"mp7,omitempty"`
	Mp8          string            `json:"mp8,omitempty"`
	Mp9          string            `json:"mp9,omitempty"`
	Nameserver   string            `json:"nameserver,omitempty"`
	Nets         map[string]string `json:"-"` // internal helper for Net0..9
	Net0         string            `json:"net0,omitempty"`
	Net1         string            `json:"net1,omitempty"`
	Net2         string            `json:"net2,omitempty"`
	Net3         string            `json:"net3,omitempty"`
	Net4         string            `json:"net4,omitempty"`
	Net5         string            `json:"net5,omitempty"`
	Net6         string            `json:"net6,omitempty"`
	Net7         string            `json:"net7,omitempty"`
	Net8         string            `json:"net8,omitempty"`
	Net9         string            `json:"net9,omitempty"`
	OnBoot       IntOrBool         `json:"onboot,omitempty"`
	OSType       string            `json:"ostype,omitempty"`
	Protection   IntOrBool         `json:"protection,omitempty"`
	RootFS       string            `json:"rootfs,omitempty"`
	SearchDomain string            `json:"searchdomain,omitempty"`
	Startup      string            `json:"startup,omitempty"`
	Swap         int               `json:"swap,omitempty"` // FIXME(issue-199): PVE default 512; use *int so unset doesn't override.
	TagsSlice    []string          `json:"-"`              // internal helper to manage tags easier
	Tags         string            `json:"tags,omitempty"`
	Template     IntOrBool         `json:"template,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	TTY          int               `json:"tty,omitempty"` // FIXME(issue-199): PVE default 2; use *int so unset doesn't override.
	Unprivileged IntOrBool         `json:"unprivileged,omitempty"`
	Unuseds      map[string]string `json:"-"` // internal helper
	Unused0      string            `json:"unused0,omitempty"`
	Unused1      string            `json:"unused1,omitempty"`
	Unused2      string            `json:"unused2,omitempty"`
	Unused3      string            `json:"unused3,omitempty"`
	Unused4      string            `json:"unused4,omitempty"`
	Unused5      string            `json:"unused5,omitempty"`
	Unused6      string            `json:"unused6,omitempty"`
	Unused7      string            `json:"unused7,omitempty"`
	Unused8      string            `json:"unused8,omitempty"`
	Unused9      string            `json:"unused9,omitempty"`
}

// indexedDeviceMaps mirrors VirtualMachineConfig.indexedDeviceMaps for the
// container side: dev0..dev255, mp0..mp255, net0..net31, unused0..unused255.
func (cc *ContainerConfig) indexedDeviceMaps() map[string]*map[string]string {
	return map[string]*map[string]string{
		"dev":    &cc.Devs,
		"mp":     &cc.Mps,
		"net":    &cc.Nets,
		"unused": &cc.Unuseds,
	}
}

func (cc *ContainerConfig) UnmarshalJSON(data []byte) error {
	type tmpContainerConfig ContainerConfig

	// create a struct and embed temporary alias of ContainerConfig to avoid recursion
	// this will also populate the rest of the fields using the built in unmarshal function
	tmp := &struct {
		*tmpContainerConfig
	}{
		tmpContainerConfig: (*tmpContainerConfig)(cc),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	// Split the tags on TagSeparator and populate TagsSlice
	cc.TagsSlice = strings.Split(cc.Tags, TagSeperator)

	// Walk the raw JSON object once and route every "<prefix><digits>" key
	// into its target map; covers indices beyond the explicit Net0..Net9 etc.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	routes := cc.indexedDeviceMaps()
	for k, v := range raw {
		prefix, ok := indexedDeviceKey(k)
		if !ok {
			continue
		}
		target, ok := routes[prefix]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			continue
		}
		if *target == nil {
			*target = make(map[string]string)
		}
		(*target)[k] = s
	}

	return nil
}

// ContainerOptions A key/value pair used to modify a container(LXC) config
// Refer to https://pve.proxmox.com/pve-docs/api-viewer/#/nodes/{node}/lxc/{vmid}/config for a list of valid values
type ContainerOptions []*ContainerOption
type ContainerOption struct {
	Name  string
	Value interface{}
}

type Appliances []*Appliance
type Appliance struct {
	client       *Client
	Node         string `json:",omitempty"`
	Os           string
	Source       string
	Type         string
	SHA512Sum    string
	Package      string
	Template     string
	Architecture string
	InfoPage     string
	Description  string
	ManageURL    string
	Version      string
	Section      string
	Headline     string
}

type Storages []*Storage
type Storage struct {
	client       *Client
	Node         string
	Name         string `json:"storage"`
	Enabled      int
	UsedFraction float64 `json:"used_fraction"`
	Active       int
	Content      string
	Shared       int
	Avail        uint64
	Type         string
	Used         uint64
	Total        uint64
	Storage      string `json:"-"` // Deprecated: Use Name instead. Excluded from JSON to prevent marshal/unmarshal issues.
}

// UnmarshalJSON implements custom unmarshaling for Storage to handle large values
// that may be returned as floats in scientific notation (e.g., values > 1PB)
func (s *Storage) UnmarshalJSON(b []byte) error {
	// Temporary struct to capture raw JSON with json.Number for numeric fields
	aux := &struct {
		Node         string      `json:"node,omitempty"`
		Name         string      `json:"storage,omitempty"`
		Enabled      json.Number `json:"enabled,omitempty"`
		UsedFraction json.Number `json:"used_fraction,omitempty"`
		Active       json.Number `json:"active,omitempty"`
		Content      string      `json:"content,omitempty"`
		Shared       json.Number `json:"shared,omitempty"`
		Avail        json.Number `json:"avail,omitempty"`
		Type         string      `json:"type,omitempty"`
		Used         json.Number `json:"used,omitempty"`
		Total        json.Number `json:"total,omitempty"`
	}{}

	// Decode with UseNumber to preserve precision
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&aux); err != nil {
		return err
	}

	// Copy string fields
	s.Node = aux.Node
	s.Name = aux.Name
	s.Storage = aux.Name // Storage field gets same value as Name
	s.Content = aux.Content
	s.Type = aux.Type

	// Convert json.Number values to appropriate types
	if aux.Enabled != "" {
		if val, err := aux.Enabled.Int64(); err == nil {
			s.Enabled = int(val)
		}
	}

	if aux.Active != "" {
		if val, err := aux.Active.Int64(); err == nil {
			s.Active = int(val)
		}
	}

	if aux.Shared != "" {
		if val, err := aux.Shared.Int64(); err == nil {
			s.Shared = int(val)
		}
	}

	if aux.Avail != "" {
		// Try int64 first, then fall back to float64 for scientific notation
		if val, err := aux.Avail.Int64(); err == nil {
			s.Avail = uint64(val)
		} else if val, err := aux.Avail.Float64(); err == nil {
			s.Avail = uint64(val)
		}
	}

	if aux.Used != "" {
		// Try int64 first, then fall back to float64 for scientific notation
		if val, err := aux.Used.Int64(); err == nil {
			s.Used = uint64(val)
		} else if val, err := aux.Used.Float64(); err == nil {
			s.Used = uint64(val)
		}
	}

	if aux.Total != "" {
		// Try int64 first, then fall back to float64 for scientific notation
		if val, err := aux.Total.Int64(); err == nil {
			s.Total = uint64(val)
		} else if val, err := aux.Total.Float64(); err == nil {
			s.Total = uint64(val)
		}
	}

	if aux.UsedFraction != "" {
		if val, err := aux.UsedFraction.Float64(); err == nil {
			s.UsedFraction = val
		}
	}

	return nil
}

// UnmarshalJSON implements custom unmarshaling for Storages slice
func (storages *Storages) UnmarshalJSON(b []byte) error {
	var items []*Storage
	if err := json.Unmarshal(b, &items); err != nil {
		return err
	}
	*storages = items
	return nil
}

type ClusterBackups []*ClusterBackup

// ClusterBackup is a single configured cluster-wide backup schedule
// (a vzdump job). See https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/backup
type ClusterBackup struct {
	client *Client

	ID               string    `json:"id,omitempty"`
	Schedule         string    `json:"schedule,omitempty"`
	Enabled          IntOrBool `json:"enabled,omitempty"`
	RepeatMissed     IntOrBool `json:"repeat-missed,omitempty"`
	All              IntOrBool `json:"all,omitempty"`
	NotesTemplate    string    `json:"notes-template,omitempty"`
	MailNotification string    `json:"mailnotification,omitempty"`
	MailTo           string    `json:"mailto,omitempty"`
	Mode             string    `json:"mode,omitempty"`
	Type             string    `json:"type,omitempty"`
	NextRun          uint64    `json:"next-run,omitempty"`
	Storage          string    `json:"storage,omitempty"`
	VMID             string    `json:"vmid,omitempty"`
	Exclude          string    `json:"exclude,omitempty"`
	Node             string    `json:"node,omitempty"`
	Pool             string    `json:"pool,omitempty"`
	BwLimit          uint64    `json:"bwlimit,omitempty"`
	Comment          string    `json:"comment,omitempty"`
	PruneBackups     string    `json:"prune-backups,omitempty"`
}

// ClusterBackupOptions is the request body for POST /cluster/backup
// (create) and PUT /cluster/backup/{id} (update). All fields are optional;
// see the PVE API docs for semantics.
type ClusterBackupOptions struct {
	All                bool   `json:"all,omitempty"`
	BwLimit            uint64 `json:"bwlimit,omitempty"`
	Comment            string `json:"comment,omitempty"`
	Compress           string `json:"compress,omitempty"`
	Dow                string `json:"dow,omitempty"`
	DumpDir            string `json:"dumpdir,omitempty"`
	Enabled            bool   `json:"enabled,omitempty"`
	Exclude            string `json:"exclude,omitempty"`
	ExcludePath        string `json:"exclude-path,omitempty"`
	ID                 string `json:"id,omitempty"`
	IoNice             uint   `json:"ionice,omitempty"`
	LockWait           uint   `json:"lockwait,omitempty"`
	MailNotification   string `json:"mailnotification,omitempty"`
	MailTo             string `json:"mailto,omitempty"`
	MaxFiles           uint   `json:"maxfiles,omitempty"`
	Mode               string `json:"mode,omitempty"`
	Node               string `json:"node,omitempty"`
	NotesTemplate      string `json:"notes-template,omitempty"`
	NotificationMode   string `json:"notification-mode,omitempty"`
	NotificationPolicy string `json:"notification-policy,omitempty"`
	NotificationTarget string `json:"notification-target,omitempty"`
	Performance        string `json:"performance,omitempty"`
	Pigz               int    `json:"pigz,omitempty"`
	Pool               string `json:"pool,omitempty"`
	Protected          bool   `json:"protected,omitempty"`
	PruneBackups       string `json:"prune-backups,omitempty"`
	Quiet              bool   `json:"quiet,omitempty"`
	Remove             bool   `json:"remove,omitempty"`
	RepeatMissed       bool   `json:"repeat-missed,omitempty"`
	Schedule           string `json:"schedule,omitempty"`
	Script             string `json:"script,omitempty"`
	StdExcludes        bool   `json:"stdexcludes,omitempty"`
	Stop               bool   `json:"stop,omitempty"`
	StopWait           uint   `json:"stopwait,omitempty"`
	Storage            string `json:"storage,omitempty"`
	TmpDir             string `json:"tmpdir,omitempty"`
	VMID               string `json:"vmid,omitempty"`
	Zstd               uint   `json:"zstd,omitempty"`
}

type ClusterStorages []*ClusterStorage

type ClusterStorage struct {
	client   *Client
	Content  string
	Digest   string
	Storage  string
	Type     string
	Shared   int    `json:",omitempty"`
	Nodes    string `json:",omitempty"`
	Thinpool string `json:",omitempty"`
	Path     string `json:",omitempty"`
	VgName   string `json:",omitempty"`
}

type ClusterStorageOptions struct {
	Name  string
	Value string
}
type Volume interface {
	Delete() error
}

type ISOs []*ISO
type ISO struct{ Content }

type VzTmpls []*VzTmpl
type VzTmpl struct{ Content }

type Imports []*Import
type Import struct{ Content }

type Backups []*Backup
type Backup struct{ Content }

type Content struct {
	client  *Client
	URL     string
	Node    string
	Storage string `json:",omitempty"`
	Content string `json:",omitempty"`
	VolID   string `json:",omitempty"`
	CTime   uint64 `json:",omitempty"`
	Format  string
	Size    StringOrUint64
	Used    StringOrUint64 `json:",omitempty"`
	Path    string         `json:",omitempty"`
	Notes   string         `json:",omitempty"`
}

type IsTemplate bool

func (it *IsTemplate) UnmarshalJSON(b []byte) error {
	*it = true
	if string(b) == "\"\"" {
		*it = false
	}

	return nil
}

type StringOrInt int

func (d *StringOrInt) UnmarshalJSON(b []byte) error {
	str := strings.ReplaceAll(string(b), "\"", "")
	// Empty string and JSON null both yield the zero value. Proxmox returns
	// null for fields that are simply absent on the resource (e.g. PID on a
	// stopped VM template — see issue #198).
	if str == "" || str == "null" {
		*d = StringOrInt(0)
		return nil
	}

	if !isFloat.MatchString(str) {
		return fmt.Errorf("failed to match %s: %s", isFloat.String(), str)
	}

	parsed, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*d = StringOrInt(math.Trunc(parsed)) // truncate to make an int
	return nil
}

type StringOrUint64 uint64

func (d *StringOrUint64) UnmarshalJSON(b []byte) error {
	str := strings.ReplaceAll(string(b), "\"", "")
	if str == "" || str == "null" {
		*d = StringOrUint64(0)
		return nil
	}

	if !isFloat.MatchString(str) {
		return fmt.Errorf("failed to match %s: %s", isFloat.String(), str)
	}

	parsed, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	*d = StringOrUint64(math.Trunc(parsed)) // truncate to make an int

	return nil
}

type StringOrFloat64 float64

func (d *StringOrFloat64) UnmarshalJSON(b []byte) error {
	str := strings.ReplaceAll(string(b), "\"", "")
	if str == "" || str == "null" {
		*d = StringOrFloat64(0)
		return nil
	}

	if !isFloat.MatchString(str) {
		return fmt.Errorf("failed to match %s: %s", isFloat.String(), str)
	}

	parsed, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}
	*d = StringOrFloat64(parsed)
	return nil
}

type IntOrBool bool

func (b *IntOrBool) UnmarshalJSON(i []byte) error {
	parsed, err := strconv.ParseBool(string(i))
	if err != nil {
		return err
	}
	*b = IntOrBool(parsed)
	return nil
}

func (b *IntOrBool) MarshalJSON() ([]byte, error) {
	if *b {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

type NodeNetworks []*NodeNetwork
type NodeNetwork struct {
	client  *Client
	Node    string `json:"-"`
	NodeAPI *Node  `json:"-"`

	Iface     string `json:"iface,omitempty"`
	Autostart int    `json:"autostart,omitempty"`

	CIDR               string `json:"cidr,omitempty"`
	CIDR6              string `json:"cidr6,omitempty"`
	Gateway            string `json:"gateway,omitempty"`
	Gateway6           string `json:"gateway6,omitempty"`
	MTU                string `json:"mtu,omitempty"`
	Netmask            string `json:"netmask,omitempty"`
	Netmask6           string `json:"netmask6,omitempty"`
	VLANID             string `json:"vlan-id,omitempty"`
	VLANRawDevice      string `json:"vlan-raw-device,omitempty"`
	BridgeVLANAware    int    `json:"bridge_vlan_aware,omitempty"`
	BridgePorts        string `json:"bridge_ports,omitempty"`
	BridgeStp          string `json:"bridge_stp,omitempty"` // not in current docs, deprecated?
	BridgeFd           string `json:"bridge_fd,omitempty"`  // not in current docs, deprecated?
	Comments           string `json:"comments,omitempty"`
	Comments6          string `json:"comments6,omitempty"`
	BondPrimary        string `json:"bond-primary,omitempty"`
	BondMode           string `json:"bond_mode,omitempty"`
	BondXmit           string `json:"bond_xmit,omitempty"`
	BondXmitHashPolicy string `json:"bond_xmit_hash_policy,omitempty"`

	OVSBonds   string `json:"ovs_bonds,omitempty"`
	OVSBridge  string `json:"ovs_bridge,omitempty"`
	OVSOptions string `json:"ovs_options,omitempty"`
	OVSPorts   string `json:"ovs_ports,omitempty"`
	OVSTags    string `json:"ovs_tag,omitempty"`

	Slaves   string      `json:"slaves,omitempty"`
	Address  string      `json:"address,omitempty"`
	Address6 string      `json:"address6,omitempty"`
	Type     string      `json:"type,omitempty"`
	Active   StringOrInt `json:"active,omitempty"`
	Method   string      `json:"method,omitempty"`
	Method6  string      `json:"method6,omitempty"`
	Priority int         `json:"priority,omitempty"`
}

type AgentNetworkIPAddress struct {
	IPAddressType string `json:"ip-address-type"` // ipv4 ipv6
	IPAddress     string `json:"ip-address"`
	Prefix        int    `json:"prefix"`
	MacAddress    string `json:"mac-address"`
}

type AgentHostName struct {
	HostName string `json:"host-name"`
}

type AgentNetworkIface struct {
	Name            string                   `json:"name"`
	HardwareAddress string                   `json:"hardware-address"`
	IPAddresses     []*AgentNetworkIPAddress `json:"ip-addresses"`
}

type AgentOsInfo struct {
	Version       string `json:"version"`
	VersionID     string `json:"version-id"`
	ID            string `json:"id"`
	Machine       string `json:"machine"`
	PrettyName    string `json:"pretty-name"`
	Name          string `json:"name"`
	KernelRelease string `json:"kernel-release"`
	KernelVersion string `json:"kernel-version"`
}

type AgentExecStatus struct {
	Exited       int       `json:"exited"`
	ErrData      string    `json:"err-data"`
	ErrTruncated bool      `json:"err-truncated"`
	ExitCode     int       `json:"exitcode"`
	OutData      string    `json:"out-data"`
	OutTruncated IntOrBool `json:"out-truncated"`
	Signal       bool      `json:"signal"`
}

// AgentFileRead is the response from /agent/file-read. PVE returns the file
// body inline alongside a truncation flag — no `result` envelope here, unlike
// most other agent endpoints.
type AgentFileRead struct {
	Content   string    `json:"content"`
	Truncated IntOrBool `json:"truncated,omitempty"`
}

// AgentFsInfo mirrors qga's "guest-get-fsinfo" filesystem entry. Each
// element of AgentGetFsInfo.Result describes one mounted filesystem inside
// the guest.
type AgentFsInfo struct {
	Name           string                  `json:"name"`
	Mountpoint     string                  `json:"mountpoint"`
	Type           string                  `json:"type"`
	UsedBytes      uint64                  `json:"used-bytes,omitempty"`
	TotalBytes     uint64                  `json:"total-bytes,omitempty"`
	Disk           []*AgentFsInfoDisk      `json:"disk,omitempty"`
}

type AgentFsInfoDisk struct {
	Serial  string             `json:"serial,omitempty"`
	BusType string             `json:"bus-type,omitempty"`
	Bus     int                `json:"bus,omitempty"`
	Unit    int                `json:"unit,omitempty"`
	Target  int                `json:"target,omitempty"`
	PciController *AgentPciCtrl `json:"pci-controller,omitempty"`
	Dev     string             `json:"dev,omitempty"`
}

type AgentPciCtrl struct {
	Domain   int `json:"domain"`
	Bus      int `json:"bus"`
	Slot     int `json:"slot"`
	Function int `json:"function"`
}

// AgentTime represents the guest's wall-clock time in nanoseconds since
// epoch, as returned by qga's guest-get-time.
type AgentTime int64

// AgentUser describes one logged-in user from qga's guest-get-users. The
// LoginTime field is unix-epoch seconds with sub-second precision.
type AgentUser struct {
	User      string  `json:"user"`
	Domain    string  `json:"domain,omitempty"`
	LoginTime float64 `json:"login-time"`
}

// AgentVCPU represents one logical CPU from qga's guest-get-vcpus. PVE
// passes the QGA payload through verbatim.
type AgentVCPU struct {
	LogicalID int  `json:"logical-id"`
	Online    bool `json:"online"`
	CanOffline bool `json:"can-offline,omitempty"`
}

// AgentInfo describes the guest-agent itself: version + supported commands.
type AgentInfo struct {
	Version           string              `json:"version"`
	SupportedCommands []*AgentCommandInfo `json:"supported_commands,omitempty"`
}

type AgentCommandInfo struct {
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	SuccessResponse bool   `json:"success-response"`
}

// AgentMemoryBlock describes one hot-pluggable memory block as reported by
// qga's guest-get-memory-blocks.
type AgentMemoryBlock struct {
	PhysIndex   int  `json:"phys-index"`
	Online      bool `json:"online"`
	CanOffline  bool `json:"can-offline,omitempty"`
}

// AgentMemoryBlockInfo is the response payload from qga's
// guest-get-memory-block-info — currently just the per-block size in bytes.
type AgentMemoryBlockInfo struct {
	Size uint64 `json:"size"`
}

// AgentCommandIndexEntry is one entry in the GET /agent command index. PVE
// only documents `{}` items, but exposes the subroute as the link's "name",
// so we accept it as an open struct and surface the name when present.
type AgentCommandIndexEntry struct {
	Name string `json:"name,omitempty"`
}

// AgentFsfreezeStatus is the freeze state string ("thawed" or "frozen")
// returned by qga's guest-fsfreeze-status.
type AgentFsfreezeStatus string

type FirewallSecurityGroup struct {
	client  *Client
	Group   string          `json:"group,omitempty"`
	Comment string          `json:"comment,omitempty"`
	Rules   []*FirewallRule `json:"rules,omitempty"`
}
type FirewallRule struct {
	Type     string `json:"type,omitempty"`
	Action   string `json:"action,omitempty"`
	Pos      int    `json:"pos,omitempty"`
	Comment  string `json:"comment,omitempty"`
	Dest     string `json:"dest,omitempty"`
	Dport    string `json:"dport,omitempty"`
	Enable   int    `json:"enable,omitempty"`
	IcmpType string `json:"icmp_type,omitempty"`
	Iface    string `json:"iface,omitempty"`
	Log      string `json:"log,omitempty"`
	Macro    string `json:"macro,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Source   string `json:"source,omitempty"`
	Sport    string `json:"sport,omitempty"`
}

func (r *FirewallRule) IsEnable() bool {
	return r.Enable == 1
}

// PVE's three-gate firewall design: cluster + node + VM. Node-level ships
// enabled by default (Enable=1) so flipping the cluster gate activates node
// rules; per-VM stays opt-in (FirewallVirtualMachineOption.Enable defaults to 0).
// The asymmetric defaults are intentional — see PVE wiki "Proxmox VE Firewall".
type FirewallNodeOption struct {
	Enable                           bool   `json:"enable,omitempty"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would silently disable an already-enabled node firewall).
	LogLevelIn                       string `json:"log_level_in,omitempty"`
	LogLevelOut                      string `json:"log_level_out,omitempty"`
	LogNfConntrack                   bool   `json:"log_nf_conntrack,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	Ntp                              bool   `json:"ntp,omitempty"`
	NFConntrackAllowInvalid          bool   `json:"nf_conntrack_allow_invalid,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	NFConntrackMax                   int    `json:"nf_conntrack_max,omitempty"`                       // FIXME(issue-199): PVE default 262144; use *int so unset doesn't override.
	NFConntrackTCPTimeoutEstablished int    `json:"nf_conntrack_tcp_timeout_established,omitempty"`   // FIXME(issue-199): PVE default 432000; use *int so unset doesn't override.
	NFConntrackTCPTimeoutSynRecv     int    `json:"nf_conntrack_tcp_timeout_syn_recv,omitempty"`      // FIXME(issue-199): PVE default 60; use *int so unset doesn't override.
	Nosmurfs                         bool   `json:"nosmurfs,omitempty"`           // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	ProtectionSynflood               bool   `json:"protection_synflood,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	ProtectionSynfloodBurst          int    `json:"protection_synflood_burst,omitempty"` // FIXME(issue-199): PVE default 1000; use *int so unset doesn't override.
	ProtectionSynfloodRate           int    `json:"protection_synflood_rate,omitempty"`  // FIXME(issue-199): PVE default 200; use *int so unset doesn't override.
	SmurfLogLevel                    string `json:"smurf_log_level,omitempty"`
	TCPFlagsLogLevel                 string `json:"tcp_flags_log_level,omitempty"`
	TCPflags                         bool   `json:"tcpflags,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
}

// Per-VM firewall is opt-in (Enable defaults to 0) by design, in contrast to
// FirewallNodeOption which ships enabled. See the doc comment on
// FirewallNodeOption for the three-gate model.
type FirewallVirtualMachineOption struct {
	Enable      bool   `json:"enable,omitempty"`   // FIXME(issue-178): schema "boolean"; use IntOrBool (default 0 — VM firewall is opt-in by design, no pointer needed).
	Dhcp        bool   `json:"dhcp,omitempty"`     // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	Ipfilter    bool   `json:"ipfilter,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	LogLevelIn  string `json:"log_level_in,omitempty"`
	LogLevelOut string `json:"log_level_out,omitempty"`
	Macfilter   bool   `json:"macfilter,omitempty"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would disable MAC filtering).
	Ntp         bool   `json:"ntp,omitempty"`
	PolicyIn    string `json:"policy_in,omitempty"`
	PolicyOut   string `json:"policy_out,omitempty"`
	Radv        bool   `json:"radv,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
}

type Snapshot struct {
	Name        string
	Vmstate     int
	Description string
	Snaptime    int64
	Parent      string
	Snapstate   string
}

type Pools []*Pool
type Pool struct {
	client  *Client
	PoolID  string            `json:"poolid,omitempty"`
	Comment string            `json:"comment,omitempty"`
	Members []ClusterResource `json:"members,omitempty"`
}

type PoolUpdateOption struct {
	Comment string `json:"comment,omitempty"`
	// Delete objects rather than adding them
	Delete bool `json:"delete,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	// AllowMove permits adding a guest that already belongs to another pool;
	// the guest is silently moved instead of the request being rejected.
	AllowMove bool `json:"allow-move,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	// Comma separated lists of Storage names to add/delete to the pool
	Storage string `json:"storage,omitempty"`
	// Comma separated lists of Virtual Machine IDs to add/delete to the pool
	VirtualMachines string `json:"vms,omitempty"`
}

type DomainType string

const (
	DomainTypeAD     = DomainType("ad")
	DomainTypeLDAP   = DomainType("ldap")
	DomainTypeOpenID = DomainType("openid")
	DomainTypePam    = DomainType("pam")
	DomainTypePVE    = DomainType("pve")
)

type Domains []*Domain
type Domain struct {
	client *Client
	Realm  string `json:",omitempty"`
	Type   string `json:",omitempty"`

	// options https://pve.proxmox.com/pve-docs/api-viewer/#/access/domains
	ACRValues      string    `json:"acr-values,omitempty"`
	AutoCreate     IntOrBool `json:"autocreate,omitempty"`
	BaseDN         string    `json:"base_dn,omitempty"`
	BindDN         string    `json:"bind_dn,omitempty"`
	CAPath         string    `json:"capath,omitempty"`
	CaseSensitive  IntOrBool `json:"case-sensitive,omitempty"`
	Cert           string    `json:"cert,omitempty"`
	CertKey        string    `json:"certkey,omitempty"`
	ClientID       string    `json:"client-id,omitempty"`
	ClientKey      string    `json:"client-key,omitempty"`
	Comment        string    `json:"comment,omitempty"`
	Default        IntOrBool `json:"default,omitempty"`
	DeleteList     string    `json:"delete,omitempty"` // a list of settings you want to delete?
	Digest         string    `json:"digest,omitempty"`
	Domain         string    `json:"domain,omitempty"`
	Filter         string    `json:"filter,omitempty"`
	GroupClasses   string    `json:"group_classes,omitempty"`
	GroupDN        string    `json:"group_dn,omitempty"`
	GroupFilter    string    `json:"group_filter,omitempty"`
	GroupName      string    `json:"group_name,omitempty"`
	IssuerURL      string    `json:"issuer-url,omitempty"`
	Mode           string    `json:"mode,omitempty"` // ldap, ldaps,ldap+starttls
	Password       string    `json:"password,omitempty"`
	Port           int       `json:"port,omitempty"`
	Prompt         string    `json:"prompt,omitempty"`
	Scopes         string    `json:"scopes,omitempty"`
	Secure         IntOrBool `json:"secure,omitempty"`
	Server1        string    `json:"server1,omitempty"`
	Server2        string    `json:"server2,omitempty"`
	SSLVersion     string    `json:"sslversion,omitempty"`
	SyncDefaults   string    `json:"sync-defaults,omitempty"`
	SyncAttributes string    `json:"sync_attributes,omitempty"`
	TFA            string    `json:"tfa,omitempty"`
	UserAttr       string    `json:"user_attr,omitempty"`
	UserClasses    string    `json:"user_classes,omitempty"`
	Verify         IntOrBool `json:"verify"`
}

// DomainSyncOptions see details https://pve.proxmox.com/pve-docs/api-viewer/#/access/domains/{realm}/sync
type DomainSyncOptions struct {
	DryRun         bool   `json:"dry-run,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	EnableNew      bool   `json:"enable-new,omitempty"`      // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs (unset would disable newly synced users).
	RemoveVanished string `json:"remove-vanished,omitempty"` // FIXME(issue-199): PVE default "none"; use *string so unset doesn't override.
	Scope          string `json:"scope,omitempty"` // users, groups, both
}

type Groups []*Group
type Group struct {
	client  *Client
	GroupID string   `json:"groupid,omitempty"`
	Comment string   `json:"comment,omitempty"`
	Users   string   `json:"users,omitempty"`   // only populated via Groups lister
	Members []string `json:"members,omitempty"` // only populated via Group read
}

type Users []*User
type User struct {
	client         *Client
	UserID         string           `json:"userid,omitempty"`
	Comment        string           `json:"comment,omitempty"`
	Email          string           `json:"email,omitempty"`
	Enable         IntOrBool        `json:"enable"`
	Expire         int              `json:"expire,omitempty"`
	Firstname      string           `json:"firstname,omitempty"`
	Lastname       string           `json:"lastname,omitempty"`
	Groups         []string         `json:"groups,omitempty"`
	Keys           string           `json:"keys,omitempty"`
	Tokens         map[string]Token `json:"tokens,omitempty"`
	RealmType      string           `json:"realm-type,omitempty"`
	TFALockedUntil string           `json:"tfa-locked-until,omitempty"`
	TOTPLocked     IntOrBool        `json:"totp-locked,omitempty"`
}

type UserOptions struct {
	Append    IntOrBool `json:"append,omitempty"`
	Comment   string    `json:"comment,omitempty"`
	Email     string    `json:"email,omitempty"`
	Enable    IntOrBool `json:"enable"` // FIXME(issue-199): PVE default 1, schema "boolean"; type already correct — use *IntOrBool so unset doesn't disable accounts.
	Expire    int       `json:"expire,omitempty"`
	Firstname string    `json:"firstname,omitempty"`
	Groups    []string  `json:"groups,omitempty"`
	Keys      string    `json:"keys,omitempty"`
	Lastname  string    `json:"lastname,omitempty"`
}

type Tokens []*Token
type Token struct {
	TokenID string    `json:"tokenid,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Expire  int       `json:"expire,omitempty"`
	Privsep IntOrBool `json:"privsep"`
}

type Roles []*Role
type Role struct {
	client  *Client
	RoleID  string    `json:"roleid,omitempty"`
	Privs   string    `json:"privs,omitempty"`
	Special IntOrBool `json:"special,omitempty"`
}

type ACLs []*ACL
type ACL struct {
	Path      string    `json:"path,omitempty"`
	RoleID    string    `json:"roleid,omitempty"`
	Type      string    `json:"type,omitempty"`
	UGID      string    `json:"ugid,omitempty"`
	Propagate IntOrBool `json:"propagate,omitempty"`
}

type ACLOptions struct {
	Path      string    `json:"path,omitempty"`
	Roles     string    `json:"roles,omitempty"` // comma separated list of roles
	Groups    string    `json:"groups,omitempty"`
	Users     string    `json:"users,omitempty"`
	Tokens    string    `json:"tokens,omitempty"`
	Propagate IntOrBool `json:"propagate"`        // Default is true, omitempty would never send false
	Delete    IntOrBool `json:"delete,omitempty"` // true to delete the ACL
}

type StorageDownloadURLOptions struct {
	Content            string    `json:"content,omitempty"`
	Filename           string    `json:"filename,omitempty"`
	Node               string    `json:"node,omitempty"`
	Storage            string    `json:"storage,omitempty"`
	URL                string    `json:"url,omitempty"`
	Checksum           string    `json:"checksum,omitempty"`
	ChecksumAlgorithm  string    `json:"checksum-algorithm,omitempty"`
	Compression        string    `json:"compression,omitempty"`
	VerifyCertificates IntOrBool `json:"verify-certificates"` // FIXME(issue-199): PVE default 1, schema "boolean"; type already correct — use *IntOrBool so unset doesn't silently disable certificate verification.
}

type StorageContent struct {
	Format       string         `json:"format,omitempty"`
	Size         uint64         `json:"size,omitempty"`
	Volid        string         `json:"volid,omitempty"`
	Ctime        StringOrUint64 `json:"ctime,omitempty"`
	Encryption   string         `json:"encryption,omitempty"`
	Notes        string         `json:"notes,omitempty"`
	Parent       string         `json:"parent,omitempty"`
	Protection   IntOrBool      `json:"protection,omitempty"`
	Used         uint64         `json:"used,omitempty"`
	Verification string         `json:"verification,omitempty"`
	VMID         uint64         `json:"vmid,omitempty"`
}

// StoragePruneBackupsOptions filters which backups PreviewPruneBackups and
// PruneBackups operate on. The zero value means "use the storage's configured
// retention spec and apply to every backup".
type StoragePruneBackupsOptions struct {
	// PruneBackups overrides the storage's configured retention spec for this
	// call only. Example: "keep-last=3,keep-monthly=4". Empty uses the storage default.
	PruneBackups string
	// Type filters by guest type: "qemu" or "lxc". Empty considers both.
	Type string
	// VMID filters to a single guest. Zero considers all guests.
	VMID uint64
}

// PruneBackupItem is one row in the dryrun listing returned by
// Storage.PreviewPruneBackups. Mark indicates what PruneBackups would do with
// this volume: "keep", "remove", "protected" (retained by a protection flag),
// or "renamed" (retained because its name doesn't match the standard scheme).
type PruneBackupItem struct {
	Volid string         `json:"volid"`
	Ctime StringOrUint64 `json:"ctime"`
	Mark  string         `json:"mark"`
	Type  string         `json:"type"`
	VMID  uint64         `json:"vmid,omitempty"`
}

// ImportMetadata is the result of Storage.ImportMetadata for an external disk
// volume on an "import"-capable storage (e.g. an ESXi-imported guest). It
// describes how Proxmox interprets the source and supplies ready-to-use
// arguments for creating a guest from it.
type ImportMetadata struct {
	Type       string                  `json:"type"`
	Source     string                  `json:"source"`
	CreateArgs map[string]interface{}  `json:"create-args"`
	Disks      map[string]string       `json:"disks,omitempty"`
	Net        map[string]interface{}  `json:"net,omitempty"`
	Warnings   []ImportMetadataWarning `json:"warnings,omitempty"`
}

type ImportMetadataWarning struct {
	Type  string `json:"type"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// NodeDNS represents the resolver configuration for a single node, as
// returned by GET /nodes/{node}/dns and accepted by PUT /nodes/{node}/dns.
// Search is required on update; the three DNS slots are individually optional.
type NodeDNS struct {
	Search string `json:"search,omitempty"`
	DNS1   string `json:"dns1,omitempty"`
	DNS2   string `json:"dns2,omitempty"`
	DNS3   string `json:"dns3,omitempty"`
}

type NodeCertificates []*NodeCertificate

type NodeCertificate struct {
	Filename      string   `json:"filename,omitempty"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	Issuer        string   `json:"issuer,omitempty"`
	NotAfter      string   `json:"not-after,omitempty"`
	NotBefore     string   `json:"not-before,omitempty"`
	Pem           string   `json:"pem,omitempty"`
	PublicKeyBits int      `json:"public-key-bits,omitempty"`
	PublicKeyType string   `json:"public-key-type,omitempty"`
	San           []string `json:"san,omitempty"`
	Subject       string   `json:"subject,omitempty"`
}

type CustomCertificate struct {
	Certificates string `json:"certificates,omitempty"` // PEM encoded certificate (chain)
	Force        bool   `json:"force,omitempty"`        // overwrite existing certificate
	Key          string `json:"key,omitempty"`          // PEM encoded private key
	Restart      bool   `json:"restart,omitempty"`      // restart pveproxy
}

type NewUser struct {
	UserID    string   `json:"userid"`
	Comment   string   `json:"comment,omitempty"`
	Email     string   `json:"email,omitempty"`
	Enable    bool     `json:"enable"`
	Expire    int      `json:"expire,omitempty"`
	Firstname string   `json:"firstname,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Keys      []string `json:"keys,omitempty"`
	Lastname  string   `json:"lastname,omitempty"`
	Password  string   `json:"password,omitempty"`
}

type TFA struct {
	Realm string   `json:"realm,omitempty"`
	Types []string `json:"types,omitempty"`
	User  string   `json:"user,omitempty"`
}

type NewAPIToken struct {
	FullTokenID string      `json:"full-tokenid,omitempty"`
	Info        interface{} `json:"info,omitempty"`
	Value       string      `json:"value,omitempty"`
}

type VNCProxyOptions struct {
	Websocket string `json:"websocket,omitempty"`
	Height    int    `json:"height,omitempty"`
	Width     int    `json:"width,omitempty"`
}

type ContainerSnapshot struct {
	Description          string `json:"description,omitempty"`
	Name                 string `json:"snapname,omitempty"`
	Parent               string `json:"parent,omitempty"`
	SnapshotCreationTime int64  `json:"snaptime,omitempty"`
}

type Firewall struct {
	Aliases []*FirewallAlias    `json:"aliases,omitempty"`
	Ipset   []*FirewallIPSet    `json:"ipset,omitempty"`
	Rules   []*FirewallRule     `json:"rules,omitempty"`
	Options *FirewallNodeOption `json:"options,omitempty"`
	// Refs 	map[string]string `json:"refs,omitempty"`
}

type FirewallAlias struct {
	Cidr    string `json:"cidr,omitempty"`
	Digest  string `json:"digest,omitempty"`
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type FirewallIPSet struct {
	Name    string `json:"name,omitempty"`
	Digest  string `json:"digest,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type FirewallIPSetCreationOption struct {
	Name    string `json:"name"`
	Digest  string `json:"digest,omitempty"`
	Comment string `json:"comment,omitempty"`
	Rename  string `json:"rename,omitempty"`
}

type FirewallIPSetEntry struct {
	CIDR    string `json:"cidr,omitempty"`
	Digest  string `json:"digest,omitempty"`
	Comment string `json:"comment,omitempty"`
	NoMatch bool   `json:"nomatch,omitempty"`
}

type FirewallIPSetEntryCreationOption struct {
	CIDR    string `json:"cidr"`
	Comment string `json:"comment,omitempty"`
	NoMatch bool   `json:"nomatch,omitempty"`
}

type FirewallIPSetEntryUpdateOption struct {
	Comment string `json:"comment,omitempty"`
	Digest  string `json:"digest,omitempty"`
	NoMatch bool   `json:"nomatch,omitempty"`
}

type (
	VirtualMachineBackupMode               = string
	VirtualMachineBackupCompress           = string
	VirtualMachineBackupNotificationPolicy = string
)

const (
	VirtualMachineBackupModeSnapshot = VirtualMachineBackupMode("snapshot")
	VirtualMachineBackupModeSuspend  = VirtualMachineBackupMode("suspend")
	VirtualMachineBackupModeStop     = VirtualMachineBackupMode("stop")

	VirtualMachineBackupCompressZero = VirtualMachineBackupCompress("0")
	VirtualMachineBackupCompressOne  = VirtualMachineBackupCompress("1")
	VirtualMachineBackupCompressGzip = VirtualMachineBackupCompress("gzip")
	VirtualMachineBackupCompressLzo  = VirtualMachineBackupCompress("lzo")
	VirtualMachineBackupCompressZstd = VirtualMachineBackupCompress("zstd")

	VirtualMachineBackupNotificationPolicyAlways  = VirtualMachineBackupNotificationPolicy("always")
	VirtualMachineBackupNotificationPolicyFailure = VirtualMachineBackupNotificationPolicy("failure")
	VirtualMachineBackupNotificationPolicyNever   = VirtualMachineBackupNotificationPolicy("never")
)

type VirtualMachineBackupOptions struct {
	All                bool                                   `json:"all,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	BwLimit            uint                                   `json:"bwlimit,omitempty"`
	Compress           VirtualMachineBackupCompress           `json:"compress,omitempty"` // FIXME(issue-199): PVE default "0" (no compression); empty string already omits, but the constant for "0" should be exported so callers can opt in explicitly.
	DumpDir            string                                 `json:"dumpDir,omitempty"`
	Exclude            string                                 `json:"exclude,omitempty"`
	ExcludePath        []string                               `json:"exclude-path,omitempty"`
	IoNice             uint                                   `json:"ionice,omitempty"`   // FIXME(issue-199): PVE default 7; use *uint so unset doesn't override.
	LockWait           uint                                   `json:"lockwait,omitempty"` // FIXME(issue-199): PVE default 180 (seconds); use *uint so unset doesn't override.
	MailTo             string                                 `json:"mailto,omitempty"`
	Mode               VirtualMachineBackupMode               `json:"mode,omitempty"` // FIXME(issue-199): PVE default "snapshot"; empty string already omits, but consider explicit handling for callers.
	Node               string                                 `json:"node,omitempty"`
	NotesTemplate      string                                 `json:"notes-template,omitempty"`
	NotificationPolicy VirtualMachineBackupNotificationPolicy `json:"notification-policy,omitempty"`
	NotificationTarget string                                 `json:"notification-target,omitempty"`
	Performance        string                                 `json:"performance,omitempty"`
	Pigz               int                                    `json:"pigz,omitempty"`
	Pool               string                                 `json:"pool,omitempty"`
	Protected          string                                 `json:"protected,omitempty"`
	PruneBackups       string                                 `json:"prune-backups,omitempty"` // FIXME(issue-199): PVE default "keep-all=1"; use *string so unset doesn't override the prune policy.
	Quiet              bool                                   `json:"quiet,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	Remove             bool                                   `json:"remove"`      // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs. Note: no omitempty, so current `false` zero-value IS sent and disables pruning.
	Script             string                                 `json:"script,omitempty"`
	StdExcludes        bool                                   `json:"stdexcludes"` // FIXME(issue-199+178): PVE default 1, schema "boolean"; use *IntOrBool — type mismatch + default differs. Note: no omitempty, so current `false` zero-value IS sent and disables tmp/log exclusion.
	StdOut             bool                                   `json:"stdout,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool to match (defaults match — no pointer needed).
	Stop               bool                                   `json:"stop,omitempty"`   // FIXME(issue-178): schema "boolean"; use IntOrBool to match (default 0 matches Go zero — no pointer needed).
	StopWait           uint                                   `json:"stopwait,omitempty"` // FIXME(issue-199): PVE default 10 (minutes); use *uint so unset doesn't override.
	Storage            string                                 `json:"storage,omitempty"`
	TmpDir             string                                 `json:"tmpdir,omitempty"`
	VMID               uint64                                 `json:"vmid,omitempty"`
	Zstd               uint                                   `json:"zstd,omitempty"` // FIXME(issue-199): PVE default 1 (zstd thread count); use *uint so unset doesn't override.
}

type Separator = string

const (
	StringSeparator = Separator("\n")
	FieldSeparator  = Separator(":")
	SpaceSeparator  = Separator(" ")
)

type VzdumpConfig struct {
	Boot       string `json:"boot"`
	CiPassword string `json:"cipassword"`
	CiUser     string `json:"ciuser"`
	Cores      uint64 `json:"cores,string"`
	Memory     uint64 `json:"memory,string"`
	Meta       string `json:"meta"`
	Numa       string `json:"numa"`
	OsType     string `json:"ostype"`
	Scsihw     string `json:"scsihw"`
	Sockets uint64 `json:"sockets,string"`
	// SSHKeys is reflected back from VzDump's recorded VM config; if you
	// round-trip this into a VirtualMachineConfig.SSHKeys, the value is
	// already PVE-encoded — re-encoding it would double-encode. See #144.
	SSHKeys string `json:"sshkeys"`
	VmgenID string `json:"vmgenid"`

	IDE0 string `json:"ide0,omitempty"`
	IDE1 string `json:"ide1,omitempty"`
	IDE2 string `json:"ide2,omitempty"`
	IDE3 string `json:"ide3,omitempty"`

	SCSI0  string `json:"scsi0,omitempty"`
	SCSI1  string `json:"scsi1,omitempty"`
	SCSI2  string `json:"scsi2,omitempty"`
	SCSI3  string `json:"scsi3,omitempty"`
	SCSI4  string `json:"scsi4,omitempty"`
	SCSI5  string `json:"scsi5,omitempty"`
	SCSI6  string `json:"scsi6,omitempty"`
	SCSI7  string `json:"scsi7,omitempty"`
	SCSI8  string `json:"scsi8,omitempty"`
	SCSI9  string `json:"scsi9,omitempty"`
	SCSI10 string `json:"scsi10,omitempty"`
	SCSI11 string `json:"scsi11,omitempty"`
	SCSI12 string `json:"scsi12,omitempty"`
	SCSI13 string `json:"scsi13,omitempty"`
	SCSI14 string `json:"scsi14,omitempty"`
	SCSI15 string `json:"scsi15,omitempty"`
	SCSI16 string `json:"scsi16,omitempty"`
	SCSI17 string `json:"scsi17,omitempty"`
	SCSI18 string `json:"scsi18,omitempty"`
	SCSI19 string `json:"scsi19,omitempty"`
	SCSI20 string `json:"scsi20,omitempty"`
	SCSI21 string `json:"scsi21,omitempty"`
	SCSI22 string `json:"scsi22,omitempty"`
	SCSI23 string `json:"scsi23,omitempty"`
	SCSI24 string `json:"scsi24,omitempty"`
	SCSI25 string `json:"scsi25,omitempty"`
	SCSI26 string `json:"scsi26,omitempty"`
	SCSI27 string `json:"scsi27,omitempty"`
	SCSI28 string `json:"scsi28,omitempty"`
	SCSI29 string `json:"scsi29,omitempty"`
	SCSI30 string `json:"scsi30,omitempty"`

	SATA0 string `json:"sata0,omitempty"`
	SATA1 string `json:"sata1,omitempty"`
	SATA2 string `json:"sata2,omitempty"`
	SATA3 string `json:"sata3,omitempty"`
	SATA4 string `json:"sata4,omitempty"`
	SATA5 string `json:"sata5,omitempty"`

	VirtIO0  string `json:"virtio0,omitempty"`
	VirtIO1  string `json:"virtio1,omitempty"`
	VirtIO2  string `json:"virtio2,omitempty"`
	VirtIO3  string `json:"virtio3,omitempty"`
	VirtIO4  string `json:"virtio4,omitempty"`
	VirtIO5  string `json:"virtio5,omitempty"`
	VirtIO6  string `json:"virtio6,omitempty"`
	VirtIO7  string `json:"virtio7,omitempty"`
	VirtIO8  string `json:"virtio8,omitempty"`
	VirtIO9  string `json:"virtio9,omitempty"`
	VirtIO10 string `json:"virtio10,omitempty"`
	VirtIO11 string `json:"virtio11,omitempty"`
	VirtIO12 string `json:"virtio12,omitempty"`
	VirtIO13 string `json:"virtio13,omitempty"`
	VirtIO14 string `json:"virtio14,omitempty"`
	VirtIO15 string `json:"virtio15,omitempty"`

	Unused0 string `json:"unused0,omitempty"`
	Unused1 string `json:"unused1,omitempty"`
	Unused2 string `json:"unused2,omitempty"`
	Unused3 string `json:"unused3,omitempty"`
	Unused4 string `json:"unused4,omitempty"`
	Unused5 string `json:"unused5,omitempty"`
	Unused6 string `json:"unused6,omitempty"`
	Unused7 string `json:"unused7,omitempty"`
	Unused8 string `json:"unused8,omitempty"`
	Unused9 string `json:"unused9,omitempty"`

	// Network devices
	Net0 string `json:"net0,omitempty"`
	Net1 string `json:"net1,omitempty"`
	Net2 string `json:"net2,omitempty"`
	Net3 string `json:"net3,omitempty"`
	Net4 string `json:"net4,omitempty"`
	Net5 string `json:"net5,omitempty"`
	Net6 string `json:"net6,omitempty"`
	Net7 string `json:"net7,omitempty"`
	Net8 string `json:"net8,omitempty"`
	Net9 string `json:"net9,omitempty"`

	// NUMA topology
	Numa0 string `json:"numa0,omitempty"`
	Numa1 string `json:"numa1,omitempty"`
	Numa2 string `json:"numa2,omitempty"`
	Numa3 string `json:"numa3,omitempty"`
	Numa4 string `json:"numa4,omitempty"`
	Numa5 string `json:"numa5,omitempty"`
	Numa6 string `json:"numa6,omitempty"`
	Numa7 string `json:"numa7,omitempty"`
	Numa8 string `json:"numa8,omitempty"`
	Numa9 string `json:"numa9,omitempty"`

	// Host PCI devices
	HostPCI0 string `json:"hostpci0,omitempty"`
	HostPCI1 string `json:"hostpci1,omitempty"`
	HostPCI2 string `json:"hostpci2,omitempty"`
	HostPCI3 string `json:"hostpci3,omitempty"`
	HostPCI4 string `json:"hostpci4,omitempty"`
	HostPCI5 string `json:"hostpci5,omitempty"`
	HostPCI6 string `json:"hostpci6,omitempty"`
	HostPCI7 string `json:"hostpci7,omitempty"`
	HostPCI8 string `json:"hostpci8,omitempty"`
	HostPCI9 string `json:"hostpci9,omitempty"`

	// Serial devices
	Serial0 string `json:"serial0,omitempty"`
	Serial1 string `json:"serial1,omitempty"`
	Serial2 string `json:"serial2,omitempty"`
	Serial3 string `json:"serial3,omitempty"`

	// USB devices
	USB0  string `json:"usb0,omitempty"`
	USB1  string `json:"usb1,omitempty"`
	USB2  string `json:"usb2,omitempty"`
	USB3  string `json:"usb3,omitempty"`
	USB4  string `json:"usb4,omitempty"`
	USB5  string `json:"usb5,omitempty"`
	USB6  string `json:"usb6,omitempty"`
	USB7  string `json:"usb7,omitempty"`
	USB8  string `json:"usb8,omitempty"`
	USB9  string `json:"usb9,omitempty"`
	USB10 string `json:"usb10,omitempty"`
	USB11 string `json:"usb11,omitempty"`
	USB12 string `json:"usb12,omitempty"`
	USB13 string `json:"usb13,omitempty"`
	USB14 string `json:"usb14,omitempty"`

	Parallel0 string `json:"parallel0,omitempty"`
	Parallel1 string `json:"parallel1,omitempty"`
	Parallel2 string `json:"parallel2,omitempty"`

	// Cloud-init
	IPConfig0 string `json:"ipconfig0,omitempty"`
	IPConfig1 string `json:"ipconfig1,omitempty"`
	IPConfig2 string `json:"ipconfig2,omitempty"`
	IPConfig3 string `json:"ipconfig3,omitempty"`
	IPConfig4 string `json:"ipconfig4,omitempty"`
	IPConfig5 string `json:"ipconfig5,omitempty"`
	IPConfig6 string `json:"ipconfig6,omitempty"`
	IPConfig7 string `json:"ipconfig7,omitempty"`
	IPConfig8 string `json:"ipconfig8,omitempty"`
	IPConfig9 string `json:"ipconfig9,omitempty"`
}

type PendingConfiguration []PendingConfigItem

type PendingConfigItem struct {
	Key    string `json:"key,omitempty"`
	Delete *int   `json:"delete,omitempty"`
	// Proxmox API doc says "Pending" & "Value" fields return string but in reality it could be anything
	Pending interface{} `json:"pending,omitempty"`
	Value   interface{} `json:"value,omitempty"`
}

type VNet struct {
	Name      string `json:"vnet,omitempty"`
	Type      string `json:"type,omitempty"`
	Zone      string `json:"zone,omitempty"`
	Alias     string `json:"alias,omitempty"`
	VlanAware int    `json:"vlanaware,omitempty"`
	Tag       uint32 `json:"tag,omitempty"`
}

type VNetOptions struct {
	Name         string `json:"vnet"`
	Zone         string `json:"zone"`
	Alias        string `json:"alias,omitempty"`
	IsolatePorts bool   `json:"isolate-ports,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	Tag          uint32 `json:"tag,omitempty"`           // Could be a VLAN or VXLAN tag
	Type         string `json:"type,omitempty"`          // Type must be set to "vnet"
	VlanAware    bool   `json:"vlanaware,omitempty"`     // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
}
type NetRange struct {
	StartAddress string `json:"start-address,omitempty"`
	EndAddress   string `json:"end-address,omitempty"`
}
type VNetSubnet struct {
	CIDR      string     `json:"cidr,omitempty"`
	Gateway   string     `json:"gateway,omitempty"`
	Netmask   string     `json:"mask,omitempty"`
	Type      string     `json:"type,omitempty"`
	Zone      string     `json:"zone,omitempty"`
	VNet      string     `json:"vnet,omitempty"`
	SNAT      int        `json:"snat,omitempty"`
	Network   string     `json:"network,omitempty"`
	ID        string     `json:"id,omitempty"`
	DhcpRange []NetRange `json:"dhcp-range,omitempty"`
}
type IPAM struct {
	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	Mac      string `json:"mac,omitempty"`
	Subnet   string `json:"subnet,omitempty"`
	VMID     string `json:"vmid,omitempty"`
	VNet     string `json:"vnet,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Gateway  int    `json:"gateway,omitempty"`
}

type SDNZone struct {
	Name       string   `json:"zone"`
	Type       string   `json:"type"`
	DHCP       string   `json:"dhcp,omitempty"`
	DNS        string   `json:"dns,omitempty"`
	DNSZone    string   `json:"dnszone,omitempty"`
	IPAM       string   `json:"ipam,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	Nodes      []string `json:"nodes,omitempty"`
	Pending    bool     `json:"pending,omitempty"`
	ReverseDNS string   `json:"reversedns,omitempty"`
	State      string   `json:"state,omitempty"`
}

type SDNZoneOptions struct {
	Name                     string `json:"zone"`
	Type                     string `json:"type"`
	AdvertiseSubnets         bool   `json:"advertise-subnets,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	Bridge                   string `json:"bridge,omitempty"`
	BridgeDisableMACLearning bool   `json:"bridge-disable-mac-learning,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	Controller               string `json:"controller,omitempty"`
	DHCP                     string `json:"dhcp,omitempty"`
	DisableARPNDSuppression  bool   `json:"disable-arp-nd-suppression,omitempty"` // FIXME(issue-178): schema "boolean"; use IntOrBool (defaults match — no pointer needed).
	DNS                      string `json:"dns,omitempty"`
	DNSZone                  string `json:"dnszone,omitempty"`
	DPID                     int    `json:"dpid,omitempty"`
	ExitNodes                string `json:"exit-nodes,omitempty"`
	ExitNodesLocalRouting    bool   `json:"exit-nodes-local-routing,omitempty"`
	ExitNodesPrimary         string `json:"exit-nodes-primary,omitempty"`
	Fabric                   string `json:"fabric,omitempty"`
	IPAM                     string `json:"ipam,omitempty"`
	MAC                      string `json:"mac,omitempty"`
	MTU                      int    `json:"mtu,omitempty"`
	Nodes                    string `json:"nodes,omitempty"`
	Peers                    string `json:"peers,omitempty"`
	ReverseDNS               string `json:"reversedns,omitempty"`
	RTImport                 string `json:"rt-import,omitempty"`
	Tag                      uint   `json:"tag,omitempty"`
	VLANProtocol             string `json:"vlan-protocol,omitempty"` // FIXME(issue-199): PVE default "802.1q"; use *string so unset doesn't override (e.g. drop 802.1ad zones).
	VRFVXLAN                 int    `json:"vrf-vxlan,omitempty"`
	VXLANPort                uint16 `json:"vxlan-port,omitempty"` // FIXME(issue-199): PVE default 4789; use *uint16 so unset doesn't send 0.
}

// ClusterMetricServers is the list payload returned by GET /cluster/metrics/server.
type ClusterMetricServers []*ClusterMetricServerSummary

// ClusterMetricServerSummary is the trimmed shape returned by the list endpoint.
// The detailed GET /cluster/metrics/server/{id} returns the full plugin config in
// ClusterMetricServer.
type ClusterMetricServerSummary struct {
	ID      string    `json:"id,omitempty"`
	Type    string    `json:"type,omitempty"`
	Server  string    `json:"server,omitempty"`
	Port    int       `json:"port,omitempty"`
	Disable IntOrBool `json:"disable,omitempty"`
}

// ClusterMetricServer is the union of fields PVE returns for a single configured
// metric server. PVE multiplexes graphite / influxdb / opentelemetry plugins
// behind one config-id; per-plugin fields are populated only when relevant.
type ClusterMetricServer struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	Server        string    `json:"server,omitempty"`
	Port          int       `json:"port,omitempty"`
	Disable       IntOrBool `json:"disable,omitempty"`
	APIPathPrefix string    `json:"api-path-prefix,omitempty"`
	Bucket        string    `json:"bucket,omitempty"`
	InfluxDBProto string    `json:"influxdbproto,omitempty"`
	MaxBodySize   uint64    `json:"max-body-size,omitempty"`
	MTU           uint      `json:"mtu,omitempty"`
	Organization  string    `json:"organization,omitempty"`
	Path          string    `json:"path,omitempty"`
	Proto         string    `json:"proto,omitempty"`
	Timeout       uint      `json:"timeout,omitempty"`
	Token         string    `json:"token,omitempty"`
	// OpenTelemetry-specific knobs.
	OtelCompression        string    `json:"otel-compression,omitempty"`
	OtelHeaders            string    `json:"otel-headers,omitempty"`
	OtelMaxBodySize        uint64    `json:"otel-max-body-size,omitempty"`
	OtelPath               string    `json:"otel-path,omitempty"`
	OtelProtocol           string    `json:"otel-protocol,omitempty"`
	OtelResourceAttributes string    `json:"otel-resource-attributes,omitempty"`
	OtelTimeout            uint      `json:"otel-timeout,omitempty"`
	OtelVerifySSL          IntOrBool `json:"otel-verify-ssl,omitempty"`
	VerifyCertificate      IntOrBool `json:"verify-certificate,omitempty"`
	Digest                 string    `json:"digest,omitempty"`
}

// ClusterMetricServerOptions is the create/update payload. POST requires id+type;
// PUT uses id from the URL and accepts a "delete" comma list to unset keys.
type ClusterMetricServerOptions struct {
	ID            string `json:"id,omitempty"`
	Type          string `json:"type,omitempty"` // graphite | influxdb | opentelemetry — POST only
	Server        string `json:"server,omitempty"`
	Port          int    `json:"port,omitempty"`
	Disable       *bool  `json:"disable,omitempty"`
	APIPathPrefix string `json:"api-path-prefix,omitempty"`
	Bucket        string `json:"bucket,omitempty"`
	InfluxDBProto string `json:"influxdbproto,omitempty"`
	MaxBodySize   uint64 `json:"max-body-size,omitempty"`
	MTU           uint   `json:"mtu,omitempty"`
	Organization  string `json:"organization,omitempty"`
	Path          string `json:"path,omitempty"`
	Proto         string `json:"proto,omitempty"`
	Timeout       uint   `json:"timeout,omitempty"`
	Token         string `json:"token,omitempty"`
	// OpenTelemetry-specific knobs.
	OtelCompression        string `json:"otel-compression,omitempty"`
	OtelHeaders            string `json:"otel-headers,omitempty"`
	OtelMaxBodySize        uint64 `json:"otel-max-body-size,omitempty"`
	OtelPath               string `json:"otel-path,omitempty"`
	OtelProtocol           string `json:"otel-protocol,omitempty"`
	OtelResourceAttributes string `json:"otel-resource-attributes,omitempty"`
	OtelTimeout            uint   `json:"otel-timeout,omitempty"`
	OtelVerifySSL          *bool  `json:"otel-verify-ssl,omitempty"`     // PVE default true; pointer so unset doesn't flip server-side
	VerifyCertificate      *bool  `json:"verify-certificate,omitempty"`  // PVE default true; pointer to avoid silently disabling TLS verification
	Digest                 string `json:"digest,omitempty"`              // PUT only — optimistic concurrency
	Delete                 string `json:"delete,omitempty"`              // PUT only — comma-separated keys to clear
}

// --- /cluster/jobs ---------------------------------------------------------

// ClusterJobIndexEntry is one row in the /cluster/jobs directory index.
type ClusterJobIndexEntry struct {
	SubDir string `json:"subdir,omitempty"`
}

// ClusterScheduleEvent is one firing in the schedule-analyze preview — a
// human-readable UTC timestamp + UNIX epoch.
type ClusterScheduleEvent struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	UTC       string `json:"utc,omitempty"`
}

// ClusterRealmSyncJob is the GET shape for a realm-sync job. PVE returns
// Enabled / EnableNew as integers; using IntOrBool to stay safe.
type ClusterRealmSyncJob struct {
	ID             string    `json:"id,omitempty"`
	Comment        string    `json:"comment,omitempty"`
	EnableNew      IntOrBool `json:"enable-new,omitempty"`
	Enabled        IntOrBool `json:"enabled,omitempty"`
	LastRun        int64     `json:"last-run,omitempty"`
	NextRun        int64     `json:"next-run,omitempty"`
	Realm          string    `json:"realm,omitempty"`
	RemoveVanished string    `json:"remove-vanished,omitempty"`
	Schedule       string    `json:"schedule,omitempty"`
	Scope          string    `json:"scope,omitempty"`
}

// ClusterRealmSyncJobOptions is the body for both POST (create) and PUT
// (update). Pointer fields preserve PVE's defaults when unset.
type ClusterRealmSyncJobOptions struct {
	Comment        string `json:"comment,omitempty"`
	EnableNew      *bool  `json:"enable-new,omitempty"` // PVE default true; pointer so unset doesn't flip
	Enabled        *bool  `json:"enabled,omitempty"`    // PVE default true; pointer so unset doesn't flip
	Realm          string `json:"realm,omitempty"`      // POST only — identifies the auth realm
	RemoveVanished string `json:"remove-vanished,omitempty"`
	Schedule       string `json:"schedule,omitempty"` // required on create
	Scope          string `json:"scope,omitempty"`
	Delete         string `json:"delete,omitempty"` // PUT only — comma-separated keys to clear
}

// --- /nodes/{node}/disks ---------------------------------------------------

// Disk is one row returned by GET /nodes/{node}/disks/list. Fields are
// best-effort optional — PVE omits keys that don't apply to a given device.
type Disk struct {
	DevPath     string    `json:"devpath,omitempty"`
	Used        string    `json:"used,omitempty"`
	GPT         IntOrBool `json:"gpt,omitempty"`
	Size        uint64    `json:"size,omitempty"`
	Health      string    `json:"health,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Type        string    `json:"type,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	WWN         string    `json:"wwn,omitempty"`
	ByIDLink    string    `json:"by_id_link,omitempty"`
	Wearout     string    `json:"wearout,omitempty"`
	OSDID       int       `json:"osdid,omitempty"`
	OSDEncrypted IntOrBool `json:"osdencrypted,omitempty"`
	Parent      string    `json:"parent,omitempty"`
	RPM         int       `json:"rpm,omitempty"`
	BLKSize     int       `json:"blocksize,omitempty"`
	MountPoint  string    `json:"mounted,omitempty"`
	Vendor2     string    `json:"vendor2,omitempty"`
}

// DiskSMART is the response from GET /nodes/{node}/disks/smart.
type DiskSMART struct {
	Health     string                 `json:"health,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Text       string                 `json:"text,omitempty"`
	Attributes []map[string]any       `json:"attributes,omitempty"`
}

// NodeDirectory is one row returned by GET /nodes/{node}/disks/directory.
type NodeDirectory struct {
	Device  string `json:"device,omitempty"`
	Options string `json:"options,omitempty"`
	Path    string `json:"path,omitempty"`
	Type    string `json:"type,omitempty"`
	UUID    string `json:"unitfile,omitempty"`
}

// NodeDirectoryOptions is the POST body for /nodes/{node}/disks/directory.
type NodeDirectoryOptions struct {
	Name       string `json:"name"`
	Device     string `json:"device"`
	Filesystem string `json:"filesystem,omitempty"` // PVE default ext4
	AddStorage IntOrBool `json:"add_storage,omitempty"`
}

// NodeLVMTree is the nested response from GET /nodes/{node}/disks/lvm. Each
// child is a volume group whose own children are the constituent physical
// volumes.
type NodeLVMTree struct {
	Children []NodeLVMVolumeGroup `json:"children,omitempty"`
	Leaf     IntOrBool            `json:"leaf,omitempty"`
}

type NodeLVMVolumeGroup struct {
	Name     string             `json:"name,omitempty"`
	Size     uint64             `json:"size,omitempty"`
	Free     uint64             `json:"free,omitempty"`
	Leaf     IntOrBool          `json:"leaf,omitempty"`
	Children []NodeLVMPhysical  `json:"children,omitempty"`
}

type NodeLVMPhysical struct {
	Name string `json:"name,omitempty"`
	Size uint64 `json:"size,omitempty"`
	Free uint64 `json:"free,omitempty"`
	Leaf IntOrBool `json:"leaf,omitempty"`
}

// NodeLVMOptions is the POST body for /nodes/{node}/disks/lvm.
type NodeLVMOptions struct {
	Name       string    `json:"name"`
	Device     string    `json:"device"`
	AddStorage IntOrBool `json:"add_storage,omitempty"`
}

// NodeLVMThin is one row from GET /nodes/{node}/disks/lvmthin.
type NodeLVMThin struct {
	LV           string `json:"lv,omitempty"`
	LVSize       uint64 `json:"lv_size,omitempty"`
	MetadataSize uint64 `json:"metadata_size,omitempty"`
	MetadataUsed uint64 `json:"metadata_used,omitempty"`
	Used         uint64 `json:"used,omitempty"`
}

// NodeLVMThinOptions is the POST body for /nodes/{node}/disks/lvmthin.
type NodeLVMThinOptions struct {
	Name       string    `json:"name"`
	Device     string    `json:"device"`
	AddStorage IntOrBool `json:"add_storage,omitempty"`
}

// NodeZFSPoolSummary is one row from GET /nodes/{node}/disks/zfs.
type NodeZFSPoolSummary struct {
	Name    string  `json:"name,omitempty"`
	Health  string  `json:"health,omitempty"`
	Size    uint64  `json:"size,omitempty"`
	Alloc   uint64  `json:"alloc,omitempty"`
	Free    uint64  `json:"free,omitempty"`
	Frag    int     `json:"frag,omitempty"`
	Dedup   float64 `json:"dedup,omitempty"`
}

// NodeZFSPool is the detailed pool status from GET /nodes/{node}/disks/zfs/{name}.
type NodeZFSPool struct {
	Name     string           `json:"name,omitempty"`
	State    string           `json:"state,omitempty"`
	Status   string           `json:"status,omitempty"`
	Action   string           `json:"action,omitempty"`
	Scan     string           `json:"scan,omitempty"`
	Errors   string           `json:"errors,omitempty"`
	Children []NodeZFSVdev    `json:"children,omitempty"`
}

type NodeZFSVdev struct {
	Name     string        `json:"name,omitempty"`
	State    string        `json:"state,omitempty"`
	Read     uint64        `json:"read,omitempty"`
	Write    uint64        `json:"write,omitempty"`
	Cksum    uint64        `json:"cksum,omitempty"`
	Msg      string        `json:"msg,omitempty"`
	Children []NodeZFSVdev `json:"children,omitempty"`
	Leaf     IntOrBool     `json:"leaf,omitempty"`
}

// NodeZFSPoolOptions is the POST body for /nodes/{node}/disks/zfs.
type NodeZFSPoolOptions struct {
	Name        string    `json:"name"`
	Devices     string    `json:"devices"` // space-separated device list per PVE
	RaidLevel   string    `json:"raidlevel"`
	Ashift      int       `json:"ashift,omitempty"`
	Compression string    `json:"compression,omitempty"`
	DraidConfig string    `json:"draid-config,omitempty"`
	AddStorage  IntOrBool `json:"add_storage,omitempty"`
}

// --- Ceph OSD (Object Storage Daemons) -------------------------------------

// CephOSD is the operations handle for a single OSD on a node, returned by
// Node.CephOSD(id). It carries no data fields — instance methods (In/Out/
// Scrub/Delete/LVInfo/Metadata) call back into the API when invoked.
type CephOSD struct {
	client *Client
	Node   string `json:"-"`
	ID     int    `json:"-"`
}

// CephOSDTree is the response from GET /nodes/{node}/ceph/osd — the CRUSH
// tree top-level plus any cluster-wide OSD flags. The CRUSH bucket structure
// is recursive and per-node properties (status, weight, in, usage, latencies,
// etc.) vary by bucket type, so Root is kept as a raw map.
type CephOSDTree struct {
	Flags string                 `json:"flags,omitempty"`
	Root  map[string]interface{} `json:"root,omitempty"`
}

// CephOSDDetails is the response from GET /nodes/{node}/ceph/osd/{osdid}/metadata
// — daemon-level info plus the list of backing devices.
type CephOSDDetails struct {
	OSD     CephOSDMetadata `json:"osd"`
	Devices []CephOSDDevice `json:"devices,omitempty"`
}

// CephOSDMetadata is the "osd" sub-object inside CephOSDDetails.
type CephOSDMetadata struct {
	BackAddr       string `json:"back_addr,omitempty"`
	Encrypted      bool   `json:"encrypted,omitempty"`
	FrontAddr      string `json:"front_addr,omitempty"`
	HBBackAddr     string `json:"hb_back_addr,omitempty"`
	HBFrontAddr    string `json:"hb_front_addr,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	ID             int    `json:"id"`
	MemUsage       int64  `json:"mem_usage,omitempty"`
	OSDData        string `json:"osd_data,omitempty"`
	OSDObjectStore string `json:"osd_objectstore,omitempty"`
	PID            int    `json:"pid,omitempty"`
	Version        string `json:"version,omitempty"`
}

// CephOSDDevice is one row in CephOSDDetails.Devices.
type CephOSDDevice struct {
	DevNode        string `json:"dev_node,omitempty"`
	Device         string `json:"device,omitempty"` // block|db|wal
	PhysicalDevice string `json:"physical_device,omitempty"`
	Size           uint64 `json:"size,omitempty"`
	SupportDiscard bool   `json:"support_discard,omitempty"`
	Type           string `json:"type,omitempty"` // hdd|ssd
}

// CephOSDLVInfo is the response from GET /nodes/{node}/ceph/osd/{osdid}/lv-info
// — LVM details for the OSD's block / db / wal logical volume.
type CephOSDLVInfo struct {
	CreationTime string `json:"creation_time,omitempty"`
	LVName       string `json:"lv_name,omitempty"`
	LVPath       string `json:"lv_path,omitempty"`
	LVSize       uint64 `json:"lv_size,omitempty"`
	LVUUID       string `json:"lv_uuid,omitempty"`
	VGName       string `json:"vg_name,omitempty"`
}

// CephOSDCreateOptions is the POST body for /nodes/{node}/ceph/osd.
// Dev is required. DBDevSize requires DBDev; WALDevSize requires WALDev.
// OSDsPerDevice is mutually exclusive with DBDev/WALDev.
type CephOSDCreateOptions struct {
	Dev              string    `json:"dev"`
	CrushDeviceClass string    `json:"crush-device-class,omitempty"`
	DBDev            string    `json:"db_dev,omitempty"`
	DBDevSize        float64   `json:"db_dev_size,omitempty"`
	Encrypted        IntOrBool `json:"encrypted,omitempty"`
	OSDsPerDevice    int       `json:"osds-per-device,omitempty"`
	WALDev           string    `json:"wal_dev,omitempty"`
	WALDevSize       float64   `json:"wal_dev_size,omitempty"`
}

// --- ACME (Let's Encrypt-style automated certificate issuance) -------------

// ACMEDirectory is one row in GET /cluster/acme/directories — a friendly name
// + URL for a known ACME CA endpoint.
type ACMEDirectory struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// ACMEMeta is the metadata document returned by GET /cluster/acme/meta — what
// the CA itself advertises about its capabilities and policies.
type ACMEMeta struct {
	CAAIdentities           []string `json:"caaIdentities,omitempty"`
	ExternalAccountRequired IntOrBool `json:"externalAccountRequired,omitempty"`
	TermsOfService          string   `json:"termsOfService,omitempty"`
	Website                 string   `json:"website,omitempty"`
}

// ACMEChallengeSchema is one entry in GET /cluster/acme/challenge-schema —
// the catalog of DNS plugin types PVE understands. Schema is the per-plugin
// parameter spec (left as raw map since each plugin defines its own).
type ACMEChallengeSchema struct {
	ID     string                 `json:"id,omitempty"`
	Name   string                 `json:"name,omitempty"`
	Type   string                 `json:"type,omitempty"`
	Schema map[string]interface{} `json:"schema,omitempty"`
}

// ACMEAccountIndex is one row in GET /cluster/acme/account — just the name.
type ACMEAccountIndex struct {
	Name string `json:"name,omitempty"`
}

// ACMEAccount is the full account record from GET /cluster/acme/account/{name}.
// `Account` is the raw CA-returned JSON (status, contacts, etc.) — left
// untyped because RFC 8555 leaves the shape extensible.
type ACMEAccount struct {
	Account   map[string]interface{} `json:"account,omitempty"`
	Directory string                 `json:"directory,omitempty"`
	Location  string                 `json:"location,omitempty"`
	TOS       string                 `json:"tos,omitempty"`
}

// ACMEAccountOptions is the POST body for creating an account. Contact is the
// only required field; PVE defaults Name to "default" and Directory to LE prod.
// EABKid + EABHMACKey are for External Account Binding (e.g. ZeroSSL).
type ACMEAccountOptions struct {
	Contact     string `json:"contact"`
	Directory   string `json:"directory,omitempty"`
	EABKid      string `json:"eab-kid,omitempty"`
	EABHMACKey  string `json:"eab-hmac-key,omitempty"`
	Name        string `json:"name,omitempty"`
	TOSURL      string `json:"tos_url,omitempty"`
}

// ACMEPlugin is the read shape from GET /cluster/acme/plugins[/{id}]. The
// per-provider parameters live in Data (a base64-encoded blob per PVE).
type ACMEPlugin struct {
	ID              string    `json:"plugin,omitempty"`
	Type            string    `json:"type,omitempty"`
	API             string    `json:"api,omitempty"`
	Data            string    `json:"data,omitempty"`
	Disable         IntOrBool `json:"disable,omitempty"`
	Nodes           string    `json:"nodes,omitempty"`
	ValidationDelay int       `json:"validation-delay,omitempty"`
	Digest          string    `json:"digest,omitempty"`
}

// ACMEPluginOptions is the create/update payload. POST requires ID + Type;
// PUT identifies the plugin by URL and accepts Delete to clear keys.
type ACMEPluginOptions struct {
	ID              string `json:"id,omitempty"`
	Type            string `json:"type,omitempty"` // "dns" | "standalone" — POST only
	API             string `json:"api,omitempty"`
	Data            string `json:"data,omitempty"`
	Disable         *bool  `json:"disable,omitempty"`
	Nodes           string `json:"nodes,omitempty"`
	ValidationDelay *int   `json:"validation-delay,omitempty"` // PVE default 30; pointer so unset doesn't reset to 0
	Digest          string `json:"digest,omitempty"`           // PUT only
	Delete          string `json:"delete,omitempty"`           // PUT only
}

// ClusterMappings is the directory index returned by GET /cluster/mapping.
type ClusterMappings []*ClusterMappingIndexEntry

// ClusterMappingIndexEntry is one row in the top-level mapping index.
type ClusterMappingIndexEntry struct {
	Name string `json:"name,omitempty"`
}

// ClusterMappingCheck captures the optional per-node diagnostic returned when
// the list endpoints are called with check-node set.
type ClusterMappingCheck struct {
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// ClusterDirMappings is the list payload returned by GET /cluster/mapping/dir.
type ClusterDirMappings []*ClusterDirMapping

// ClusterDirMapping describes a single directory mapping. The "map" field is
// a list of PVE property-strings ("node=...,path=...") rather than structured
// objects — that's what the API returns.
type ClusterDirMapping struct {
	ID          string                 `json:"id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Map         []string               `json:"map,omitempty"`
	Checks      []*ClusterMappingCheck `json:"checks,omitempty"`
	Digest      string                 `json:"digest,omitempty"`
}

// ClusterDirMappingOptions is the create/update payload for dir mappings.
type ClusterDirMappingOptions struct {
	ID          string   `json:"id,omitempty"`
	Description string   `json:"description,omitempty"`
	Map         []string `json:"map,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      string   `json:"delete,omitempty"`
}

// ClusterPCIMappings is the list payload returned by GET /cluster/mapping/pci.
type ClusterPCIMappings []*ClusterPCIMapping

// ClusterPCIMapping describes a logical PCI device mapping.
type ClusterPCIMapping struct {
	ID                   string                 `json:"id,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Map                  []string               `json:"map,omitempty"`
	Checks               []*ClusterMappingCheck `json:"checks,omitempty"`
	MDev                 IntOrBool              `json:"mdev,omitempty"`
	LiveMigrationCapable IntOrBool              `json:"live-migration-capable,omitempty"`
	Digest               string                 `json:"digest,omitempty"`
}

// ClusterPCIMappingOptions is the create/update payload for PCI mappings.
// mdev / live-migration-capable both default to false on PVE, so plain bool
// with omitempty is safe.
type ClusterPCIMappingOptions struct {
	ID                   string   `json:"id,omitempty"`
	Description          string   `json:"description,omitempty"`
	Map                  []string `json:"map,omitempty"`
	MDev                 bool     `json:"mdev,omitempty"`
	LiveMigrationCapable bool     `json:"live-migration-capable,omitempty"`
	Digest               string   `json:"digest,omitempty"`
	Delete               string   `json:"delete,omitempty"`
}

// ClusterUSBMappings is the list payload returned by GET /cluster/mapping/usb.
type ClusterUSBMappings []*ClusterUSBMapping

// ClusterUSBMapping describes a logical USB device mapping. USB uses "error"
// instead of "checks" in the list response (PVE quirk — not normalised).
type ClusterUSBMapping struct {
	ID          string                 `json:"id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Map         []string               `json:"map,omitempty"`
	Error       []*ClusterMappingCheck `json:"error,omitempty"`
	Digest      string                 `json:"digest,omitempty"`
}

// ClusterUSBMappingOptions is the create/update payload for USB mappings.
type ClusterUSBMappingOptions struct {
	ID          string   `json:"id,omitempty"`
	Description string   `json:"description,omitempty"`
	Map         []string `json:"map,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      string   `json:"delete,omitempty"`
}

// --- notifications ----------------------------------------------------------

// ClusterNotificationIndex is the top-level directory under /cluster/notifications.
type ClusterNotificationIndex []*ClusterNotificationIndexEntry

// ClusterNotificationIndexEntry is one row in the notifications index.
type ClusterNotificationIndexEntry struct {
	Name string `json:"name,omitempty"`
}

// ClusterNotificationMatcherField is a row from /cluster/notifications/matcher-fields.
type ClusterNotificationMatcherField struct {
	Name string `json:"name,omitempty"`
}

// ClusterNotificationMatcherFieldValue is a row from
// /cluster/notifications/matcher-field-values.
type ClusterNotificationMatcherFieldValue struct {
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// ClusterNotificationTarget is a row from /cluster/notifications/targets — a
// flattened view across all endpoint plugin types (sendmail/gotify/smtp/webhook).
type ClusterNotificationTarget struct {
	Name    string    `json:"name,omitempty"`
	Type    string    `json:"type,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Origin  string    `json:"origin,omitempty"`
	Disable IntOrBool `json:"disable,omitempty"`
}

// ClusterNotificationMatcher is a single matcher.
type ClusterNotificationMatcher struct {
	Name          string    `json:"name,omitempty"`
	Comment       string    `json:"comment,omitempty"`
	Mode          string    `json:"mode,omitempty"` // all | any
	Disable       IntOrBool `json:"disable,omitempty"`
	InvertMatch   IntOrBool `json:"invert-match,omitempty"`
	MatchCalendar []string  `json:"match-calendar,omitempty"`
	MatchField    []string  `json:"match-field,omitempty"`
	MatchSeverity []string  `json:"match-severity,omitempty"`
	Target        []string  `json:"target,omitempty"`
	Origin        string    `json:"origin,omitempty"`
	Digest        string    `json:"digest,omitempty"`
}

// ClusterNotificationMatcherOptions is the create/update payload for matchers.
// Delete is an array of keys (not a comma-string) per PVE schema.
type ClusterNotificationMatcherOptions struct {
	Name          string   `json:"name,omitempty"`
	Comment       string   `json:"comment,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	Disable       *bool    `json:"disable,omitempty"`
	InvertMatch   *bool    `json:"invert-match,omitempty"`
	MatchCalendar []string `json:"match-calendar,omitempty"`
	MatchField    []string `json:"match-field,omitempty"`
	MatchSeverity []string `json:"match-severity,omitempty"`
	Target        []string `json:"target,omitempty"`
	Digest        string   `json:"digest,omitempty"`
	Delete        []string `json:"delete,omitempty"`
}

// ClusterNotificationGotifyEndpoint is a Gotify endpoint configuration. The
// `token` field is write-only on PVE — GET never returns it.
type ClusterNotificationGotifyEndpoint struct {
	Name    string    `json:"name,omitempty"`
	Server  string    `json:"server,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Disable IntOrBool `json:"disable,omitempty"`
	Origin  string    `json:"origin,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}

// ClusterNotificationGotifyOptions is the create/update payload for gotify.
type ClusterNotificationGotifyOptions struct {
	Name    string   `json:"name,omitempty"`
	Server  string   `json:"server,omitempty"`
	Token   string   `json:"token,omitempty"`
	Comment string   `json:"comment,omitempty"`
	Disable *bool    `json:"disable,omitempty"`
	Digest  string   `json:"digest,omitempty"`
	Delete  []string `json:"delete,omitempty"`
}

// ClusterNotificationSendmailEndpoint is a sendmail endpoint.
type ClusterNotificationSendmailEndpoint struct {
	Name        string    `json:"name,omitempty"`
	Author      string    `json:"author,omitempty"`
	FromAddress string    `json:"from-address,omitempty"`
	MailTo      []string  `json:"mailto,omitempty"`
	MailToUser  []string  `json:"mailto-user,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Disable     IntOrBool `json:"disable,omitempty"`
	Origin      string    `json:"origin,omitempty"`
	Digest      string    `json:"digest,omitempty"`
}

// ClusterNotificationSendmailOptions is the create/update payload for sendmail.
type ClusterNotificationSendmailOptions struct {
	Name        string   `json:"name,omitempty"`
	Author      string   `json:"author,omitempty"`
	FromAddress string   `json:"from-address,omitempty"`
	MailTo      []string `json:"mailto,omitempty"`
	MailToUser  []string `json:"mailto-user,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Disable     *bool    `json:"disable,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      []string `json:"delete,omitempty"`
}

// ClusterNotificationSMTPEndpoint is an SMTP endpoint. Password is write-only.
type ClusterNotificationSMTPEndpoint struct {
	Name        string    `json:"name,omitempty"`
	Server      string    `json:"server,omitempty"`
	Port        int       `json:"port,omitempty"`
	Mode        string    `json:"mode,omitempty"` // insecure | starttls | tls
	Username    string    `json:"username,omitempty"`
	FromAddress string    `json:"from-address,omitempty"`
	Author      string    `json:"author,omitempty"`
	MailTo      []string  `json:"mailto,omitempty"`
	MailToUser  []string  `json:"mailto-user,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Disable     IntOrBool `json:"disable,omitempty"`
	Origin      string    `json:"origin,omitempty"`
	Digest      string    `json:"digest,omitempty"`
}

// ClusterNotificationSMTPOptions is the create/update payload for smtp.
type ClusterNotificationSMTPOptions struct {
	Name        string   `json:"name,omitempty"`
	Server      string   `json:"server,omitempty"`
	Port        int      `json:"port,omitempty"`
	Mode        string   `json:"mode,omitempty"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	FromAddress string   `json:"from-address,omitempty"`
	Author      string   `json:"author,omitempty"`
	MailTo      []string `json:"mailto,omitempty"`
	MailToUser  []string `json:"mailto-user,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Disable     *bool    `json:"disable,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      []string `json:"delete,omitempty"`
}

// ClusterNotificationWebhookEndpoint is a webhook endpoint. The header / secret
// arrays use PVE property-string format ("name=...,value=<base64>"); body is
// already base64-encoded on the wire.
type ClusterNotificationWebhookEndpoint struct {
	Name    string    `json:"name,omitempty"`
	URL     string    `json:"url,omitempty"`
	Method  string    `json:"method,omitempty"` // post | put | get
	Header  []string  `json:"header,omitempty"`
	Body    string    `json:"body,omitempty"`
	Secret  []string  `json:"secret,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Disable IntOrBool `json:"disable,omitempty"`
	Origin  string    `json:"origin,omitempty"`
	Digest  string    `json:"digest,omitempty"`
}

// ClusterNotificationWebhookOptions is the create/update payload for webhook.
type ClusterNotificationWebhookOptions struct {
	Name    string   `json:"name,omitempty"`
	URL     string   `json:"url,omitempty"`
	Method  string   `json:"method,omitempty"`
	Header  []string `json:"header,omitempty"`
	Body    string   `json:"body,omitempty"`
	Secret  []string `json:"secret,omitempty"`
	Comment string   `json:"comment,omitempty"`
	Disable *bool    `json:"disable,omitempty"`
	Digest  string   `json:"digest,omitempty"`
	Delete  []string `json:"delete,omitempty"`
}

// --- /nodes/{node}/qemu/{vmid} directory indexes --------------------------

// VirtualMachineDirIndexEntry is one row in the per-VM directory index
// (GET /nodes/{node}/qemu/{vmid}) — each entry names a child resource
// (config, status, snapshot, firewall, agent, …).
type VirtualMachineDirIndexEntry struct {
	Subdir string `json:"subdir,omitempty"`
}

// VirtualMachineStatusIndexEntry is one row in the VM status directory index
// (GET /nodes/{node}/qemu/{vmid}/status) — each entry names a status
// sub-command (current, start, stop, reboot, …).
type VirtualMachineStatusIndexEntry struct {
	Subdir string `json:"subdir,omitempty"`
}

// VirtualMachineSnapshotIndexEntry is one row in the per-snapshot directory
// index (GET /nodes/{node}/qemu/{vmid}/snapshot/{snapname}) — each entry
// names a sub-resource on the snapshot (config, rollback).
type VirtualMachineSnapshotIndexEntry struct {
	Subdir string `json:"subdir,omitempty"`
}

// --- /nodes/{node}/qemu/{vmid}/mtunnel ------------------------------------

// VirtualMachineMigrationTunnel is the response from POST
// /nodes/{node}/qemu/{vmid}/mtunnel — a Unix socket path plus an
// authentication ticket the caller can use with the mtunnelwebsocket
// endpoint. PVE marks this endpoint as "for internal use by VM migration".
type VirtualMachineMigrationTunnel struct {
	Socket string `json:"socket,omitempty"`
	Ticket string `json:"ticket,omitempty"`
	UPID   string `json:"upid,omitempty"`
}

// VirtualMachineMigrationTunnelOptions is the request body for POST
// /nodes/{node}/qemu/{vmid}/mtunnel.
type VirtualMachineMigrationTunnelOptions struct {
	// Bridges is a comma-separated list of network bridges to check
	// availability for. Optional.
	Bridges string `json:"bridges,omitempty"`
	// Storages is a comma-separated list of storages to check permission
	// and availability for. Optional.
	Storages string `json:"storages,omitempty"`
}

// --- /nodes/{node}/ceph/{mon,mgr,mds} daemon registries -------------------
//
// These types are the per-node daemon-registry entries returned by the
// "list" GETs under /nodes/{node}/ceph/{mon,mgr,mds}. They are distinct
// from the cluster-wide ClusterCephMon (used inside ClusterCephStatus.Monmap.Mons)
// and CephMgrMap (the active manager map in cluster status) — those describe
// what the Ceph cluster sees, while these describe what PVE has configured
// on this node (including stopped/unknown daemons).
//
// Each carries unexported `client` and exported `Node` fields that the
// Node.CephMons / Node.CephMon (and Mgr/MDS equivalents) accessors populate so
// instance methods (`.Delete()`) can call back into the API without the caller
// re-supplying Node + client.

// CephMon is one row from GET /nodes/{node}/ceph/mon AND the operations handle
// returned by Node.CephMon(name). "Name" is the monid; "State" mixes cluster
// reality (running) with PVE config state (stopped/unknown).
type CephMon struct {
	client           *Client
	Node             string    `json:"-"`
	Addr             string    `json:"addr,omitempty"`
	CephVersion      string    `json:"ceph_version,omitempty"`
	CephVersionShort string    `json:"ceph_version_short,omitempty"`
	DirExists        IntOrBool `json:"direxists,omitempty"`
	Host             string    `json:"host,omitempty"`
	Name             string    `json:"name,omitempty"`
	Quorum           IntOrBool `json:"quorum,omitempty"`
	Rank             int       `json:"rank,omitempty"`
	Service          IntOrBool `json:"service,omitempty"`
	State            string    `json:"state,omitempty"`
}

// CephMonOptions is the POST body for /nodes/{node}/ceph/mon/{monid}. monid
// is set via the URL path; MonAddress overrides the autodetected monitor IP
// address(es), must be on Ceph's public network.
type CephMonOptions struct {
	MonAddress string `json:"mon-address,omitempty"`
}

// CephMgr is one row from GET /nodes/{node}/ceph/mgr AND the operations handle
// returned by Node.CephMgr(id). Distinct from CephMgrMap (the active manager
// map in cluster-status snapshots).
type CephMgr struct {
	client           *Client
	Node             string    `json:"-"`
	Addr             string    `json:"addr,omitempty"`
	CephVersion      string    `json:"ceph_version,omitempty"`
	CephVersionShort string    `json:"ceph_version_short,omitempty"`
	DirExists        IntOrBool `json:"direxists,omitempty"`
	Host             string    `json:"host,omitempty"`
	Name             string    `json:"name,omitempty"`
	Service          IntOrBool `json:"service,omitempty"`
	State            string    `json:"state,omitempty"`
}

// CephMDS is one row from GET /nodes/{node}/ceph/mds AND the operations handle
// returned by Node.CephMDS(name).
type CephMDS struct {
	client           *Client
	Node             string    `json:"-"`
	Addr             string    `json:"addr,omitempty"`
	CephVersion      string    `json:"ceph_version,omitempty"`
	CephVersionShort string    `json:"ceph_version_short,omitempty"`
	DirExists        IntOrBool `json:"direxists,omitempty"`
	FSName           string    `json:"fs_name,omitempty"`
	Host             string    `json:"host,omitempty"`
	Name             string    `json:"name,omitempty"`
	Rank             int       `json:"rank,omitempty"`
	Service          IntOrBool `json:"service,omitempty"`
	StandbyReplay    IntOrBool `json:"standby_replay,omitempty"`
	State            string    `json:"state,omitempty"`
}

// CephMDSOptions is the POST body for /nodes/{node}/ceph/mds/{name}. Hot
// standby has the daemon poll and replay an active MDS' log for faster
// failover at the cost of always-on idle resources.
type CephMDSOptions struct {
	HotStandby IntOrBool `json:"hotstandby,omitempty"`
}
