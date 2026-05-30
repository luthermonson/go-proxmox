package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_SDNControllers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	controllers, err := cluster.SDNControllers(context.Background(), "")
	assert.Nil(t, err)
	assert.Len(t, controllers, 2)
	assert.Equal(t, "ctrl1", controllers[0].Controller)
	assert.Equal(t, "evpn", controllers[0].Type)
}

func TestCluster_SDNController(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	c := cluster.SDNController("ctrl1")
	assert.NotNil(t, c)
	err := c.Read(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint32(65000), c.ASN)
	assert.Equal(t, "VTEP", c.PeerGroupName)
}

func TestCluster_NewSDNController(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	err := cluster.NewSDNController(context.Background(), &SDNControllerOptions{Controller: "ctrl3", Type: "bgp", ASN: 65002})
	assert.Nil(t, err)

	err = cluster.NewSDNController(context.Background(), nil)
	assert.NotNil(t, err)

	err = cluster.NewSDNController(context.Background(), &SDNControllerOptions{Controller: "ctrl3"})
	assert.NotNil(t, err)
}

func TestSDNController_UpdateDelete(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	c := cluster.SDNController("ctrl1")
	err := c.Update(context.Background(), &SDNControllerOptions{ASN: 65010})
	assert.Nil(t, err)

	err = c.Delete(context.Background())
	assert.Nil(t, err)
}

func TestCluster_SDNControllersFiltered(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	controllers, err := cluster.SDNControllers(context.Background(), "evpn")
	assert.Nil(t, err)
	assert.Len(t, controllers, 1)
	assert.Equal(t, "evpn", controllers[0].Type)
}

func TestSDNController_UpdateNilOpts(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	c := cluster.SDNController("ctrl1")
	// nil opts should be replaced with zero-value internally
	err := c.Update(context.Background(), nil)
	assert.Nil(t, err)
}

func TestSDNController_EmptyName_Errors(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()

	c := &SDNController{client: client}
	assert.Error(t, c.Read(context.Background()))
	assert.Error(t, c.Update(context.Background(), &SDNControllerOptions{}))
	assert.Error(t, c.Delete(context.Background()))
}

func TestSDNController_Read_NotFound(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	c := cluster.SDNController("missing")
	err := c.Read(context.Background())
	assert.Error(t, err)
}
