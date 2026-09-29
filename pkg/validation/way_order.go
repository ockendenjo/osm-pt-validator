package validation

import (
	"context"
	"fmt"
	"slices"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func extractWayMembers(re osm.Relation) ([]int64, []osm.Member) {
	var wayIds []int64
	var ways []osm.Member
	for _, member := range re.Members {
		if member.Type == "way" && member.Role == "" {
			wayIds = append(wayIds, member.Ref)
			ways = append(ways, member)
		}
	}
	return wayIds, ways
}

func validateWaysMap(waysMap map[int64]*osm.Way) error {
	for k, way := range waysMap {
		if way == nil {
			return fmt.Errorf("failed to load way %d", k)
		}
	}
	return nil
}

func matchWay(wayElem osm.Way, allowedNodes map[int64]bool) (int, map[int64]bool, wayTraversal) {
	wayDir := traverseAny
	nextAllowedNodes := map[int64]bool{}
	matches := 0
	for an := range allowedNodes {
		if wayElem.IsCircular() {
			if slices.Contains(wayElem.Nodes, an) {
				nextAllowedNodes = mapFromNodes(wayElem.Nodes)
				matches++
			}
		} else if an == wayElem.GetFirstNode() {
			nextAllowedNodes[wayElem.GetLastNode()] = true
			wayDir = traverseForward
			matches++
		} else if an == wayElem.GetLastNode() {
			nextAllowedNodes[wayElem.GetFirstNode()] = true
			wayDir = traverseReverse
			matches++
		}
	}
	return matches, nextAllowedNodes, wayDir
}

func (v *Validator) checkOnewayViolations(wayDirects []wayDirection) []ValidationError {
	var errors []ValidationError
	for _, d := range wayDirects {
		if !v.checkOneway(d.wayElem, d.direction) {
			errors = append(errors, ValidationError{URL: d.wayElem.GetElementURL(), Message: "way with oneway tag is traversed in wrong direction"})
		}
	}
	return errors
}

func (v *Validator) validateWayOrder(ctx context.Context, re osm.Relation) ([]ValidationError, []wayDirection, error) {
	wayIds, ways := extractWayMembers(re)
	waysMap := v.osmClient.LoadWays(ctx, wayIds)

	if err := validateWaysMap(waysMap); err != nil {
		return nil, nil, err
	}

	var validationErrors []ValidationError
	allowedNodes := map[int64]bool{}
	var wayDirects []wayDirection
	hasGap := false

	for _, relationMemberWay := range ways {
		wayElem := *waysMap[relationMemberWay.Ref]

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

		matches, nextAllowedNodes, wayDir := matchWay(wayElem, allowedNodes)

		switch matches {
		case 0:
			validationErrors = append(validationErrors, ValidationError{URL: wayElem.GetElementURL(), Message: "ways are incorrectly ordered"})
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
		return validationErrors, nil, nil
	}

	wayDirects = fillInMissingWayDirects(wayDirects)
	validationErrors = append(validationErrors, v.checkOnewayViolations(wayDirects)...)

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

func getDirectionJoinCircular(circularWay osm.Way, joiningWay osm.Way) wayTraversal {
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

func getDirectionJoinLinear(secondWay osm.Way, direction wayTraversal, joiningWay osm.Way) wayTraversal {
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

func (v *Validator) checkOneway(way osm.Way, direction wayTraversal) bool {
	onewayTag := getOnewayTag(way)
	if onewayTag == "" {
		return true
	}
	if v.config.IsWayDirectionIgnored(way.ID) {
		return true
	}
	return isDirectionAllowed(onewayTag, direction)
}

func isDirectionAllowed(onewayTag string, direction wayTraversal) bool {
	switch onewayTag {
	case "no", "alternating", "reversible":
		return true
	case "yes", "true", "1":
		return direction == traverseForward || direction == traverseAny
	case "-1", "directionReverse":
		return direction == traverseReverse || direction == traverseAny
	default:
		return false
	}
}

func getOnewayTag(way osm.Way) string {
	if tag, found := way.Tags["oneway:psv"]; found {
		return tag
	}
	if tag, found := way.Tags["oneway:bus"]; found {
		return tag
	}
	if tag, found := way.Tags["oneway"]; found {
		return tag
	}
	if tag := way.Tags["junction"]; tag == "roundabout" {
		return "yes"
	}
	return ""
}

type wayDirection struct {
	wayElem   osm.Way
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
