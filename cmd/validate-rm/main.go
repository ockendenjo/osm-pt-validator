package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	sqsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/uuid"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/events"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/snsEvents"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"

	"log/slog"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")
	topicArn := handler.MustGetEnv("TOPIC_ARN")
	userAgent := handler.MustGetEnv("USER_AGENT")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[sqsEvents.SQSEvent, sqsEvents.SQSEventResponse] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		snsClient := sns.NewFromConfig(awsConfig)
		osmClient := osm.NewClient(userAgent).WithXRay()

		return handler.GetSQSHandler(buildProcessRecord(sqsClient.SendMessageBatch, queueUrl, osmClient, snsClient.Publish, topicArn))
	})
}

func buildProcessRecord(sendMessageBatch sendMessageBatchApi, queueUrl string, osmClient *osm.OSMClient, publish publishApi, topicArn string) handler.SQSRecordProcessor {
	return func(ctx context.Context, record sqsEvents.SQSMessage) error {
		var event events.CheckRelationEvent
		err := json.Unmarshal([]byte(record.Body), &event)
		if err != nil {
			return err
		}

		validator := validation.NewValidator(event.Config, osmClient)

		logger := handler.GetLogger(ctx).With("relationID", event.RelationID)
		relation, err := osmClient.GetRelation(ctx, event.RelationID)
		if err != nil {
			var hse osm.HttpStatusError
			if errors.As(err, &hse) && hse.StatusCode == http.StatusGone {
				goneErr := handleGone(ctx, event.RelationID, publish, topicArn)
				return goneErr
			}
			return err
		}

		logger.Info("processing relation", "type", relation.Tags["type"])

		if relation.Tags["type"] == "route_master" {
			return handleRouteMaster(ctx, logger, validator, relation, sendMessageBatch, queueUrl, publish, topicArn)
		}
		if relation.Tags["type"] == "route" {
			return handleRoute(ctx, logger, relation, event.Config, sendMessageBatch, queueUrl)
		}
		return nil
	}
}

func handleGone(ctx context.Context, relationId int64, publish publishApi, topicArn string) error {
	outputEvent := snsEvents.InvalidRelationEvent{
		RelationID:       relationId,
		ValidationErrors: []validation.ValidationError{{Message: "relation no longer exists"}},
	}
	bytes, err := json.Marshal(outputEvent)
	if err != nil {
		return err
	}

	_, err = publish(ctx, &sns.PublishInput{
		Message:  jsii.String(string(bytes)),
		Subject:  jsii.String(fmt.Sprintf("Unknown relation %d", relationId)),
		TopicArn: jsii.String(topicArn),
	})
	if err != nil {
		return err
	}
	return nil
}

func handleRoute(ctx context.Context, logger *slog.Logger, element osm.Relation, config validation.Config, sendMessageBatch sendMessageBatchApi, queueUrl string) error {
	logger.Info("processing route relation")
	messages := []types.SendMessageBatchRequestEntry{}

	outEvent := events.CheckRelationEvent{RelationID: element.ID, Config: config}
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

func handleRouteMaster(ctx context.Context, logger *slog.Logger, validator *validation.Validator, element osm.Relation, sendMessageBatch sendMessageBatchApi, queueUrl string, publish publishApi, topicArn string) error {
	logger.Info("processing route_master relation")
	messages := []types.SendMessageBatchRequestEntry{}

	validationErrors := validator.RouteMaster(element)
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
			outEvent := events.CheckRelationEvent{RelationID: member.Ref, Config: validator.GetConfig()}
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
