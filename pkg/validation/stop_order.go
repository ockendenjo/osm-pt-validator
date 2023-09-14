package validation

import (
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func validateStopOrder(wayDirects []wayDirection, re osm.RelationElement) []string {
	stops := []osm.Member{}
	stopMap := map[int64][]int{}
	validationErrors := []string{}

	for _, member := range re.Members {
		if member.Type == "node" && member.Role == "stop" {
			stops = append(stops, member)
			stopMap[member.Ref] = nil
		}
	}
	stopCount := len(stops)
	if stopCount < 2 {
		return []string{}
	}

	//For each stop, record the node index in the route
	nodes := getAllNodesInOrder(wayDirects)
	for i, node := range nodes {
		slice, found := stopMap[node]
		if found {
			slice = append(slice, i)
			stopMap[node] = slice
		}
	}

	lastIndex := -1
	for _, stop := range stops {
		indices := stopMap[stop.Ref]
		if len(indices) < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("stop is not on route - %s", stop.GetElementURL()))
			continue
		}

		indices = filterGt(indices, lastIndex)
		if len(indices) < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("stop is incorrectly ordered - %s", stop.GetElementURL()))
			continue
		}

		lastIndex = indices[0]
	}

	return validationErrors
}

func getNodesInOrder(direction wayTraversal, we osm.WayElement) []int64 {
	if direction == traverseForward || direction == traverseAny {
		return we.Nodes
	}
	count := len(we.Nodes)
	reversed := make([]int64, count)
	for i, node := range we.Nodes {
		reversed[count-1-i] = node
	}
	return reversed
}

func getAllNodesInOrder(wayDirects []wayDirection) []int64 {
	allNodes := []int64{}
	for _, direct := range wayDirects {
		thisNodes := getNodesInOrder(direct.direction, direct.wayElem)
		allNodes = append(allNodes, thisNodes...)
	}
	return allNodes
}

func filterGt(indices []int, threshold int) []int {
	newIndices := []int{}
	for _, node := range indices {
		if node > threshold {
			newIndices = append(newIndices, node)
		}
	}
	return newIndices
}
