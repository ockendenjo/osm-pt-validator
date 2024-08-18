package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
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
		Minute: jsii.String("23"),
		Hour:   jsii.String("5"),
		Day:    jsii.String("*"),
	})

	bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
	})

	userAgent, err := getUserAgent()
	if err != nil {
		panic(err)
	}

	alarmTopic := awssns.NewTopic(stack, jsii.String("AlarmTopic"), &awssns.TopicProps{
		TopicName: jsii.String("osm-pt-alarms"),
	})

	NewLambda(stack, "Trigger", "build/trigger").
		WithQueuePublish(rmQueues, "QUEUE_URL").
		WithS3Read(bucket, "S3_BUCKET_NAME").
		Build(alarmTopic).
		RunAtFixedRate("OSMDailyValidate", schedule, nil)

	NewLambda(stack, "SplitRelation", "build/validate-rm").
		WithQueuePublish(routeQueues, "QUEUE_URL").
		WithTopicPublish(topic, "TOPIC_ARN").
		SetUserAgent(userAgent).
		Build(alarmTopic).
		AddSQSBatchTrigger(rmQueues)

	NewLambda(stack, "ValidateRoute", "build/validate-route").
		WithTopicPublish(topic, "TOPIC_ARN").
		SetUserAgent(userAgent).
		Build(alarmTopic).
		AddSQSBatchTrigger(routeQueues)

	awscdk.Tags_Of(stack).Add(jsii.String("application"), jsii.String("osm-pt-validator"), nil)

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

func getUserAgent() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	hash := strings.Replace(string(out), "\n", "", 1)
	userAgent := fmt.Sprintf("osm-pt-validator/%s https://github.com/ockendenjo/osm-pt-validator", hash)
	return userAgent, nil
}
