// Package endpoints provides mage targets that snapshot the upstream Proxmox VE
// REST API surface (apidoc.js) and flatten it into a (method, path) list, so
// the library's wrapper coverage can be diffed against the canonical schema.
package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	upstreamURL = "https://pve.proxmox.com/pve-docs/api-viewer/apidoc.js"
	cacheDir    = ".cache/pve-api"
)

type methodInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type node struct {
	Text     string                `json:"text"`
	Info     map[string]methodInfo `json:"info"`
	Children []node                `json:"children"`
}

type endpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Sync fetches the upstream API schema and writes endpoints.json + endpoints.txt
// into .cache/pve-api/. This is the normal entry point.
func Sync() error {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return err
	}
	fmt.Printf("Fetching %s\n", upstreamURL)
	if err := download(upstreamURL, filepath.Join(cacheDir, "apidoc.js")); err != nil {
		return err
	}
	return Extract()
}

// Extract parses the cached apidoc.js (no network) and emits endpoints.json
// and endpoints.txt. Useful for re-running the parser without re-downloading.
func Extract() error {
	cache := filepath.Join(cacheDir, "apidoc.js")
	raw, err := os.ReadFile(cache)
	if err != nil {
		return fmt.Errorf("read %s: %w (run `mage endpoints:sync` first)", cache, err)
	}
	jsonBlob, err := extractSchemaJSON(raw)
	if err != nil {
		return err
	}
	var schema []node
	if err := json.Unmarshal(jsonBlob, &schema); err != nil {
		return fmt.Errorf("parse apidoc.js schema: %w", err)
	}

	out := flatten(schema)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "endpoints.json"), jsonBytes, 0o644); err != nil {
		return err
	}
	var txt bytes.Buffer
	for _, e := range out {
		fmt.Fprintf(&txt, "%-7s %s\n", e.Method, e.Path)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "endpoints.txt"), txt.Bytes(), 0o644); err != nil {
		return err
	}
	printSummary(out)
	return nil
}

// extractSchemaJSON peels the `const apiSchema = [ ... ];` array out of the
// upstream JS file. We anchor on `\n]\n;` rather than the last `]` because
// the JS handler code below the schema contains template literals like
// `${m2c[method]}` whose `]` would be matched first.
func extractSchemaJSON(raw []byte) ([]byte, error) {
	start := bytes.IndexByte(raw, '[')
	endIdx := bytes.Index(raw, []byte("\n]\n;"))
	if start < 0 || endIdx <= start {
		return nil, fmt.Errorf("could not locate JSON array boundaries in apidoc.js")
	}
	return raw[start : endIdx+2], nil
}

func flatten(schema []node) []endpoint {
	var out []endpoint
	var walk func(n node, path string)
	walk = func(n node, path string) {
		here := path
		if n.Text != "" {
			here = path + "/" + n.Text
		}
		for method, meta := range n.Info {
			desc := meta.Description
			if i := strings.IndexByte(desc, '\n'); i >= 0 {
				desc = desc[:i]
			}
			out = append(out, endpoint{
				Method:      method,
				Path:        here,
				Name:        meta.Name,
				Description: strings.TrimSpace(desc),
			})
		}
		for _, c := range n.Children {
			walk(c, here)
		}
	}
	for _, top := range schema {
		walk(top, "")
	}
	return out
}

func download(url, dst string) (err error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close response body: %w", cerr)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", dst, cerr)
		}
	}()
	_, err = io.Copy(f, resp.Body)
	return err
}

func printSummary(out []endpoint) {
	uniqPaths := map[string]struct{}{}
	areas := map[string]int{}
	methodCounts := map[string]int{}
	for _, e := range out {
		uniqPaths[e.Path] = struct{}{}
		methodCounts[e.Method]++
		areas[topArea(e.Path)]++
	}
	fmt.Printf("\nProxmox VE API snapshot — %s\n", time.Now().Format("2006-01-02"))
	fmt.Printf("  endpoints   : %d\n", len(out))
	fmt.Printf("  unique paths: %d\n", len(uniqPaths))
	fmt.Printf("  methods     :")
	for _, m := range sortedKeys(methodCounts) {
		fmt.Printf(" %s=%d", m, methodCounts[m])
	}
	fmt.Println()
	fmt.Printf("  top areas   :")
	for _, a := range sortedKeys(areas) {
		fmt.Printf(" %s=%d", a, areas[a])
	}
	fmt.Println()
	fmt.Printf("\nwrote:\n  %s/endpoints.json\n  %s/endpoints.txt\n", cacheDir, cacheDir)
}

func topArea(path string) string {
	p := strings.TrimPrefix(path, "/")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return p
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
