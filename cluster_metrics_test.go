package proxmox

import (
	"context"
	"testing"

	"github.com/luthermonson/go-proxmox/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCluster_MetricServers(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	servers, err := cluster.MetricServers(ctx)
	assert.Nil(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, "influx1", servers[0].ID)
	assert.Equal(t, "influxdb", servers[0].Type)
	assert.Equal(t, 8086, servers[0].Port)
}

func TestCluster_MetricServer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	server, err := cluster.MetricServer(ctx, "influx1")
	assert.Nil(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, "influx1", server.ID)
	assert.Equal(t, "influxdb", server.Type)
	assert.Equal(t, "metrics.example.com", server.Server)
	assert.Equal(t, "http", server.InfluxDBProto)
	assert.Equal(t, "abc123", server.Digest)
}

func TestCluster_MetricServer_EmptyID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	server, err := cluster.MetricServer(ctx, "")
	assert.Nil(t, server)
	assert.Error(t, err)
}

func TestCluster_NewMetricServer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewMetricServer(ctx, &ClusterMetricServerOptions{
		ID:     "influx1",
		Type:   "influxdb",
		Server: "metrics.example.com",
		Port:   8086,
	})
	assert.Nil(t, err)
}

func TestCluster_NewMetricServer_EmptyID(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.NewMetricServer(ctx, &ClusterMetricServerOptions{Type: "influxdb"})
	assert.Error(t, err)
	err = cluster.NewMetricServer(ctx, nil)
	assert.Error(t, err)
}

func TestCluster_UpdateMetricServer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.UpdateMetricServer(ctx, "influx1", &ClusterMetricServerOptions{Port: 9000})
	assert.Nil(t, err)
	// Nil opts gets defaulted internally — also acceptable.
	err = cluster.UpdateMetricServer(ctx, "influx1", nil)
	assert.Nil(t, err)
	err = cluster.UpdateMetricServer(ctx, "", &ClusterMetricServerOptions{})
	assert.Error(t, err)
}

func TestCluster_MetricsSubdirs(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	subs, err := cluster.MetricsSubdirs(context.Background())
	assert.Nil(t, err)
	assert.Contains(t, subs, "server")
	assert.Contains(t, subs, "export")
}

func TestCluster_MetricsExport(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	export, err := cluster.MetricsExport(context.Background(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, export)
	assert.Len(t, export.Data, 2)
	assert.Equal(t, "node/node1", export.Data[0].ID)
	assert.Equal(t, "gauge", export.Data[0].Type)
}

func TestCluster_MetricsExport_WithOptions(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	cluster, _ := mockClient().Cluster(context.Background())

	export, err := cluster.MetricsExport(context.Background(), &MetricsExportOptions{
		History:   true,
		LocalOnly: true,
		StartTime: 1700000000,
		NodeList:  "node1,node2",
	})
	assert.Nil(t, err)
	assert.NotNil(t, export)
}

func TestCluster_DeleteMetricServer(t *testing.T) {
	mocks.On(mockConfig)
	defer mocks.Off()
	client := mockClient()
	ctx := context.Background()

	cluster, err := client.Cluster(ctx)
	assert.Nil(t, err)

	err = cluster.DeleteMetricServer(ctx, "influx1")
	assert.Nil(t, err)
	err = cluster.DeleteMetricServer(ctx, "")
	assert.Error(t, err)
}
