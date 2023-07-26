package main

import (
	"context"
	"encoding/json"
	"fmt"

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
		logger := handler.GetLogger(ctx)

		var event handler.CheckRelationEvent
		err := json.Unmarshal([]byte(record.Body), &event)
		if err != nil {
			return err
		}
		logger.Info("validating relation", "relationID", event.RelationID)

		relation, err := client.GetRelation(ctx, event.RelationID)
		if err != nil {
			return err
		}

		validationErrors, err := osm.ValidateRelation(ctx, client, relation)
		if err != nil {
			return err
		}

		if len(validationErrors) > 0 {
			outputEvent := invalidRelationEvent{
				RelationID:       event.RelationID,
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
		return nil
	}
}

type publishApi func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)

type invalidRelationEvent struct {
	RelationID       int64    `json:"relationID"`
	ValidationErrors []string `json:"validationErrors"`
}
