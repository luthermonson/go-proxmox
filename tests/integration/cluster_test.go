package integration

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Cluster(t *testing.T) {
	client := ClientFromLogins()
	ctx := context.Background()
	cluster, err := client.Cluster(ctx)
	assert.NoError(t, err)
	fmt.Println(cluster)
}

func TestClusterResources(t *testing.T) {
	client := ClientFromLogins()
	ctx := context.Background()

	// Check a call without parameters
	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	rs, err := cluster.Resources(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(rs), 1)

	re := regexp.MustCompile("^(pool|qemu|lxc|node|storage)$")
	for _, r := range rs {
		// Check types against known values
		assert.Regexp(t, re, r.Type)
	}

	// Check a call with all the valid filter values
	for _, rsType := range []string{"vm", "storage", "node", "sdn"} {
		rs, err = cluster.Resources(ctx, rsType)
		assert.Nil(t, err)

		// vm and sdn may be empty as it is absolutely not mandatory
		if rsType == "sdn" || rsType == "vm" {
			assert.GreaterOrEqual(t, len(rs), 0)
		} else {
			assert.GreaterOrEqual(t, len(rs), 1)
		}

		var s interface{}
		// api v2 returns type = qemu or lxc when filtering on vm
		if rsType == "vm" {
			s = []string{"qemu", "lxc"}
		} else {
			s = rsType
		}

		// Check that every resource returned if of the asked type
		for _, r := range rs {
			assert.Contains(t, s, r.Type)
		}
	}

	// Check a call with more than one parameter
	_, err = cluster.Resources(ctx, "bad", "call")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "value 'badcall' does not have a value in the enumeration 'vm, storage, node, sdn'")

	// Check a call with a string parameter which is not a single word
	_, err = cluster.Resources(ctx, "bad filter")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "value 'badfilter' does not have a value in the enumeration 'vm, storage, node, sdn'")

	// Check a call with a string parameter which is a word
	_, err = cluster.Resources(ctx, "unknownword")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "bad request: 400 Parameter verification failed")
}
