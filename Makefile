ci:
	golangci-lint run
	go test

.DEFAULT_GOAL := ci