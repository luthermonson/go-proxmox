package integration

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/luthermonson/go-proxmox"
	"github.com/stretchr/testify/assert"
)

type TestingData struct {
	client    *proxmox.Client
	node      *proxmox.Node
	storage   *proxmox.Storage
	appliance *proxmox.Appliance

	username    string
	password    string
	tokenID     string
	secret      string
	otp         string
	nodeName    string
	nodeStorage string
	isoURL      string
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

	tinycoreURL     = "https://github.com/luthermonson/go-proxmox/releases/download/tests/tinycore.iso"
	ubuntuURL       = "https://releases.ubuntu.com/20.04.3/ubuntu-20.04.3-desktop-amd64.iso"
	alpineAppliance = "http://download.proxmox.com/images/system/alpine-3.17-default_20221129_amd64.tar.xz"
)

func init() {
	td.username = os.Getenv("PROXMOX_USERNAME")
	td.password = os.Getenv("PROXMOX_PASSWORD")
	td.otp = os.Getenv("PROXMOX_OTP")
	td.tokenID = os.Getenv("PROXMOX_TOKENID")
	td.secret = os.Getenv("PROXMOX_SECRET")
	td.nodeName = os.Getenv("PROXMOX_NODE_NAME")
	td.nodeStorage = os.Getenv("PROXMOX_NODE_STORAGE")
	td.isoURL = os.Getenv("PROXMOX_ISO_URL") // https://dl-cdn.alpinelinux.org/alpine/v3.14/releases/x86_64/alpine-virt-3.14.1-x86_64.iso

	if td.nodeName == "" {
		return
	}

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

func nameGenerator(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	rstr := fmt.Sprintf("%x", b)[:length]
	return fmt.Sprintf("go-proxmox-%s", rstr)
}

func downloadFile(src, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func createTestISO(file string) error {
	//making iso
	blocksize := int64(2048)
	iso, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, os.FileMode(0700))
	if err != nil {
		return err
	}
	defer iso.Close()

	fs, err := iso9660.Create(iso, 0, 0, blocksize, "")
	if err != nil {
		return err
	}

	err = fs.Mkdir("/")
	if err != nil {
		return err
	}

	return fs.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		VolumeIdentifier: "cidata",
	})
}

func ClientFromEnv() *proxmox.Client {
	return proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithClient(&insecureHTTPClient),
		proxmox.WithLogger(&logger),
	)
}

func ClientFromLogins() *proxmox.Client {
	client := proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithClient(&insecureHTTPClient),
		proxmox.WithLogins(td.username, td.password),
		proxmox.WithLogger(&logger),
	)

	return client
}

func ClientFromToken() *proxmox.Client {
	return proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithClient(&insecureHTTPClient),
		proxmox.WithAPIToken(td.tokenID, td.secret),
		proxmox.WithLogger(&logger),
	)
}

func TestVersion(t *testing.T) {
	client := ClientFromLogins()
	version, err := client.Version()
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
