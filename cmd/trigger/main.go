package main

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/util"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")
	bucketName := handler.MustGetEnv("S3_BUCKET_NAME")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		s3Client := s3.NewFromConfig(awsConfig)
		listObjects := buildListObjectKeys(s3Client.ListObjectsV2, bucketName)
		getRelationIDs := buildGetRelationIDs(s3Client.GetObject, bucketName)

		batchSend := util.NewSQSBatchSender(sqsClient.SendMessageBatch, queueUrl)

		return buildHandler(listObjects, getRelationIDs, batchSend)
	})
}

func buildHandler(listObjects listObjects, getRelationIDs getRelationIDs, batchSend util.SQSBatchSender) handler.Handler[any, any] {
	return func(ctx context.Context, _ any) (any, error) {
		objectKeys, err := listObjects(ctx)
		if err != nil {
			return nil, err
		}

		relationIDs, err := getRelationIDs(ctx, objectKeys)
		if err != nil {
			return nil, err
		}

		entries := []sqsTypes.SendMessageBatchRequestEntry{}
		for id := range relationIDs {
			outEvent := handler.CheckRelationEvent{RelationID: id}
			body, err := json.Marshal(outEvent)
			if err != nil {
				return nil, err
			}

			entries = append(entries, sqsTypes.SendMessageBatchRequestEntry{
				Id:          jsii.String(uuid.New().String()),
				MessageBody: jsii.String(string(body)),
			})
		}

		err = batchSend(ctx, entries)
		if err != nil {
			handler.GetLogger(ctx).Info("trigger checking for route_master relations", "count", len(entries))
		}
		return nil, err
	}
}

type Route struct {
	Name       string `json:"name"`
	RelationID int64  `json:"relation_id"`
}

type listObjectsV2Api func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
type listObjects func(ctx context.Context) ([]string, error)
type getObjectApi func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
type getRelationIDs func(ctx context.Context, objectKeys []string) (map[int64]bool, error)

func buildListObjectKeys(listObjects listObjectsV2Api, bucketName string) listObjects {
	keys := []string{}

	return func(ctx context.Context) ([]string, error) {
		logger := handler.GetLogger(ctx)
		var token *string
		for {
			result, err := listObjects(ctx, &s3.ListObjectsV2Input{Bucket: &bucketName, ContinuationToken: token, Prefix: jsii.String("routes")})
			if err != nil {
				return nil, err
			}
			logger.Info("listObjects", "count", len(result.Contents))

			for _, content := range result.Contents {
				key := *content.Key
				if strings.HasSuffix(key, ".json") {
					logger.Info("S3 object", "key", key)
					keys = append(keys, key)
				}
			}

			token = result.NextContinuationToken
			if result.NextContinuationToken == nil {
				break
			}
		}
		return keys, nil
	}
}

func buildGetRelationIDs(getObject getObjectApi, bucketName string) getRelationIDs {
	maxConcurrent := 10

	return func(ctx context.Context, objectKeys []string) (map[int64]bool, error) {
		c := make(chan readResult, len(objectKeys))
		relationIDs := map[int64]bool{}
		var err error
		remaining := 0

		readChannel := func() {
			result := <-c
			remaining--
			if result.err != nil {
				err = result.err
				return
			}
			for _, id := range result.relationIDs {
				relationIDs[id] = true
			}
		}

		for i, key := range objectKeys {
			go readFile(ctx, getObject, bucketName, key, c)
			remaining++

			if i >= maxConcurrent {
				readChannel()
			}
		}

		for remaining > 0 {
			readChannel()
		}

		return relationIDs, err
	}
}

func readFile(ctx context.Context, getObject getObjectApi, bucketName, objectKey string, c chan readResult) {
	result, err := getObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		c <- readResult{err: err}
		return
	}

	bytes, err := io.ReadAll(result.Body)
	if err != nil {
		c <- readResult{err: err}
		return
	}

	var routesFile map[string][]Route
	err = json.Unmarshal(bytes, &routesFile)
	if err != nil {
		c <- readResult{err: err}
		return
	}

	relationIDs := []int64{}
	for _, routes := range routesFile {
		for _, route := range routes {
			if route.RelationID == 0 {
				continue
			}
			relationIDs = append(relationIDs, route.RelationID)
		}
	}

	c <- readResult{relationIDs: relationIDs}
}

type readResult struct {
	relationIDs []int64
	err         error
}
