package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type OSMPTStackProps struct {
	awscdk.StackProps
}

func NewStack(scope constructs.Construct, id string, props *OSMPTStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	qb := NewQueueBuilder(stack)

	topic := awssns.NewTopic(stack, jsii.String("InvalidRelationTopic"), &awssns.TopicProps{})
	queues := qb.NewQueueWithDLQ("RelationValidation")

	NewLambda(stack, "SplitRelation", "build/validate-rm").
		WithQueuePublish(queues, "QUEUE_URL").
		Build()

	NewLambda(stack, "ValidateRoute", "build/validate-route").
		WithTopicPublish(topic, "TOPIC_ARN").
		Build().
		AddSQSBatchTrigger(queues)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewStack(app, "OSMPTValidatorStack", &OSMPTStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
