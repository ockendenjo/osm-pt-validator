package osm

import (
	"context"
)

func (c *OSMClient) LoadNodes(ctx context.Context, nodeIds []int64) map[int64]*Node {
	nodeMap := map[int64]*Node{}

	for _, nodeID := range nodeIds {
		nodeRes := loadNode(ctx, c, nodeID)
		nodeMap[nodeRes.nodeID] = nodeRes.node
	}

	return nodeMap
}

func loadNode(ctx context.Context, client *OSMClient, wayId int64) nodeResult {
	node, err := client.GetNode(ctx, wayId)
	if err != nil {
		return nodeResult{
			nodeID: wayId,
			node:   nil,
		}
	}
	return nodeResult{
		nodeID: wayId,
		node:   &node,
	}
}

type nodeResult struct {
	nodeID int64
	node   *Node
}
