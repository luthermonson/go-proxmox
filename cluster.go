package proxmox

import (
	"errors"
	"fmt"
	"regexp"
)

// ClusterResources retrieves a summary list of all resources in the cluster.
// It calls /cluster/resources api v2 endpoint with an optional "type" parameter
// to filter searched values.
// It returns a list of ClusterResources.
func (c *Client) ClusterResources(filter ...string) (rs ClusterResources, err error) {
	url := "/cluster/resources"

	if len(filter) > 1 {
		return rs,
			errors.New("ClusterResources accepts maximum one parameter: type")
	} else if len(filter) == 1 {
		ok, _ := regexp.Match(`^[a-z]+$`, []byte(filter[0]))
		if ok == false {
			return rs,
				errors.New("ClusterResources accepts only a single word for type parameter")
		}
		url = fmt.Sprintf("%s?type=%s", url, filter[0])
	}
	return rs, c.Get(url, &rs)
}
