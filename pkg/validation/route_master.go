package validation

import (
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteMaster(r osm.Relation) []ValidationError {
	validationErrors := []ValidationError{}

	relCount := 0
	for _, member := range r.Members {
		if member.Type != "relation" {
			validationErrors = append(validationErrors, ValidationError{URL: member.GetElementURL(), Message: "member is not a relation"})
		} else {
			relCount++
		}
	}

	minVar := v.config.MinimumRouteVariants
	if minVar > 0 && relCount < minVar {
		validationErrors = append(validationErrors, ValidationError{URL: r.GetElementURL(), Message: "not enough route variants"})
	}

	tagMissingErrors := checkTagsPresent(r, "name", "ref", "operator")
	validationErrors = append(validationErrors, tagMissingErrors...)
	return validationErrors
}
