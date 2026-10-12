package proxmox

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Sentinels for API error classification. Match them with errors.Is, or use
// the Is* helpers below. ErrNotFound and ErrNotAuthorized live in proxmox.go
// and predate *APIError; *APIError matches them too.
var (
	// ErrForbidden matches HTTP 403 responses. A 403 also matches
	// ErrNotAuthorized, which has always covered both 401 and 403.
	ErrForbidden = errors.New("permission denied")

	// ErrBadRequest matches HTTP 400 responses, which PVE returns for
	// parameter verification failures ("Parameter verification failed.")
	// and other raise_param_exc errors. The per-parameter detail is in
	// APIError.Errors.
	ErrBadRequest = errors.New("bad request")

	// ErrRateLimited matches HTTP 429 responses. APIError.RetryAfter carries
	// the server's Retry-After hint when one was sent.
	ErrRateLimited = errors.New("rate limited")

	// ErrNodeUnreachable matches the 595 and 596 pseudo-statuses pveproxy
	// returns when it cannot proxy a request to the target cluster node,
	// typically "595 Connection refused", "595 No route to host" or
	// "596 Connection timed out". They come from AnyEvent::HTTP, which
	// pveproxy uses to forward requests between nodes: 595 is a failure
	// while connecting, 596 a failure during TLS negotiation, sending the
	// request or reading the response headers.
	ErrNodeUnreachable = errors.New("cluster node unreachable")
)

// Pseudo-statuses surfaced by pveproxy when forwarding a request to another
// cluster node fails. See ErrNodeUnreachable.
const (
	StatusNodeConnectFailed = 595
	StatusNodeRequestFailed = 596
)

// APIError is returned for every non-2xx response from the Proxmox API.
// Use errors.As to inspect it:
//
//	var apiErr *proxmox.APIError
//	if errors.As(err, &apiErr) {
//		log.Printf("%s %s: %d %s %v", apiErr.Method, apiErr.Path,
//			apiErr.StatusCode, apiErr.Message, apiErr.Errors)
//	}
//
// It also matches the package sentinels through errors.Is: ErrNotFound,
// ErrNotAuthorized, ErrForbidden, ErrBadRequest, ErrRateLimited and
// ErrNodeUnreachable.
type APIError struct {
	// StatusCode is the HTTP status code, e.g. 500.
	StatusCode int
	// Status is the raw HTTP status line, e.g. "500 Configuration file
	// 'nodes/pve/qemu-server/100.conf' does not exist". PVE puts the first
	// line of its error message in the reason phrase.
	Status string
	// Message is PVE's error message. It comes from the "message" key of the
	// JSON body when present (that copy is not truncated to one line),
	// otherwise from the reason phrase of the status line.
	Message string
	// Errors holds per-parameter messages from the "errors" key of the JSON
	// body, as returned with 400 "Parameter verification failed.".
	Errors map[string]string
	// Method and Path identify the request. Path excludes the query string.
	Method string
	Path   string
	// RetryAfter is the parsed Retry-After header, or zero when absent.
	RetryAfter time.Duration

	// detail is the text Error() appends to the status for 400 responses,
	// kept for byte-compatibility with the pre-APIError message.
	detail string
}

// Error renders the status line first, so code that matched on the old
// errors.New(res.Status) text keeps matching:
//
//	500 Configuration file 'nodes/pve/qemu-server/100.conf' does not exist
//	bad request: 400 Parameter verification failed. - {"vmid":"..."}
//	not authorized to access endpoint: 401 No ticket
//
// For 400 the legacy "bad request: <status> - <errors>" form is kept as-is.
// For 401/403 the text of ErrNotAuthorized prefixes the status, which is what
// these responses returned before. Otherwise any part of Message not already
// in the status line (PVE truncates the reason phrase to one line) is
// appended after a colon.
func (e *APIError) Error() string {
	switch e.StatusCode {
	case http.StatusBadRequest:
		if e.detail != "" {
			return "bad request: " + e.Status + " - " + e.detail
		}
		return "bad request: " + e.withMessage()
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrNotAuthorized.Error() + ": " + e.withMessage()
	}
	return e.withMessage()
}

func (e *APIError) withMessage() string {
	status := e.Status
	if status == "" {
		status = strconv.Itoa(e.StatusCode) + " " + http.StatusText(e.StatusCode)
	}
	msg := strings.TrimSpace(e.Message)
	if reason := e.reason(); reason != "" {
		msg = strings.TrimSpace(strings.TrimPrefix(msg, reason))
	}
	if msg == "" {
		return status
	}
	return status + ": " + msg
}

// reason returns the reason phrase of Status, without the leading code.
func (e *APIError) reason() string {
	_, reason, _ := strings.Cut(e.Status, " ")
	return strings.TrimSpace(reason)
}

// Is reports whether the error matches one of the package sentinels.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.isNotFound()
	case ErrNotAuthorized:
		return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrNodeUnreachable:
		return e.StatusCode == StatusNodeConnectFailed || e.StatusCode == StatusNodeRequestFailed
	}
	return false
}

