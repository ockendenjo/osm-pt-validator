package util

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/osm-pt-validator/pkg/handler"
)

type SendMessageBatchApi func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)

type SQSBatchSender func(ctx context.Context, entries []types.SendMessageBatchRequestEntry) error

func NewSQSBatchSender(sendMessage SendMessageBatchApi, queueUrl string) SQSBatchSender {
	chunkSize := 10

	return func(ctx context.Context, entries []types.SendMessageBatchRequestEntry) error {
		size := len(entries)
		logger := handler.GetLogger(ctx)

		for i := 0; i < size; i += chunkSize {
			end := i + chunkSize
			if end > size {
				end = size
			}
			subSlice := entries[i:end]

			_, err := sendMessage(ctx, &sqs.SendMessageBatchInput{
				Entries:  subSlice,
				QueueUrl: jsii.String(queueUrl),
			})
			if err != nil {
				return err
			}
			logger.Info("Sent messages to SQS", "count", len(subSlice))
		}
		return nil
	}
}
