package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/snsEvents"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"

	"log/slog"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")
	topicArn := handler.MustGetEnv("TOPIC_ARN")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[events.SQSEvent, events.SQSEventResponse] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		snsClient := sns.NewFromConfig(awsConfig)
		osmClient := osm.NewClient().WithXRay()

		return handler.GetSQSHandler(buildProcessRecord(sqsClient.SendMessageBatch, queueUrl, osmClient, snsClient.Publish, topicArn))
	})
}

func buildProcessRecord(sendMessageBatch sendMessageBatchApi, queueUrl string, osmClient *osm.OSMClient, publish publishApi, topicArn string) handler.SQSRecordProcessor {
	return func(ctx context.Context, record events.SQSMessage) error {
		var event handler.CheckRelationEvent
		err := json.Unmarshal([]byte(record.Body), &event)
		if err != nil {
			return err
		}

		validator := validation.NewValidator(event.Config, osmClient)

		logger := handler.GetLogger(ctx).With("relationID", event.RelationID)
		relation, err := osmClient.GetRelation(ctx, event.RelationID)
		if err != nil {
			return err
		}
		element := relation.Elements[0]
		logger.Info("processing relation", "type", element.Tags["type"])

		if element.Tags["type"] == "route_master" {
			return handleRouteMaster(ctx, logger, validator, element, sendMessageBatch, queueUrl, publish, topicArn)
		}
		if element.Tags["type"] == "route" {
			return handleRoute(ctx, logger, element, event.Config, sendMessageBatch, queueUrl)
		}
		return nil
	}
}

func handleRoute(ctx context.Context, logger *slog.Logger, element osm.RelationElement, config validation.Config, sendMessageBatch sendMessageBatchApi, queueUrl string) error {
	logger.Info("processing route relation")
	messages := []types.SendMessageBatchRequestEntry{}

	outEvent := handler.CheckRelationEvent{RelationID: element.ID, Config: config}
	bytes, err := json.Marshal(outEvent)
	if err != nil {
		return err
	}

	message := types.SendMessageBatchRequestEntry{
		MessageBody: jsii.String(string(bytes)),
		Id:          jsii.String(uuid.New().String()),
	}
	messages = append(messages, message)
	_, err = sendMessageBatch(ctx, &sqs.SendMessageBatchInput{QueueUrl: jsii.String(queueUrl), Entries: messages})
	return err
}

func handleRouteMaster(ctx context.Context, logger *slog.Logger, validator *validation.Validator, element osm.RelationElement, sendMessageBatch sendMessageBatchApi, queueUrl string, publish publishApi, topicArn string) error {
	logger.Info("processing route_master relation")
	messages := []types.SendMessageBatchRequestEntry{}

	validationErrors := validator.RouteMasterElement(element)
	if len(validationErrors) > 0 {
		logger.Error("relation is invalid", "validationErrors", validationErrors)

		outputEvent := snsEvents.InvalidRelationEvent{
			RelationID:       element.ID,
			RelationName:     element.Tags["name"],
			ValidationErrors: validationErrors,
		}
		bytes, err := json.Marshal(outputEvent)
		if err != nil {
			return err
		}

		_, err = publish(ctx, &sns.PublishInput{
			Message:  jsii.String(string(bytes)),
			Subject:  jsii.String(fmt.Sprintf("Invalid relation %d", element.ID)),
			TopicArn: jsii.String(topicArn),
		})
		if err != nil {
			return err
		}
	}

	for _, member := range element.Members {
		if member.Type == "relation" {
			logger.Info("relation contains relation", "subRelationID", member.Ref)
			outEvent := handler.CheckRelationEvent{RelationID: member.Ref}
			bytes, err := json.Marshal(outEvent)
			if err != nil {
				return err
			}

			message := types.SendMessageBatchRequestEntry{
				MessageBody: jsii.String(string(bytes)),
				Id:          jsii.String(uuid.New().String()),
			}
			messages = append(messages, message)
		}
	}
	if len(messages) > 0 {
		_, err := sendMessageBatch(ctx, &sqs.SendMessageBatchInput{QueueUrl: jsii.String(queueUrl), Entries: messages})
		return err
	}
	return nil
}

type sendMessageBatchApi func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
type publishApi func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
