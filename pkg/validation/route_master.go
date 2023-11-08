package validation

import (
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) RouteMaster(r osm.Relation) []string {
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
