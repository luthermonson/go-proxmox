// Package endpoints provides mage targets that snapshot the upstream Proxmox VE
// REST API surface (apidoc.js) and flatten it into a (method, path) list, so
// the library's wrapper coverage can be diffed against the canonical schema.
package endpoints

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	upstreamURL = "https://pve.proxmox.com/pve-docs/api-viewer/apidoc.js"
	cacheDir    = ".cache/pve-api"
)

// intentionallyExcluded enumerates PVE endpoints we deliberately do not wrap
// in Go, with the rationale for each. Coverage stats subtract these from the
// denominator so the % reflects design intent rather than a mechanical count
// that would never reach 100.
//
// Keys are `key(method, normalizedPath)` form — lowercase verb + space + path
// with `{var}`/`%s` segments collapsed to `{}`. See `normalizePath` + `key`.
//
// **What does NOT belong here:** directory-index ("diridx") GETs that simply
// enumerate subresources (`[{"subdir":"x"}, ...]`). They look useless at first
// glance, but PVE filters the response by the caller's ACLs, so they're a
// supported permission-introspection / capability-discovery probe. We wrap
// them; see (*Node).sdnDiridx and (*Node).capabilitiesDiridx for the helper
// shape and the *Index family of methods that call into them.
//
// **What belongs here:** endpoints whose shape doesn't fit the generic
// `Get/Post/Put/Delete(ctx, path, ...)` wrapper at all and that a caller would
// need a custom code path for anyway. Two cases today:
//
//   - mtunnelwebsocket — the library exposes a URL-builder
//     (`MigrationTunnelWebSocketPath`) that returns the signed URL with
//     ticket; the caller plumbs that into their own websocket dialer. There
//     is no in-library HTTP call to scan, and the canonical "wrapper" is the
//     path helper, not a GET.
//
//   - file-restore/download — streams a tar of restored files. Doesn't fit
//     the JSON-decoding `Get` helper; needs an `io.Reader` flow we haven't
//     plumbed.
//
// Adding a new entry requires updating both the map AND the call site that
// proves the wrapper exists in some non-scannable form. Removing an entry is
// preferable when we add a real wrapper.
var intentionallyExcluded = map[string]string{
	"get /nodes/{}/qemu/{}/mtunnelwebsocket":         "URL-builder helper (MigrationTunnelWebSocketPath); caller drives the websocket dialer",
	"get /nodes/{}/lxc/{}/mtunnelwebsocket":          "URL-builder helper (MigrationTunnelWebSocketPath); caller drives the websocket dialer",
	"get /nodes/{}/storage/{}/file-restore/download": "streaming binary download; needs custom io.Reader plumbing, not the JSON Get helper",
}

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

