package proxmox

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *Client) Cluster() (*Cluster, error) {
	cluster := &Cluster{
		client: c,
	}
	return cluster, c.Get("/cluster/status", cluster)
}

func (cl *Cluster) NextID() (int, error) {
	var ret string
	if err := cl.client.Get("/cluster/nextid", &ret); err != nil {
		return 0, err
	}
	return strconv.Atoi(ret)
}

// Resources retrieves a summary list of all resources in the cluster.
// It calls /cluster/resources api v2 endpoint with an optional "type" parameter
// to filter searched values.
// It returns a list of ClusterResources.
func (cl *Cluster) Resources(filters ...string) (rs ClusterResources, err error) {
	url := "/cluster/resources"

	// filters are variadic because they're optional, munging everything passed into one big string to make
	// a good request and the api will error out if there's an issue
	if f := strings.Replace(strings.Join(filters, ""), " ", "", -1); f != "" {
		url = fmt.Sprintf("%s?type=%s", url, f)
	}

	return rs, cl.client.Get(url, &rs)
}
