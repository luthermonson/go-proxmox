package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_TFAUsers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	users, err := mockClient().TFAUsers(context.Background())
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "alice@pve", users[0].UserID)
	assert.True(t, users[0].TOTP)
}

func TestClient_TFAEntries(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entries, err := mockClient().TFAEntries(context.Background(), "alice@pve")
	assert.Nil(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "totp-1", entries[0].ID)

	_, err = mockClient().TFAEntries(context.Background(), "")
	assert.NotNil(t, err)
}

func TestClient_TFAEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	entry, err := mockClient().TFAEntry(context.Background(), "alice@pve", "totp-1")
	assert.Nil(t, err)
	assert.Equal(t, "phone", entry.Description)

	_, err = mockClient().TFAEntry(context.Background(), "", "totp-1")
	assert.NotNil(t, err)
}

func TestClient_NewTFAEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	id, err := mockClient().NewTFAEntry(context.Background(), "alice@pve", &TFAEntryOptions{
		Type:        "totp",
		Description: "phone",
		TOTP:        "otpauth://totp/Proxmox:alice?secret=ABCD&issuer=Proxmox",
		Value:       "123456",
	})
	assert.Nil(t, err)
	assert.Equal(t, "totp-2", id)

	_, err = mockClient().NewTFAEntry(context.Background(), "alice@pve", &TFAEntryOptions{})
	assert.NotNil(t, err)

	_, err = mockClient().NewTFAEntry(context.Background(), "", nil)
	assert.NotNil(t, err)
}

func TestClient_UpdateTFAEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	enable := false
	err := mockClient().UpdateTFAEntry(context.Background(), "alice@pve", "totp-1", &TFAEntryUpdateOptions{Enable: &enable})
	assert.Nil(t, err)

	err = mockClient().UpdateTFAEntry(context.Background(), "", "totp-1", nil)
	assert.NotNil(t, err)
}

func TestClient_DeleteTFAEntry(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := mockClient().DeleteTFAEntry(context.Background(), "alice@pve", "totp-1", "currentpw")
	assert.Nil(t, err)

	err = mockClient().DeleteTFAEntry(context.Background(), "alice@pve", "totp-1", "")
	assert.Nil(t, err)
}

func TestClient_UnlockUserTFA(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	err := mockClient().UnlockUserTFA(context.Background(), "alice@pve")
	assert.Nil(t, err)

	err = mockClient().UnlockUserTFA(context.Background(), "")
	assert.NotNil(t, err)
}
