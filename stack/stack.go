package main

import (
	"strconv"
	"strings"

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
	queues := qb.NewQueueWithDLQ("RelationValidation")

	splitRelation := NewLambda(stack, "SplitRelation", "build/validate-rm").
		WithQueuePublish(queues, "QUEUE_URL").
		Build()

	NewLambda(stack, "ValidateRoute", "build/validate-route").
		WithTopicPublish(topic, "TOPIC_ARN").
		Build().
		AddSQSBatchTrigger(queues)

	edinburghRms := map[int64]string{
		11702779: "1",
		11087988: "3",
		310090:   "4",
		7643291:  "8",
		2190289:  "10",
		11221362: "11",
		10529100: "12",
		10529294: "16",
		1358434:  "27",
		11300139: "37",
		11014193: "41",
	}
	i := 0
	for rm, name := range edinburghRms {
		input := awsevents.RuleTargetInput_FromObject(map[string]interface{}{
			"relationID": rm,
		})
		ruleName := "OSM-Validate-Edinburgh-Route-" + strings.ReplaceAll(name, " ", "-")
		schedule := awsevents.Schedule_Cron(&awsevents.CronOptions{
			Minute: jsii.String(strconv.Itoa(i)),
			Hour:   jsii.String("12"),
			Day:    jsii.String("*"),
		})
		splitRelation.RunAtFixedRate(ruleName, schedule, input)
		i++
	}
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
