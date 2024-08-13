package proxmox

import (
	"context"
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
	ID         string  `jsont:"id"`
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
	MHZ     interface{}
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
	MaxCPU  int
	MaxMem  uint64
	Disk    int
	MaxDisk uint64
	Time    uint64
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
	VMGenID     string `json:"vmgenid,omitempty"`
	Hookscript  string `json:"hookscript,omitempty"`
	Hotplug     string `json:"hotplug,omitempty"`
	Template    int    `json:"template,omitempty"`
	Agent       string `json:"agent,omitempty"`
	Autostart   int    `json:"autostart,omitempty"`
	Tablet      int    `json:"tablet,omitempty"`
	KVM         int    `json:"kvm,omitempty"`

	Tags      string   `json:"tags,omitempty"`
	TagsSlice []string `json:"-"` // internal helper to manage tags easier

	Protection int    `json:"protection,omitempty"`
	Lock       string `json:"lock,omitempty"`

	// Boot configuration
	Boot   string `json:"boot,omitempty"`
	OnBoot int    `json:"onboot,omitempty"`

	// Qemu general specs
	OSType  string `json:"ostype,omitempty"`
	Machine string `json:"machine,omitempty"`
	Args    string `json:"args,omitempty"`

	// Qemu firmware specs
	Bios     string `json:"bios,omitempty"`
	EFIDisk0 string `json:"efidisk0,omitempty"`
	SMBios1  string `json:"smbios1,omitempty"`
	Acpi     int    `json:"acpi,omitempty"`

	// Qemu CPU specs
	Sockets  int             `json:"sockets,omitempty"`
	Cores    int             `json:"cores,omitempty"`
	CPU      string          `json:"cpu,omitempty"`
	CPULimit StringOrFloat64 `json:"cpulimit,omitempty"`
	CPUUnits int             `json:"cpuunits,omitempty"`
	Vcpus    int             `json:"vcpus,omitempty"`
	Affinity string          `json:"affinity,omitempty"`

	// Qemu memory specs
	Numa      int         `json:"numa,omitempty"`
	Memory    StringOrInt `json:"memory,omitempty"` // See commit 7f8c808772979f274cdfac1dc7264771a3b7a7ae on qemu-server
	Hugepages string      `json:"hugepages,omitempty"`
	Balloon   int         `json:"balloon,omitempty"`

	// Other Qemu devices
	VGA       string `json:"vga,omitempty"`
	SCSIHW    string `json:"scsihw,omitempty"`
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
	SSHKeys      string `json:"sshkeys,omitempty"`
	CICustom     string `json:"cicustom,omitempty"`

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

type VirtualMachineMigrateOptions struct {
	Target           string    `json:"target"`
	BWLimit          uint64    `json:"bwlimit,omitempty"`
	Force            IntOrBool `json:"force,omitempty"`
	MigrationNetwork string    `json:"migration_network,omitempty"`
	MigrationType    string    `json:"migration_type,omitempty"`
	Online           IntOrBool `json:"online,omitempty"`
	TargetStorage    string    `json:"targetstorage,omitempty"`
	WithLocalDisks   IntOrBool `json:"with-local-disks,omitempty"`
}

type ContainerMigrateOptions struct {
	Target  string    `json:"target"`
	BWLimit uint64    `json:"bwlimit,omitempty"`
	Online  IntOrBool `json:"online,omitempty"`
	Restart IntOrBool `json:"restart,omitempty"`
}

type VirtualMachineCloneOptions struct {
	NewID       int    `json:"newid"`
	BWLimit     uint64 `json:"bwlimit,omitempty"`
	Description string `json:"description,omitempty"`
	Format      string `json:"format,omitempty"`
	Full        uint8  `json:"full,omitempty"`
	Name        string `json:"name,omitempty"`
	Pool        string `json:"pool,omitempty"`
	SnapName    string `json:"snapname,omitempty"`
	Storage     string `json:"storage,omitempty"`
	Target      string `json:"target,omitempty"`
}

type VirtualMachineMoveDiskOptions struct {
	Disk         string `json:"disk"`
	BWLimit      uint64 `json:"bwlimit,omitempty"`
	Delete       uint8  `json:"delete,omitempty"`
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
	Tags    string
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
	BWLimit     uint64 `json:"bwlimit,omitempty"`
	Description string `json:"description,omitempty"`
	Full        uint8  `json:"full,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Pool        string `json:"pool,omitempty"`
	SnapName    string `json:"snapname,omitempty"`
	Storage     string `json:"storage,omitempty"`
	Target      string `json:"target,omitempty"`
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
	if str == "" {
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
	str := strings.Replace(string(b), "\"", "", -1)
	if str == "" {
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
	str := strings.Replace(string(b), "\"", "", -1)
	if str == "" {
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
	if *b == true {
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

	Slaves   string `json:"slaves,omitempty"`
	Address  string `json:"address,omitempty"`
	Address6 string `json:"address6,omitempty"`
	Type     string `json:"type,omitempty"`
	Active   int    `json:"active,omitempty"`
	Method   string `json:"method,omitempty"`
	Method6  string `json:"method6,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

type AgentNetworkIPAddress struct {
	IPAddressType string `json:"ip-address-type"` // ipv4 ipv6
	IPAddress     string `json:"ip-address"`
	Prefix        int    `json:"prefix"`
	MacAddress    string `json:"mac-address"`
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
	Exited       int    `json:"exited"`
	ErrData      string `json:"err-data"`
	ErrTruncated bool   `json:"err-truncated"`
	ExitCode     int    `json:"exitcode"`
	OutData      string `json:"out-data"`
	OutTruncated string `json:"out-truncated"`
	Signal       bool   `json:"signal"`
}

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
	return 1 == r.Enable
}

type FirewallNodeOption struct {
	Enable                           bool   `json:"enable,omitempty"`
	LogLevelIn                       string `json:"log_level_in,omitempty"`
	LogLevelOut                      string `json:"log_level_out,omitempty"`
	LogNfConntrack                   bool   `json:"log_nf_conntrack,omitempty"`
	Ntp                              bool   `json:"ntp,omitempty"`
	NFConntrackAllowInvalid          bool   `json:"nf_conntrack_allow_invalid,omitempty"`
	NFConntrackMax                   int    `json:"nf_conntrack_max,omitempty"`
	NFConntrackTCPTimeoutEstablished int    `json:"nf_conntrack_tcp_timeout_established,omitempty"`
	NFConntrackTCPTimeoutSynRecv     int    `json:"nf_conntrack_tcp_timeout_syn_recv,omitempty"`
	Nosmurfs                         bool   `json:"nosmurfs,omitempty"`
	ProtectionSynflood               bool   `json:"protection_synflood,omitempty"`
	ProtectionSynfloodBurst          int    `json:"protection_synflood_burst,omitempty"`
	ProtectionSynfloodRate           int    `json:"protection_synflood_rate,omitempty"`
	SmurfLogLevel                    string `json:"smurf_log_level,omitempty"`
	TCPFlagsLogLevel                 string `json:"tcp_flags_log_level,omitempty"`
	TCPflags                         bool   `json:"tcpflags,omitempty"`
}

type FirewallVirtualMachineOption struct {
	Enable      bool   `json:"enable,omitempty"`
	Dhcp        bool   `json:"dhcp,omitempty"`
	Ipfilter    bool   `json:"ipfilter,omitempty"`
	LogLevelIn  string `json:"log_level_in,omitempty"`
	LogLevelOut string `json:"log_level_out,omitempty"`
	Macfilter   bool   `json:"macfilter,omitempty"`
	Ntp         bool   `json:"ntp,omitempty"`
	PolicyIn    string `json:"policy_in,omitempty"`
	PolicyOut   string `json:"policy_out,omitempty"`
	Radv        bool   `json:"radv,omitempty"`
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
	context context.Context
	PoolID  string            `json:"poolid,omitempty"`
	Comment string            `json:"comment,omitempty"`
	Members []ClusterResource `json:"members,omitempty"`
}

type PoolUpdateOption struct {
	Comment string `json:"comment,omitempty"`
	// Delete objects rather than adding them
	Delete bool `json:"delete,omitempty"`
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
	Verify         IntOrBool `json:"verify,omitempty"`
}

// DomainSyncOptions see details https://pve.proxmox.com/pve-docs/api-viewer/#/access/domains/{realm}/sync
type DomainSyncOptions struct {
	DryRun         bool   `json:"dry-run,omitempty"`
	EnableNew      bool   `json:"enable-new,omitempty"`
	RemoveVanished string `json:"remove-vanished,omitempty"`
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
	Enable         IntOrBool        `json:"enable,omitempty"`
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

type Tokens []*Token
type Token struct {
	TokenID string    `json:"tokenid,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Expire  int       `json:"expire,omitempty"`
	Privsep IntOrBool `json:"privsep,omitempty"`
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
	Path      string    `json:",omitempty"`
	RoleID    string    `json:",omitempty"`
	Type      string    `json:",omitempty"`
	UGID      string    `json:",omitempty"`
	Propagate IntOrBool `json:",omitempty"`
}

type ACLOptions struct {
	Path      string    `json:",omitempty"`
	Roles     string    `json:",omitempty"`
	Groups    string    `json:",omitempty"`
	Users     string    `json:",omitempty"`
	Tokens    string    `json:",omitempty"`
	Propagate IntOrBool `json:",omitempty"`
	Delete    IntOrBool `json:",omitempty"` // true to delete the ACL
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
	VerifyCertificates IntOrBool `json:"verify-certificates,omitempty"`
}

type StorageContent struct {
	Format       string `json:"format,omitempty"`
	Size         uint64 `json:"size,omitempty"`
	Volid        string `json:"volid,omitempty"`
	Ctime        uint64 `json:"ctime,omitempty"`
	Encryption   string `json:"encryption,omitempty"`
	Notes        string `json:"notes,omitempty"`
	Parent       string `json:"parent,omitempty"`
	Protection   bool   `json:"protection,omitempty"`
	Used         uint64 `json:"used,omitempty"`
	Verification string `json:"verification,omitempty"`
	VMID         uint64 `json:"vmid,omitempty"`
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
	Enable    bool     `json:"enable,omitempty"`
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
	All                bool                                   `json:"all,omitempty"`
	BwLimit            uint                                   `json:"bwlimit,omitempty"`
	Compress           VirtualMachineBackupCompress           `json:"compress,omitempty"`
	DumpDir            string                                 `json:"dumpDir,omitempty"`
	Exclude            string                                 `json:"exclude,omitempty"`
	ExcludePath        []string                               `json:"exclude-path,omitempty"`
	IoNice             uint                                   `json:"ionice,omitempty"`
	LockWait           uint                                   `json:"lockwait,omitempty"`
	MailTo             string                                 `json:"mailto,omitempty"`
	Mode               VirtualMachineBackupMode               `json:"mode,omitempty"`
	Node               string                                 `json:"node,omitempty"`
	NotesTemplate      string                                 `json:"notes-template,omitempty"`
	NotificationPolicy VirtualMachineBackupNotificationPolicy `json:"notification-policy,omitempty"`
	NotificationTarget string                                 `json:"notification-target,omitempty"`
	Performance        string                                 `json:"performance,omitempty"`
	Pigz               int                                    `json:"pigz,omitempty"`
	Pool               string                                 `json:"pool,omitempty"`
	Protected          string                                 `json:"protected,omitempty"`
	PruneBackups       string                                 `json:"prune-backups,omitempty"`
	Quiet              bool                                   `json:"quiet,omitempty"`
	Remove             bool                                   `json:"remove,omitempty"`
	Script             string                                 `json:"script,omitempty"`
	StdExcludes        bool                                   `json:"stdexcludes,omitempty"`
	StdOut             bool                                   `json:"stdout,omitempty"`
	Stop               bool                                   `json:"stop,omitempty"`
	StopWait           uint                                   `json:"stopwait,omitempty"`
	Storage            string                                 `json:"storage,omitempty"`
	TmpDir             string                                 `json:"tmpdir,omitempty"`
	VMID               uint64                                 `json:"vmid,omitempty"`
	Zstd               uint                                   `json:"zstd,omitempty"`
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
	Sockets    uint64 `json:"sockets,string"`
	SSHKeys    string `json:"sshkeys"`
	VmgenID    string `json:"vmgenid"`

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
