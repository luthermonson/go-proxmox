package record

import (
	"fmt"
	"os"
	"strconv"
)

// Config is the env-driven configuration for the recording pipeline. All
// fields are required for `mage record:pveN` targets; missing values cause an
// early, descriptive error rather than a partially-completed nested VM.
//
// Naming is consistent with the existing magefile envConfig where possible
// (PROXMOX_URL, PROXMOX_TOKENID, PROXMOX_SECRET, PROXMOX_NODE_NAME,
// PROXMOX_NODE_STORAGE) so a single .env covers the integration tests and
// the recording pipeline.
type Config struct {
	// OuterURL is the Proxmox API URL of the outer host that runs the
	// nested VM (PROXMOX_URL).
	OuterURL string
	// OuterTokenID is the API token id on the outer host
	// (PROXMOX_TOKENID), e.g. "automation@pve!recorder".
	OuterTokenID string
	// OuterSecret is the API token secret on the outer host
	// (PROXMOX_SECRET).
	OuterSecret string
	// OuterNode is the outer-host node name to spin nested VMs on
	// (PROXMOX_NODE_NAME).
	OuterNode string
	// OuterStorage is the outer-host storage to put ISOs on
	// (PROXMOX_NODE_STORAGE), typically "local".
	OuterStorage string
	// OuterInsecure skips TLS verification on the outer-host API client
	// (PROXMOX_INSECURE). Defaults to true because Proxmox ships
	// self-signed certs.
	OuterInsecure bool

	// SSHHost is the hostname used to SSH into the outer host so we can
	// run proxmox-auto-install-assistant (PROXMOX_RECORDER_SSH_HOST).
	// Defaults to the host portion of OuterURL.
	SSHHost string
	// SSHUser is the unix user to SSH as (PROXMOX_RECORDER_SSH_USER).
	// Defaults to "root".
	SSHUser string
	// SSHKey is the path to the private key used for SSH
	// (PROXMOX_RECORDER_SSH_KEY). Defaults to ~/.ssh/id_ed25519.
	SSHKey string

	// NestedVMID is the VMID used for the nested PVE VM
	// (PROXMOX_RECORDER_VMID). Single VMID is reused across pveN runs
	// because the VM is destroyed after each. Default 9999.
	NestedVMID int
	// NestedName is the VM name on the outer host
	// (PROXMOX_RECORDER_NAME). Default "go-proxmox-recorder".
	NestedName string
	// NestedBridge is the bridge the nested VM connects to
	// (PROXMOX_RECORDER_BRIDGE). Default "vmbr0".
	NestedBridge string
	// NestedIP is the static IPv4 (with prefix) the nested VM is
	// configured with via answer.toml (PROXMOX_RECORDER_IP).
	// Default "192.0.2.10/24" — TEST-NET-1 per RFC 5737.
	NestedIP string
	// NestedGateway is the gateway the nested VM uses
	// (PROXMOX_RECORDER_GATEWAY). Default "192.0.2.1".
	NestedGateway string
	// NestedRootPassword is the password set on the nested PVE root
	// account (PROXMOX_RECORDER_ROOT_PASSWORD). Default "recorder!1A".
	// This is throwaway; the nested VM is destroyed after recording.
	NestedRootPassword string
	// NestedTokenSecret is the token secret created during seed for the
	// recorder@pve user (PROXMOX_RECORDER_TOKEN_SECRET). Default
	// "00000000-0000-0000-0000-000000000000". Scrubbed from cassettes.
	NestedTokenSecret string
	// NestedCPU and NestedRAM size the nested VM (PROXMOX_RECORDER_CPU,
	// PROXMOX_RECORDER_RAM_MB). Defaults 4 / 4096.
	NestedCPU int
	NestedRAM int
	// NestedDiskGB sizes the boot disk (PROXMOX_RECORDER_DISK_GB).
	// Default 32.
	NestedDiskGB int
}

// LoadConfig reads the env into a Config and applies defaults. Missing
// required values produce a descriptive error.
func LoadConfig() (*Config, error) {
	c := &Config{
		OuterURL:           os.Getenv("PROXMOX_URL"),
		OuterTokenID:       os.Getenv("PROXMOX_TOKENID"),
		OuterSecret:        os.Getenv("PROXMOX_SECRET"),
		OuterNode:          os.Getenv("PROXMOX_NODE_NAME"),
		OuterStorage:       os.Getenv("PROXMOX_NODE_STORAGE"),
		SSHHost:            os.Getenv("PROXMOX_RECORDER_SSH_HOST"),
		SSHUser:            envOr("PROXMOX_RECORDER_SSH_USER", "root"),
		SSHKey:             envOr("PROXMOX_RECORDER_SSH_KEY", defaultSSHKey()),
		NestedName:         envOr("PROXMOX_RECORDER_NAME", "go-proxmox-recorder"),
		NestedBridge:       envOr("PROXMOX_RECORDER_BRIDGE", "vmbr0"),
		NestedIP:           envOr("PROXMOX_RECORDER_IP", "192.0.2.10/24"),
		NestedGateway:      envOr("PROXMOX_RECORDER_GATEWAY", "192.0.2.1"),
		NestedRootPassword: envOr("PROXMOX_RECORDER_ROOT_PASSWORD", "recorder!1A"),
		NestedTokenSecret:  envOr("PROXMOX_RECORDER_TOKEN_SECRET", "00000000-0000-0000-0000-000000000000"),
	}
	c.OuterInsecure = envBool("PROXMOX_INSECURE", true)
	c.NestedVMID = envInt("PROXMOX_RECORDER_VMID", 9999)
	c.NestedCPU = envInt("PROXMOX_RECORDER_CPU", 4)
	c.NestedRAM = envInt("PROXMOX_RECORDER_RAM_MB", 4096)
	c.NestedDiskGB = envInt("PROXMOX_RECORDER_DISK_GB", 32)

	// Default SSH host to the outer URL's hostname (not host:port — SSH
	// runs on :22, not the API's :8006). NewSSHClient will append :22.
	if c.SSHHost == "" {
		c.SSHHost = stripPort(hostOnly(c.OuterURL))
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// Validate returns an error if a required field is missing.
func (c *Config) Validate() error {
	missing := []string{}
	if c.OuterURL == "" {
		missing = append(missing, "PROXMOX_URL")
	}
	if c.OuterTokenID == "" {
		missing = append(missing, "PROXMOX_TOKENID")
	}
	if c.OuterSecret == "" {
		missing = append(missing, "PROXMOX_SECRET")
	}
	if c.OuterNode == "" {
		missing = append(missing, "PROXMOX_NODE_NAME")
	}
	if c.OuterStorage == "" {
		missing = append(missing, "PROXMOX_NODE_STORAGE")
	}
	if len(missing) > 0 {
		return fmt.Errorf("recorder: missing required env vars: %v", missing)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func defaultSSHKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.ssh/id_ed25519"
}

// stripPort removes ":<port>" from a host:port string, leaving only the host.
// Returns the input unchanged if no colon is present.
func stripPort(host string) string {
	for i := 0; i < len(host); i++ {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

// hostOnly extracts host:port from a URL like https://host:port/api2/json.
func hostOnly(u string) string {
	if u == "" {
		return ""
	}
	s := u
	for _, p := range []string{"https://", "http://"} {
		if len(s) >= len(p) && s[:len(p)] == p {
			s = s[len(p):]
			break
		}
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return s[:i]
		}
	}
	return s
}
