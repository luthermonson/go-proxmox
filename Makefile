ci:
	golangci-lint run
	go test -tags "nodes containers vms"

.DEFAULT_GOAL := ci