//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	//mage:import install
	"github.com/luthermonson/go-proxmox/mage/install"

	//mage:import test
	"github.com/luthermonson/go-proxmox/mage/test"
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

var Aliases = map[string]interface{}{
	"test":    test.Unit,
	"install": install.Dependencies,
}

// run everything for ci process (install deps, lint, coverage, build)
func Ci() error {
	fmt.Println("Running Continuous Integration...")
	mg.Deps(
		install.Dependencies,
		Lint,
		test.Coverage,
		test.Build)
	return nil
}

// run the linter
func Lint() error {
	mg.Deps(install.Golangcilint)
	fmt.Println("Running Linter...")
	return sh.RunV("golangci-lint", "run")
}

// validate env vars to run the testing suite
func Env() error {
	for k := range envConfig {
		if strings.Contains(strings.ToLower(k), "password") || strings.Contains(strings.ToLower(k), "secret") {
			fmt.Printf("%s: %s\n", k, strings.Repeat("*", len(os.Getenv(k))))
		} else {
			fmt.Printf("%s: %s\n", k, os.Getenv(k))
		}
	}

	return nil
}
