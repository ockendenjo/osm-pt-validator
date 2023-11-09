package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func TestValidationRouteMasterMembers(t *testing.T) {
	testcases := []struct {
		name        string
		members     []osm.Member
		tags        map[string]string
		setupConfig func(c *Config)
		checkFn     func(t *testing.T, validationErrors []string)
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
		{
			name: "should return validation error if not enough route variants",
			members: []osm.Member{
				{
					Type: "relation",
					Ref:  34567,
				},
			},
			setupConfig: func(c *Config) {
				c.MinimumRouteVariants = 2
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "not enough route variants - https://www.openstreetmap.org/relation/1234")
			},
		},
		{
			name: "should not have validation errors if enough route variants",
			members: []osm.Member{
				{
					Type: "relation",
					Ref:  34567,
				},
			},
			setupConfig: func(c *Config) {
				c.MinimumRouteVariants = 1
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
			c := Config{}
			if tc.setupConfig != nil {
				tc.setupConfig(&c)
			}
			validator := NewValidator(c, nil)
			validationErrors := validator.RouteMaster(osm.Relation{Members: tc.members, Tags: tc.tags, ID: 1234})
			tc.checkFn(t, validationErrors)
		})
	}
}
