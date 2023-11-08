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
