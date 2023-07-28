package osm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func Test_validateREMemberOrder(t *testing.T) {

	testcases := []struct {
		name    string
		members []Member
		checkFn func(t *testing.T, validationErrors []string)
	}{
		{
			name: "members in correct order",
			members: []Member{
				{
					Type: "node",
					Ref:  1234,
					Role: "stop",
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Empty(t, validationErrors)
			},
		},
		{
			name: "way before stops",
			members: []Member{
				{
					Type: "way",
					Ref:  1234,
					Role: "",
				},
				{
					Type: "node",
					Ref:  1234,
					Role: "stop",
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route way appears before stop/platform")
			},
		},
		{
			name: "stop after ways",
			members: []Member{
				{
					Type: "node",
					Ref:  1234,
					Role: "stop",
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
				{
					Type: "node",
					Ref:  1234,
					Role: "platform",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "stop/platform appears after route ways")
			},
		},
		{
			name: "node with missing role",
			members: []Member{
				{
					Type: "node",
					Ref:  1234,
					Role: "",
				},
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "stop/platform with empty role")
			},
		},
		{
			name: "missing stop/platforms",
			members: []Member{
				{
					Type: "way",
					Ref:  34567,
					Role: "",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route does not contain a stop/platform")
			},
		},
		{
			name: "missing route ways",
			members: []Member{
				{
					Type: "node",
					Ref:  34567,
					Role: "platform_exit_only",
				},
			},
			checkFn: func(t *testing.T, validationErrors []string) {
				assert.Contains(t, validationErrors, "route does not contain any route ways")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors := validateREMemberOrder(RelationElement{Members: tc.members})
			tc.checkFn(t, validationErrors)
		})
	}
}

func Test_validateRelationRoute(t *testing.T) {

	testcases := []struct {
		name    string
		members []Member
		checkFn func(t *testing.T, validationErrors []string, err error)
	}{
		{
			name: "valid route",
			members: []Member{
				{Ref: 1, Role: "", Type: "way"},
				{Ref: 2, Role: "", Type: "way"},
				{Ref: 3, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Empty(t, validationErrors)
			},
		},
		{
			name: "invalid route",
			members: []Member{
				{Ref: 1, Role: "", Type: "way"},
				{Ref: 3, Role: "", Type: "way"},
				{Ref: 2, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Contains(t, validationErrors, "ways are incorrectly ordered - way 3")
			},
		},
		{
			name: "route with circular way in middle",
			members: []Member{
				{Ref: 3, Role: "", Type: "way"},
				{Ref: 4, Role: "", Type: "way"},
				{Ref: 5, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Empty(t, validationErrors)
			},
		},
		{
			name: "valid route starting with circular way",
			members: []Member{
				{Ref: 4, Role: "", Type: "way"},
				{Ref: 5, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Empty(t, validationErrors)
			},
		},
		{
			name: "invalid route starting with circular way",
			members: []Member{
				{Ref: 4, Role: "", Type: "way"},
				{Ref: 1, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Contains(t, validationErrors, "ways are incorrectly ordered - way 1")
			},
		},
		{
			name: "valid route entering and leaving circular way at the same node",
			members: []Member{
				{Ref: 3, Role: "", Type: "way"},
				{Ref: 4, Role: "", Type: "way"},
				{Ref: 3, Role: "", Type: "way"},
			},
			checkFn: func(t *testing.T, validationErrors []string, err error) {
				assert.Nil(t, err)
				assert.Empty(t, validationErrors)
			},
		},
	}

	svr, err := setupTestServer()
	if err != nil {
		t.Fatal(err)
	}
	defer svr.Close()

	osmClient := NewClient().WithBaseUrl(svr.URL)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			validationErrors, err := validateRelationRoute(context.Background(), osmClient, RelationElement{Members: tc.members})
			tc.checkFn(t, validationErrors, err)
		})
	}
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
