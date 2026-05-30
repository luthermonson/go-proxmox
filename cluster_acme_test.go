package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_ACMEDirectories(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	dirs, err := cluster.ACMEDirectories(ctx)
	assert.Nil(t, err)
	assert.Len(t, dirs, 2)
	assert.Equal(t, "Let's Encrypt V2", dirs[0].Name)
	assert.Contains(t, dirs[0].URL, "acme-v02.api.letsencrypt.org")
}

func TestCluster_ACMEChallengeSchema(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	schemas, err := cluster.ACMEChallengeSchema(ctx)
	assert.Nil(t, err)
	assert.Len(t, schemas, 2)
	assert.Equal(t, "dns-01", schemas[0].ID)
}

func TestCluster_ACMETermsOfService(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	tosURL, err := cluster.ACMETermsOfService(ctx, "")
	assert.Nil(t, err)
	assert.Contains(t, tosURL, "letsencrypt.org")
}

func TestCluster_ACMEMeta(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	meta, err := cluster.ACMEMeta(ctx, "https://acme-v02.api.letsencrypt.org/directory")
	assert.Nil(t, err)
	assert.NotNil(t, meta)
	assert.Contains(t, meta.CAAIdentities, "letsencrypt.org")
	assert.Contains(t, meta.Website, "letsencrypt.org")
}

func TestCluster_ACMEAccounts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	accounts, err := cluster.ACMEAccounts(ctx)
	assert.Nil(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "default", accounts[0].Name)
}

func TestCluster_ACMEAccount(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	account, err := cluster.ACMEAccount(ctx, "")
	assert.Nil(t, err)
	assert.NotNil(t, account)
	assert.Contains(t, account.Directory, "letsencrypt.org")
}

func TestCluster_NewACMEAccount(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	task, err := cluster.NewACMEAccount(ctx, &ACMEAccountOptions{Contact: "mailto:admin@example.com"})
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "acme-register", task.Type)

	_, err = cluster.NewACMEAccount(ctx, nil)
	assert.NotNil(t, err)
}

func TestCluster_UpdateACMEAccount(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	task, err := cluster.UpdateACMEAccount(ctx, "", "mailto:newadmin@example.com")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "acme-update", task.Type)
}

func TestCluster_DeleteACMEAccount(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	task, err := cluster.DeleteACMEAccount(ctx, "")
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "acme-deactivate", task.Type)
}

func TestCluster_ACMEPlugins(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	plugins, err := cluster.ACMEPlugins(ctx, "")
	assert.Nil(t, err)
	assert.Len(t, plugins, 1)
	assert.Equal(t, "cloudflare", plugins[0].ID)
	assert.Equal(t, "dns", plugins[0].Type)
}

func TestCluster_ACMEPlugin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	plugin, err := cluster.ACMEPlugin(ctx, "cloudflare")
	assert.Nil(t, err)
	assert.NotNil(t, plugin)
	assert.Equal(t, "cloudflare", plugin.ID)
	assert.Equal(t, "cf", plugin.API)
}

func TestCluster_NewACMEPlugin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewACMEPlugin(ctx, &ACMEPluginOptions{ID: "cloudflare", Type: "dns", API: "cf"})
	assert.Nil(t, err)

	err = cluster.NewACMEPlugin(ctx, nil)
	assert.NotNil(t, err)

	err = cluster.NewACMEPlugin(ctx, &ACMEPluginOptions{ID: "no-type"})
	assert.NotNil(t, err)
}

func TestCluster_UpdateACMEPlugin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateACMEPlugin(ctx, "cloudflare", &ACMEPluginOptions{API: "cf"})
	assert.Nil(t, err)

	err = cluster.UpdateACMEPlugin(ctx, "", nil)
	assert.NotNil(t, err)
}

func TestCluster_ACMETermsOfService_WithDirectory(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	tos, err := cluster.ACMETermsOfService(ctx, "https://acme-staging-v02.api.letsencrypt.org/directory")
	assert.Nil(t, err)
	assert.Contains(t, tos, "letsencrypt.org")
}

func TestCluster_ACMEPlugins_WithType(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	plugins, err := cluster.ACMEPlugins(ctx, "dns")
	assert.Nil(t, err)
	assert.NotNil(t, plugins)
}

func TestCluster_UpdateACMEPlugin_NilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	// nil opts is allowed; method should still PUT without panicking.
	assert.Nil(t, cluster.UpdateACMEPlugin(ctx, "cloudflare", nil))
}

func TestCluster_NewACMEAccount_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	// opts without contact -> error
	_, err = cluster.NewACMEAccount(ctx, &ACMEAccountOptions{})
	assert.Error(t, err)
}

func TestCluster_DeleteACMEPlugin(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	ctx := context.Background()
	cluster, err := mockClient().Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteACMEPlugin(ctx, "cloudflare")
	assert.Nil(t, err)

	err = cluster.DeleteACMEPlugin(ctx, "")
	assert.NotNil(t, err)
}
