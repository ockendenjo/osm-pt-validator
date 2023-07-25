package osm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateRelation(t *testing.T) {
	bytes, err := os.ReadFile("testdata/relation.json")
	if err != nil {
		t.Fatal(err)
	}
	var relation Relation
	err = json.Unmarshal(bytes, &relation)
	if err != nil {
		t.Fatal(err)
	}

	validationErrors, err := ValidateRelation(relation)
	if err != nil {
		t.Fatal(err)
	}
	assert.Empty(t, validationErrors)
}

func Test_validateRETags(t *testing.T) {

	testcases := []struct {
		name    string
		tags    map[string]string
		element RelationElement
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
			validationErrors := validateRETags(RelationElement{Tags: tc.tags})
			tc.checkFn(t, validationErrors)
		})
	}
}
