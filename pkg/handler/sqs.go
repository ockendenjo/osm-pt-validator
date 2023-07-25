package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type SQSRecordProcessor func(ctx context.Context, record events.SQSMessage) error

// GetSQSHandler returns a lambda handler that will process each SQS message in parallel using the provided processRecord function
func GetSQSHandler(processRecord SQSRecordProcessor) Handler[events.SQSEvent, events.SQSEventResponse] {

	process := func(ctx context.Context, record events.SQSMessage, c chan string) {
		err := processRecord(ctx, record)
		if err != nil {
			logger := GetLogger(ctx)
			logger.Error("sqs messaging processing failed", "error", err.Error(), "body", record.Body)
			c <- record.ReceiptHandle
		} else {
			c <- ""
		}
	}

	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		failures := []events.SQSBatchItemFailure{}
		count := len(event.Records)
		c := make(chan string, count)

		for _, record := range event.Records {
			go process(ctx, record, c)
		}

		for i := 0; i < count; i++ {
			r := <-c
			if r != "" {
				failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: r})
			}
		}

		return events.SQSEventResponse{BatchItemFailures: failures}, nil
	}
}
