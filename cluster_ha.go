package proxmox

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	PrioritySeparator = ":"
	NodeSeparator     = ","
)

type haGroupConfiguration struct {
	Nodes      string  `json:"nodes"`
	Group      string  `json:"group"`
	Comment    *string `json:"comment,omitempty"`
	NoFailback *int    `json:"nofailback,omitempty"`
	Restricted *int    `json:"restricted,omitempty"`
	Type       HAType  `json:"type,omitempty"`
}

func (cl *Cluster) HAGroupCreate(ctx context.Context, groupConfiguration *HAGroupConfiguration) error {
	if groupConfiguration == nil {
		return fmt.Errorf("empty ha group configuration")
	}

	haGroupCfg := haGroupConfiguration{
		Group:      groupConfiguration.Group,
		Comment:    groupConfiguration.Comment,
		NoFailback: groupConfiguration.NoFailback,
		Restricted: groupConfiguration.Restricted,
		Type:       HATypeGroup,
	}

	var nodes []string
	for _, haNode := range groupConfiguration.HaNodes {
		if haNode.Priority != nil {
			nodes = append(nodes,
				fmt.Sprintf("%s%s%d", haNode.Node, PrioritySeparator, *haNode.Priority))
		} else {
			nodes = append(nodes,
				haNode.Node)
		}
	}

	haGroupCfg.Nodes = strings.Join(nodes, NodeSeparator)

	if err := cl.client.Post(ctx, "/cluster/ha/groups", &haGroupCfg, nil); err != nil {
		return err
	}

	return nil
}

func (cl *Cluster) HAGroupDelete(ctx context.Context, groupName string) error {
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/ha/groups/%s", groupName), nil)
}

func (cl *Cluster) HAGroups(ctx context.Context) ([]HAGroupConfiguration, error) {
	var haConfigurations []haGroupConfiguration
	if err := cl.client.Get(ctx, "/cluster/ha/groups", &haConfigurations); err != nil {
		return nil, err
	}

	haGroupConfigurations := make([]HAGroupConfiguration, 0, len(haConfigurations))

	for _, groupCfg := range haConfigurations {
		haNodes, err := prepareHANodeList(groupCfg.Nodes)
		if err != nil {
			return nil, fmt.Errorf("prepare ha node list : %w", err)
		}

		haGroupConfigurations = append(haGroupConfigurations, HAGroupConfiguration{
			Group:      groupCfg.Group,
			HaNodes:    haNodes,
			Comment:    groupCfg.Comment,
			NoFailback: groupCfg.NoFailback,
			Restricted: groupCfg.Restricted,
		})
	}

	return haGroupConfigurations, nil
}

func prepareHANodeList(haNodeStr string) ([]HANodes, error) {
	nodeStrs := strings.Split(haNodeStr, NodeSeparator)

	haNodes := make([]HANodes, 0, len(nodeStrs))

	for _, nodeStr := range nodeStrs {
		haNode, err := splitToHANode(nodeStr)
		if err != nil {
			return nil, fmt.Errorf("split node string : %w", err)
		}

		haNodes = append(haNodes, haNode)
	}

	return haNodes, nil
}

func splitToHANode(haNodeStr string) (HANodes, error) {
	haNodeParts := strings.Split(haNodeStr, PrioritySeparator)

	haNode := HANodes{
		Node: haNodeParts[0],
	}

	if len(haNodeParts) > 1 {
		priority, err := strconv.ParseUint(haNodeParts[1], 10, 32)
		if err != nil {
			return HANodes{}, fmt.Errorf("cannot parse priority : %w", err)
		}

		haNode.Priority = AsPtr[uint](uint(priority))
	}

	return haNode, nil
}
