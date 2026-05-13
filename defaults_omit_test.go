package proxmox

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests are the load-bearing safety net for issue #199 + #178: they
// prove that when a caller leaves an "optional with non-zero PVE default"
// field unset, the marshalled JSON does NOT contain that field — so PVE
// applies its own server-side default instead of being overridden by Go's
// zero value. Each test covers both directions: unset → omitted, explicit
// value → present.

func marshalToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	var m map[string]any
	assert.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestContainerConfig_DefaultsOmittedWhenUnset(t *testing.T) {
	// Bare struct — none of the pointer-typed fields should appear.
	m := marshalToMap(t, ContainerConfig{})
	for _, k := range []string{"arch", "cmode", "console", "cpuunits", "memory", "swap", "tty"} {
		_, present := m[k]
		assert.Falsef(t, present, "%q should be omitted when unset so PVE applies its default", k)
	}
}

func TestContainerConfig_ExplicitValuesAppearOnTheWire(t *testing.T) {
	cfg := ContainerConfig{
		Arch:     Ptr("amd64"),
		CMode:    Ptr("tty"),
		Console:  Ptr(IntOrBool(false)), // explicit disable — must reach PVE as 0, not be silently swallowed
		CPUUnits: Ptr(1024),
		Memory:   Ptr(512),
		Swap:     Ptr(0), // explicit no-swap — must reach PVE as 0, not be omitted
		TTY:      Ptr(2),
	}
	m := marshalToMap(t, cfg)
	assert.Equal(t, "amd64", m["arch"])
	assert.Equal(t, "tty", m["cmode"])
	assert.Equal(t, float64(0), m["console"]) // IntOrBool false marshals to 0
	assert.Equal(t, float64(1024), m["cpuunits"])
	assert.Equal(t, float64(512), m["memory"])
	assert.Equal(t, float64(0), m["swap"])
	assert.Equal(t, float64(2), m["tty"])
}

func TestVirtualMachineConfig_DefaultsOmittedWhenUnset(t *testing.T) {
	m := marshalToMap(t, VirtualMachineConfig{})
	// Pointer-typed: omitted means caller didn't set, server applies default.
	for _, k := range []string{
		"vmgenid", "hotplug", "tablet", "kvm", "ostype", "bios", "acpi",
		"sockets", "cores", "cpulimit", "cpuunits", "scsihw", "ciupgrade",
	} {
		_, present := m[k]
		assert.Falsef(t, present, "%q should be omitted when unset so PVE applies its default", k)
	}
	// Type-only IntOrBool fields (defaults match Go zero): they're still
	// dropped by omitempty when zero — that's the documented behavior.
	for _, k := range []string{"template", "autostart", "protection", "onboot", "numa"} {
		_, present := m[k]
		assert.Falsef(t, present, "%q (default 0) should be omitted at zero", k)
	}
}

func TestVirtualMachineConfig_ExplicitFalseSurvivesForBooleanFields(t *testing.T) {
	// The regression we're guarding against: a caller wanting to *disable*
	// KVM or ACPI explicitly. Pre-fix, KVM was `int` with omitempty, so
	// setting it to 0 was indistinguishable from not setting it and the
	// field was dropped — caller's intent lost. With *IntOrBool, an
	// explicit Ptr(false) marshals as 0 and reaches the server.
	cfg := VirtualMachineConfig{
		KVM:       Ptr(IntOrBool(false)),
		Acpi:      Ptr(IntOrBool(false)),
		CIUpgrade: Ptr(IntOrBool(false)),
	}
	m := marshalToMap(t, cfg)
	assert.Equal(t, float64(0), m["kvm"], "explicit KVM=false must reach the server, not be silently dropped")
	assert.Equal(t, float64(0), m["acpi"], "explicit ACPI=false must reach the server")
	assert.Equal(t, float64(0), m["ciupgrade"], "explicit CIUpgrade=false must reach the server")
}

func TestVirtualMachineBackupOptions_RemoveAndStdExcludesNoLongerSendFalse(t *testing.T) {
	// Pre-fix bug: Remove and StdExcludes were declared `bool` with NO
	// omitempty, so every call to Vzdump shipped `"remove":false` and
	// `"stdexcludes":false` whether the caller wanted that or not — silently
	// disabling backup pruning and tmp/log exclusion. Now with *IntOrBool
	// + omitempty, leaving them unset omits them and PVE applies its
	// documented default of 1.
	m := marshalToMap(t, VirtualMachineBackupOptions{VMID: 100, Storage: "local"})
	_, hasRemove := m["remove"]
	assert.False(t, hasRemove, "Remove must be omitted when unset; pre-fix it silently sent false")
	_, hasStdExcludes := m["stdexcludes"]
	assert.False(t, hasStdExcludes, "StdExcludes must be omitted when unset; pre-fix it silently sent false")
}

func TestVirtualMachineBackupOptions_ExplicitDisablesStillReachTheWire(t *testing.T) {
	// Caller explicitly opts out of pruning / tmp-log exclusion — must
	// reach PVE as 0, not be swallowed by omitempty.
	opts := VirtualMachineBackupOptions{
		VMID:        100,
		Remove:      Ptr(IntOrBool(false)),
		StdExcludes: Ptr(IntOrBool(false)),
	}
	m := marshalToMap(t, opts)
	assert.Equal(t, float64(0), m["remove"])
	assert.Equal(t, float64(0), m["stdexcludes"])
}

