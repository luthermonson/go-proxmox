package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_Nodes(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	nodes, err := client.Nodes(ctx)
	assert.Nil(t, err)
	for _, n := range nodes {
		assert.Contains(t, n.Node, "node")
		assert.Equal(t, n.Type, "node")
	}
	//assert.Equal(t, 6, len(testData))
}

func TestClient_Node(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)
	assert.Equal(t, "node1", node.Name)
	assert.NotNil(t, node.client)

	v, err := node.Version(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "9.1", v.Release)

	_, err = client.Node(ctx, "doesntexist")
	assert.NotNil(t, err)
}

func TestNode_Report(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	report, err := node.Report(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, report)
	assert.Contains(t, report, "pve-manager")
	assert.Contains(t, report, "kernel")
}

func TestNode_TermProxy(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	term, err := node.TermProxy(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, term)
	assert.Greater(t, term.Port, 0)
	assert.NotEmpty(t, term.Ticket)
	assert.NotEmpty(t, term.User)
}

func TestNode_Appliances(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	appliances, err := node.Appliances(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, appliances)
	assert.GreaterOrEqual(t, len(appliances), 1)

	// Check first appliance
	assert.NotEmpty(t, appliances[0].Template)
	assert.NotEmpty(t, appliances[0].Type)
	assert.NotEmpty(t, appliances[0].Os)
}

func TestNode_DownloadAppliance(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	ret, err := node.DownloadAppliance(ctx, "ubuntu-22.04-standard", "local")
	assert.Nil(t, err)
	assert.NotEmpty(t, ret)
	assert.Contains(t, ret, "UPID")
}

func TestNode_Storages(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	storages, err := node.Storages(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, storages)
	assert.GreaterOrEqual(t, len(storages), 1)

	// Check first storage
	storage := storages[0]
	assert.NotEmpty(t, storage.Name)
	assert.NotEmpty(t, storage.Type)
	assert.Equal(t, node.Name, storage.Node)
	assert.NotNil(t, storage.client)
}

func TestNode_Storage(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	storage, err := node.Storage(ctx, "local")
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, "local", storage.Name)
	assert.Equal(t, node.Name, storage.Node)
	assert.NotNil(t, storage.client)
}

func TestNode_VzTmpls(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	templates, err := node.VzTmpls(ctx, "local")
	assert.Nil(t, err)
	assert.NotEmpty(t, templates)
	assert.GreaterOrEqual(t, len(templates), 1)

	// Check first template
	assert.NotEmpty(t, templates[0].VolID)
	assert.Equal(t, "vztmpl", templates[0].Content.Content)
}

func TestNode_VzTmpl(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	template, err := node.VzTmpl(ctx, "ubuntu-22.04-standard_22.04-1_amd64.tar.zst", "local")
	assert.Nil(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "local:vztmpl/ubuntu-22.04-standard_22.04-1_amd64.tar.zst", template.VolID)

	// Test non-existent template
	template, err = node.VzTmpl(ctx, "nonexistent.tar.zst", "local")
	assert.NotNil(t, err)
	assert.Nil(t, template)
}

func TestNode_StorageDownloadURL(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	options := &StorageDownloadURLOptions{
		Storage:  "local",
		Content:  "iso",
		Filename: "test.iso",
		URL:      "http://example.com/test.iso",
	}

	ret, err := node.StorageDownloadURL(ctx, options)
	assert.Nil(t, err)
	assert.NotEmpty(t, ret)
	assert.Contains(t, ret, "UPID")
}

func TestNode_StorageByContent(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	// Test StorageISO
	storage, err := node.StorageISO(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Contains(t, storage.Content, "iso")

	// Test StorageVZTmpl
	storage, err = node.StorageVZTmpl(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Contains(t, storage.Content, "vztmpl")

	// Test StorageBackup
	storage, err = node.StorageBackup(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Contains(t, storage.Content, "backup")

	// Test StorageRootDir
	storage, err = node.StorageRootDir(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Contains(t, storage.Content, "rootdir")

	// Test StorageImages
	storage, err = node.StorageImages(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, storage)
	assert.Contains(t, storage.Content, "images")
}

func TestNode_FirewallOptionGet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	options, err := node.FirewallOptionGet(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, options)
}

func TestNode_FirewallOptionSet(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	options := &FirewallNodeOption{
		Enable: true,
	}

	err = node.FirewallOptionSet(ctx, options)
	assert.Nil(t, err)
}

func TestNode_FirewallGetRules(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	rules, err := node.FirewallGetRules(ctx)
	assert.Nil(t, err)
	assert.NotEmpty(t, rules)
	assert.GreaterOrEqual(t, len(rules), 1)

	// Check first rule
	rule := rules[0]
	assert.NotNil(t, rule.Pos)
	assert.NotEmpty(t, rule.Type)
	assert.NotEmpty(t, rule.Action)
}

func TestNode_FirewallRulesCreate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	rule := &FirewallRule{
		Type:   "in",
		Action: "ACCEPT",
		Enable: 1,
		Proto:  "tcp",
		Dport:  "22",
	}

	err = node.FirewallRulesCreate(ctx, rule)
	assert.Nil(t, err)
}

func TestNode_FirewallRulesUpdate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	rule := &FirewallRule{
		Pos:    0,
		Type:   "in",
		Action: "DROP",
		Enable: 1,
		Proto:  "tcp",
		Dport:  "22",
	}

	err = node.FirewallRulesUpdate(ctx, rule)
	assert.Nil(t, err)
}

func TestNode_FirewallRulesDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	err = node.FirewallRulesDelete(ctx, 0)
	assert.Nil(t, err)
}

func TestNode_GetCustomCertificates(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	certs, err := node.GetCustomCertificates(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, certs)
}

func TestNode_UploadCustomCertificate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	cert := &CustomCertificate{
		Certificates: "-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKZx...\n-----END CERTIFICATE-----",
		Key:          "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0B...\n-----END PRIVATE KEY-----",
	}

	err = node.UploadCustomCertificate(ctx, cert)
	assert.Nil(t, err)
}

func TestNode_DeleteCustomCertificate(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	err = node.DeleteCustomCertificate(ctx)
	assert.Nil(t, err)
}

func TestNode_Vzdump(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	params := &VirtualMachineBackupOptions{
		VMID:          100,
		Storage:       "local",
		Mode:          "snapshot",
		Compress:      "zstd",
		Remove:        false,
		All:           false,
		NotesTemplate: "",
	}

	task, err := node.Vzdump(ctx, params)
	assert.Nil(t, err)
	assert.NotNil(t, task)
	assert.NotEmpty(t, task.UPID)
}

func TestNode_VzdumpExtractConfig(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	node, err := client.Node(ctx, "node1")
	assert.Nil(t, err)

	config, err := node.VzdumpExtractConfig(ctx, "local:backup/vzdump-lxc-100-2024_01_01-00_00_00.tar.zst")
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, uint64(2), config.Cores)
	assert.Equal(t, "debian", config.OsType)
}
