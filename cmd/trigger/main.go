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
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")
	bucketName := handler.MustGetEnv("S3_BUCKET_NAME")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		s3Client := s3.NewFromConfig(awsConfig)
		listObjects := buildListObjectKeys(s3Client.ListObjectsV2, bucketName)
		readFile := getFileReader(s3Client.GetObject, bucketName)

		batchSend := util.NewSQSBatchSender(sqsClient.SendMessageBatch, queueUrl)

		return buildHandler(listObjects, readFile, batchSend)
	})
}

func buildHandler(listObjects listObjects, readFile fileReader, batchSend util.SQSBatchSender) handler.Handler[any, any] {
	return func(ctx context.Context, _ any) (any, error) {

		objectKeys, err := listObjects(ctx)
		if err != nil {
			return nil, err
		}

		c := make(chan readResult, len(objectKeys))
		remaining := 0

		for _, key := range objectKeys {
			go readFile(ctx, key, c)
			remaining++
		}

		events := []handler.CheckRelationEvent{}
		for remaining > 0 {
			result := <-c
			remaining--
			if result.err != nil {
				return nil, err
			}
			events = append(events, result.events...)
		}

		entries := []sqsTypes.SendMessageBatchRequestEntry{}
		for _, event := range events {

			body, err := json.Marshal(event)
			if err != nil {
				return nil, err
			}

			entries = append(entries, sqsTypes.SendMessageBatchRequestEntry{
				Id:          jsii.String(uuid.New().String()),
				MessageBody: jsii.String(string(body)),
			})
		}

		err = batchSend(ctx, entries)
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
type fileReader func(ctx context.Context, key string, ch chan readResult)

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

func getFileReader(getObject getObjectApi, bucketName string) fileReader {
	return func(ctx context.Context, objectKey string, c chan readResult) {
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

		var file RoutesFile
		err = json.Unmarshal(bytes, &file)
		if err != nil {
			c <- readResult{err: err}
			return
		}

		events := []handler.CheckRelationEvent{}
		for _, routes := range file.Routes {
			for _, route := range routes {
				if route.RelationID != 0 {
					outEvent := handler.CheckRelationEvent{RelationID: route.RelationID, Config: file.Config}
					events = append(events, outEvent)
				}
			}
		}
		c <- readResult{events: events}
	}
}

type readResult struct {
	err    error
	events []handler.CheckRelationEvent
}

type RoutesFile struct {
	Config validation.Config  `json:"config"`
	Routes map[string][]Route `json:"routes"`
}
