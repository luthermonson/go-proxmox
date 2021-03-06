package proxmox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	client := ClientFromEnv()
	_, err := client.Version()
	assert.Equal(t, err, ErrNotAuthorized)

	err = client.Login(td.username, td.password)
	assert.Nil(t, err)

	version, err := client.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, version.Version)
}

func TestAPIToken(t *testing.T) {
	client := ClientFromEnv()
	_, err := client.Version()
	assert.Equal(t, err, ErrNotAuthorized)

	client.APIToken(td.tokenID, td.secret)
	version, err := client.Version()
	assert.Nil(t, err)
	assert.NotNil(t, version.Version)
}
