package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
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
	routeQueues := qb.NewQueueWithDLQ("ValidateRouteEvents")
	rmQueues := qb.NewQueueWithDLQ("ValidateRMEvents")

	schedule := awsevents.Schedule_Cron(&awsevents.CronOptions{
		Minute: jsii.String("0"),
		Hour:   jsii.String("5"),
		Day:    jsii.String("*"),
	})

	NewLambda(stack, "Trigger", "build/trigger").
		WithQueuePublish(rmQueues, "QUEUE_URL").
		Build().
		RunAtFixedRate("OSMDailyValidate", schedule, nil)

	NewLambda(stack, "SplitRelation", "build/validate-rm").
		WithQueuePublish(routeQueues, "QUEUE_URL").
		Build().
		AddSQSBatchTrigger(rmQueues)

	NewLambda(stack, "ValidateRoute", "build/validate-route").
		WithTopicPublish(topic, "TOPIC_ARN").
		Build().
		AddSQSBatchTrigger(routeQueues)

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
