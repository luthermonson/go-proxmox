package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func diridxCluster() *Cluster {
	return &Cluster{client: mockClient()}
}

func TestCluster_Subdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().Subdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "replication")
	assert.Contains(t, subdirs, "ha")
	assert.Contains(t, subdirs, "sdn")
	assert.Contains(t, subdirs, "ceph")
}

func TestCluster_ACMESubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().ACMESubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "plugins")
	assert.Contains(t, subdirs, "account")
	assert.Contains(t, subdirs, "directories")
}

func TestCluster_FirewallSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().FirewallSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "groups")
	assert.Contains(t, subdirs, "rules")
	assert.Contains(t, subdirs, "options")
}

func TestCluster_SDNSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().SDNSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "vnets")
	assert.Contains(t, subdirs, "zones")
	assert.Contains(t, subdirs, "controllers")
}

func TestCluster_CephSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().CephSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"metadata", "status", "flags"}, subdirs)
}

func TestCluster_ConfigSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().ConfigSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "nodes")
	assert.Contains(t, subdirs, "join")
	assert.Contains(t, subdirs, "totem")
}

func TestCluster_HASubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().HASubdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"groups", "resources", "status", "rules"}, subdirs)
}

func TestCluster_HAStatusSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().HAStatusSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"current", "manager_status"}, subdirs)
}

func TestCluster_QEMUSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().QEMUSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"100", "101", "200"}, subdirs)
}
