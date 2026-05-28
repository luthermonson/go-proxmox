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

func TestCluster_ClusterIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().ClusterIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "replication")
	assert.Contains(t, subdirs, "ha")
	assert.Contains(t, subdirs, "sdn")
	assert.Contains(t, subdirs, "ceph")
}

func TestCluster_ACMEIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().ACMEIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "plugins")
	assert.Contains(t, subdirs, "account")
	assert.Contains(t, subdirs, "directories")
}

func TestCluster_FirewallIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().FirewallIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "groups")
	assert.Contains(t, subdirs, "rules")
	assert.Contains(t, subdirs, "options")
}

func TestCluster_SDNIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().SDNIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "vnets")
	assert.Contains(t, subdirs, "zones")
	assert.Contains(t, subdirs, "controllers")
}

func TestCluster_CephIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().CephIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"metadata", "status", "flags"}, subdirs)
}

func TestCluster_ConfigIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().ConfigIndex(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subdirs, "nodes")
	assert.Contains(t, subdirs, "join")
	assert.Contains(t, subdirs, "totem")
}

func TestCluster_HAIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().HAIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"groups", "resources", "status", "rules"}, subdirs)
}

func TestCluster_HAStatusIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().HAStatusIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"current", "manager_status"}, subdirs)
}

func TestCluster_QEMUIndex(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	subdirs, err := diridxCluster().QEMUIndex(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []string{"100", "101", "200"}, subdirs)
}
