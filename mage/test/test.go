package test

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

func Unit() error {
	fmt.Println("Running Tests...")
	return sh.RunV("go", "test")
}

func Build() error {
	fmt.Println("Running Build...")
	return sh.RunV("go", "build", "-tags", "test")
}

func Coverage() error {
	fmt.Println("Running Tests with Coverage...")
	return sh.RunV("go", "test", "-race", "-coverprofile=coverage.txt", "-covermode=atomic")
}

func Integration() error {
	fmt.Println("Running Integration Tests against a PVE Cluster...")
	return sh.RunV("go", "test", "./tests/integration", "-tags", "nodes containers vms")
}