// Coverage scans the library source for HTTP call sites, normalizes their
// path templates, and reports the fraction of the upstream PVE API surface
// that has at least one Go wrapper. Auto-runs Sync if the cached schema is
// missing. Writes coverage.txt (uncovered endpoints) and coverage_by_area.txt
// into .cache/pve-api/.
func Coverage() error {
	endpointsFile := filepath.Join(cacheDir, "endpoints.json")
	if _, err := os.Stat(endpointsFile); err != nil {
		if err := Sync(); err != nil {
			return err
		}
	}

	raw, err := os.ReadFile(endpointsFile)
	if err != nil {
		return err
	}
	var schema []endpoint
	if err := json.Unmarshal(raw, &schema); err != nil {
		return err
	}

	schemaKey := make(map[string]struct{}, len(schema))
	schemaByArea := map[string]map[string]struct{}{}
	for _, e := range schema {
		k := key(e.Method, normalizePath(e.Path))
		schemaKey[k] = struct{}{}
		a := areaOf(normalizePath(e.Path))
		if schemaByArea[a] == nil {
			schemaByArea[a] = map[string]struct{}{}
		}
		schemaByArea[a][k] = struct{}{}
	}

	// Reconcile the excluded list against the live schema. An entry that no
	// longer exists upstream (PVE removed/renamed the endpoint) is a real bug
	// in the exclusion table, not a silent no-op — fail loudly so the table
	// stays accurate.
	for k := range intentionallyExcluded {
		if _, ok := schemaKey[k]; !ok {
			return fmt.Errorf("intentionallyExcluded entry %q is not in the upstream schema — update or remove it", k)
		}
	}

	calls, err := scanCallSites(".")
	if err != nil {
		return err
	}

	covered := map[string]struct{}{}
	var unmatched []extractedCall
	matched, missed := 0, 0
	for _, c := range calls {
		k := key(c.Method, c.Path)
		if _, ok := schemaKey[k]; ok {
			matched++
			covered[k] = struct{}{}
		} else {
			missed++
			unmatched = append(unmatched, c)
		}
	}
	sort.Slice(unmatched, func(i, j int) bool {
		if unmatched[i].File != unmatched[j].File {
			return unmatched[i].File < unmatched[j].File
		}
		return unmatched[i].Line < unmatched[j].Line
	})

	var uncovered []string
	excludedHits := 0
	for _, e := range schema {
		k := key(e.Method, normalizePath(e.Path))
		if _, ok := covered[k]; ok {
			continue
		}
		if _, ok := intentionallyExcluded[k]; ok {
			excludedHits++
			continue
		}
		uncovered = append(uncovered, fmt.Sprintf("%-7s %s", e.Method, e.Path))
	}
	sort.Strings(uncovered)
	if err := os.WriteFile(
		filepath.Join(cacheDir, "coverage.txt"),
		[]byte(strings.Join(uncovered, "\n")+"\n"),
		0o644,
	); err != nil {
		return err
	}

	type areaRow struct {
		area                         string
		total, cov, excluded, missed int
	}
	var rows []areaRow
	for area, keys := range schemaByArea {
		row := areaRow{area: area, total: len(keys)}
		for k := range keys {
			if _, ok := covered[k]; ok {
				row.cov++
				continue
			}
			if _, ok := intentionallyExcluded[k]; ok {
				row.excluded++
			}
		}
		row.missed = row.total - row.cov - row.excluded
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].total != rows[j].total {
			return rows[i].total > rows[j].total
		}
		return rows[i].area < rows[j].area
	})

	var areaBuf bytes.Buffer
	fmt.Fprintf(&areaBuf, "%-25s %8s %8s %8s %8s %8s\n", "area", "covered", "total", "excluded", "missing", "pct")
	for _, r := range rows {
		// pct is covered / (total - excluded) so excluded endpoints don't
		// drag the area's number down. covered/total without that adjustment
		// would punish areas that have endpoints we've deliberately decided
		// not to wrap (e.g. streaming downloads).
		denom := r.total - r.excluded
		fmt.Fprintf(&areaBuf, "%-25s %8d %8d %8d %8d %7.1f%%\n",
			r.area, r.cov, r.total, r.excluded, r.missed, pct(r.cov, denom))
	}
	if err := os.WriteFile(
		filepath.Join(cacheDir, "coverage_by_area.txt"),
		areaBuf.Bytes(),
		0o644,
	); err != nil {
		return err
	}

	fmt.Printf("\ngo-proxmox endpoint coverage — %s\n", time.Now().Format("2006-01-02"))
	fmt.Printf("  schema endpoints   : %d\n", len(schema))
	fmt.Printf("  call sites scanned : %d (matched=%d, no-match=%d → match rate %.1f%%)\n",
		len(calls), matched, missed, pct(matched, len(calls)))
	// "Considered" is total minus the deliberately-unwrappable endpoints
	// (URL-builder helpers, streaming binary downloads, etc.). Reporting the
	// covered % against this denominator answers "of the endpoints we set
	// out to wrap, how many have we wrapped" rather than against a moving
	// upstream target that includes things we never intended to cover.
	considered := len(schema) - excludedHits
	fmt.Printf("  unique covered     : %d / %d  →  %.1f%%  (excluded by design: %d)\n",
		len(covered), considered, pct(len(covered), considered), excludedHits)
	fmt.Println()
	fmt.Print(areaBuf.String())
	if len(unmatched) > 0 {
		fmt.Printf("\ncall sites that don't resolve to a schema endpoint (%d):\n", len(unmatched))
		fmt.Println("  (real coverage findings — either stale Go code, or one Go call covering multiple schema endpoints)")
		for _, c := range unmatched {
			fmt.Printf("  %s:%d  %s %s\n", c.File, c.Line, c.Method, c.Path)
		}
	}
	fmt.Printf("\nwrote:\n  %s/coverage.txt           # %d uncovered endpoints\n  %s/coverage_by_area.txt\n",
		cacheDir, len(uncovered), cacheDir)
	return nil
}