func TestFirewallNodeOption_EnableUnsetDoesNotDisableFirewall(t *testing.T) {
	// Per-node firewall ships enabled (PVE default 1). Pre-fix the field
	// was `bool` with omitempty, so an unset value sent nothing — that
	// was actually safe here because omitempty dropped the false. BUT
	// the field was also `bool` not `IntOrBool`, so setting `true`
	// marshalled as `true` instead of `1`. With *IntOrBool, unset
	// stays omitted and set values marshal as 0/1 as PVE expects.
	m := marshalToMap(t, FirewallNodeOption{})
	_, present := m["enable"]
	assert.False(t, present, "Enable must be omitted when unset so PVE keeps its default of 1 (firewall enabled)")
}

func TestFirewallNodeOption_ExplicitEnableMarshalsAsOne(t *testing.T) {
	// Type-correctness: schema declares boolean, PVE expects 0/1 on the
	// wire. IntOrBool emits 0/1; plain bool would emit true/false.
	opts := FirewallNodeOption{Enable: Ptr(IntOrBool(true))}
	m := marshalToMap(t, opts)
	assert.Equal(t, float64(1), m["enable"], "schema is boolean — wire format must be 0/1, not true/false")
}

func TestFirewallClusterOption_DefaultsOmittedWhenUnset(t *testing.T) {
	// Cluster-wide firewall is the TOP gate of PVE's three-gate model: this
	// is the only place where Enable=0 disables everything regardless of
	// node/VM settings. Ebtables defaults to 1 — silently shipping 0
	// disables bridge-level filtering across the whole cluster, so the
	// pointer + omitempty discipline matters here as much as anywhere.
	m := marshalToMap(t, FirewallClusterOption{})
	for _, k := range []string{"enable", "ebtables", "log_ratelimit", "policy_in", "policy_out", "policy_forward"} {
		_, present := m[k]
		assert.Falsef(t, present, "%q should be omitted when unset so PVE keeps its server-side default", k)
	}
}

func TestFirewallClusterOption_ExplicitEbtablesDisableSurvives(t *testing.T) {
	// The regression we're guarding: a caller wanting to disable ebtables
	// cluster-wide. Pre-pointer, declaring Ebtables as IntOrBool with
	// omitempty would drop an explicit false. With *IntOrBool, an explicit
	// Ptr(false) marshals as 0 and reaches the server.
	opts := FirewallClusterOption{Ebtables: Ptr(IntOrBool(false))}
	m := marshalToMap(t, opts)
	assert.Equal(t, float64(0), m["ebtables"], "explicit ebtables=false must reach the server, not be swallowed by omitempty")
}

func TestReplicationJobOptions_ScheduleOmittedWhenUnset(t *testing.T) {
	// PVE default schedule is "*/15" (every 15 minutes). A nil Schedule
	// pointer must omit, so the server keeps that default instead of being
	// overridden by an empty string.
	m := marshalToMap(t, ReplicationJobOptions{ID: "100-0", Target: "node2", Type: "local"})
	_, present := m["schedule"]
	assert.False(t, present, "schedule must be omitted when unset so PVE applies its */15 default")
}

func TestHAResource_DefaultsOmittedWhenUnset(t *testing.T) {
	// PVE defaults are: state="started", failback=1, max_relocate=1,
	// max_restart=1. All four are pointer-typed; bare struct must not ship
	// them so the server keeps its documented defaults.
	m := marshalToMap(t, HAResource{SID: "vm:100"})
	for _, k := range []string{"state", "failback", "max_relocate", "max_restart"} {
		_, present := m[k]
		assert.Falsef(t, present, "%q should be omitted when unset so PVE applies its default", k)
	}
}

func TestHAResource_ExplicitDisableFailbackSurvives(t *testing.T) {
	// Regression we're guarding: caller wants to *disable* automatic failback
	// (PVE default is 1=on). Plain IntOrBool with omitempty would drop the
	// false. *IntOrBool keeps it.
	r := HAResource{
		SID:         "vm:100",
		Failback:    Ptr(IntOrBool(false)),
		MaxRelocate: Ptr(0),
	}
	m := marshalToMap(t, r)
	assert.Equal(t, float64(0), m["failback"], "explicit failback=false must reach PVE, not be dropped")
	assert.Equal(t, float64(0), m["max_relocate"], "explicit max_relocate=0 must reach PVE")
}

func TestFirewallClusterOption_ExplicitValuesMarshal(t *testing.T) {
	// Wire-format correctness: Enable is integer per schema, Ebtables is
	// boolean. Both must land as 0/1 numerics — never `true`/`false`. The
	// *IntOrBool method guarantees that for Ebtables.
	opts := FirewallClusterOption{Enable: 1, Ebtables: Ptr(IntOrBool(true))}
	m := marshalToMap(t, opts)
	assert.Equal(t, float64(1), m["enable"])
	assert.Equal(t, float64(1), m["ebtables"], "schema is boolean — wire format must be 0/1")
}

func TestStorageDownloadURLOptions_VerifyCertificatesDefaultsToServerDefault(t *testing.T) {
	// Highest-stakes field in the audit: pre-fix, NOT setting
	// VerifyCertificates silently sent 0 (because the field had no
	// omitempty), which disabled certificate verification. Post-fix it's
	// *IntOrBool + omitempty, so unset → omitted → PVE default of 1
	// (verify) applies.
	m := marshalToMap(t, StorageDownloadURLOptions{Storage: "local", URL: "https://example.com/iso"})
	_, present := m["verify-certificates"]
	assert.Falsef(t, present, "verify-certificates must be omitted when unset; pre-fix it silently sent 0 and disabled cert verification")
}
