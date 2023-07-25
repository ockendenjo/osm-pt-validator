package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestGetSQSHandler(t *testing.T) {

	sqsEvent := events.SQSEvent{Records: []events.SQSMessage{
		{ReceiptHandle: "5a3e8884-4ff1-46f1-8617-b3f483a79956"},
		{ReceiptHandle: "2ecc59ae-ea1a-462a-8fca-d835858fc470"},
	}}

	testcases := []struct {
		name          string
		processRecord SQSRecordProcessor
		checkResult   func(t *testing.T, result events.SQSEventResponse)
	}{
		{
			name: "All messages processed",
			processRecord: func(ctx context.Context, record events.SQSMessage) error {
				return nil
			},
			checkResult: func(t *testing.T, result events.SQSEventResponse) {
				expected := events.SQSEventResponse{BatchItemFailures: []events.SQSBatchItemFailure{}}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "Some messages fail",
			processRecord: func(ctx context.Context, record events.SQSMessage) error {
				if record.ReceiptHandle == "2ecc59ae-ea1a-462a-8fca-d835858fc470" {
					return errors.New("something bad happened")
				}
				return nil
			},
			checkResult: func(t *testing.T, result events.SQSEventResponse) {
				expected := events.SQSEventResponse{BatchItemFailures: []events.SQSBatchItemFailure{
					{ItemIdentifier: "2ecc59ae-ea1a-462a-8fca-d835858fc470"},
				}}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "All messages fail",
			processRecord: func(ctx context.Context, record events.SQSMessage) error {
				return errors.New("something bad happened")
			},
			checkResult: func(t *testing.T, result events.SQSEventResponse) {
				errorMap := map[string]bool{}
				for _, failure := range result.BatchItemFailures {
					errorMap[failure.ItemIdentifier] = true
				}
				assert.True(t, errorMap["5a3e8884-4ff1-46f1-8617-b3f483a79956"])
				assert.True(t, errorMap["2ecc59ae-ea1a-462a-8fca-d835858fc470"])
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			handler := GetSQSHandler(tc.processRecord)
			result, err := handler(context.Background(), sqsEvent)
			assert.Nil(t, err)
			tc.checkResult(t, result)
		})
	}
}
