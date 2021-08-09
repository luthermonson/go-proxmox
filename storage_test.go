// +build nodes

package proxmox

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Windows_1909.iso
func TestStorage_ISO(t *testing.T) {
	_, err := td.storage.ISO("doesnt-exist")
	assert.Contains(t, err.Error(), "unable to parse directory volume name 'iso/doesnt-exist'")

	isoName := "Windows_1909.iso"
	iso, err := td.storage.ISO(isoName)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(iso.Path, isoName))
}

func TestStorage_VzTmpl(t *testing.T) {
	_, err := td.storage.VzTmpl("doesnt-exist")
	assert.Contains(t, err.Error(), "unable to parse directory volume name 'vztmpl/doesnt-exist'")

	assert.NotNil(t, td.appliance)
	vztmpl, err := td.storage.VzTmpl(td.appliance.Template)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(vztmpl.Path, td.appliance.Template))

	assert.Nil(t, vztmpl.Delete())
}
