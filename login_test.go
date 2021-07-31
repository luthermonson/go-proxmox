package proxmox

import (
	"fmt"
	"testing"
)

func init() {

}

func TestLogin(t *testing.T) {
	session, err := client.Login(creds)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v", session)
}
