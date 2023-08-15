package osm

import (
	"context"
	"fmt"
)

func validateWayOrder(ctx context.Context, client *OSMClient, re RelationElement) ([]string, []wayDirection, error) {
	wayIds := []int64{}
	ways := []Member{}
	validationErrors := []string{}

	for _, member := range re.Members {
		if member.Type == "way" && member.Role == "" {
			wayIds = append(wayIds, member.Ref)
			ways = append(ways, member)
		}
	}

	waysMap := loadWays(ctx, client, wayIds)

	//Check for any nil ways
	for k, way := range waysMap {
		if way == nil {
			return []string{}, nil, fmt.Errorf("failed to load way %d", k)
		}
	}

	allowedNodes := map[int64]bool{}
	var wayDirects []wayDirection
	hasGap := false

	for _, relationMemberWay := range ways {
		wayElem := (*waysMap[relationMemberWay.Ref]).Elements[0]

		if len(allowedNodes) == 0 {
			if wayElem.IsCircular() {
				allowedNodes = mapFromNodes(wayElem.Nodes)
				wayDirects = append(wayDirects, wayDirection{wayElem: wayElem, direction: "any"})
			} else {
				allowedNodes = map[int64]bool{wayElem.GetFirstNode(): true, wayElem.GetLastNode(): true}
				wayDirects = append(wayDirects, wayDirection{wayElem: wayElem, direction: "tbc"})
			}
			continue
		}

		wayDir := traverseAny
		nextAllowedNodes := map[int64]bool{}
		matches := 0
		for an, _ := range allowedNodes {
			if wayElem.IsCircular() {
				for _, node := range wayElem.Nodes {
					if node == an {
						nextAllowedNodes = mapFromNodes(wayElem.Nodes)
						matches++
						break
					}
				}
			} else if an == wayElem.GetFirstNode() {
				if wayElem.IsCircular() {
					nextAllowedNodes = mapFromNodes(wayElem.Nodes)
				} else {
					nextAllowedNodes[wayElem.GetLastNode()] = true
					wayDir = traverseForward
				}
				matches++
			} else if an == wayElem.GetLastNode() {
				if wayElem.IsCircular() {
					nextAllowedNodes = mapFromNodes(wayElem.Nodes)
					delete(nextAllowedNodes, wayElem.GetLastNode())
				} else {
					nextAllowedNodes[wayElem.GetFirstNode()] = true
					wayDir = traverseReverse
				}
				matches++
			}
		}

		switch matches {
		case 0:
			validationErrors = append(validationErrors, fmt.Sprintf("ways are incorrectly ordered - https://www.openstreetmap.org/way/%d", wayElem.ID))
			allowedNodes = mapFromNodes(wayElem.Nodes)
			hasGap = true
		case 1:
			allowedNodes = nextAllowedNodes
		default:
			wayDir = traverseTBC
			allowedNodes = nextAllowedNodes
		}

		wayDirects = append(wayDirects, wayDirection{wayElem: wayElem, direction: wayDir})
	}

	if hasGap {
		//Don't bother checking one-way traversal
		return validationErrors, nil, nil
	}

	wayDirects = fillInMissingWayDirects(wayDirects)

	for _, d := range wayDirects {
		wayElem := d.wayElem
		if !checkOneway(wayElem, d.direction) {
			validationErrors = append(validationErrors, fmt.Sprintf("way with oneway tag is traversed in wrong direction - https://www.openstreetmap.org/way/%d", wayElem.ID))
		}
	}

	return validationErrors, wayDirects, nil
}

func fillInMissingWayDirects(wayDirects []wayDirection) []wayDirection {

	var previousWD wayDirection
	for i := (len(wayDirects) - 1); i >= 0; i-- {
		if wayDirects[i].direction == "tbc" {
			pw := previousWD.wayElem
			if pw.IsCircular() {
				wayDirects[i].direction = getDirectionJoinCircular(pw, wayDirects[i].wayElem)
			} else {
				wayDirects[i].direction = getDirectionJoinLinear(pw, previousWD.direction, wayDirects[i].wayElem)
			}
		}
		previousWD = wayDirects[i]
	}
	return wayDirects
}

func mapFromNodes(nodes []int64) map[int64]bool {
	nodeMap := map[int64]bool{}
	for _, node := range nodes {
		nodeMap[node] = true
	}
	return nodeMap
}

func getDirectionJoinCircular(circularWay WayElement, joiningWay WayElement) wayTraversal {
	startNode := joiningWay.GetFirstNode()
	lastNode := joiningWay.GetLastNode()

	for _, nid := range circularWay.Nodes {
		if nid == startNode {
			return traverseReverse
		}
		if nid == lastNode {
			return traverseForward
		}
	}
	return traverseError
}

func getDirectionJoinLinear(secondWay WayElement, direction wayTraversal, joiningWay WayElement) wayTraversal {
	lastNode := joiningWay.GetLastNode()
	compareNode := secondWay.GetFirstNode()
	if direction == traverseReverse {
		compareNode = secondWay.GetLastNode()
	}

	if compareNode == lastNode {
		return traverseForward
	}
	return traverseReverse
}

func isIgnoredWay(wayId int64) bool {
	//Roadworks in Haymarket
	ignoredWays := map[int64]bool{
		61883421:  true,
		4871756:   true,
		9234350:   true,
		224830909: true,
	}
	_, found := ignoredWays[wayId]
	return found
}

func checkOneway(way WayElement, direction wayTraversal) bool {
	onewayTag := getOnewayTag(way)
	if onewayTag == "" {
		//No oneway restrictions
		return true
	}

	if isIgnoredWay(way.ID) {
		return true
	}

	if onewayTag == "no" || onewayTag == "alternating" || onewayTag == "reversible" {
		return true
	}

	if onewayTag == "yes" || onewayTag == "true" || onewayTag == "1" {
		return direction == traverseForward || direction == traverseAny
	}

	if onewayTag == "-1" || onewayTag == "directionReverse" {
		return direction == traverseReverse || direction == traverseAny
	}

	return false
}

func getOnewayTag(way WayElement) string {
	if tag, found := way.Tags["oneway:psv"]; found {
		return tag
	}
	if tag, found := way.Tags["oneway:bus"]; found {
		return tag
	}
	if tag, found := way.Tags["oneway"]; found {
		return tag
	}
	return ""
}

type wayDirection struct {
	wayElem   WayElement
	direction wayTraversal
}

type wayTraversal string

const (
	traverseForward wayTraversal = "forward"
	traverseReverse wayTraversal = "reverse"
	traverseAny     wayTraversal = "any"
	traverseError   wayTraversal = "error"
	traverseTBC     wayTraversal = "tbc"
)
