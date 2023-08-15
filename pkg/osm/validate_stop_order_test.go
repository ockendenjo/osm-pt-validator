package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateStopOrder(t *testing.T) {
	testcases := []struct {
		name       string
		wayDirects []wayDirection
		relation   RelationElement
		checkFn    func(t *testing.T, validationErrors []string)
	}{
		{
			name:     "stops in correct order",
			relation: makeRelation(102, 104),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103),
				makeWayWithDirection(traverseForward, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []string) {
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Len(t, validationErrors, 1)
				assert.Contains(t, validationErrors, "stop is incorrectly ordered - https://www.openstreetmap.org/node/104")
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
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Len(t, validationErrors, 2)
				assert.Contains(t, validationErrors, "stop is incorrectly ordered - https://www.openstreetmap.org/node/104")
				assert.Contains(t, validationErrors, "stop is incorrectly ordered - https://www.openstreetmap.org/node/105")
			},
		},
		{
			name:     "multiple stops in correct order on same way",
			relation: makeRelation(102, 104),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "multiple stops in correct order on reversed way",
			relation: makeRelation(104, 102),
			wayDirects: []wayDirection{
				makeWayWithDirection("backward", 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:     "stop not on route",
			relation: makeRelation(102, 109),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 101, 102, 103, 104, 105),
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Len(t, validationErrors, 1)
				assert.Contains(t, validationErrors, "stop is not on route - https://www.openstreetmap.org/node/109")
			},
		},
		{
			name:     "stop on repeated way",
			relation: makeRelation(101, 103, 107),
			wayDirects: []wayDirection{
				makeWayWithDirection(traverseForward, 100, 101, 102),
				makeWayWithDirection(traverseForward, 102, 103, 104),
				makeWayWithDirection(traverseForward, 104, 105, 106, 104),
				makeWayWithDirection(traverseReverse, 102, 103, 104),
				makeWayWithDirection(traverseForward, 102, 107, 108),
			},
			checkFn: func(t *testing.T, validationErrors []string) {
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

func makeRelation(stops ...int64) RelationElement {
	members := make([]Member, len(stops))
	for i, stop := range stops {
		members[i] = Member{Type: "node", Role: "stop", Ref: stop}
	}

	return RelationElement{
		Members: members,
	}
}

func makeWayWithDirection(direction wayTraversal, nodes ...int64) wayDirection {
	return wayDirection{
		direction: direction,
		wayElem: WayElement{
			Nodes: nodes,
		},
	}
}
