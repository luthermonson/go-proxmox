package install

import (
	"fmt"
	"os/exec"

	"github.com/magefile/mage/mg"
)

const (
	GolangCILintVersion = "v1.50.1"
)

func Dependencies() error {
	fmt.Println("Installing Dependencies...")
	mg.Deps(Golangcilint)

	return nil
}

func Golangcilint() error {
	fmt.Println("Installing GolangCI Lint...")
	cmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@"+GolangCILintVersion)
	return cmd.Run()
}
