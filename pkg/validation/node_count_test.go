package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func TestValidator_validateNodeMembersCount(t *testing.T) {

	testcases := []struct {
		name           string
		config         Config
		relElement     osm.RelationElement
		expectedResult bool
	}{
		{
			name:           "should return true if config ignores node count",
			config:         Config{MinimumNodeMembers: 0},
			relElement:     osm.RelationElement{Members: []osm.Member{}},
			expectedResult: true,
		},
		{
			name:           "should return true if relation element has enough nodes",
			config:         Config{MinimumNodeMembers: 1},
			relElement:     osm.RelationElement{Members: []osm.Member{{Type: "node"}}},
			expectedResult: true,
		},
		{
			name:           "should return false if relation element has too few nodes",
			config:         Config{MinimumNodeMembers: 1},
			relElement:     osm.RelationElement{Members: []osm.Member{{Type: "way"}}},
			expectedResult: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := &Validator{
				config: tc.config,
			}

			result := v.validateNodeMembersCount(tc.relElement)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
