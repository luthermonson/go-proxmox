package integration

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"

	proxmox "github.com/luthermonson/go-proxmox"
	"github.com/stretchr/testify/assert"
)

type TestingData struct {
	client    *proxmox.Client
	node      *proxmox.Node
	storage   *proxmox.Storage
	appliance *proxmox.Appliance

	username        string
	password        string
	tokenID         string
	secret          string
	otp             string
	nodeName        string
	nodeStorage     string
	isoURL          string
	appliancePrefix string
	vncVmId         int
}

var (
	td     TestingData
	logger = proxmox.LeveledLogger{Level: proxmox.LevelDebug}

	insecureHTTPClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func init() {
	var err error

	td.username = os.Getenv("PROXMOX_USERNAME")
	td.password = os.Getenv("PROXMOX_PASSWORD")
	td.otp = os.Getenv("PROXMOX_OTP")
	td.tokenID = os.Getenv("PROXMOX_TOKENID")
	td.secret = os.Getenv("PROXMOX_SECRET")
	td.nodeName = os.Getenv("PROXMOX_NODE_NAME")
	td.nodeStorage = os.Getenv("PROXMOX_NODE_STORAGE")
	td.isoURL = os.Getenv("PROXMOX_ISO_URL") // https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/x86_64/alpine-virt-3.14.1-x86_64.iso
	td.appliancePrefix = "alpine-virt-3.14.1"
	vncVmId, err := strconv.Atoi(os.Getenv("PROXMOX_VNC_VMID"))
	if err == nil {
		td.vncVmId = vncVmId
	}

	if td.nodeName == "" {
		return
	}

	td.client = ClientFromLogins()
	ctx := context.Background()

	td.node, err = td.client.Node(ctx, td.nodeName)
	if err != nil {
		panic(err)
	}

	td.storage, err = td.node.Storage(ctx, td.nodeStorage)
	if err != nil {
		panic(err)
	}
}

func ClientFromEnv() *proxmox.Client {
	return proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithLogger(&logger),
	)
}

func ClientFromLogins() *proxmox.Client {
	client := proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithCredentials(&proxmox.Credentials{
			Username: td.username,
			Password: td.password,
		}),
		proxmox.WithLogger(&logger),
	)

	return client
}

func ClientFromToken() *proxmox.Client {
	return proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(td.tokenID, td.secret),
		proxmox.WithLogger(&logger),
	)
}

func ClientFromTicket() *proxmox.Client {
	return proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(td.tokenID, td.secret),
		proxmox.WithLogger(&logger),
	)
}

func TestVersion(t *testing.T) {
	client := ClientFromLogins()
	ctx := context.Background()
	version, err := client.Version(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, version.Version)
}

func TestLogUnmarshall(t *testing.T) {
	payload := `[{"t":"starting file import from: /var/tmp/pveupload-f07236b021b8decb513b8735d302b6e0","n":1},{"t":"target node: i7","n":2},{"t":"target file: /var/lib/vz/template/iso/69adc234b08d.iso","n":3},{"n":4,"t":"file size is: 43018"},{"t":"command: cp -- /var/tmp/pveupload-f07236b021b8decb513b8735d302b6e0 /var/lib/vz/template/iso/69adc234b08d.iso","n":5},{"t":"finished file import successfully","n":6},{"t":"TASK OK","n":7}]`
	var log proxmox.Log
	assert.Nil(t, json.Unmarshal([]byte(payload), &log))
	assert.Equal(t, log[0], "starting file import from: /var/tmp/pveupload-f07236b021b8decb513b8735d302b6e0")
	assert.Equal(t, log[6], "TASK OK")
}
