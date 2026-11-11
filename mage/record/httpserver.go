package record

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// FileServer is a short-lived HTTP server bound to the workstation's LAN IP
// that hosts the auto-install answer file and first-boot script during a
// recording session.
//
// The nested PVE installer fetches answer.toml during install
// (proxmox-auto-install-assistant prepare-iso baked the URL into the ISO at
// re-master time). The freshly-installed system then fetches first-boot.sh
// on first boot. Both fetches happen over plain HTTP across the lab LAN —
// fine for a throwaway nested VM, but the answer.toml does contain the
// nested host's root password (which is itself throwaway and randomised
// per-record by default).
type FileServer struct {
	listener net.Listener
	server   *http.Server
	baseURL  string
}

// PlannedURLs is the addressing the FileServer will use, computed from
// config without binding a listener. Callers use this to build the
// answer.toml content (which must reference its own first-boot URL) before
// starting the server.
type PlannedURLs struct {
	BaseURL      string
	AnswerURL    string
	FirstBootURL string
}

// PlanURLs picks the bind IP from config and computes the URLs the
// FileServer will serve from. Pure — no listener bound.
func PlanURLs(cfg *Config) (*PlannedURLs, error) {
	bindIP, err := chooseBindIP(cfg)
	if err != nil {
		return nil, err
	}
	base := fmt.Sprintf("http://%s:%d", bindIP, cfg.HTTPPort)
	return &PlannedURLs{
		BaseURL:      base,
		AnswerURL:    base + "/answer.toml",
		FirstBootURL: base + "/first-boot.sh",
	}, nil
}

// NewFileServer binds the workstation HTTP server and starts serving in a
// background goroutine. Returns once the listener is ready. The caller must
// invoke Stop() when the install + first-boot phase is complete; the
// recorder pipeline does this via defer.
func NewFileServer(cfg *Config, answer, firstBoot []byte) (*FileServer, error) {
	urls, err := PlanURLs(cfg)
	if err != nil {
		return nil, err
	}
	addr := strings.TrimPrefix(urls.BaseURL, "http://")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("http bind %s: %w", addr, err)
	}

	baseURL := urls.BaseURL

	mux := http.NewServeMux()
	mux.HandleFunc("/answer.toml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/toml")
		_, _ = w.Write(answer)
	})
	mux.HandleFunc("/first-boot.sh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(firstBoot)
	})

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fs := &FileServer{
		listener: listener,
		server:   srv,
		baseURL:  baseURL,
	}

	go func() {
		// Serve always returns a non-nil error; ignore the expected
		// closed-listener error after Stop.
		_ = srv.Serve(listener)
	}()

	return fs, nil
}

// BaseURL returns the http://host:port the server is bound to. Callers
// append /answer.toml or /first-boot.sh for specific assets.
func (fs *FileServer) BaseURL() string { return fs.baseURL }

// AnswerURL is the URL to hand to proxmox-auto-install-assistant.
func (fs *FileServer) AnswerURL() string { return fs.baseURL + "/answer.toml" }

// FirstBootURL is the URL the answer.toml [first-boot] section points at.
func (fs *FileServer) FirstBootURL() string { return fs.baseURL + "/first-boot.sh" }

// Stop shuts the server down gracefully with a 5-second grace period.
func (fs *FileServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return fs.server.Shutdown(ctx)
}

// chooseBindIP picks the workstation IP the nested VM can route to. With
// PROXMOX_RECORDER_HTTP_HOST set, that wins. Otherwise we scan local
// interfaces for the *most specific* subnet that contains the gateway
// (longest prefix wins) — that's almost always the real LAN interface
// rather than a WSL/Docker/Hyper-V bridge that happens to have a wide
// mask covering it incidentally.
//
// This keeps us off 127.0.0.1 (unreachable from the nested VM) without
// requiring the user to know their own LAN IP up front.
func chooseBindIP(cfg *Config) (string, error) {
	if cfg.HTTPHost != "" {
		return cfg.HTTPHost, nil
	}
	gw := net.ParseIP(cfg.NestedGateway)
	if gw == nil {
		return "", fmt.Errorf("invalid PROXMOX_RECORDER_GATEWAY %q", cfg.NestedGateway)
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var bestIP net.IP
	bestPrefix := -1
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			if !ipnet.Contains(gw) {
				continue
			}
			ones, _ := ipnet.Mask.Size()
			if ones > bestPrefix {
				bestPrefix = ones
				bestIP = ipnet.IP
			}
		}
	}
	if bestIP == nil {
		return "", fmt.Errorf("no workstation interface found in subnet containing gateway %s; "+
			"set PROXMOX_RECORDER_HTTP_HOST explicitly", gw)
	}
	return bestIP.String(), nil
}
