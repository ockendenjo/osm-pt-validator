package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ockendenjo/osm-pt-validator/pkg/snsEvents"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func main() {
	topicArn := handler.MustGetEnv("TOPIC_ARN")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[events.SQSEvent, events.SQSEventResponse] {
		snsClient := sns.NewFromConfig(awsConfig)
		osmClient := osm.NewClient().WithXRay()

		processRecord := buildProcessRecord(osmClient, snsClient.Publish, topicArn)
		return handler.GetSQSHandler(processRecord)
	})
}

func buildProcessRecord(client *osm.OSMClient, publish publishApi, topicArn string) handler.SQSRecordProcessor {
	return func(ctx context.Context, record events.SQSMessage) error {

		var event handler.CheckRelationEvent
		err := json.Unmarshal([]byte(record.Body), &event)
		if err != nil {
			return err
		}
		logger := handler.GetLogger(ctx).With("relationID", event.RelationID)
		logger.Info("validating relation")

		relation, err := client.GetRelation(ctx, event.RelationID)
		if err != nil {
			return err
		}

		validationErrors, err := osm.ValidateRelation(ctx, client, relation)
		if err != nil {
			return err
		}

		if len(validationErrors) > 0 {
			logger.Error("relation is invalid", "validationErrors", validationErrors)

			outputEvent := snsEvents.InvalidRelationEvent{
				RelationID:       event.RelationID,
				RelationName:     relation.Elements[0].Tags["name"],
				ValidationErrors: validationErrors,
			}
			bytes, err := json.Marshal(outputEvent)
			if err != nil {
				return err
			}

			_, err = publish(ctx, &sns.PublishInput{
				Message:  jsii.String(string(bytes)),
				Subject:  jsii.String(fmt.Sprintf("Invalid relation %d", event.RelationID)),
				TopicArn: jsii.String(topicArn),
			})
			if err != nil {
				return err
			}
		}
		logger.Info("relation is valid")
		return nil
	}
}

type publishApi func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
