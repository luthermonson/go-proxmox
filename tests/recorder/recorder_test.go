package recorder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	proxmox "github.com/luthermonson/go-proxmox"
	"github.com/luthermonson/go-proxmox/tests/recorder"
)

// TestRecorder_Smoke is the canary that proves the record→scrub→replay loop is
// wired correctly. It calls Client.Version, which is the cheapest non-trivial
// Proxmox endpoint (one GET /version, no auth, no body), and asserts a known
// version string from the checked-in cassette.
//
// In replay mode (PROXMOX_URL unset) the recorder serves testdata/TestRecorder_Smoke.yaml
// and the test must pass without touching the network.
//
// In record mode (PROXMOX_URL set) the recorder will hit the live PVE host,
// scrub the cassette via the BeforeSaveHook chain, and write it back. The
// scrubbed cassette should produce a stable diff across re-records.
func TestRecorder_Smoke(t *testing.T) {
	r := recorder.New(t)

	client := proxmox.NewClient(recorder.FakeURL,
		proxmox.WithHTTPClient(r.GetDefaultClient()))

	v, err := client.Version(context.Background())
	require.NoError(t, err, "Version() must succeed against a recorded cassette")
	require.NotNil(t, v)
	require.NotEmpty(t, v.Version, "cassette must populate Version")
	require.NotEmpty(t, v.Release, "cassette must populate Release")
}
