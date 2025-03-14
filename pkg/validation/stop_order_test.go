package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func Test_validateStopOrder(t *testing.T) {
	testcases := []struct {
		name       string
		wayDirects []wayDirection
		relation   osm.Relation
		checkFn    func(t *testing.T, validationErrors []ValidationError)
	}{
		{
			name:     "stops in correct order",
			relation: makeRelation(102, 104),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103),
				makeWayWithDirection(traverseForward, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "stops in incorrect order",
			relation: makeRelation(104, 102),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103),
				makeWayWithDirection(traverseForward, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/node/102",
					Message: "stop is incorrectly ordered",
				}
				assert.Equal(t, []ValidationError{exp}, validationErrors)
			},
		},
		{
			name:     "multiple stops in incorrect order",
			relation: makeRelation(104, 102, 105, 103),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103),
				makeWayWithDirection(traverseForward, 103, 104, 105),
				makeWayWithDirection(traverseForward, 105, 106, 107),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp1 := ValidationError{
					URL:     "https://www.openstreetmap.org/node/102",
					Message: "stop is incorrectly ordered",
				}
				exp2 := ValidationError{
					URL:     "https://www.openstreetmap.org/node/103",
					Message: "stop is incorrectly ordered",
				}
				assert.Equal(t, []ValidationError{exp1, exp2}, validationErrors)
			},
		},
		{
			name:     "multiple stops in correct order on same way",
			relation: makeRelation(102, 104),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "multiple stops in correct order on reversed way",
			relation: makeRelation(104, 102),
			wayDirects: []wayDirection{
				makeWayWithDirection("backward", 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "stop not on route",
			relation: makeRelation(102, 109),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/node/109",
					Message: "stop is not on route",
				}
				assert.Equal(t, []ValidationError{exp}, validationErrors)
			},
		},
		{
			name:     "stop on repeated way",
			relation: makeRelation(101, 103, 109, 107),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 100, 101, 102),
				makeWayWithDirection(traverseForward, 102, 109, 103, 104),
				makeWayWithDirection(traverseForward, 104, 105, 106, 104),
				makeWayWithDirection(traverseReverse, 102, 109, 103, 104),
				makeWayWithDirection(traverseForward, 102, 107, 108),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "stop at start and end of loop",
			relation: makeRelation(101, 104, 101),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103),
				makeWayWithDirection(traverseForward, 103, 104, 105),
				makeWayWithDirection(traverseForward, 106, 107, 101),
			},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors := validateStopOrder(tc.wayDirects, tc.relation)
			tc.checkFn(t, validationErrors)
		})
	}
}

func makeRelation(stops ...int64) osm.Relation {
	members := make([]osm.Member, len(stops))
	roles := []string{osm.RoleStop, osm.RoleStopEntryOnly, osm.RoleStopExitOnly}
	for i, stop := range stops {
		members[i] = osm.Member{Type: "node", Role: roles[i%3], Ref: stop}
	}

	return osm.Relation{
		Members: members,
	}
}

func makeWayWithDirection(direction wayTraversal, nodes ...int64) wayDirection {
	return wayDirection{
		direction: direction,
		wayElem: osm.Way{
			Nodes: nodes,
		},
	}
}
