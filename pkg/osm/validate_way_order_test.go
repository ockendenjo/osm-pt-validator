package osm

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func Test_validateWayOrder(t *testing.T) {

	expectedValid := func(t *testing.T, validationErrors []string, err error) {
		assert.Nil(t, err)
		assert.Empty(t, validationErrors)
	}

	expectedOneWayError := func(wayId int64) func(t *testing.T, validationErrors []string, err error) {
		return func(t *testing.T, validationErrors []string, err error) {
			assert.Nil(t, err)
			errStr := fmt.Sprintf("way with oneway tag is traversed in wrong direction - https://www.openstreetmap.org/way/%d", wayId)
			assert.Contains(t, validationErrors, errStr)
		}
	}

	testcases := []struct {
		name    string
		members []Member
		checkFn func(t *testing.T, validationErrors []string, err error)
	}{
		{
			name:    "valid route",
			members: setupWays(1, 2, 3),
			checkFn: expectedValid,
		},
		{
			name:    "invalid route",
			members: setupWays(1, 3, 2),
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Contains(t, validationErrors, "ways are incorrectly ordered - https://www.openstreetmap.org/way/3")
			},
		},
		{
			name:    "route with circular way in middle",
			members: setupWays(3, 4, 5),
			checkFn: expectedValid,
		},
		{
			name:    "valid route starting with circular way",
			members: setupWays(4, 5),
			checkFn: expectedValid,
		},
		{
			name:    "invalid route starting with circular way",
			members: setupWays(4, 1),
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Contains(t, validationErrors, "ways are incorrectly ordered - https://www.openstreetmap.org/way/1")
			},
		},
		{
			name:    "valid route entering and leaving circular way at the same node",
			members: setupWays(3, 4, 3),
			checkFn: expectedValid,
		},
		{
			name:    "route with oneway way traversed in correct direction",
			members: setupWays(5, 6),
			checkFn: expectedValid,
		},
		{
			name:    "route with oneway way traversed in wrong direction",
			members: setupWays(5, 7),
			checkFn: expectedOneWayError(7),
		},
		{
			name:    "route starting with oneway way traversed in correct direction",
			members: setupWays(8, 5),
			checkFn: expectedValid,
		},
		{
			name:    "route starting with oneway way traversed in wrong direction",
			members: setupWays(6, 5),
			checkFn: expectedOneWayError(6),
		},
		{
			name:    "route with oneway=yes way traversed in wrong direction, but allowed because of oneway:psv=no",
			members: setupWays(5, 8),
			checkFn: expectedValid,
		},
		{
			name:    "route where first and seconds ways have same end nodes (permutation 1)",
			members: setupWays(9, 1, 2),
			checkFn: expectedValid,
		},
		{
			name:    "route where first and seconds ways have same end nodes (permutation 2)",
			members: setupWays(1, 9, 2),
			checkFn: expectedValid,
		},
	}

	svr, err := setupTestServer()
	if err != nil {
		t.Fatal(err)
	}
	defer svr.Close()

	osmClient := NewClient().WithBaseUrl(svr.URL)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			validationErrors, _, err := validateWayOrder(context.Background(), osmClient, RelationElement{Members: tc.members})
			tc.checkFn(t, validationErrors, err)
		})
	}
}

func setupWays(ids ...int64) []Member {
	members := []Member{}
	for _, id := range ids {
		members = append(members, Member{Ref: id, Role: "", Type: "way"})
	}
	return members
}

func setupTestServer() (*httptest.Server, error) {
	files, err := loadWayFiles()
	if err != nil {
		return nil, err
	}
	handlerFn := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		name := request.RequestURI
		name = strings.Replace(name, "/way/", "", 1)
		name = strings.Replace(name, ".json", "", 1)
		bytes, found := files[name]
		if !found {
			writer.WriteHeader(404)
			return
		}
		writer.Write(bytes)
	})
	return httptest.NewServer(handlerFn), nil
}

func loadWayFiles() (map[string][]byte, error) {
	dir, err := os.ReadDir("testdata")
	if err != nil {
		return nil, err
	}

	fileMap := map[string][]byte{}

	for _, entry := range dir {
		name := entry.Name()
		if strings.HasPrefix(name, "way_") {
			name = strings.Replace(name, "way_", "", 1)
			name = strings.Replace(name, ".json", "", 1)
			bytes, err := os.ReadFile("testdata/" + entry.Name())
			if err != nil {
				return nil, err
			}
			fileMap[name] = bytes
		}
	}

	return fileMap, nil
}
