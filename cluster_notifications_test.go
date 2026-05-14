package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_Notifications(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	entries, err := cluster.Notifications(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(entries), 3)
}

func TestCluster_NotificationMatcherFields(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	fields, err := cluster.NotificationMatcherFields(ctx)
	assert.Nil(t, err)
	assert.Len(t, fields, 2)
	assert.Equal(t, "type", fields[0].Name)
}

func TestCluster_NotificationMatcherFieldValues(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	values, err := cluster.NotificationMatcherFieldValues(ctx)
	assert.Nil(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "type", values[0].Field)
	assert.Equal(t, "vzdump", values[0].Value)
}

// --- targets ---------------------------------------------------------------

func TestCluster_NotificationTargets(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	targets, err := cluster.NotificationTargets(ctx)
	assert.Nil(t, err)
	assert.Len(t, targets, 2)
	assert.Equal(t, "mail-to-root", targets[0].Name)
	assert.Equal(t, "sendmail", targets[0].Type)
}

func TestCluster_TestNotificationTarget(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.TestNotificationTarget(ctx, "mail-to-root")
	assert.Nil(t, err)
	err = cluster.TestNotificationTarget(ctx, "")
	assert.Error(t, err)
}

// --- matchers --------------------------------------------------------------

func TestCluster_NotificationMatchers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	matchers, err := cluster.NotificationMatchers(ctx)
	assert.Nil(t, err)
	assert.Len(t, matchers, 1)
	assert.Equal(t, "default-matcher", matchers[0].Name)
}

