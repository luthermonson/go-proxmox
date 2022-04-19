package proxmox

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

var (
	vmConfigRegexpIDE    *regexp.Regexp
	vmConfigRegexpSCSI   *regexp.Regexp
	vmConfigRegexpSATA   *regexp.Regexp
	vmConfigRegexpNet    *regexp.Regexp
	vmConfigRegexpUnused *regexp.Regexp
)

func init() {
	vmConfigRegexpIDE, _ = regexp.Compile("^IDE[\\d]+$")
	vmConfigRegexpSCSI, _ = regexp.Compile("^SCSI[\\d]+$")
	vmConfigRegexpSATA, _ = regexp.Compile("^SATAIDE[\\d]+$")
	vmConfigRegexpNet, _ = regexp.Compile("^Net[\\d]+$")
	vmConfigRegexpUnused, _ = regexp.Compile("^Unused[\\d]+$")
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Otp      string `json:"otp,omitempty"`
	Path     string `json:"path,omitempty"`
	Privs    string `json:"privs,omitempty"`
	Realm    string `json:"realm,omitempty"`
}

type Session struct {
	Username            string `json:"username"`
	CsrfPreventionToken string `json:"CSRFPreventionToken,omitempty"`
	ClusterName         string `json:"clustername,omitempty"`
	Ticket              string `json:"ticket,omitempty"`
}

type Version struct {
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
	Version string `json:"version"`
}

type VNC struct {
	Cert   string
	Port   StringOrInt
	Ticket string
	UPID   string
	User   string
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
	ID         string  `jsont:"id"`
	Type       string  `json:"type"`
	Content    string  `json:",omitempty"`
	CPU        float64 `json:",omitempty"`
	Disk       uint64  `json:",omitempty"` // documented as string but this is an int
	HAstate    string  `json:",omitempty"`
	Level      string  `json:",omitempty"`
	MaxCPU     uint64  `json:",omitempty"`
	MaxDisk    uint64  `json:",omitempty"`
	MaxMem     uint64  `json:",omitempty"`
	Mem        uint64  `json:",omitempty"` // documented as string but this is an int
	Name       string  `json:",omitempty"`
	Node       string  `json:",omitempty"`
	PluginType string  `json:",omitempty"`
	Pool       string  `json:",omitempty"`
	Status     string  `json:",omitempty"`
	Storage    string  `json:",omitempty"`
	Uptime     uint64  `json:",omitempty"`
}

