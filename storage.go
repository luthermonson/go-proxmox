package proxmox

func NewISO(url string) *ISO {
	return &ISO{
		File{
			URL:    url,
			Format: "iso",
		},
	}
}

// TODO https://192.168.1.6:8006/api2/extjs/nodes/i7/storage/local/content//local:vztmpl/alpine-3.11-default_20200425_amd64.tar.xz?delay=5

func (c *File) Delete() error {
	return nil
}
