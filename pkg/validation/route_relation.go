package validation

import (
	"context"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteRelation(ctx context.Context, r osm.Relation) ([]string, error) {
	validationErrors := []string{}
	for _, relationElement := range r.Elements {
		ve, err := v.validationRelationElement(ctx, relationElement)
		if err != nil {
			return []string{}, err
		}
		validationErrors = append(validationErrors, ve...)
	}
	return validationErrors, nil
}

func (v *Validator) validationRelationElement(ctx context.Context, re osm.RelationElement) ([]string, error) {
	allErrors := []string{}

	tagValidationErrors := validateRETags(re)
	allErrors = append(allErrors, tagValidationErrors...)

	memberOrderErrors := validateREMemberOrder(re)
	allErrors = append(allErrors, memberOrderErrors...)

	nodeErrors, err := v.validateRelationNodes(ctx, re)
	allErrors = append(allErrors, nodeErrors...)
	if err != nil {
		return allErrors, err
	}

	routeErrors, wayDirects, err := v.validateWayOrder(ctx, re)
	allErrors = append(allErrors, routeErrors...)

	if len(routeErrors) == 0 {
		stopErrors := validateStopOrder(wayDirects, re)
		allErrors = append(allErrors, stopErrors...)
	}

	return allErrors, err
}

func validateREMemberOrder(re osm.RelationElement) []string {
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

func validateRETags(re osm.RelationElement) []string {
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