func TestCluster_NotificationMatcher(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	m, err := cluster.NotificationMatcher(ctx, "default-matcher")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "default-matcher", m.Name)
	assert.Contains(t, m.MatchSeverity, "warning")
	assert.Equal(t, "m1", m.Digest)
	_, err = cluster.NotificationMatcher(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewNotificationMatcher(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewNotificationMatcher(ctx, &ClusterNotificationMatcherOptions{
		Name:          "default-matcher",
		Mode:          "all",
		MatchSeverity: []string{"warning", "error"},
		Target:        []string{"mail-to-root"},
	})
	assert.Nil(t, err)
	err = cluster.NewNotificationMatcher(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateNotificationMatcher(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateNotificationMatcher(ctx, "default-matcher", &ClusterNotificationMatcherOptions{
		Comment: "updated",
	})
	assert.Nil(t, err)
	err = cluster.UpdateNotificationMatcher(ctx, "default-matcher", nil)
	assert.Nil(t, err)
	err = cluster.UpdateNotificationMatcher(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteNotificationMatcher(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteNotificationMatcher(ctx, "default-matcher")
	assert.Nil(t, err)
	err = cluster.DeleteNotificationMatcher(ctx, "")
	assert.Error(t, err)
}

// --- gotify ----------------------------------------------------------------

func TestCluster_NotificationGotifyEndpoints(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	endpoints, err := cluster.NotificationGotifyEndpoints(ctx)
	assert.Nil(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "gotify1", endpoints[0].Name)
}

func TestCluster_NotificationGotifyEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	e, err := cluster.NotificationGotifyEndpoint(ctx, "gotify1")
	assert.Nil(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, "gotify1", e.Name)
	assert.Equal(t, "g1", e.Digest)
	_, err = cluster.NotificationGotifyEndpoint(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewNotificationGotifyEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewNotificationGotifyEndpoint(ctx, &ClusterNotificationGotifyOptions{
		Name:   "gotify1",
		Server: "https://gotify.example.com",
		Token:  "tok",
	})
	assert.Nil(t, err)
	err = cluster.NewNotificationGotifyEndpoint(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateNotificationGotifyEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateNotificationGotifyEndpoint(ctx, "gotify1", &ClusterNotificationGotifyOptions{Comment: "x"})
	assert.Nil(t, err)
	err = cluster.UpdateNotificationGotifyEndpoint(ctx, "gotify1", nil)
	assert.Nil(t, err)
	err = cluster.UpdateNotificationGotifyEndpoint(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteNotificationGotifyEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteNotificationGotifyEndpoint(ctx, "gotify1")
	assert.Nil(t, err)
	err = cluster.DeleteNotificationGotifyEndpoint(ctx, "")
	assert.Error(t, err)
}

// --- sendmail --------------------------------------------------------------

func TestCluster_NotificationSendmailEndpoints(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	endpoints, err := cluster.NotificationSendmailEndpoints(ctx)
	assert.Nil(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "mail-to-root", endpoints[0].Name)
}

func TestCluster_NotificationSendmailEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	e, err := cluster.NotificationSendmailEndpoint(ctx, "mail-to-root")
	assert.Nil(t, err)
	assert.Equal(t, "mail-to-root", e.Name)
	assert.Equal(t, "s1", e.Digest)
	_, err = cluster.NotificationSendmailEndpoint(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewNotificationSendmailEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewNotificationSendmailEndpoint(ctx, &ClusterNotificationSendmailOptions{
		Name:       "mail-to-root",
		MailToUser: []string{"root@pam"},
	})
	assert.Nil(t, err)
	err = cluster.NewNotificationSendmailEndpoint(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateNotificationSendmailEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateNotificationSendmailEndpoint(ctx, "mail-to-root", &ClusterNotificationSendmailOptions{Comment: "x"})
	assert.Nil(t, err)
	err = cluster.UpdateNotificationSendmailEndpoint(ctx, "mail-to-root", nil)
	assert.Nil(t, err)
	err = cluster.UpdateNotificationSendmailEndpoint(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteNotificationSendmailEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteNotificationSendmailEndpoint(ctx, "mail-to-root")
	assert.Nil(t, err)
	err = cluster.DeleteNotificationSendmailEndpoint(ctx, "")
	assert.Error(t, err)
}

// --- smtp ------------------------------------------------------------------

func TestCluster_NotificationSMTPEndpoints(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	endpoints, err := cluster.NotificationSMTPEndpoints(ctx)
	assert.Nil(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "smtp1", endpoints[0].Name)
	assert.Equal(t, "starttls", endpoints[0].Mode)
}

func TestCluster_NotificationSMTPEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	e, err := cluster.NotificationSMTPEndpoint(ctx, "smtp1")
	assert.Nil(t, err)
	assert.Equal(t, "smtp1", e.Name)
	assert.Equal(t, 587, e.Port)
	assert.Equal(t, "st1", e.Digest)
	_, err = cluster.NotificationSMTPEndpoint(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewNotificationSMTPEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewNotificationSMTPEndpoint(ctx, &ClusterNotificationSMTPOptions{
		Name:        "smtp1",
		Server:      "smtp.example.com",
		Port:        587,
		Mode:        "starttls",
		FromAddress: "alerts@example.com",
	})
	assert.Nil(t, err)
	err = cluster.NewNotificationSMTPEndpoint(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateNotificationSMTPEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateNotificationSMTPEndpoint(ctx, "smtp1", &ClusterNotificationSMTPOptions{Password: "secret"})
	assert.Nil(t, err)
	err = cluster.UpdateNotificationSMTPEndpoint(ctx, "smtp1", nil)
	assert.Nil(t, err)
	err = cluster.UpdateNotificationSMTPEndpoint(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteNotificationSMTPEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteNotificationSMTPEndpoint(ctx, "smtp1")
	assert.Nil(t, err)
	err = cluster.DeleteNotificationSMTPEndpoint(ctx, "")
	assert.Error(t, err)
}

// --- webhook ---------------------------------------------------------------

func TestCluster_NotificationWebhookEndpoints(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	endpoints, err := cluster.NotificationWebhookEndpoints(ctx)
	assert.Nil(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "wh1", endpoints[0].Name)
	assert.Equal(t, "post", endpoints[0].Method)
}

func TestCluster_NotificationWebhookEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	e, err := cluster.NotificationWebhookEndpoint(ctx, "wh1")
	assert.Nil(t, err)
	assert.Equal(t, "wh1", e.Name)
	assert.Equal(t, "https://hook.example.com/alert", e.URL)
	assert.Equal(t, "w1", e.Digest)
	_, err = cluster.NotificationWebhookEndpoint(ctx, "")
	assert.Error(t, err)
}

func TestCluster_NewNotificationWebhookEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewNotificationWebhookEndpoint(ctx, &ClusterNotificationWebhookOptions{
		Name:   "wh1",
		URL:    "https://hook.example.com/alert",
		Method: "post",
	})
	assert.Nil(t, err)
	err = cluster.NewNotificationWebhookEndpoint(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateNotificationWebhookEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateNotificationWebhookEndpoint(ctx, "wh1", &ClusterNotificationWebhookOptions{Comment: "x"})
	assert.Nil(t, err)
	err = cluster.UpdateNotificationWebhookEndpoint(ctx, "wh1", nil)
	assert.Nil(t, err)
	err = cluster.UpdateNotificationWebhookEndpoint(ctx, "", nil)
	assert.Error(t, err)
}

func TestCluster_DeleteNotificationWebhookEndpoint(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteNotificationWebhookEndpoint(ctx, "wh1")
	assert.Nil(t, err)
	err = cluster.DeleteNotificationWebhookEndpoint(ctx, "")
	assert.Error(t, err)
}
