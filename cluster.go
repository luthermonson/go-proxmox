package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) Cluster(ctx context.Context) (*Cluster, error) {
	cluster := &Cluster{
		client: c,
	}

	// requires (/, Sys.Audit), do not error out if no access to still get the cluster
	if err := cluster.Status(ctx); !IsNotAuthorized(err) {
		return cluster, err
	}

	return cluster, nil
}

func (cl *Cluster) New(c *Client) *Cluster {
	return &Cluster{
		client: c,
	}
}

func (cl *Cluster) Status(ctx context.Context) error {
	return cl.client.Get(ctx, "/cluster/status", cl)
}

func (cl *Cluster) NextID(ctx context.Context) (int, error) {
	var ret string
	if err := cl.client.Get(ctx, "/cluster/nextid", &ret); err != nil {
		return 0, err
	}
	return strconv.Atoi(ret)
}

// CheckID checks if the given vmid is free.
// CheckID calls the /cluster/nextid endpoint with the "vmid" parameter.
// The API documentation describes the check as: "Pass a VMID to assert that its free (at time of check)."
// Returns true if the vmid is free, false otherwise.
func (cl *Cluster) CheckID(ctx context.Context, vmid int) (bool, error) {
	var ret string
	err := cl.client.Get(ctx, fmt.Sprintf("/cluster/nextid?vmid=%d", vmid), ret)
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf("VM %d already exists", vmid)) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Resources retrieves a summary list of all resources in the cluster.
// It calls /cluster/resources api v2 endpoint with an optional "type" parameter
// to filter searched values.
// It returns a list of ClusterResources.
func (cl *Cluster) Resources(ctx context.Context, filters ...string) (rs ClusterResources, err error) {
	u := url.URL{Path: "/cluster/resources"}

	// filters are variadic because they're optional, munging everything passed into one big string to make
	// a good request and the api will error out if there's an issue
	if f := strings.ReplaceAll(strings.Join(filters, ""), " ", ""); f != "" {
		params := url.Values{}
		params.Add("type", f)
		u.RawQuery = params.Encode()
	}

	return rs, cl.client.Get(ctx, u.String(), &rs)
}

func (cl *Cluster) Tasks(ctx context.Context) (Tasks, error) {
	var tasks Tasks

	if err := cl.client.Get(ctx, "/cluster/tasks", &tasks); err != nil {
		return nil, err
	}

	for index := range tasks {
		tasks[index].client = cl.client
	}

	return tasks, nil
}

func (cl *Cluster) Ceph(ctx context.Context) (*Ceph, error) {
	ceph := &Ceph{
		client: cl.client,
	}

	// TODO?
	//// requires (/, Sys.Audit), do not error out if no access to still get the ceph
	//if err := ceph.Status(ctx); !IsNotAuthorized(err) {
	//	return ceph, err
	//}

	return ceph, nil
}
