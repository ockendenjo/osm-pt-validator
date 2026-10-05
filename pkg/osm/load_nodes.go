package osm

import (
	"context"
)

func (c *OSMClient) LoadNodes(ctx context.Context, nodeIds []int64) map[int64]*Node {
	nodeMap := map[int64]*Node{}

	for _, nodeID := range nodeIds {
		nodeObj := loadNode(ctx, c, nodeID)
		nodeMap[nodeID] = nodeObj
	}

	return nodeMap
}

func loadNode(ctx context.Context, client *OSMClient, wayId int64) *Node {
	node, err := client.GetNode(ctx, wayId)
	if err != nil {
		return nil
	}
	return &node
}
