package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/events"
	"github.com/ockendenjo/osm-pt-validator/pkg/routes"
	"github.com/ockendenjo/osm-pt-validator/pkg/util"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
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
						{Key: aws.String(".")},
						{Key: aws.String("foo.csv")},
						{Key: aws.String("edinburgh.json")},
						{Key: aws.String("bar/baz.txt")},
						{Key: aws.String("bar/baz.txt")},
						{Key: aws.String("bar/lancs.json")},
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
							{Key: aws.String("edinburgh.json")},
						}
						return &s3.ListObjectsV2Output{Contents: contents, NextContinuationToken: aws.String("token")}, nil
					}

					assert.Equal(t, "token", *params.ContinuationToken)
					contents := []types.Object{
						{Key: aws.String("glasgow.json")},
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
			objectKeys, err := listObjectKeys(handler.Get(t.Context()))
			tc.checkFn(t, objectKeys, err)
		})
	}
}

func Test_readFile(t *testing.T) {

	testcases := []struct {
		name      string
		getObject getObjectApi
		checkFn   func(t *testing.T, res readResult)
	}{
		{
			name: "should read file",
			getObject: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				routeGroup := []routes.Route{{RelationID: 1}, {RelationID: 2}}
				routeFile := routes.RoutesFile{Routes: map[string][]routes.Route{"foo": routeGroup}, Config: validation.Config{NaptanPlatformTags: true}}
				b, err := json.Marshal(routeFile)
				assert.NoError(t, err)
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(b))}, nil
			},
			checkFn: func(t *testing.T, res readResult) {
				assert.Nil(t, res.err)
				expected := []events.CheckRelationEvent{
					{RelationID: 1, Config: validation.Config{NaptanPlatformTags: true}},
					{RelationID: 2, Config: validation.Config{NaptanPlatformTags: true}},
				}
				assert.Equal(t, expected, res.events)
			},
		},
		{
			name: "should ignore relations with zero-value relation IDs",
			getObject: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				routeGroup := []routes.Route{{RelationID: 0}}
				routeFile := routes.RoutesFile{Routes: map[string][]routes.Route{"foo": routeGroup}, Config: validation.Config{NaptanPlatformTags: true}}
				b, err := json.Marshal(routeFile)
				assert.NoError(t, err)
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(b))}, nil
			},
			checkFn: func(t *testing.T, res readResult) {
				assert.Nil(t, res.err)
				assert.Len(t, res.events, 0)
			},
		},
		{
			name: "should return error if unmarshalling file body fails",
			getObject: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("this_is_not_a_json_file")))}, nil
			},
			checkFn: func(t *testing.T, res readResult) {
				assert.NotNil(t, res.err)
			},
		},
		{
			name: "should return error if GetObject API returns an error",
			getObject: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, errors.New("something bad happened")
			},
			checkFn: func(t *testing.T, res readResult) {
				assert.NotNil(t, res.err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			c := make(chan readResult)
			readFile := getFileReader(tc.getObject, "bucketName")
			go readFile(context.Background(), "routes.json", c)
			res := <-c
			tc.checkFn(t, res)
		})
	}
}

func Test_handler(t *testing.T) {

	testcases := []struct {
		name      string
		readFile  fileReader
		sqsSender func(t *testing.T) util.SQSBatchSender
	}{
		{
			name: "should send events to SQS",
			readFile: func(ctx context.Context, key string, ch chan readResult) {
				if key == "foo.json" {
					ch <- readResult{err: nil, events: []events.CheckRelationEvent{
						{RelationID: 1, Config: validation.Config{NaptanPlatformTags: true}},
						{RelationID: 2, Config: validation.Config{NaptanPlatformTags: true}},
					}}
					return
				}
				ch <- readResult{err: nil, events: []events.CheckRelationEvent{
					{RelationID: 3, Config: validation.Config{NaptanPlatformTags: false}},
					{RelationID: 4, Config: validation.Config{NaptanPlatformTags: false}},
				}}
			},
			sqsSender: func(t *testing.T) util.SQSBatchSender {
				return func(ctx context.Context, entries []sqsTypes.SendMessageBatchRequestEntry) error {
					assert.Len(t, entries, 4)
					return nil
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			listObjectsFn := func(ctx *handler.Context) ([]string, error) {
				return []string{"foo.json", "bar.json"}, nil
			}

			handlerFn := buildHandler(listObjectsFn, tc.readFile, tc.sqsSender(t))
			_, err := handlerFn(handler.Get(t.Context()), nil)
			assert.Nil(t, err)
		})
	}
}
