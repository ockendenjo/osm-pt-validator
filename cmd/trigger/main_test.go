package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/stretchr/testify/assert"
)

func Test_listObjectKeys(t *testing.T) {
	testcases := []struct {
		name           string
		getListObjects func(t *testing.T) listObjectsV2Api
		checkFn        func(t *testing.T, objectKeys []string, err error)
	}{
		{
			name: "should filter-out non JSON object keys",
			getListObjects: func(t *testing.T) listObjectsV2Api {
				return func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					assert.Equal(t, "bucketName", *params.Bucket)
					assert.Equal(t, "routes", *params.Prefix)

					contents := []types.Object{
						{Key: jsii.String(".")},
						{Key: jsii.String("foo.csv")},
						{Key: jsii.String("edinburgh.json")},
						{Key: jsii.String("bar/baz.txt")},
						{Key: jsii.String("bar/baz.txt")},
						{Key: jsii.String("bar/lancs.json")},
					}
					return &s3.ListObjectsV2Output{Contents: contents}, nil
				}
			},
			checkFn: func(t *testing.T, objectKeys []string, err error) {
				assert.NoError(t, err)
				exp := []string{"edinburgh.json", "bar/lancs.json"}
				assert.Equal(t, exp, objectKeys)
			},
		},
		{
			name: "should return error if listObjects returns error",
			getListObjects: func(t *testing.T) listObjectsV2Api {
				return func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					return nil, errors.New("something bad happened")
				}
			},
			checkFn: func(t *testing.T, objectKeys []string, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should use continuation-token to make multiple listObjectV2 requests",
			getListObjects: func(t *testing.T) listObjectsV2Api {
				i := 0
				return func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					i++

					if i == 1 {
						assert.Nil(t, params.ContinuationToken)
						contents := []types.Object{
							{Key: jsii.String("edinburgh.json")},
						}
						return &s3.ListObjectsV2Output{Contents: contents, NextContinuationToken: jsii.String("token")}, nil
					}

					assert.Equal(t, "token", *params.ContinuationToken)
					contents := []types.Object{
						{Key: jsii.String("glasgow.json")},
					}
					return &s3.ListObjectsV2Output{Contents: contents}, nil
				}
			},
			checkFn: func(t *testing.T, objectKeys []string, err error) {
				assert.NoError(t, err)
				exp := []string{"edinburgh.json", "glasgow.json"}
				assert.Equal(t, exp, objectKeys)
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			listObjectKeys := buildListObjectKeys(tc.getListObjects(t), "bucketName")
			objectKeys, err := listObjectKeys(context.Background())
			tc.checkFn(t, objectKeys, err)
		})
	}
}

func Test_getRelationIDs(t *testing.T) {

	testcases := []struct {
		name          string
		numObjectKeys int
		files         map[string][]int64
		checkFn       func(t *testing.T, ids map[int64]bool, err error)
	}{
		{
			name:          "should load multiple files",
			numObjectKeys: 2,
			files: map[string][]int64{
				"1.json": {11702779, 0, 11087988},
				"2.json": {11087988, 310090},
			},
			checkFn: func(t *testing.T, ids map[int64]bool, err error) {
				assert.NoError(t, err)
				exp := map[int64]bool{310090: true, 11087988: true, 11702779: true}
				assert.Equal(t, exp, ids)
			},
		},
		{
			name:          "should return error if getObject returns error",
			numObjectKeys: 2,
			checkFn: func(t *testing.T, ids map[int64]bool, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:          "should read lots of files",
			numObjectKeys: 12,
			files: map[string][]int64{
				"1.json":  {1},
				"2.json":  {2},
				"3.json":  {3},
				"4.json":  {4},
				"5.json":  {5},
				"6.json":  {6},
				"7.json":  {7},
				"8.json":  {8},
				"9.json":  {9},
				"10.json": {10},
				"11.json": {11},
				"12.json": {12},
			},
			checkFn: func(t *testing.T, ids map[int64]bool, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 12, len(ids))
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			getObject := func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				ids, found := tc.files[*params.Key]
				if !found {
					return nil, errors.New("something bad happened")
				}

				routeGroup := []Route{}
				for _, id := range ids {
					routeGroup = append(routeGroup, Route{RelationID: id})
				}
				routeFile := RoutesFile{Routes: map[string][]Route{"foo": routeGroup}, Config: Config{Naptan: true}}

				b, err := json.Marshal(routeFile)
				assert.NoError(t, err)
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(b))}, nil
			}

			getRelationIDs := buildGetRelationIDs(getObject, "bucketName")
			objectKeys := make([]string, tc.numObjectKeys)
			for i := 0; i < tc.numObjectKeys; i++ {
				objectKeys[i] = fmt.Sprintf("%d.json", i+1)
			}
			ids, err := getRelationIDs(context.Background(), objectKeys)
			tc.checkFn(t, ids, err)
		})
	}
}
