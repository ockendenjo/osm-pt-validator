package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func Test_validateRETags(t *testing.T) {

	testcases := []struct {
		name    string
		tags    map[string]string
		element osm.Relation
		checkFn func(t *testing.T, validationErrors []string)
	}{
		{
			name: "not a route",
			tags: map[string]string{"type": "multipolygon"},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "tag 'type' should have value 'route'")
			},
		},
		{
			name: "missing type",
			tags: map[string]string{},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "missing tag 'type'")
			},
		},
		{
			name: "wrong public_transport:version",
			tags: map[string]string{"public_transport:version": "1"},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "tag 'public_transport:version' should have value '2'")
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

func Test_validateREMemberOrder(t *testing.T) {

	testcases := []struct {
		name    string
		members []osm.Member
		checkFn func(t *testing.T, validationErrors []string)
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
			checkFn: func(t *testing.T, validationErrors []string) {
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route way appears before stop/platform")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "stop/platform appears after route ways")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "stop/platform with empty role")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route does not contain a stop/platform")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route does not contain any route ways")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors[0], "element has unexpected role 'forward'")
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
