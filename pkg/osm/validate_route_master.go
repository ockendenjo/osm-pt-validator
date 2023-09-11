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
			validationErrors = append(validationErrors, fmt.Sprintf("member is not a relation - %s", member.GetElementURL()))
		}
	}

	tagMissingErrors := checkTagsPresent(r, "name", "ref", "operator")
	validationErrors = append(validationErrors, tagMissingErrors...)
	return validationErrors
}
