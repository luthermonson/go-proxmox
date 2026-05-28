package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// MetricsSubdirs enumerates the children of /cluster/metrics ("server",
// "export"). ACL-filtered.
//
// GET /cluster/metrics
func (cl *Cluster) MetricsSubdirs(ctx context.Context) ([]string, error) {
	return cl.metricsDiridx(ctx, "/cluster/metrics")
}

// MetricsExportOptions filters the /cluster/metrics/export response.
type MetricsExportOptions struct {
	// History returns the full historic series instead of just the latest
	// observation. PVE default false; matches Go zero, so plain bool is safe.
	History bool
	// LocalOnly restricts the output to the current node instead of the whole
	// cluster. PVE default false; matches Go zero.
	LocalOnly bool
	// StartTime: only include metrics with timestamp > start-time (unix
	// seconds). 0 disables the filter.
	StartTime int64
	// NodeList: comma-separated node names to scope the export to. Empty =
	// all nodes (or local-only when LocalOnly is true).
	NodeList string
}

// MetricsExport retrieves cluster metrics. opts is optional.
//
// GET /cluster/metrics/export
func (cl *Cluster) MetricsExport(ctx context.Context, opts *MetricsExportOptions) (export *MetricsExportResponse, err error) {
	path := "/cluster/metrics/export"
	if opts != nil {
		q := url.Values{}
		if opts.History {
			q.Set("history", "1")
		}
		if opts.LocalOnly {
			q.Set("local-only", "1")
		}
		if opts.StartTime > 0 {
			q.Set("start-time", strconv.FormatInt(opts.StartTime, 10))
		}
		if opts.NodeList != "" {
			q.Set("node-list", opts.NodeList)
		}
		if enc := q.Encode(); enc != "" {
			path = path + "?" + enc
		}
	}
	export = &MetricsExportResponse{}
	err = cl.client.Get(ctx, path, export)
	return
}

// MetricServers lists configured external metric servers (graphite / influxdb /
// opentelemetry). See https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/metrics/server
func (cl *Cluster) MetricServers(ctx context.Context) (servers ClusterMetricServers, err error) {
	err = cl.client.Get(ctx, "/cluster/metrics/server", &servers)
	return
}

// MetricServer reads the full configuration of a single metric server.
func (cl *Cluster) MetricServer(ctx context.Context, id string) (server *ClusterMetricServer, err error) {
	if id == "" {
		err = errors.New("metric server id can not be empty")
		return
	}
	server = &ClusterMetricServer{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/metrics/server/%s", id), server); err != nil {
		return
	}
	if server.ID == "" {
		server.ID = id
	}
	return
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
