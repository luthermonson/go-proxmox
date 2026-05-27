package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Node-level RRD + external-network queries — graph rendering, time-series,
// URL/OCI pre-flight, vzdump defaults.

// --- /nodes/{node}/rrd[data] ---------------------------------------------

// RRD asks PVE to render a single-graph PNG and returns its on-disk path
// (lives in PVE's rrdcached dir). Most callers want RRDData for numbers;
// this exists for parity with the web UI's graph rendering. ds is a
// comma-separated list of datasources (cpu/mem/diskread/...). cf is
// optional — empty defaults to AVERAGE server-side.
func (n *Node) RRD(ctx context.Context, ds string, timeframe Timeframe, cf ConsolidationFunction) (rrd *NodeRRDImage, err error) {
	if ds == "" {
		return nil, errors.New("ds is required")
	}
	q := url.Values{}
	q.Set("ds", ds)
	q.Set("timeframe", string(timeframe))
	if cf != "" {
		q.Set("cf", string(cf))
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/rrd?%s", n.Name, q.Encode()), &rrd)
	return
}

// RRDData returns the node's historical cpu/mem/disk/net timeseries. cf is
// optional — empty defaults to AVERAGE server-side.
func (n *Node) RRDData(ctx context.Context, timeframe Timeframe, cf ConsolidationFunction) (data []*RRDData, err error) {
	q := url.Values{}
	q.Set("timeframe", string(timeframe))
	if cf != "" {
		q.Set("cf", string(cf))
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/rrddata?%s", n.Name, q.Encode()), &data)
	return
}

// --- /nodes/{node}/query-url-metadata -------------------------------------

// QueryURLMetadata HEADs the URL and returns filename / mimetype / size.
// Used to pre-flight a download-url request without committing storage.
// verifyTLS=nil uses the PVE default (true); pass false to skip cert
// validation against self-signed servers.
func (n *Node) QueryURLMetadata(ctx context.Context, fileURL string, verifyTLS *bool) (meta *NodeURLMetadata, err error) {
	if fileURL == "" {
		return nil, errors.New("url is required")
	}
	q := url.Values{}
	q.Set("url", fileURL)
	if verifyTLS != nil && !*verifyTLS {
		q.Set("verify-certificates", "0")
	}
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/query-url-metadata?%s", n.Name, q.Encode()), &meta)
	return
}

// --- /nodes/{node}/query-oci-repo-tags ------------------------------------

// QueryOCIRepoTags lists all tags advertised by an OCI registry for the
// given repo reference (e.g. "docker.io/library/alpine").
func (n *Node) QueryOCIRepoTags(ctx context.Context, reference string) (tags []string, err error) {
	if reference == "" {
		return nil, errors.New("reference is required")
	}
	q := url.Values{}
	q.Set("reference", reference)
	err = n.client.Get(ctx, fmt.Sprintf("/nodes/%s/query-oci-repo-tags?%s", n.Name, q.Encode()), &tags)
	return
}

// --- /nodes/{node}/vzdump/defaults ----------------------------------------

// VzdumpDefaults returns the effective default vzdump options for this
// node, optionally narrowed to a specific backup storage. The schema mirrors
// POST /nodes/{node}/vzdump — wrapped as a map for forward compatibility.
func (n *Node) VzdumpDefaults(ctx context.Context, storage string) (defaults map[string]any, err error) {
	path := fmt.Sprintf("/nodes/%s/vzdump/defaults", n.Name)
	if storage != "" {
		q := url.Values{}
		q.Set("storage", storage)
		path = path + "?" + q.Encode()
	}
	err = n.client.Get(ctx, path, &defaults)
	return
}
