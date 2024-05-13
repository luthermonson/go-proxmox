package proxmox

import (
	"context"
	"fmt"
	"strings"
)

func (cl *Cluster) HAGroup(ctx context.Context, groupConfiguration *HAGroupConfiguration) error {
	if groupConfiguration == nil {
		return fmt.Errorf("empty ha group configuration")
	}

	haGroupConfiguration := struct {
		Nodes string `json:"nodes"`
		HAGroupConfiguration
	}{
		HAGroupConfiguration: HAGroupConfiguration{
			Group:      groupConfiguration.Group,
			Comment:    groupConfiguration.Comment,
			NoFailback: groupConfiguration.NoFailback,
			Restricted: groupConfiguration.Restricted,
			Type:       HATypeGroup,
		},
	}

	var nodes []string
	for _, haNode := range groupConfiguration.HaNodes {
		if haNode.Priority != nil {
			nodes = append(nodes,
				fmt.Sprintf("%s:%d", haNode.Node, haNode.Priority))
		} else {
			nodes = append(nodes,
				haNode.Node)
		}
	}

	haGroupConfiguration.Nodes = strings.Join(nodes, ",")

	if err := cl.client.Post(ctx, "cluster/ha/groups", &haGroupConfiguration, nil); err != nil {
		return err
	}

	return nil
}
