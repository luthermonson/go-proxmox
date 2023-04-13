package proxmox

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

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
	// PVE Metadata
	Digest      string `json:"digest"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Meta        string `json:"meta,omitempty"`
	VMGenID     string `json:"vmgenid,omitempty"`
	Hookscript  string `json:"hookscript,omitempty"`
	Hotplug     string `json:"hotplug,omitempty"`
	Template    int    `json:"template,omitempty"`

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
	Sockets  int    `json:"sockets,omitempty"`
	Cores    int    `json:"cores,omitempty"`
	CPU      string `json:"cpu,omitempty"`
	CPULimit int    `json:"cpulimit,omitempty"`
	CPUUnits int    `json:"cpuunits,omitempty"`
	Vcpus    int    `json:"vcpus,omitempty"`
	Affinity string `json:"affinity,omitempty"`

	// Qemu memory specs
	Numa      int    `json:"numa,omitempty"`
	Memory    int    `json:"memory,omitempty"`
	Hugepages string `json:"hugepages,omitempty"`
	Balloon   int    `json:"balloon,omitempty"`

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

	numeric := regexp.MustCompile(`\d`).MatchString(str)
	if !numeric {
		str = "0"
	}

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

	numeric := regexp.MustCompile(`\d`).MatchString(str)
	if !numeric {
		str = "0"
	}

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
	NodeAPI *Node   `json:"-"`

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

type AgentNetworkIPAddress struct {
	IPAddressType string `json:"ip-address-type"` //ipv4 ipv6
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
	Exited       bool   `json:"exited"`
	ErrData      string `json:"err-data"`
	ErrTruncated bool   `json:"err-truncated"`
	ExitCode     int    `json:"exit-code"`
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
