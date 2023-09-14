package validation

import (
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteMaster(r osm.Relation) []string {
	validationErrors := []string{}
	for _, relationElement := range r.Elements {
		validationErrors = append(validationErrors, v.RouteMasterElement(relationElement)...)
	}
	return validationErrors
}

func (v *Validator) RouteMasterElement(r osm.RelationElement) []string {
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