// scanCallSites walks .go files (skipping tests, examples, mage, .claude, .cache)
// and extracts `c.Get|Post|Put|Delete(ctx, "/path"|fmt.Sprintf("/path", ...), ...)`
// call sites. Also recognizes the `url.URL{Path: ...}.String()` indirection
// used by call sites that need to attach query parameters, e.g.
//
//	u := url.URL{Path: fmt.Sprintf("/nodes/%s/qemu/%d/rrddata", v.Node, v.VMID)}
//	u.RawQuery = params.Encode()
//	err := v.client.Get(ctx, u.String(), &out)
//
// Additionally recognizes the websocket call signature used by *.VNCWebSocket
// and *.TermWebSocket — these wrap GET /…/vncwebsocket and do not take a
// context, so their first argument is the path:
//
//	p := fmt.Sprintf("/nodes/%s/qemu/%d/vncwebsocket?…", …)
//	return v.client.VNCWebSocket(p, vnc)
//
// Returns normalized (method, path) pairs.
func scanCallSites(root string) ([]extractedCall, error) {
	var out []extractedCall
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", ".cache", "examples", "tests", "audit", "mage":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Fprintln(os.Stderr, "close:", cerr)
			}
		}()
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 1<<16), 1<<20)
		// Per-file map of `<var>` → path for `<var> := url.URL{Path: ...}`
		// declarations. Reset per file so that vars in different files don't
		// bleed. Reused vars within a file simply track the most recent binding,
		// which is correct because the corresponding Get/Post call always
		// follows the assignment in the same function.
		urlVars := map[string]string{}
		// Per-file map of `<var>` → path for `<var> := fmt.Sprintf("/…", …)` or
		// `<var> := "/…"` bindings, used by call sites that pass the path as a
		// pre-built variable rather than inline (e.g. the websocket helpers).
		pathVars := map[string]string{}
		// pendingMethod/Line are set when an Upload (or similar) call lands its
		// `(` at end of line and the path literal is on the following line(s).
		// Empty pendingMethod means no continuation is in flight.
		var pendingMethod string
		var pendingLine int
		for line := 1; s.Scan(); line++ {
			text := s.Text()

			// Continuation from a previous line: try to pull the path off this
			// line. Skip blank-only and pure-comment lines. Clear state once a
			// non-empty line is processed regardless of whether we extracted a
			// path, so we don't accidentally grab a path many lines downstream.
			if pendingMethod != "" {
				trimmed := strings.TrimSpace(text)
				if trimmed == "" || strings.HasPrefix(trimmed, "//") {
					continue
				}
				if pathLit := resolvePath(text, urlVars, pathVars); pathLit != "" {
					out = append(out, extractedCall{
						Method: pendingMethod,
						Path:   normalizePath(pathLit),
						File:   filepath.ToSlash(path),
						Line:   pendingLine,
					})
				}
				pendingMethod = ""
				pendingLine = 0
				// fall through — this line might also start a new call
			}

			if am := urlAssignRE.FindStringSubmatch(text); am != nil {
				if p := extractPathExpr(am[2]); p != "" {
					urlVars[am[1]] = p
				}
			}
			if pm := pathAssignRE.FindStringSubmatch(text); pm != nil {
				if p := extractPathExpr(pm[2]); p != "" {
					pathVars[pm[1]] = p
				}
			}

			if m := callRE.FindStringSubmatch(text); m != nil {
				method := m[1]
				rest := m[2]
				pathLit := resolvePath(rest, urlVars, pathVars)
				if pathLit != "" {
					out = append(out, extractedCall{
						Method: method,
						Path:   normalizePath(pathLit),
						File:   filepath.ToSlash(path),
						Line:   line,
					})
				}
				continue
			}

			if wm := wsCallRE.FindStringSubmatch(text); wm != nil {
				rest := wm[2]
				pathLit := resolvePath(rest, urlVars, pathVars)
				if pathLit == "" {
					continue
				}
				// VNCWebSocket and TermWebSocket both wrap GET /…/vncwebsocket.
				out = append(out, extractedCall{
					Method: "GET",
					Path:   normalizePath(pathLit),
					File:   filepath.ToSlash(path),
					Line:   line,
				})
				continue
			}

			if dm := diridxRE.FindStringSubmatch(text); dm != nil {
				pathLit := resolvePath(dm[1], urlVars, pathVars)
				if pathLit == "" {
					continue
				}
				out = append(out, extractedCall{
					Method: "GET",
					Path:   normalizePath(pathLit),
					File:   filepath.ToSlash(path),
					Line:   line,
				})
				continue
			}

			if um := uploadRE.FindStringSubmatch(text); um != nil {
				if pathLit := resolvePath(um[1], urlVars, pathVars); pathLit != "" {
					out = append(out, extractedCall{
						Method: "POST",
						Path:   normalizePath(pathLit),
						File:   filepath.ToSlash(path),
						Line:   line,
					})
					continue
				}
				// `(` at end of line; the path is on a subsequent line. Mark
				// state so the next non-blank, non-comment line is treated as
				// the path expression. The path is always the first arg to
				// `client.Upload[Reader](path, …)`, so the very next code line
				// is the right place to look.
				pendingMethod = "POST"
				pendingLine = line
			}
		}
		return s.Err()
	})
	return out, err
}

