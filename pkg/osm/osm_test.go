package osm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_getRelation(t *testing.T) {
	bytes, err := os.ReadFile("testdata/relation.json")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name      string
		handlerFn func(t *testing.T) func(w http.ResponseWriter, r *http.Request)
		checkFn   func(t *testing.T, r Relation, err error)
	}{
		{
			name: "HTTP 200",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/relation/3411082864.json", r.RequestURI)
					_, err := w.Write(bytes)
					if err != nil {
						t.Fatal(err)
					}
				}
			},
			checkFn: func(t *testing.T, r Relation, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(3411082864), r.ID)
			},
		},
		{
			name: "HTTP 404",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/relation/3411082864.json", r.RequestURI)
					w.WriteHeader(404)
				}
			},
			checkFn: func(t *testing.T, r Relation, err error) {
				assert.EqualError(t, err, "HTTP status code 404")
				var hse HttpStatusError
				assert.ErrorAs(t, err, &hse)
				assert.Equal(t, 404, hse.StatusCode)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			handlerFn := http.HandlerFunc(tc.handlerFn(t))
			svr := httptest.NewServer(handlerFn)
			defer svr.Close()

			client := NewClient("unit-test/0.0").WithBaseUrl(svr.URL)
			relation, err := client.GetRelation(context.Background(), 3411082864)
			tc.checkFn(t, relation, err)
		})
	}
}

func Test_getWay(t *testing.T) {
	bytes, err := os.ReadFile("testdata/way.json")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name      string
		handlerFn func(t *testing.T) func(w http.ResponseWriter, r *http.Request)
		checkFn   func(t *testing.T, w Way, err error)
	}{
		{
			name: "HTTP 200",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/way/2154620362.json", r.RequestURI)
					_, err := w.Write(bytes)
					if err != nil {
						t.Fatal(err)
					}
				}
			},
			checkFn: func(t *testing.T, w Way, err error) {
				assert.Nil(t, err)
				assert.Equal(t, int64(2154620362), w.ID)
			},
		},
		{
			name: "HTTP 404",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/way/2154620362.json", r.RequestURI)
					w.WriteHeader(404)
				}
			},
			checkFn: func(t *testing.T, w Way, err error) {
				assert.EqualError(t, err, "HTTP status code 404")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			handlerFn := http.HandlerFunc(tc.handlerFn(t))
			svr := httptest.NewServer(handlerFn)
			defer svr.Close()

			client := NewClient("unit-test/0.0").WithBaseUrl(svr.URL)
			way, err := client.GetWay(context.Background(), 2154620362)
			tc.checkFn(t, way, err)
		})
	}
}

func Test_getRelationRelations(t *testing.T) {
	bytes, err := os.ReadFile("testdata/relation_relations.json")
	if err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name      string
		handlerFn func(t *testing.T) func(w http.ResponseWriter, r *http.Request)
		checkFn   func(t *testing.T, r []Relation, err error)
	}{
		{
			name: "HTTP 200",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "/relation/11562232/relations.json", r.RequestURI)
					_, err := w.Write(bytes)
					if err != nil {
						t.Fatal(err)
					}
				}
			},
			checkFn: func(t *testing.T, r []Relation, err error) {
				require.Nil(t, err)
				require.Len(t, r, 1)
				require.Equal(t, int64(3009058), r[0].ID)
			},
		},
		{
			name: "HTTP 404",
			handlerFn: func(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
				}
			},
			checkFn: func(t *testing.T, r []Relation, err error) {
				require.EqualError(t, err, "HTTP status code 404")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			handlerFn := http.HandlerFunc(tc.handlerFn(t))
			svr := httptest.NewServer(handlerFn)
			defer svr.Close()

			client := NewClient("unit-test/0.0").WithBaseUrl(svr.URL)
			relation, err := client.GetRelationRelations(context.Background(), 11562232)
			tc.checkFn(t, relation, err)
		})
	}
}
