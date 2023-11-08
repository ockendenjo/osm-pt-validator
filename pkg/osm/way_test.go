package osm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_unmarshallWay(t *testing.T) {
	bytes, err := os.ReadFile("testdata/way.json")
	if err != nil {
		t.Fatal(err)
	}

	var way wayResponse
	err = json.Unmarshal(bytes, &way)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(2154620362), way.Elements[0].ID)
}
