package validation

import (
	"context"
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteRelation(ctx context.Context, r osm.Relation) ([]ValidationError, error) {
	ve, err := v.validationRelationElement(ctx, r)
	return ve, err
}

func (v *Validator) validationRelationElement(ctx context.Context, re osm.Relation) ([]ValidationError, error) {
	allErrors := []ValidationError{}

	if !re.IsPTv2() {
		ve := ValidationError{URL: re.GetElementURL(), Message: "tag 'public_transport:version' should have value '2'"}
		return []ValidationError{ve}, nil
	}

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

	if !v.validateNodeMembersCount(re) {
		ve := ValidationError{URL: re.GetElementURL(), Message: "relation does not have enough node members"}
		allErrors = append(allErrors, ve)
	}
	return allErrors, err
}

func validateREMemberOrder(re osm.Relation) []ValidationError {
	startedStops := false
	startedRoute := false
	routeBeforeStops := false
	stopAfterRoute := false
	validationErrors := []ValidationError{}

	roles := map[string]bool{
		osm.RoleStop:              true,
		osm.RoleStopEntryOnly:     true,
		osm.RoleStopExitOnly:      true,
		osm.RolePlatform:          true,
		osm.RolePlatformEntryOnly: true,
		osm.RolePlatformExitOnly:  true,
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
			ve := ValidationError{URL: member.GetElementURL(), Message: "stop/platform with empty role"}
			validationErrors = append(validationErrors, ve)
		}

		if member.Role != "" && !roles[member.Role] {
			ve := ValidationError{URL: member.GetElementURL(), Message: fmt.Sprintf("element has unexpected role '%s'", member.Role)}
			validationErrors = append(validationErrors, ve)
		}
	}

	if routeBeforeStops {
		validationErrors = append(validationErrors, ValidationError{Message: "route way appears before stop/platform"})
	}
	if stopAfterRoute {
		validationErrors = append(validationErrors, ValidationError{Message: "stop/platform appears after route ways"})
	}
	if !startedStops {
		validationErrors = append(validationErrors, ValidationError{Message: "route does not contain a stop/platform"})
	}
	if !startedRoute {
		validationErrors = append(validationErrors, ValidationError{Message: "route does not contain any route ways"})
	}

	return validationErrors
}

func validateRETags(re osm.Relation) []ValidationError {
	validationErrors := []ValidationError{}

	missingTagErrors := checkTagsPresent(re, "from", "to", "name", "operator", "ref")
	validationErrors = append(validationErrors, missingTagErrors...)

	for k, v := range map[string]string{
		"type":                     "route",
		"public_transport:version": "2",
	} {
		ve := checkTagValue(re, k, v)
		if ve != nil {
			validationErrors = append(validationErrors, *ve)
		}
	}

	return validationErrors
}
