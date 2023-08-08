package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/util"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		batchSend := util.NewSQSBatchSender(sqsClient.SendMessageBatch, queueUrl)

		return buildHandler(batchSend)
	})
}

func buildHandler(batchSend util.SQSBatchSender) handler.Handler[any, any] {
	return func(ctx context.Context, _ any) (any, error) {
		bytes, err := os.ReadFile("routes.json")
		if err != nil {
			return nil, err
		}

		var routesFile map[string][]Route
		err = json.Unmarshal(bytes, &routesFile)
		if err != nil {
			return nil, err
		}

		entries := []sqsTypes.SendMessageBatchRequestEntry{}
		for _, routes := range routesFile {
			for _, route := range routes {
				if route.RelationID == 0 {
					continue
				}

				outEvent := handler.CheckRelationEvent{RelationID: route.RelationID}
				body, err := json.Marshal(outEvent)
				if err != nil {
					return nil, err
				}

				entries = append(entries, sqsTypes.SendMessageBatchRequestEntry{
					Id:          jsii.String(uuid.New().String()),
					MessageBody: jsii.String(string(body)),
				})
			}
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
