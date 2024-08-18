package validation

import (
	"fmt"
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func Test_validateRETags(t *testing.T) {

	testcases := []struct {
		name    string
		tags    map[string]string
		element osm.Relation
		checkFn func(t *testing.T, validationErrors []ValidationError)
	}{
		{
			name: "not a route",
			tags: map[string]string{"type": "multipolygon"},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{URL: "https://www.openstreetmap.org/relation/0", Message: "tag 'type' should have value 'route'"}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
		{
			name: "missing type",
			tags: map[string]string{},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/relation/0",
					Message: "missing tag 'type'",
				}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
		{
			name: "wrong public_transport:version",
			tags: map[string]string{"public_transport:version": "1"},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/relation/0",
					Message: "tag 'public_transport:version' should have value '2'",
				}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors := validateRETags(osm.Relation{Tags: tc.tags})
			tc.checkFn(t, validationErrors)
		})
	}
}

func assertContainsValidationError(t *testing.T, list []ValidationError, ves ...ValidationError) bool {

outer:
	for _, ve := range ves {
		for _, i := range list {
			if i == ve {
				continue outer
			}
		}

		return assert.Fail(t, fmt.Sprintf("%#v does not contain %#v", list, ve))
	}

	return true
}

func Test_validateREMemberOrder(t *testing.T) {

	testcases := []struct {
		name    string
		members []osm.Member
		checkFn func(t *testing.T, validationErrors []ValidationError)
	}{
		{
			name: "members in correct order",
			members: []osm.Member{
				{
					Type: "node",
					Ref:  1234,
					Role: osm.RoleStop,
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name: "way before stops",
			members: []osm.Member{
				{
					Type: "way",
					Ref:  1234,
					Role: "",
				},
				{
					Type: "node",
					Ref:  1234,
					Role: osm.RoleStop,
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Contains(t, validationErrors[0].Message, "route way appears before stop/platform")
			},
		},
		{
			name: "stop after ways",
			members: []osm.Member{
				{
					Type: "node",
					Ref:  1234,
					Role: osm.RoleStop,
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
				{
					Type: "node",
					Ref:  1234,
					Role: osm.RolePlatform,
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Contains(t, validationErrors[0].Message, "stop/platform appears after route ways")
			},
		},
		{
			name: "node with missing role",
			members: []osm.Member{
				{
					Type: "node",
					Ref:  1234,
					Role: "",
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/node/1234",
					Message: "stop/platform with empty role",
				}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
		{
			name: "missing stop/platforms",
			members: []osm.Member{
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{Message: "route does not contain a stop/platform"}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
		{
			name: "missing route ways",
			members: []osm.Member{
				{
					Type: "node",
					Ref:  34567,
					Role: osm.RolePlatformExitOnly,
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Contains(t, validationErrors[0].Message, "route does not contain any route ways")
			},
		},
		{
			name: "unexpected way role",
			members: []osm.Member{
				{
					Type: "node",
					Ref:  12345,
					Role: osm.RoleStopEntryOnly,
				},
				{
					Type: "way",
					Ref:  98712,
					Role: "forward",
				},
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/way/98712",
					Message: "element has unexpected role 'forward'",
				}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors := validateREMemberOrder(osm.Relation{Members: tc.members})
			tc.checkFn(t, validationErrors)
		})
	}
}
