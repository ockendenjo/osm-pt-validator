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
		checkFn     func(t *testing.T, validationErrors []ValidationError)
	}{
		{
			name: "non-relation member",
			members: []osm.Member{
				{
					Type: "way",
					Ref:  34567,
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{URL: "https://www.openstreetmap.org/way/34567", Message: "member is not a relation"}
				assertContainsValidationError(t, validationErrors, exp)
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
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp1 := ValidationError{URL: "https://www.openstreetmap.org/relation/1234", Message: "missing tag 'name'"}
				exp2 := ValidationError{URL: "https://www.openstreetmap.org/relation/1234", Message: "missing tag 'operator'"}
				exp3 := ValidationError{URL: "https://www.openstreetmap.org/relation/1234", Message: "missing tag 'ref'"}
				assertContainsValidationError(t, validationErrors, exp1, exp2, exp3)
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
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
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
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/relation/1234",
					Message: "not enough route variants",
				}
				assertContainsValidationError(t, validationErrors, exp)
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
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
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
