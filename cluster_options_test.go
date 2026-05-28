package proxmox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_ClusterOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	opts, err := cluster.ClusterOptions(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, opts)
	assert.Equal(t, "html5", opts.Console)
	assert.Equal(t, "en", opts.Language)
	assert.Equal(t, 4, opts.MaxWorkers)

	// Extra picks up the long-tail keys.
	assert.NotNil(t, opts.Extra)
	assert.Contains(t, opts.Extra, "ha")
	assert.Contains(t, opts.Extra, "next-id")
}

func TestCluster_UpdateClusterOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.UpdateClusterOptions(context.Background(), &ClusterOptionsUpdate{
		Console:  "xtermjs",
		Language: "en",
	})
	assert.Nil(t, err)

	err = cluster.UpdateClusterOptions(context.Background(), nil)
	assert.Nil(t, err)
}

func TestClusterOptionsUpdate_MarshalIncludesExtra(t *testing.T) {
	u := ClusterOptionsUpdate{
		Console: "xtermjs",
		Extra: map[string]any{
			"ha":      "shutdown_policy=conditional",
			"next-id": "lower=100,upper=1000000",
		},
	}
	b, err := json.Marshal(u)
	assert.Nil(t, err)
	m := map[string]any{}
	assert.Nil(t, json.Unmarshal(b, &m))
	assert.Equal(t, "xtermjs", m["console"])
	assert.Equal(t, "shutdown_policy=conditional", m["ha"])
	assert.Equal(t, "lower=100,upper=1000000", m["next-id"])
}
