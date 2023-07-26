package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
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

	initFn := awslambda.NewFunction(stack, jsii.String("StartValidationFunction"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("function"),
		Code:         awslambda.Code_FromAsset(jsii.String("build/validate-rm"), nil),
		FunctionName: jsii.String("StartValidationFunction"),
		Environment: &map[string]*string{
			"QUEUE_URL": queues.Queue.QueueUrl(),
		},
		Timeout:    awscdk.Duration_Seconds(jsii.Number(10)),
		MemorySize: jsii.Number(1024),
		Tracing:    awslambda.Tracing_ACTIVE,
	})
	queues.Queue.GrantSendMessages(initFn)

	validateFn := awslambda.NewFunction(stack, jsii.String("CheckRelationFunction"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("function"),
		Code:         awslambda.Code_FromAsset(jsii.String("build/validate-route"), nil),
		FunctionName: jsii.String("CheckRelationFunction"),
		Environment: &map[string]*string{
			"TOPIC_ARN": topic.TopicArn(),
		},
		Timeout:    awscdk.Duration_Seconds(jsii.Number(10)),
		MemorySize: jsii.Number(1024),
		Tracing:    awslambda.Tracing_ACTIVE,
	})
	topic.GrantPublish(validateFn.Role())

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