// resolvePath extracts a path literal from a call argument expression. It
// understands fmt.Sprintf literals, bare string literals, url.URL{}.String()
// indirection, and previously-tracked path variables.
func resolvePath(expr string, urlVars, pathVars map[string]string) string {
	if p := extractPathExpr(expr); p != "" {
		return p
	}
	trimmed := strings.TrimSpace(expr)
	if um := urlStringRE.FindStringSubmatch(trimmed); um != nil {
		return urlVars[um[1]]
	}
	if vm := varRefRE.FindStringSubmatch(trimmed); vm != nil {
		return pathVars[vm[1]]
	}
	return ""
}

// extractPathExpr pulls the path literal out of an expression like:
//
//	"/some/path"
//	fmt.Sprintf("/some/path", args...)
//
// Returns "" if neither form is present.
func extractPathExpr(expr string) string {
	if sm := sprintfRE.FindStringSubmatch(expr); sm != nil {
		return sm[1]
	}
	if t := strings.TrimSpace(expr); strings.HasPrefix(t, `"`) {
		if end := strings.Index(t[1:], `"`); end > 0 {
			return t[1 : 1+end]
		}
	}
	return ""
}

type extractedCall struct {
	Method, Path, File string
	Line               int
}

var (
	// Captures `<recv>.Get|Post|Put|Delete(WithParams)?(ctx, ...)`. The
	// WithParams variants on *Client (GetWithParams, DeleteWithParams) take an
	// extra body/query interface but the path is still the first non-ctx arg,
	// so the same extractor handles both. The `(?:WithParams)?` is
	// non-capturing — group 1 always returns the bare HTTP verb.
	callRE   = regexp.MustCompile(`\b\w+\.(Get|Post|Put|Delete)(?:WithParams)?\s*\(\s*ctx\s*,\s*(.*?)$`)
	wsCallRE = regexp.MustCompile(`\b\w+\.(VNCWebSocket|TermWebSocket)\s*\(\s*(.*?)$`)
	// Diridx helpers — internal package methods named `<X>Diridx` or `<X>diridx`
	// that wrap a GET against a directory-index endpoint. They exist so several
	// sibling diridx wrappers (e.g. (*Node).SDNIndex, SDNZoneIndex, …) can share
	// the same `[{"subdir":"x"}, ...] → []string` unmarshalling. The actual HTTP
	// verb is always GET; scanner treats every diridx-helper call site as a GET
	// against the first non-ctx argument path.
	diridxRE = regexp.MustCompile(`\b\w+\.\w*[Dd]iridx\s*\(\s*ctx\s*,\s*(.*?)$`)
	// Upload — `*Client.Upload` and `*Client.UploadReader` both wrap the same
	// multipart POST endpoint. Signatures are `Upload(path, ...)` /
	// `UploadReader(path, ...)` (no ctx; the helper builds its own request).
	// The `client.` qualifier is required to avoid matching the unrelated
	// method `(*Storage).Upload(content, filename string)` whose first arg
	// is a content-type, not a path. The `\w*` suffix catches both variants
	// (and any future Upload* the library adds) without re-listing each.
	uploadRE    = regexp.MustCompile(`\bclient\.Upload\w*\s*\(\s*(.*?)$`)
	sprintfRE   = regexp.MustCompile(`fmt\.Sprintf\(\s*"([^"]+)"`)
	urlAssignRE = regexp.MustCompile(`\b(\w+)\s*:?=\s*url\.URL\{\s*Path:\s*(.+)\}`)
	// `<var> := fmt.Sprintf("/…", …)` or `<var> := "/…"` — bindings that later
	// flow into a Get/Post/WS call as a bare variable reference. Captures only
	// the rhs head so extractPathExpr can pull the string literal; the rest of
	// the Sprintf arglist may continue on subsequent lines.
	pathAssignRE = regexp.MustCompile(`\b(\w+)\s*:?=\s*(fmt\.Sprintf\(\s*"[^"]+"|"/[^"]*")`)
	urlStringRE  = regexp.MustCompile(`^(\w+)\.String\(\)`)
	varRefRE     = regexp.MustCompile(`^(\w+)\s*[,)]`)
	braceRE      = regexp.MustCompile(`\{[^}]+\}`)
	fmtVerbRE    = regexp.MustCompile(`%[sdvqxXt]`)
	// PVE volume identifiers serialize as `<storage>:<type>/<filename>` and are
	// represented as a single `{volume}` segment in the schema. Go code that
	// builds the path via `fmt.Sprintf(...content/%s:%s/%s, ...)` normalizes to
	// `.../content/{}:{}/{}` and must collapse to `.../content/{}` to match.
	volidRE = regexp.MustCompile(`\{\}:\{\}/\{\}`)
)

// normalizePath canonicalizes both schema and Go forms so they can be joined:
// {var}/%s/%d → {}; volid-style {}:{}/{}  → {}; query string and trailing
// slash stripped.
func normalizePath(p string) string {
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	p = braceRE.ReplaceAllString(p, "{}")
	p = fmtVerbRE.ReplaceAllString(p, "{}")
	p = volidRE.ReplaceAllString(p, "{}")
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func key(method, path string) string {
	return strings.ToLower(method) + " " + path
}

// areaOf groups endpoints for the breakdown table — top-level segment, except
// /nodes/{}/<sub> where <sub> is the meaningful area.
func areaOf(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "(root)"
	}
	if parts[0] == "nodes" && len(parts) >= 3 {
		return "nodes/" + parts[2]
	}
	return parts[0]
}

func pct(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) * 100 / float64(d)
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
