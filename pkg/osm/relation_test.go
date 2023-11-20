package osm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_unmarshallRelation(t *testing.T) {
	bytes, err := os.ReadFile("testdata/relation.json")
	if err != nil {
		t.Fatal(err)
	}

	var relation relationsResponse
	err = json.Unmarshal(bytes, &relation)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(3411082864), relation.Elements[0].ID)
}

func Test_isPTv2(t *testing.T) {

	testCases := []struct {
		name     string
		relation Relation
		expected bool
	}{
		{
			name:     "should return false if relation does not have public_transport:version tag",
			relation: Relation{Tags: map[string]string{}},
			expected: false,
		},
		{
			name: "should return false if relation has wrong tag value",
			relation: Relation{
				Tags: map[string]string{
					"public_transport:version": "1",
				},
			},
			expected: false,
		},
		{
			name: "should return true if relation has correct v2 tag value",
			relation: Relation{
				Tags: map[string]string{
					"public_transport:version": "2",
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			v2 := tc.relation.IsPTv2()
			assert.Equal(t, tc.expected, v2)
		})
	}
}
