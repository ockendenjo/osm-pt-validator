package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func TestValidationRouteMasterMembers(t *testing.T) {
	testcases := []struct {
		name    string
		members []osm.Member
		tags    map[string]string
		checkFn func(t *testing.T, validationErrors []string)
	}{
		{
			name: "non-relation member",
			members: []osm.Member{
				{
					Type: "way",
					Ref:  34567,
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "member is not a relation - https://www.openstreetmap.org/way/34567")
			},
		},
		{
			name: "missing tags",
			members: []osm.Member{
				{
					Type: "relation",
					Ref:  34567,
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "missing tag 'name' - https://www.openstreetmap.org/relation/1234")
				assert.Contains(t, validationErrors, "missing tag 'operator' - https://www.openstreetmap.org/relation/1234")
				assert.Contains(t, validationErrors, "missing tag 'ref' - https://www.openstreetmap.org/relation/1234")
			},
		},
		{
			name: "valid route master",
			members: []osm.Member{
				{
					Type: "relation",
					Ref:  34567,
				},
			},
			tags: map[string]string{
				"name":     "Route 1: A <=> B",
				"operator": "BusCo",
				"ref":      "1",
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Empty(t, validationErrors)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validator := DefaultValidator(nil)
			validationErrors := validator.RouteMaster(osm.Relation{Members: tc.members, Tags: tc.tags, ID: 1234})
			tc.checkFn(t, validationErrors)
		})
	}
}
