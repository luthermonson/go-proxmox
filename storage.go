package proxmox

func NewISO(url string) *ISO {
	return &ISO{
		File{
			URL:    url,
			Format: "iso",
		},
	}
}

func (c *File) Delete() error {
	return nil
}
