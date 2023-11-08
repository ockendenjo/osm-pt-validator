package validation

import (
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) validateNodeMembersCount(re osm.Relation) bool {

	if v.config.MinimumNodeMembers < 1 {
		return true
	}

	count := 0
	for _, member := range re.Members {
		if member.Type == "node" {
			count++
		}
	}

	return count >= v.config.MinimumNodeMembers
}
