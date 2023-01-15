//go:build nodes
// +build nodes

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStorage_ISO(t *testing.T) {
	_, err := td.storage.ISO("doesnt-exist")
	assert.Contains(t, err.Error(), "unable to parse directory volume name 'iso/doesnt-exist'")
}

func TestStorage_DownloadUrl(t *testing.T) {
	// download url
	isoName := nameGenerator(12) + ".iso"
	task, err := td.storage.DownloadURL("iso", isoName, tinycoreURL)
	assert.Nil(t, err)
	assert.Nil(t, task.Wait(time.Duration(5*time.Second), time.Duration(5*time.Minute)))

	iso, err := td.storage.ISO(isoName)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(iso.Path, isoName))
	task, err = iso.Delete()
	assert.Nil(t, err)
	task.Wait(1*time.Second, 10*time.Second)
}

func TestStorage_Upload(t *testing.T) {
	// upload from local file
	isoName := nameGenerator(12) + ".iso"
	file := filepath.Join("./", isoName)
	createTestISO(file)
	defer os.Remove(file)

	task, err := td.storage.Upload("iso", file)
	assert.Nil(t, err)
	task.Wait(1*time.Second, 5*time.Second)
	iso, err := td.storage.ISO(isoName)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(iso.Path, isoName))

	task, err = iso.Delete()
	assert.Nil(t, err)
	task.Wait(1*time.Second, 15*time.Second)
}

func TestStorage_VzTmpl(t *testing.T) {
	_, err := td.storage.VzTmpl("doesnt-exist")
	assert.Contains(t, err.Error(), "unable to parse directory volume name 'vztmpl/doesnt-exist'")

	name := nameGenerator(12) + ".tar.xz"
	task, err := td.storage.DownloadURL("vztmpl", name, alpineAppliance)
	assert.Nil(t, err)
	task.Wait(1*time.Second, 5*time.Second)

	vztmpl, err := td.storage.VzTmpl(name)
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(vztmpl.Path, "tar.xz"))

	task, err = vztmpl.Delete()
	assert.Nil(t, err)
	task.Wait(1*time.Second, 15*time.Second)
}
