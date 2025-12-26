package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/events"
	"github.com/ockendenjo/osm-pt-validator/pkg/snsEvents"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"

	sqsEvents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func main() {
	topicArn := handler.MustGetEnv("TOPIC_ARN")
	userAgent := handler.MustGetEnv("USER_AGENT")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[sqsEvents.SQSEvent, sqsEvents.SQSEventResponse] {
		snsClient := sns.NewFromConfig(awsConfig)
		osmClient := osm.NewClient(userAgent).WithXRay()

		h := &lambdaHandler{
			osmClient: osmClient,
			publish:   snsClient.Publish,
			topicArn:  topicArn,
		}

		return handler.GetSQSHandler(h, func(lp *handler.LoggerParams, event events.CheckRelationEvent) {
			lp.Add("relationID", event.RelationID)
		})
	})
}

type lambdaHandler struct {
	osmClient *osm.OSMClient
	publish   publishApi
	topicArn  string
}

func (h *lambdaHandler) ProcessSQSEvent(ctx *handler.Context, event events.CheckRelationEvent, _ map[string]sqsEvents.SQSMessageAttribute) error {
	logger := ctx.GetLogger()
	logger.Info("validating relation")

	relation, err := h.osmClient.GetRelation(ctx, event.RelationID)
	if err != nil {
		return err
	}

	validator := validation.NewValidator(event.Config, h.osmClient)
	validationErrors, err := validator.RouteRelation(ctx, relation)
	if err != nil {
		return err
	}

	if len(validationErrors) > 0 {
		logger.Error("relation is invalid", "validationErrors", validationErrors)

		outputEvent := snsEvents.InvalidRelationEvent{
			RelationID:       event.RelationID,
			RelationURL:      fmt.Sprintf("https://openstreetmap.org/relation/%d", event.RelationID),
			RelationName:     relation.Tags["name"],
			ValidationErrors: validationErrors,
		}
		bytes, err := json.MarshalIndent(outputEvent, "", "    ")
		if err != nil {
			return err
		}

		_, err = h.publish(ctx, &sns.PublishInput{
			Message:  aws.String(string(bytes)),
			Subject:  aws.String(fmt.Sprintf("Invalid relation %d", event.RelationID)),
			TopicArn: &h.topicArn,
		})
		if err != nil {
			return err
		}
	}
	logger.Info("relation is valid")
	return nil
}

type publishApi func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
