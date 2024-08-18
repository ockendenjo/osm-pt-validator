package validation

import (
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteMaster(r osm.Relation) []string {
	validationErrors := []string{}

	relCount := 0
	for _, member := range r.Members {
		if member.Type != "relation" {
			validationErrors = append(validationErrors, fmt.Sprintf("member is not a relation - %s", member.GetElementURL()))
		} else {
			relCount++
		}
	}

	minVar := v.config.MinimumRouteVariants
	if minVar > 0 && relCount < minVar {
		validationErrors = append(validationErrors, fmt.Sprintf("not enough route variants - %s", r.GetElementURL()))
	}

	tagMissingErrors := checkTagsPresent(r, "name", "ref", "operator")
	for _, ve := range tagMissingErrors {
		validationErrors = append(validationErrors, ve.String())
	}
	return validationErrors
}
