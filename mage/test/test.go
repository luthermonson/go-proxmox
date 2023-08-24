package test

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

// Unit run all unit tests
func Unit() error {
	fmt.Println("Running Tests...")
	return sh.RunV("go", "test")
}

// Build run a test build to confirm no compilation errors
func Build() error {
	fmt.Println("Running Build...")
	return sh.RunV("go", "build", "-tags", "test")
}

// Coverage run all unit tests and output coverage
func Coverage() error {
	fmt.Println("Running Tests with Coverage...")
	return sh.RunV("go", "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic")
}

// Integration run all integration tests against a pve node, see ./tests/integration
func Integration() error {
	fmt.Println("Running Integration Tests against a PVE Cluster...")
	return sh.RunV("go", "test", "./tests/integration", "-tags", "nodes containers vms")
}