// notFoundPatterns are the PVE error messages treated as "resource does not
// exist". PVE reports most missing resources as a 500 (a plain die in the
// API handler) or a 400 (raise_param_exc on the identifying parameter)
// rather than a 404, so the message is the only signal. Matching is
// case-insensitive. Sources are in the proxmox/* repositories:
//
//	pattern                       example message                                        source
//	does not exist(s)             Configuration file 'nodes/pve/qemu-server/100.conf'    pve-guest-common PVE/AbstractConfig.pm
//	                              does not exist
//	                              storage 'foo' does not exist                           pve-storage PVE/Storage.pm
//	                              pool 'x' / role 'x' / group 'x' does not exist         pve-manager API2/Pool.pm, pve-access-control API2/{Role,Group}.pm
//	                              sdn zone/vnet/subnet/ipam/dns/controller 'x' does      pve-network PVE/Network/SDN/*.pm
//	                              not exist
//	                              snapshot 'x' does not exist                            qemu-server PVE/QemuServer.pm
//	                              OSD '3' does not exists on host 'pve'                  pve-manager API2/Ceph/OSD.pm
//	no such <thing>               no such user ('foo@pve')                               pve-access-control PVE/AccessControl.pm
//	                              no such cluster node 'x'                               pve-cluster PVE/Cluster.pm
//	                              no such job 'x', No such job 'x'                       pve-manager API2/Replication*.pm, API2/Backup.pm
//	                              no such resource 'vm:100', no such ha group 'x'        pve-ha-manager API2/HA/*.pm
//	                              no such OSD / cephfs / monitor id / volume group       pve-manager API2/Ceph/*.pm, pve-storage API2/Disks/*.pm
//	                              No such storage. (400, errors.storage)                 pve-storage API2/Storage/Status.pm
//	                              no such task (400, errors.upid)                        pve-manager API2/Tasks.pm
//	<thing> not found             mapping 'x' not found, Custom cputype 'x' not found    pve-manager API2/Cluster/Mapping/*.pm, qemu-server CPUConfig.pm
//	unable to find configuration  unable to find configuration file for VM 100 on node   qemu-server PVE/QemuConfig.pm
//	file                          'pve'
var notFoundPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bdoes not exists?\b`),
	regexp.MustCompile(`(?i)\bno such \S`),
	regexp.MustCompile(`(?i)\bnot found\b`),
	regexp.MustCompile(`(?i)\bunable to find configuration file\b`),
}

// notFoundExclusions are messages that match notFoundPatterns but describe
// an invalid request rather than a missing resource.
//
//	no such option 'x'           unknown config key       pve-storage API2/Storage/Config.pm, pve-manager API2/*
//	no such standard option      schema lookup            pve-common PVE/JSONSchema.pm
//	No such timezone             invalid parameter value  pve-common PVE/INotify.pm
//	invalid ...                  e.g. "invalid bootorder: device 'x' does not exist"   qemu-server API2/Qemu.pm
//	command/executable not found missing host binary      pve-storage PBSPlugin.pm and others
var notFoundExclusions = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bno such (standard )?option\b`),
	regexp.MustCompile(`(?i)\bno such timezone\b`),
	regexp.MustCompile(`(?i)^invalid\b`),
	regexp.MustCompile(`(?i)\b(command|executable) not found\b`),
}

func matchesNotFound(msg string) bool {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return false
	}
	for _, ex := range notFoundExclusions {
		if ex.MatchString(msg) {
			return false
		}
	}
	for _, p := range notFoundPatterns {
		if p.MatchString(msg) {
			return true
		}
	}
	return false
}

func (e *APIError) isNotFound() bool {
	switch e.StatusCode {
	case http.StatusNotFound:
		return true
	case http.StatusInternalServerError:
		return matchesNotFound(e.Message) || matchesNotFound(e.reason())
	case http.StatusBadRequest:
		// raise_param_exc puts "No such storage." etc. on the parameter
		// that names the missing resource.
		for _, msg := range e.Errors {
			if matchesNotFound(msg) {
				return true
			}
		}
	}
	return false
}

// newAPIError builds an *APIError from a non-2xx response and its already
// read body. Body parsing is best-effort: a non-JSON or empty body leaves
// Message set from the status line.
func newAPIError(res *http.Response, body []byte) *APIError {
	e := &APIError{
		StatusCode: res.StatusCode,
		Status:     res.Status,
		detail:     string(body),
	}
	if res.Request != nil {
		e.Method = res.Request.Method
		if res.Request.URL != nil {
			e.Path = res.Request.URL.Path
		}
	}
	if d, ok := parseRetryAfter(res.Header.Get("Retry-After"), time.Now()); ok {
		e.RetryAfter = d
	}

	// PVE's JSON formatter emits {"data":null,"message":"...","errors":{...}}.
	var envelope map[string]json.RawMessage
	if json.Unmarshal(body, &envelope) == nil {
		var msg string
		if json.Unmarshal(envelope["message"], &msg) == nil {
			e.Message = strings.TrimSpace(msg)
		}
		if raw, ok := envelope["errors"]; ok {
			// The pre-APIError 400 message embedded the raw "errors" JSON.
			e.detail = string(raw)
			var errs map[string]json.RawMessage
			if json.Unmarshal(raw, &errs) == nil && len(errs) > 0 {
				e.Errors = make(map[string]string, len(errs))
				for k, v := range errs {
					var s string
					if json.Unmarshal(v, &s) != nil {
						s = string(v)
					}
					e.Errors[k] = strings.TrimSpace(s)
				}
			}
		}
	}
	if e.Message == "" {
		e.Message = e.reason()
	}
	return e
}

// IsForbidden reports whether err is a 403 response.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsBadRequest reports whether err is a 400 response, typically PVE's
// "Parameter verification failed." with per-parameter detail in
// APIError.Errors.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsRateLimited reports whether err is a 429 response.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsNodeUnreachable reports whether pveproxy could not reach the target
// cluster node (HTTP 595 or 596).
func IsNodeUnreachable(err error) bool {
	return errors.Is(err, ErrNodeUnreachable)
}

// AsAPIError returns the *APIError in err's chain, if any.
func AsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
