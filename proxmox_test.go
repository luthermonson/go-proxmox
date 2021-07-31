package proxmox

import (
	"crypto/tls"
	"net/http"
	"os"
)

var creds Credentials
var client *Client

func init() {
	creds.Username = os.Getenv("PROXMOX_USERNAME")
	creds.Password = os.Getenv("PROXMOX_PASSWORD")
	creds.Otp = os.Getenv("PROXMOX_OTP")
	client = NewClient(os.Getenv("PROXMOX_URL"), WithClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}))
}
