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
	if _, err := cluster.Status(ctx); !IsNotAuthorized(err) {
		return cluster, err
	}

	return cluster, nil
}

func (cl *Cluster) Status(ctx context.Context) (ClusterStatus, error) {
	var rawResponseItems []genericClusterStatusItem
	err := cl.client.Get(ctx, "/cluster/status", &rawResponseItems)

	var ci clusterInfo
	var clusterNodes []clusterNode
	for _, item := range rawResponseItems {
		if item.Type == "cluster" {
			ci = clusterInfo{
				Name:    item.Name,
				Type:    item.Type,
				Id:      item.Id,
				Version: item.Version,
				Quorate: item.Quorate,
				Nodes:   item.Nodes,
			}
		}
		if item.Type == "node" {
			cn := clusterNode{
				Name:   item.Name,
				Type:   item.Type,
				Id:     item.Id,
				Ip:     item.Ip,
				Online: item.Online,
				NodeId: item.NodeId,
				Level:  item.Level,
				Local:  item.Local,
			}
			clusterNodes = append(clusterNodes, cn)
		}
	}

	return ClusterStatus{ClusterInfo: ci, ClusterNodes: clusterNodes}, err
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
	if f := strings.Replace(strings.Join(filters, ""), " ", "", -1); f != "" {
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
