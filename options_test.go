package proxmox

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithClient(t *testing.T) {
	httpClient := http.Client{Timeout: time.Second * 10}
	client := NewClient("", WithClient(&httpClient))
	assert.Equal(t, client.httpClient, &http.Client{Timeout: time.Second * 10})
}

func TestWithLogins(t *testing.T) {
	client := NewClient("", WithLogins("root@pam", "1234"))
	assert.Equal(t, client.credentials, &Credentials{Username: "root@pam", Password: "1234"})
}

func TestWithCredentials(t *testing.T) {
	client := NewClient("", WithCredentials(&Credentials{
		Username: "root@pam",
		Password: "1234",
	}))
	assert.Equal(t, client.credentials, &Credentials{Username: "root@pam", Password: "1234"})
}

func TestWithAPIToken(t *testing.T) {
	client := NewClient("", WithAPIToken("root@pam!test", "1234"))
	assert.Equal(t, client.token, "root@pam!test=1234")
}

func TestWithSession(t *testing.T) {
	client := NewClient("", WithSession("ticket", "csrf"))
	assert.Equal(t, client.session, &Session{Ticket: "ticket", CSRFPreventionToken: "csrf"})
}

func TestWithUserAgent(t *testing.T) {
	client := NewClient("", WithUserAgent("test-ua"))
	assert.Equal(t, client.userAgent, "test-ua")
}

func TestWithLogger(t *testing.T) {
	client := NewClient("", WithLogger(&LeveledLogger{Level: 1}))
	assert.Equal(t, client.log, &LeveledLogger{Level: 1})
}
