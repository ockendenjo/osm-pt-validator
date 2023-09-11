package osm

import (
	"context"
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

	routeErrors, wayDirects, err := validateWayOrder(ctx, client, re)
	allErrors = append(allErrors, routeErrors...)

	if len(routeErrors) == 0 {
		stopErrors := validateStopOrder(wayDirects, re)
		allErrors = append(allErrors, stopErrors...)
	}

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

	missingTagErrors := checkTagsPresent(re, "from", "to", "name", "network", "operator", "ref")
	validationErrors = append(validationErrors, missingTagErrors...)

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
