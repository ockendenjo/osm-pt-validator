package osm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidationRouteMasterMembers(t *testing.T) {
	testcases := []struct {
		name    string
		members []Member
		tags    map[string]string
		checkFn func(t *testing.T, validationErrors []string)
	}{
		{
			name: "non-relation member",
			members: []Member{
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
			members: []Member{
				{
					Type: "relation",
					Ref:  34567,
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "missing tag 'name'")
				assert.Contains(t, validationErrors, "missing tag 'operator'")
				assert.Contains(t, validationErrors, "missing tag 'ref'")
			},
		},
		{
			name: "valid route master",
			members: []Member{
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
			validationErrors := ValidateRouteMasterElement(RelationElement{Members: tc.members, Tags: tc.tags})
			tc.checkFn(t, validationErrors)
		})
	}
}
