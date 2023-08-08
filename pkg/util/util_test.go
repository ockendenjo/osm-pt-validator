package util

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
)

func TestNewSQSBatchSender(t *testing.T) {

	testcases := []struct {
		name        string
		numMessages int
		expChunks   int
	}{
		{
			name:        "Handle less than 10 items",
			numMessages: 4,
			expChunks:   1,
		},
		{
			name:        "Handle exactly 10 items",
			numMessages: 10,
			expChunks:   1,
		},
		{
			name:        "Chunk more than 10 items",
			numMessages: 11,
			expChunks:   2,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			counter := 0
			sendMessageBatchApi := func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
				counter++
				return nil, nil
			}
			sendMessages := NewSQSBatchSender(sendMessageBatchApi, "https://sqs.eu-west-1.amazonaws.com/123456789012/QueueName")

			entries := []types.SendMessageBatchRequestEntry{}
			for i := 0; i < tc.numMessages; i++ {
				entry := types.SendMessageBatchRequestEntry{}
				entries = append(entries, entry)
			}

			err := sendMessages(context.Background(), entries)
			assert.Nil(t, err)
			assert.Equal(t, tc.expChunks, counter)
		})
	}
}
