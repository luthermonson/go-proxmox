//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	envConfig = map[string]struct{}{
		"PROXMOX_URL":              {},
		"PROXMOX_USERNAME":         {},
		"PROXMOX_PASSWORD":         {},
		"PROXMOX_OTP":              {},
		"PROXMOX_TOKENID":          {},
		"PROXMOX_SECRET":           {},
		"PROXMOX_NODE_NAME":        {},
		"PROXMOX_NODE_STORAGE":     {},
		"PROXMOX_APPLIANCE_PREFIX": {},
		"PROXMOX_ISO_URL":          {},
	}
)

// lint and run all tests against a proxmox server using env var config, see env for more details
func Ci() error {
	fmt.Println("Running Continuous Integration...")
	mg.Deps(Lint)
	mg.Deps(Test)
	return nil
}

func Lint() error {
	mg.Deps(InstallDeps)
	fmt.Println("Running Linter...")
	return sh.RunV("golangci-lint", "run")
}

func Test() error {
	fmt.Println("Running Tests...")
	return sh.RunV("go", "test", "-tags", "\"nodes containers vms\"")
}

// validate all env vars to run the testing suite
func Env() error {
	for k, _ := range envConfig {
		if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "secret") {
			fmt.Printf("%s: %s\n", k, strings.Repeat("*", len(os.Getenv(k))))
		} else {
			fmt.Printf("%s: %s\n", k, os.Getenv(k))
		}
	}

	return nil
}

func InstallDeps() error {
	fmt.Println("Installing Deps...")
	installs := []string{
		"github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0",
	}

	for _, pkg := range installs {
		fmt.Printf("Installing %s\n", pkg)
		cmd := exec.Command("go", "install", pkg)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
