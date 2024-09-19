package impl

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strconv"

	"github.com/luthermonson/go-proxmox"
)

var client *proxmox.Client

func Init() {
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	credentials := proxmox.Credentials{
		Username: os.Getenv("PROXMOX_USERNAME"),
		Password: os.Getenv("PROXMOX_PASSWORD"),
	}
	client = proxmox.NewClient(os.Getenv("PROXMOX_URL"),
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithCredentials(&credentials),
	)
}

func GetVm() (*proxmox.VirtualMachine, error) {
	node, err := client.Node(context.Background(), os.Getenv("PROXMOX_NODE"))
	if err != nil {
		return nil, err
	}

	vmId, err := strconv.Atoi(os.Getenv("PROXMOX_VM"))
	if err != nil {
		return nil, err
	}

	vm, err := node.VirtualMachine(context.Background(), vmId)
	if err != nil {
		return nil, err
	}

	return vm, nil
}
