package osm

import "fmt"

func validateStopOrder(wayDirects []wayDirection, re RelationElement) []string {
	stops := []Member{}
	stopMap := map[int64]int{}
	stopFound := map[int64]bool{}

	i := 0
	for _, member := range re.Members {
		if member.Type == "node" && member.Role == "stop" {
			stops = append(stops, member)
			stopMap[member.Ref] = i
			stopFound[member.Ref] = false
			i++
		}
	}

	stopCount := len(stops)
	if stopCount < 2 {
		return []string{}
	}

	expectedNextIndex := 0
	validationErrors := []string{}

	for _, direct := range wayDirects {
		//Loop through nodes and check again stops
		nodes := getNodesInOrder(direct.direction, direct.wayElem)
		for _, node := range nodes {
			nodeStopIndex, found := stopMap[node]
			if found {
				stopFound[node] = true

				if nodeStopIndex < expectedNextIndex {
					//Error already reported
					continue
				}

				if nodeStopIndex > expectedNextIndex {
					for i := expectedNextIndex; i < nodeStopIndex; i++ {
						validationErrors = append(validationErrors, fmt.Sprintf("stop is incorrectly ordered - https://www.openstreetmap.org/node/%d", stops[i].Ref))
					}
				}
				expectedNextIndex = nodeStopIndex + 1
			}
		}
	}

	for nodeId, found := range stopFound {
		if !found {
			validationErrors = append(validationErrors, fmt.Sprintf("stop is not on route - https://www.openstreetmap.org/node/%d", nodeId))
		}
	}

	return validationErrors
}

func getNodesInOrder(direction string, we WayElement) []int64 {
	if direction == "forward" || direction == "any" {
		return we.Nodes
	}
	count := len(we.Nodes)
	reversed := make([]int64, count)
	for i, node := range we.Nodes {
		reversed[count-1-i] = node
	}
	return reversed
}
