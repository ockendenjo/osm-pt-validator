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
	tagValidationErrors := validateRETags(re)
	if len(tagValidationErrors) > 0 {
		return tagValidationErrors, nil
	}
	memberValidationErrors := validateREMemberOrder(re)
	if len(memberValidationErrors) > 0 {
		return memberValidationErrors, nil
	}

	return validateRelationRoute(ctx, client, re)
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

func validateRelationRoute(ctx context.Context, client *OSMClient, re RelationElement) ([]string, error) {
	wayIds := []int64{}
	ways := []Member{}

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
	for _, relationMemberWay := range ways {
		way := (*waysMap[relationMemberWay.Ref]).Elements[0]

		if len(allowedNodes) == 0 {
			if way.IsCircular() {
				allowedNodes = mapFromNodes(way.Nodes)
				continue
			}
			allowedNodes = map[int64]bool{way.GetFirstNode(): true, way.GetLastNode(): true}
			continue
		}

		found := false
		var nextAllowedNodes map[int64]bool

		for an, _ := range allowedNodes {
			if way.IsCircular() {
				for _, node := range way.Nodes {
					if node == an {
						nextAllowedNodes = mapFromNodes(way.Nodes)
						found = true
						break
					}
				}
			} else if an == way.GetFirstNode() {
				if way.IsCircular() {
					nextAllowedNodes = mapFromNodes(way.Nodes)
				} else {
					nextAllowedNodes = map[int64]bool{way.GetLastNode(): true}
				}
				found = true
				break
			} else if an == way.GetLastNode() {
				if way.IsCircular() {
					nextAllowedNodes = mapFromNodes(way.Nodes)
					delete(nextAllowedNodes, way.GetLastNode())
				} else {
					nextAllowedNodes = map[int64]bool{way.GetFirstNode(): true}
				}
				found = true
				break
			}
		}

		if found {
			allowedNodes = nextAllowedNodes
			continue
		}

		return []string{"ways are incorrectly ordered"}, nil
	}

	return []string{}, nil
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
