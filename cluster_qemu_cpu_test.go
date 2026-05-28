package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_QEMUCPUFlags(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	flags, err := cluster.QEMUCPUFlags(context.Background(), "", "")
	assert.Nil(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "aes", flags[0].Name)
}

func TestCluster_QEMUCPUFlags_WithArchAccel(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	flags, err := cluster.QEMUCPUFlags(context.Background(), "x86_64", "kvm")
	assert.Nil(t, err)
	assert.Len(t, flags, 2)
}

func TestCluster_CustomCPUModels(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	models, err := cluster.CustomCPUModels(context.Background())
	assert.Nil(t, err)
	assert.Len(t, models, 1)
	assert.Equal(t, "custom-epyc", models[0].CPUType)
}

func TestCustomCPUModel_CRUD(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	m := cluster.CustomCPUModel("custom-epyc")
	assert.Nil(t, m.Read(context.Background()))
	assert.Equal(t, "EPYC", m.ReportedModel)

	assert.Nil(t, m.Update(context.Background(), &CustomCPUModelOptions{Flags: "+aes"}))
	assert.Nil(t, m.Delete(context.Background()))

	// blank cputype guard
	empty := cluster.CustomCPUModel("")
	assert.Error(t, empty.Read(context.Background()))
	assert.Error(t, empty.Update(context.Background(), nil))
	assert.Error(t, empty.Delete(context.Background()))
}

func TestCluster_NewCustomCPUModel(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewCustomCPUModel(context.Background(), &CustomCPUModelOptions{
		CPUType:       "custom-epyc",
		ReportedModel: "EPYC",
	})
	assert.Nil(t, err)

	assert.Error(t, cluster.NewCustomCPUModel(context.Background(), nil))
	assert.Error(t, cluster.NewCustomCPUModel(context.Background(), &CustomCPUModelOptions{CPUType: "x"}))
}
