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
		checkFn func(t *testing.T, ve []string)

		want []string
	}{
		{
			name:   "all tags present",
			object: osm.Node{Tags: map[string]string{"foo": "value", "bar": "value"}},
			checkFn: func(t *testing.T, ve []string) {
				assert.Empty(t, ve)
			},
		},
		{
			name:   "one tag missing",
			object: osm.Node{Tags: map[string]string{"foo": "value"}},
			checkFn: func(t *testing.T, ve []string) {
				assert.Contains(t, ve, "missing tag 'bar' - https://www.openstreetmap.org/node/0")
			},
		},
		{
			name:   "multiple tags missing",
			object: osm.Node{Tags: map[string]string{}},
			checkFn: func(t *testing.T, ve []string) {
				assert.Contains(t, ve, "missing tag 'foo' - https://www.openstreetmap.org/node/0")
				assert.Contains(t, ve, "missing tag 'bar' - https://www.openstreetmap.org/node/0")
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
