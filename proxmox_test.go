package proxmox

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestingData struct {
	client  *Client
	node    *Node
	storage *Storage

	username      string
	password      string
	tokenID       string
	secret        string
	otp           string
	nodeName      string
	nodeStorage   string
	applianceName string
	isoURL        string
}

var (
	td     TestingData
	logger = LeveledLogger{Level: LevelDebug}

	insecureHTTPClient = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

func init() {
	td.username = os.Getenv("PROXMOX_USERNAME")
	td.password = os.Getenv("PROXMOX_PASSWORD")
	td.otp = os.Getenv("PROXMOX_OTP")
	td.tokenID = os.Getenv("PROXMOX_TOKENID")
	td.secret = os.Getenv("PROXMOX_SECRET")
	td.nodeName = os.Getenv("PROXMOX_NODE_NAME")
	td.nodeStorage = os.Getenv("PROXMOX_NODE_STORAGE")
	td.applianceName = os.Getenv("PROXMOX_CONTAINER_TEMPLATE") // alpine-3.14-default_20210623_amd64.tar.xz
	td.isoURL = os.Getenv("PROXMOX_ISO_URL")                   // https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/x86_64/alpine-virt-3.14.1-x86_64.iso

	td.client = ClientFromLogins()
	var err error
	td.node, err = td.client.Node(td.nodeName)
	if err != nil {
		panic(err)
	}

	td.storage, err = td.node.Storage(td.nodeStorage)
	if err != nil {
		panic(err)
	}
}

func ClientFromEnv() *Client {
	return NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithLogger(&logger),
	)
}

func ClientFromLogins() *Client {
	client := NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithLogins(td.username, td.password),
		WithLogger(&logger),
	)

	return client
}

func ClientFromToken() *Client {
	return NewClient(os.Getenv("PROXMOX_URL"),
		WithClient(&insecureHTTPClient),
		WithAPIToken(td.tokenID, td.secret),
		WithLogger(&logger),
	)
}

func TestVersion(t *testing.T) {
	client := ClientFromLogins()
	version, err := client.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, version.Version)
}
