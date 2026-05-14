package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// MetricServers lists configured external metric servers (graphite / influxdb /
// opentelemetry). See https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/metrics/server
func (cl *Cluster) MetricServers(ctx context.Context) (ClusterMetricServers, error) {
	var servers ClusterMetricServers
	if err := cl.client.Get(ctx, "/cluster/metrics/server", &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

// MetricServer reads the full configuration of a single metric server.
func (cl *Cluster) MetricServer(ctx context.Context, id string) (*ClusterMetricServer, error) {
	if id == "" {
		return nil, errors.New("metric server id can not be empty")
	}
	server := &ClusterMetricServer{}
	if err := cl.client.Get(ctx, fmt.Sprintf("/cluster/metrics/server/%s", id), server); err != nil {
		return nil, err
	}
	if server.ID == "" {
		server.ID = id
	}
	return server, nil
}

// NewMetricServer creates a new external metric server entry. Requires opts.ID
// and opts.Type ("graphite" | "influxdb" | "opentelemetry").
func (cl *Cluster) NewMetricServer(ctx context.Context, opts *ClusterMetricServerOptions) error {
	if opts == nil || opts.ID == "" {
		return errors.New("metric server id can not be empty")
	}
	// PVE puts the id in the URL, not the body, on create.
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/metrics/server/%s", opts.ID), opts, nil)
}

// UpdateMetricServer mutates an existing metric server entry. The opts.Delete
// field is a comma-separated list of keys to reset (PVE quirk).
func (cl *Cluster) UpdateMetricServer(ctx context.Context, id string, opts *ClusterMetricServerOptions) error {
	if id == "" {
		return errors.New("metric server id can not be empty")
	}
	if opts == nil {
		opts = &ClusterMetricServerOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/metrics/server/%s", id), opts, nil)
}

// DeleteMetricServer removes a configured metric server.
func (cl *Cluster) DeleteMetricServer(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("metric server id can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/metrics/server/%s", id), nil)
}
