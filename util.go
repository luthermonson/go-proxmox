package proxmox

func AsPtr[T any](input T) *T {
	return &input
}
