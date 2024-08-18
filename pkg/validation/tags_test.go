package validation

import (
	"testing"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/stretchr/testify/assert"
)

func Test_checkTagsPresent(t *testing.T) {

	testcases := []struct {
		name    string
		object  Taggable
		checkFn func(t *testing.T, ve []ValidationError)

		want []string
	}{
		{
			name:   "all tags present",
			object: osm.Node{Tags: map[string]string{"foo": "value", "bar": "value"}},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name:   "one tag missing",
			object: osm.Node{Tags: map[string]string{"foo": "value"}},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				exp := ValidationError{
					URL:     "https://www.openstreetmap.org/node/0",
					Message: "missing tag 'bar'",
				}
				assertContainsValidationError(t, validationErrors, exp)
			},
		},
		{
			name:   "multiple tags missing",
			object: osm.Node{Tags: map[string]string{}},
			checkFn: func(t *testing.T, validationErrors []ValidationError) {
				expFoo := ValidationError{
					URL:     "https://www.openstreetmap.org/node/0",
					Message: "missing tag 'bar'",
				}
				expBar := ValidationError{
					URL:     "https://www.openstreetmap.org/node/0",
					Message: "missing tag 'bar'",
				}
				assertContainsValidationError(t, validationErrors, expFoo, expBar)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors := checkTagsPresent(tc.object, "foo", "bar")
			tc.checkFn(t, validationErrors)
		})
	}
}
