package osm

import (
	"context"
	"fmt"
)

func ValidateRelation(ctx context.Context, client *OSMClient, r Relation) ([]string, error) {
	validationErrors := []string{}
	for _, relationElement := range r.Elements {
		ve, err := validationRelationElement(ctx, client, relationElement)
		if err != nil {
			return []string{}, err
		}
		validationErrors = append(validationErrors, ve...)
	}
	return validationErrors, nil
}

func validationRelationElement(ctx context.Context, client *OSMClient, re RelationElement) ([]string, error) {
	allErrors := []string{}

	tagValidationErrors := validateRETags(re)
	allErrors = append(allErrors, tagValidationErrors...)

	memberOrderErrors := validateREMemberOrder(re)
	allErrors = append(allErrors, memberOrderErrors...)

	nodeErrors, err := validateRelationNodes(ctx, client, re)
	allErrors = append(allErrors, nodeErrors...)
	if err != nil {
		return allErrors, err
	}

	routeErrors, err := validateRelationWays(ctx, client, re)
	allErrors = append(allErrors, routeErrors...)

	return allErrors, err
}

const maxParallelOSMRequests = 10

func loadWays(ctx context.Context, client *OSMClient, wayIds []int64) map[int64]*Way {
	c := make(chan wayResult, len(wayIds))
	wayMap := map[int64]*Way{}

	remaining := 0
	for idx, wayId := range wayIds {
		go loadWay(ctx, client, wayId, c)
		remaining++
		if idx >= maxParallelOSMRequests {
			//Wait before starting next request
			wayResult := <-c
			remaining--
			wayMap[wayResult.WayID] = wayResult.Way
		}
	}
	for i := 0; i < remaining; i++ {
		wayResult := <-c
		wayMap[wayResult.WayID] = wayResult.Way
	}
	return wayMap
}

func loadWay(ctx context.Context, client *OSMClient, wayId int64, c chan wayResult) {
	way, err := client.GetWay(ctx, wayId)
	if err != nil {
		c <- wayResult{
			WayID: wayId,
			Way:   nil,
		}
		return
	}
	c <- wayResult{
		WayID: wayId,
		Way:   &way,
	}
}

type wayResult struct {
	WayID int64
	Way   *Way
}

func validateRelationWays(ctx context.Context, client *OSMClient, re RelationElement) ([]string, error) {
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
			return []string{}, fmt.Errorf("failed to load way %d", k)
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

		direction := "any"
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
					direction = "forward"
				}
				matches++
			} else if an == wayElem.GetLastNode() {
				if wayElem.IsCircular() {
					nextAllowedNodes = mapFromNodes(wayElem.Nodes)
					delete(nextAllowedNodes, wayElem.GetLastNode())
				} else {
					nextAllowedNodes[wayElem.GetFirstNode()] = true
					direction = "reverse"
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
			direction = "tbc"
			allowedNodes = nextAllowedNodes
		}

		wayDirects = append(wayDirects, wayDirection{wayElem: wayElem, direction: direction})
	}

	if hasGap {
		//Don't bother checking one-way traversal
		return validationErrors, nil
	}

	wayDirects = fillInMissingWayDirects(wayDirects)

	for _, d := range wayDirects {
		wayElem := d.wayElem
		if !checkOneway(wayElem, d.direction) {
			validationErrors = append(validationErrors, fmt.Sprintf("way with oneway tag is traversed in wrong direction - https://www.openstreetmap.org/way/%d", wayElem.ID))
		}
	}

	return validationErrors, nil
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

func getDirectionJoinCircular(circularWay WayElement, joiningWay WayElement) string {
	startNode := joiningWay.GetFirstNode()
	lastNode := joiningWay.GetLastNode()

	for _, nid := range circularWay.Nodes {
		if nid == startNode {
			return "reverse"
		}
		if nid == lastNode {
			return "forward"
		}
	}
	return "error"
}

func getDirectionJoinLinear(secondWay WayElement, direction string, joiningWay WayElement) string {
	lastNode := joiningWay.GetLastNode()
	compareNode := secondWay.GetFirstNode()
	if direction == "reverse" {
		compareNode = secondWay.GetLastNode()
	}

	if compareNode == lastNode {
		return "forward"
	}
	return "reverse"
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

func checkOneway(way WayElement, direction string) bool {
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
		return direction == "forward" || direction == "any"
	}

	if onewayTag == "-1" || onewayTag == "reverse" {
		return direction == "reverse" || direction == "any"
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

func mapFromNodes(nodes []int64) map[int64]bool {
	nodeMap := map[int64]bool{}
	for _, node := range nodes {
		nodeMap[node] = true
	}
	return nodeMap
}

func validateREMemberOrder(re RelationElement) []string {
	startedStops := false
	startedRoute := false
	routeBeforeStops := false
	stopAfterRoute := false
	nodeMissingRole := false
	validationErrors := []string{}

	roles := map[string]bool{
		"stop":                true,
		"stop_exit_only":      true,
		"stop_entry_only":     true,
		"platform":            true,
		"platform_exit_only":  true,
		"platform_entry_only": true,
	}

	for _, member := range re.Members {
		if roles[member.Role] {
			startedStops = true

			if startedRoute {
				stopAfterRoute = true
			}
		} else {
			startedRoute = true

			if !startedStops {
				routeBeforeStops = true
			}
		}

		if member.Type == "node" && member.Role == "" {
			nodeMissingRole = true
		}
	}

	if routeBeforeStops {
		validationErrors = append(validationErrors, "route way appears before stop/platform")
	}
	if stopAfterRoute {
		validationErrors = append(validationErrors, "stop/platform appears after route ways")
	}
	if nodeMissingRole {
		validationErrors = append(validationErrors, "stop/platform with empty role")
	}
	if !startedStops {
		validationErrors = append(validationErrors, "route does not contain a stop/platform")
	}
	if !startedRoute {
		validationErrors = append(validationErrors, "route does not contain any route ways")
	}

	return validationErrors
}

func validateRETags(re RelationElement) []string {
	validationErrors := []string{}

	for _, s := range []string{"from", "to", "name", "network", "operator", "ref"} {
		ve := checkTagPresent(re, s)
		if ve != "" {
			validationErrors = append(validationErrors, ve)
		}
	}

	for k, v := range map[string]string{
		"type":                     "route",
		"public_transport:version": "2",
	} {
		ve := checkTagValue(re, k, v)
		if ve != "" {
			validationErrors = append(validationErrors, ve)
		}
	}

	return validationErrors
}

func checkTagPresent(t Taggable, key string) string {
	_, found := t.GetTags()[key]
	if !found {
		return fmt.Sprintf("missing tag '%s'", key)
	}
	return ""
}

func checkTagValue(t Taggable, key string, expVal string) string {
	val, found := t.GetTags()[key]
	if !found {
		return fmt.Sprintf("missing tag '%s'", key)
	}
	if val != expVal {
		return fmt.Sprintf("tag '%s' should have value '%s'", key, expVal)
	}
	return ""
}

type Taggable interface {
	GetTags() map[string]string
}

type wayDirection struct {
	wayElem   WayElement
	direction string
}
