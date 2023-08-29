package osm

import (
	"fmt"
)

func ValidateRouteMaster(r Relation) []string {
	validationErrors := []string{}
	for _, relationElement := range r.Elements {
		validationErrors = append(validationErrors, ValidateRouteMasterElement(relationElement)...)
	}
	return validationErrors
}

func ValidateRouteMasterElement(r RelationElement) []string {
	validationErrors := []string{}

	for _, member := range r.Members {
		if member.Type != "relation" {
			validationErrors = append(validationErrors, fmt.Sprintf("member is not a relation - https://www.openstreetmap.org/%s/%d", member.Type, member.Ref))
		}
	}

	for _, s := range []string{"name", "network", "operator"} {
		ve := checkTagPresent(r, s)
		if ve != "" {
			validationErrors = append(validationErrors, ve)
		}
	}

	return validationErrors
}
