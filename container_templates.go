package proxmox

import "fmt"

func (c *ContainerTemplate) Download(template, storage string) (ret string, err error) {
	return ret, c.client.Post(fmt.Sprintf("/nodes/%s/aplinfo", c.Node), map[string]string{
		"template": template,
		"storage":  storage,
	}, &ret)
}

// TODO https://192.168.1.6:8006/api2/extjs/nodes/i7/storage/local/content//local:vztmpl/alpine-3.11-default_20200425_amd64.tar.xz?delay=5
func (c *ContainerTemplate) Delete() (ret string, err error) {
	return ret, err
}
