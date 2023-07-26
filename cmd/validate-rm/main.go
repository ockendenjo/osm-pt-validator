package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"golang.org/x/exp/slog"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[handler.CheckRelationEvent, any] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		osmClient := osm.NewClient().WithXRay()

		handlerFn := buildHandlerFn(sqsClient.SendMessageBatch, queueUrl, osmClient)
		return handlerFn
	})
}

func buildHandlerFn(sendMessageBatch sendMessageBatchApi, queueUrl string, osmClient *osm.OSMClient) handler.Handler[handler.CheckRelationEvent, any] {
	return func(ctx context.Context, event handler.CheckRelationEvent) (any, error) {
		logger := handler.GetLogger(ctx).With("relationID", event.RelationID)
		relation, err := osmClient.GetRelation(ctx, event.RelationID)
		if err != nil {
			return nil, err
		}
		element := relation.Elements[0]
		logger.Info("processing relation", "type", element.Tags["type"])

		if element.Tags["type"] == "route_master" {
			return handleRouteMaster(ctx, logger, element, sendMessageBatch, queueUrl)
		}
		return nil, nil
	}
}

func handleRouteMaster(ctx context.Context, logger *slog.Logger, element osm.RelationElement, sendMessageBatch sendMessageBatchApi, queueUrl string) (any, error) {
	logger.Info("processing route_master relation")
	messages := []types.SendMessageBatchRequestEntry{}
	for _, member := range element.Members {
		if member.Type == "relation" {
			logger.Info("relation contains relation", "subRelationID", member.Ref)
			outEvent := handler.CheckRelationEvent{RelationID: member.Ref}
			bytes, err := json.Marshal(outEvent)
			if err != nil {
				return nil, err
			}

			message := types.SendMessageBatchRequestEntry{
				MessageBody: jsii.String(string(bytes)),
				Id:          jsii.String(uuid.New().String()),
			}
			messages = append(messages, message)
		}
	}
	if len(messages) > 0 {
		return sendMessageBatch(ctx, &sqs.SendMessageBatchInput{QueueUrl: jsii.String(queueUrl), Entries: messages})
	}
	return nil, nil
}

type sendMessageBatchApi func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
