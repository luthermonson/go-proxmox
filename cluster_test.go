package proxmox

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterResources(t *testing.T) {
	client := ClientFromLogins()

	// Check a call without parameters
	rs, err := client.ClusterResources()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(rs), 1)

	re := regexp.MustCompile("^(pool|qemu|node|storage)$")
	for _, r := range rs {
		// Check types against known values
		assert.Regexp(t, re, r.Type)
	}

	// Check a call with all the valid filter values
	for _, rsType := range []string{"vm", "storage", "node", "sdn"} {
		rs, err = client.ClusterResources(rsType)
		assert.Nil(t, err)

		// vm and sdn may be empty as it is absolutely not mandatory
		if rsType == "sdn" || rsType == "vm" {
			assert.GreaterOrEqual(t, len(rs), 0)
		} else {
			assert.GreaterOrEqual(t, len(rs), 1)
		}

		// api v2 returns type = qemu when filtering on vm
		if rsType == "vm" {
			rsType = "qemu"
		}

		// Check that every resource returned if of the asked type
		for _, r := range rs {
			assert.Equal(t, rsType, r.Type)
		}
	}

	// Check a call with more than one parameter
	rs, err = client.ClusterResources("bad", "call")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "accepts maximum one parameter")

	// Check a call with a string parameter which is not a single word
	rs, err = client.ClusterResources("bad call")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "accepts only a single word")

	// Check a call with a string parameter which is a word
	rs, err = client.ClusterResources("unknownword")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "bad request: 400 Parameter verification failed")
}
