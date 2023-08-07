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

	edinburghRms := []Route{
		{RouteNumber: "1", RouteMasterID: 11702779},
		{RouteNumber: "2", RouteMasterID: 0},
		{RouteNumber: "3", RouteMasterID: 11087988},
		{RouteNumber: "4", RouteMasterID: 310090},
		{RouteNumber: "5", RouteMasterID: 0},
		{RouteNumber: "7", RouteMasterID: 11221396},
		{RouteNumber: "8", RouteMasterID: 7643291},
		{RouteNumber: "9", RouteMasterID: 0},
		{RouteNumber: "10", RouteMasterID: 2190289},
		{RouteNumber: "11", RouteMasterID: 11221362},
		{RouteNumber: "12", RouteMasterID: 10529100},
		{RouteNumber: "14", RouteMasterID: 11221421}, //Not a routemaster
		{RouteNumber: "15", RouteMasterID: 0},
		{RouteNumber: "16", RouteMasterID: 10529294},
		{RouteNumber: "19", RouteMasterID: 11909907},
		{RouteNumber: "21", RouteMasterID: 11146200},
		{RouteNumber: "22", RouteMasterID: 0},
		{RouteNumber: "23", RouteMasterID: 5926221},
		{RouteNumber: "24", RouteMasterID: 946201},
		{RouteNumber: "25", RouteMasterID: 7650283},
		{RouteNumber: "26", RouteMasterID: 11074556},
		{RouteNumber: "X26", RouteMasterID: 0},
		{RouteNumber: "27", RouteMasterID: 1358434},
		{RouteNumber: "29", RouteMasterID: 12561097},
		{RouteNumber: "X29", RouteMasterID: 0},
		{RouteNumber: "30", RouteMasterID: 0},
		{RouteNumber: "31", RouteMasterID: 7797069},
		{RouteNumber: "X31", RouteMasterID: 0},
		{RouteNumber: "33", RouteMasterID: 0},
		{RouteNumber: "X33", RouteMasterID: 0},
		{RouteNumber: "34", RouteMasterID: 15786785},
		{RouteNumber: "35", RouteMasterID: 8593878},
		{RouteNumber: "36", RouteMasterID: 2247597},
		{RouteNumber: "37", RouteMasterID: 11300139},
		{RouteNumber: "X37", RouteMasterID: 0},
		{RouteNumber: "38", RouteMasterID: 0},
		{RouteNumber: "41", RouteMasterID: 11014193}, //Not listed on Lothian buses??
		{RouteNumber: "43", RouteMasterID: 11105014},
		{RouteNumber: "44", RouteMasterID: 11375367},
		{RouteNumber: "X44", RouteMasterID: 0},
		{RouteNumber: "45", RouteMasterID: 0},
		{RouteNumber: "46", RouteMasterID: 0},
		{RouteNumber: "47", RouteMasterID: 11245656},
		{RouteNumber: "47B", RouteMasterID: 0},
		{RouteNumber: "48", RouteMasterID: 0},
		{RouteNumber: "49", RouteMasterID: 7654545},
		{RouteNumber: "98", RouteMasterID: 0},
		//Airport services
		{RouteNumber: "100", RouteMasterID: 11217795},
		{RouteNumber: "200", RouteMasterID: 7261862},
		{RouteNumber: "N200", RouteMasterID: 0},
		{RouteNumber: "400", RouteMasterID: 8631405},
		{RouteNumber: "N400", RouteMasterID: 0},
	}
	i := 0
	for _, rm := range edinburghRms {
		if rm.RouteMasterID == 0 {
			continue
		}

		input := awsevents.RuleTargetInput_FromObject(map[string]interface{}{
			"relationID": rm.RouteMasterID,
		})
		ruleName := "OSM-Validate-Edinburgh-Route-" + strings.ReplaceAll(rm.RouteNumber, " ", "-")
		schedule := awsevents.Schedule_Cron(&awsevents.CronOptions{
			Minute: jsii.String(strconv.Itoa(i)),
			Hour:   jsii.String("15"),
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

type Route struct {
	RouteMasterID int64
	RouteNumber   string
}
