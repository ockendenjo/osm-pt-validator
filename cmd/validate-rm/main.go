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
	"github.com/google/uuid"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/events"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/snsEvents"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
)

func main() {
	queueUrl := handler.MustGetEnv("QUEUE_URL")
	topicArn := handler.MustGetEnv("TOPIC_ARN")
	userAgent := handler.MustGetEnv("USER_AGENT")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[sqsEvents.SQSEvent, sqsEvents.SQSEventResponse] {
		sqsClient := sqs.NewFromConfig(awsConfig)
		snsClient := sns.NewFromConfig(awsConfig)
		osmClient := osm.NewClient(userAgent).WithXRay()

		h := &lambdaHandler{
			sendMessageBatch: sqsClient.SendMessageBatch,
			queueUrl:         queueUrl,
			osmClient:        osmClient,
			publish:          snsClient.Publish,
			topicArn:         topicArn,
		}
		return handler.GetSQSHandler(h, nil)
	})
}

type lambdaHandler struct {
	sendMessageBatch sendMessageBatchApi
	queueUrl         string
	osmClient        *osm.OSMClient
	publish          publishApi
	topicArn         string
}

func (h *lambdaHandler) ProcessSQSEvent(ctx *handler.Context, event events.CheckRelationEvent, _ map[string]sqsEvents.SQSMessageAttribute) error {

	validator := validation.NewValidator(event.Config, h.osmClient)

	logger := ctx.GetLogger().AddParam("relationID", event.RelationID)
	relation, err := h.osmClient.GetRelation(ctx, event.RelationID)
	if err != nil {
		var hse osm.HttpStatusError
		if errors.As(err, &hse) && hse.StatusCode == http.StatusGone {
			goneErr := h.handleGone(ctx, event.RelationID)
			return goneErr
		}
		return err
	}

	logger.Info("processing relation", "type", relation.Tags["type"])

	if relation.Tags["type"] == "route_master" {
		return h.handleRouteMaster(ctx, validator, relation)
	}
	if relation.Tags["type"] == "route" {
		return h.handleRoute(ctx, relation, event.Config)
	}
	return nil
}

func (h *lambdaHandler) handleGone(ctx context.Context, relationId int64) error {
	outputEvent := snsEvents.InvalidRelationEvent{
		RelationID:       relationId,
		ValidationErrors: []validation.ValidationError{{Message: "relation no longer exists"}},
	}
	bytes, err := json.Marshal(outputEvent)
	if err != nil {
		return err
	}

	_, err = h.publish(ctx, &sns.PublishInput{
		Message:  aws.String(string(bytes)),
		Subject:  aws.String(fmt.Sprintf("Unknown relation %d", relationId)),
		TopicArn: aws.String(h.topicArn),
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *lambdaHandler) handleRoute(ctx *handler.Context, element osm.Relation, config validation.Config) error {
	logger := ctx.GetLogger()
	logger.Info("processing route relation")
	messages := []types.SendMessageBatchRequestEntry{}

	outEvent := events.CheckRelationEvent{RelationID: element.ID, Config: config}
	bytes, err := json.Marshal(outEvent)
	if err != nil {
		return err
	}

	message := types.SendMessageBatchRequestEntry{
		MessageBody: aws.String(string(bytes)),
		Id:          aws.String(uuid.New().String()),
	}
	messages = append(messages, message)
	_, err = h.sendMessageBatch(ctx, &sqs.SendMessageBatchInput{QueueUrl: &h.queueUrl, Entries: messages})
	return err
}

func (h *lambdaHandler) handleRouteMaster(ctx *handler.Context, validator *validation.Validator, element osm.Relation) error {
	logger := ctx.GetLogger()
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

		_, err = h.publish(ctx, &sns.PublishInput{
			Message:  aws.String(string(bytes)),
			Subject:  aws.String(fmt.Sprintf("Invalid relation %d", element.ID)),
			TopicArn: aws.String(h.topicArn),
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
				MessageBody: aws.String(string(bytes)),
				Id:          aws.String(uuid.New().String()),
			}
			messages = append(messages, message)
		}
	}
	if len(messages) > 0 {
		_, err := h.sendMessageBatch(ctx, &sqs.SendMessageBatchInput{QueueUrl: &h.queueUrl, Entries: messages})
		return err
	}
	return nil
}

type sendMessageBatchApi func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
type publishApi func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
