// Command apidrift reports where this library's return types drift from the
// upstream Proxmox VE API schema. It fetches apidoc.js (or reads a local copy),
// diffs each endpoint's declared return properties against the Go struct the
// library decodes it into (per internal/apidrift's curated registry), and
// writes a human-readable report plus machine-readable JSON/Markdown artifacts.
//
// Exit codes: 0 = no drift (or -report-only), 1 = drift found, 2 = tool error.
//
// Usage:
//
//	go run ./cmd/apidrift                      # fetch upstream, print report, fail on drift
//	go run ./cmd/apidrift -schema apidoc.js    # use a local/cached schema (no network)
//	go run ./cmd/apidrift -md drift.md -json drift.json -report-only
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/luthermonson/go-proxmox/internal/apidrift"
)

const upstreamURL = "https://pve.proxmox.com/pve-docs/api-viewer/apidoc.js"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		schemaPath = flag.String("schema", "", "path to a local apidoc.js (skips the network fetch)")
		url        = flag.String("url", upstreamURL, "upstream apidoc.js URL (used when -schema is empty)")
		mdOut      = flag.String("md", "", "write the Markdown report to this path")
		jsonOut    = flag.String("json", "", "write the JSON report to this path")
		reportOnly = flag.Bool("report-only", false, "always exit 0, even when drift is found")
	)
	flag.Parse()

	raw, err := loadSchema(*schemaPath, *url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "apidrift:", err)
		return 2
	}

	endpoints, err := apidrift.ParseAPIDoc(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "apidrift:", err)
		return 2
	}

	report := apidrift.Diff(endpoints, apidrift.Registry())

	fmt.Print(report.Text())

	if *mdOut != "" {
		if err := os.WriteFile(*mdOut, []byte(report.Markdown()), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, "apidrift: write markdown:", err)
			return 2
		}
	}
	if *jsonOut != "" {
		b, err := report.JSON()
		if err != nil {
			fmt.Fprintln(os.Stderr, "apidrift: render json:", err)
			return 2
		}
		if err := os.WriteFile(*jsonOut, append(b, '\n'), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, "apidrift: write json:", err)
			return 2
		}
	}

	if report.HasDrift() && !*reportOnly {
		return 1
	}
	return 0
}

func loadSchema(path, url string) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url) //nolint:noctx // simple one-shot CLI fetch
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
