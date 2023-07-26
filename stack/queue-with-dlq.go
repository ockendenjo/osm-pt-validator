package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewQueueBuilder(scope constructs.Construct) QueueBuilder {
	return QueueBuilder{scope: scope}
}

type QueueBuilder struct {
	scope constructs.Construct
}

type QueueWithDLQ struct {
	Queue awssqs.Queue
	DLQ   awssqs.Queue
}

const maxReceiveCount = 3

// NewQueueWithDLQ creates a SQS queue with re-drive policy and dead-letter queue
func (b *QueueBuilder) NewQueueWithDLQ(id string) QueueWithDLQ {

	construct := constructs.NewConstruct(b.scope, jsii.String(id))

	dlq := b.newDLQOnConstruct(construct, id)

	queue := awssqs.NewQueue(construct, jsii.String("Queue"), &awssqs.QueueProps{
		QueueName: jsii.String(id),
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			MaxReceiveCount: jsii.Number(maxReceiveCount),
			Queue:           dlq,
		},
		ReceiveMessageWaitTime: awscdk.Duration_Seconds(jsii.Number(0)),
		VisibilityTimeout:      awscdk.Duration_Seconds(jsii.Number(10)),
	})

	return QueueWithDLQ{
		Queue: queue,
		DLQ:   dlq,
	}
}

func (b *QueueBuilder) NewDLQ(id string) awssqs.Queue {
	return b.newDLQOnConstruct(b.scope, id)
}

func (b *QueueBuilder) newDLQOnConstruct(scope constructs.Construct, id string) awssqs.Queue {
	dlq := awssqs.NewQueue(scope, jsii.String("DLQ"), &awssqs.QueueProps{
		QueueName: jsii.String(id + "__DLQ"),
	})

	awscloudwatch.NewAlarm(scope, jsii.String("Alarm"), &awscloudwatch.AlarmProps{
		Metric: awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
			MetricName: jsii.String("ApproximateNumberOfMessagesVisible"),
			Namespace:  jsii.String("AWS/SQS"),
			DimensionsMap: &map[string]*string{
				"QueueName": dlq.QueueName(),
			},
			Statistic: jsii.String("Average"),
			Period:    awscdk.Duration_Minutes(jsii.Number(1)),
		}),
		AlarmName:          jsii.String("DLQLength__" + id),
		AlarmDescription:   jsii.String("Goes into ALARM state when messages are sent to the DLQ.\n\nRemediation: Investigate why the messages are failing to be processed and then re-drive the DLQ."),
		EvaluationPeriods:  jsii.Number(1),
		Threshold:          jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_NOT_BREACHING,
	})

	return dlq
}
