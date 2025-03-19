package proxmox

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luthermonson/go-proxmox/tests/mocks"
)

func TestClusterCeph_Status(t *testing.T) {
	mocks.ProxmoxVE8x(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()
	ceph := Ceph{
		client: client,
	}

	expectedOsdNearFullCheck := CephHealthCheck{
		Detail: []CephHealthCheckDetail{
			{
				Message: "osd.3 is near full",
			},
			{
				Message: "osd.8 is near full",
			},
			{
				Message: "osd.9 is near full",
			},
		},
		Muted:    false,
		Severity: "HEALTH_WARN",
		Summary: CephHealthCheckSummary{
			Count:   3,
			Message: "3 nearfull osd(s)",
		},
	}

	actual, err := ceph.Status(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "HEALTH_WARN", actual.Health.Status)
	assert.Equal(t, []string{"proxmox-node01", "proxmox-node03", "proxmox-node02"}, actual.QuorumNames)
	assert.Equal(t, expectedOsdNearFullCheck, actual.Health.Checks["OSD_NEARFULL"])
}
