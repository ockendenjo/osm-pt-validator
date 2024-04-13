package osm

import (
	"context"
)

func (c *OSMClient) LoadNodes(ctx context.Context, nodeIds []int64) map[int64]*Node {
	ch := make(chan nodeResult, len(nodeIds))
	nodeMap := map[int64]*Node{}

	remaining := 0
	for idx, wayId := range nodeIds {
		go loadNode(ctx, c, wayId, ch)
		remaining++
		if idx >= c.parallelReqs {
			//Wait before starting next request
			nodeResult := <-ch
			remaining--
			nodeMap[nodeResult.nodeID] = nodeResult.node
		}
	}
	for i := 0; i < remaining; i++ {
		wayResult := <-ch
		nodeMap[wayResult.nodeID] = wayResult.node
	}
	return nodeMap
}

func loadNode(ctx context.Context, client *OSMClient, wayId int64, c chan nodeResult) {
	node, err := client.GetNode(ctx, wayId)
	if err != nil {
		c <- nodeResult{
			nodeID: wayId,
			node:   nil,
		}
		return
	}
	c <- nodeResult{
		nodeID: wayId,
		node:   &node,
	}
}

type nodeResult struct {
	nodeID int64
	node   *Node
}