type NodeStatuses []*NodeStatus
type NodeStatus struct {
	// shared
	Status string `json:",omitempty"`
	Level  string `json:",omitempty"`
	ID     string `json:",omitempty"` // format "node/<name>"

	// from /nodes endpoint
	Node           string  `json:",omitempty"`
	MaxCPU         int     `json:",omitempty"`
	MaxMem         uint64  `json:",omitempty"`
	Disk           uint64  `json:",omitempty"`
	SSLFingerprint string  `json:"ssl_fingerprint,omitempty"`
	MaxDisk        uint64  `json:",omitempty"`
	Mem            uint64  `json:",omitempty"`
	CPU            float64 `json:",omitempty"`

	// from /cluster endpoint
	NodeID int    `json:",omitempty"` // the internal id of the node
	Name   string `json:",omitempty"`
	IP     string `json:",omitempty"`
	Online int    `json:",omitempty"`
	Local  int    `json:",omitempty"`
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

	Name      string
	Node      string
	NetIn     uint64
	CPUs      int
	DiskWrite uint64
	Status    string
	Lock      string `json:",omitempty"`
	VMID      StringOrUint64
	PID       StringOrUint64
	Netout    uint64
	Disk      uint64
	Uptime    uint64
	Mem       uint64
	CPU       float64
	MaxMem    uint64
	MaxDisk   uint64
	DiskRead  uint64
	QMPStatus string     `json:"qmpstatus,omitempty"`
	Template  IsTemplate // empty str if a vm, int 1 if a template
	HA        HA         `json:",omitempty"`
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
	MHZ     string
	Mode    string
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
type VirtualMachineOptions []*VirtualMachineOption
type VirtualMachineOption struct {
	Name  string
	Value interface{}
}

type VirtualMachineConfig struct {
	Cores   int
	Numa    int
	Memory  int
	Sockets int
	IDE2    string
	OSType  string
	SMBios1 string
	SCSIHW  string
	Net0    string
	Digest  string
	Meta    string
	SCSI0   string
	Boot    string
	VMGenID string
	Name    string

	IDEs map[string]string
	IDE0 string
	IDE1 string
	IDE3 string
	IDE4 string
	IDE5 string
	IDE6 string
	IDE7 string
	IDE8 string
	IDE9 string

	SCSIs map[string]string
	SCSI1 string
	SCSI2 string
	SCSI3 string
	SCSI4 string
	SCSI5 string
	SCSI6 string
	SCSI7 string
	SCSI8 string
	SCSI9 string

	SATAs map[string]string
	SATA0 string
	SATA1 string
	SATA2 string
	SATA3 string
	SATA4 string
	SATA5 string
	SATA6 string
	SATA7 string
	SATA8 string
	SATA9 string

	Nets map[string]string
	Net1 string
	Net2 string
	Net3 string
	Net4 string
	Net5 string
	Net6 string
	Net7 string
	Net8 string
	Net9 string

	Unuseds map[string]string
	Unused0 string
	Unused1 string
	Unused2 string
	Unused3 string
	Unused4 string
	Unused5 string
	Unused6 string
	Unused7 string
	Unused8 string
	Unused9 string
}

func (vmc *VirtualMachineConfig) MergeIDEs() map[string]string {
	if nil == vmc.IDEs {
		vmc.IDEs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpIDE.MatchString(fn) {
				vmc.IDEs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.IDEs
}
func (vmc *VirtualMachineConfig) MergeSCSIs() map[string]string {
	if nil == vmc.SCSIs {
		vmc.SCSIs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpSCSI.MatchString(fn) {
				vmc.SCSIs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SCSIs
}

func (vmc *VirtualMachineConfig) MergeSATAs() map[string]string {
	if nil == vmc.SATAs {
		vmc.SATAs = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpSATA.MatchString(fn) {
				vmc.SATAs[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.SATAs
}
func (vmc *VirtualMachineConfig) MergeNets() map[string]string {
	if nil == vmc.Nets {
		vmc.Nets = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpNet.MatchString(fn) {
				vmc.Nets[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.Nets
}
func (vmc *VirtualMachineConfig) MergeUnuseds() map[string]string {
	if nil == vmc.Unuseds {
		vmc.Unuseds = map[string]string{}
		t := reflect.TypeOf(*vmc)
		v := reflect.ValueOf(*vmc)
		count := v.NumField()

		for i := 0; i < count; i++ {
			fn := t.Field(i).Name
			fv := v.Field(i).String()
			//fmt.Println(fn, fv)
			if "" == fv {
				continue
			}
			if vmConfigRegexpUnused.MatchString(fn) {
				vmc.Unuseds[strings.ToLower(fn)] = fv
			}
		}
	}
	return vmc.Unuseds
}

type UPID string

type Tasks []*Tasks
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
	Name    string
	Node    string
	client  *Client
	CPUs    int
	Status  string
	VMID    StringOrUint64
	Uptime  uint64
	MaxMem  uint64
	MaxDisk uint64
	MaxSwap uint64
}

type ContainerStatuses []*ContainerStatus
type ContainerStatus struct {
	Data string `json:",omitempty"`
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
	Storage      string
}

type Volume interface {
	Delete() error
}

type ISOs []*ISO
type ISO struct{ Content }

type VzTmpls []*VzTmpl
type VzTmpl struct{ Content }

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
	str := strings.Replace(string(b), "\"", "", -1)
	parsed, err := strconv.ParseUint(str, 0, 64)
	if err != nil {
		return err
	}
	*d = StringOrInt(parsed)
	return nil
}

type StringOrUint64 uint64

func (d *StringOrUint64) UnmarshalJSON(b []byte) error {
	str := strings.Replace(string(b), "\"", "", -1)
	parsed, err := strconv.ParseUint(str, 0, 64)
	if err != nil {
		return err
	}
	*d = StringOrUint64(parsed)
	return nil
}

type NodeNetworks []*NodeNetwork
type NodeNetwork struct {
	client  *Client `json:"-"`
	Node    string  `json:"-"`
	NodeApi *Node   `json:"-"`

	Iface    string `json:"iface,omitempty"`
	BondMode string `json:"bond_mode,omitempty"`

	Autostart int `json:"autostart,omitempty"`

	CIDR            string `json:"cidr,omitempty"`
	CIDR6           string `json:"cidr6,omitempty"`
	Gateway         string `json:"gateway,omitempty"`
	Gateway6        string `json:"gateway6,omitempty"`
	Netmask         string `json:"netmask,omitempty"`
	Netmask6        string `json:"netmask6,omitempty"`
	BridgeVlanAware bool   `json:"bridge_vlan_aware,omitempty"`
	BridgePorts     string `json:"bridge_ports,omitempty"`
	Comments        string `json:"comments,omitempty"`
	Comments6       string `json:"comments6,omitempty"`
	BridgeStp       string `json:"bridge_stp,omitempty"`
	BridgeFd        string `json:"bridge_fd,omitempty"`
	BondPrimary     string `json:"bond-primary,omitempty"`

	Address  string `json:"address,omitempty"`
	Address6 string `json:"address6,omitempty"`
	Type     string `json:"type,omitempty"`
	Active   int    `json:"active,omitempty"`
	Method   string `json:"method,omitempty"`
	Method6  string `json:"method6,omitempty"`
	Priority int    `json:"priority,omitempty"`
}
