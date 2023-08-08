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
		//City services
		{RouteNumber: "1", RouteMasterID: 11_702_779},
		{RouteNumber: "2", RouteMasterID: 0},
		{RouteNumber: "3", RouteMasterID: 11_087_988},
		{RouteNumber: "4", RouteMasterID: 310_090},
		{RouteNumber: "5", RouteMasterID: 0},
		{RouteNumber: "7", RouteMasterID: 11_221_396},
		{RouteNumber: "8", RouteMasterID: 7_643_291},
		{RouteNumber: "9", RouteMasterID: 0},
		{RouteNumber: "10", RouteMasterID: 2_190_289},
		{RouteNumber: "11", RouteMasterID: 11_221_362},
		{RouteNumber: "12", RouteMasterID: 10_529_100},
		{RouteNumber: "14", RouteMasterID: 11_221_421}, //Not a routemaster
		{RouteNumber: "15", RouteMasterID: 0},
		{RouteNumber: "16", RouteMasterID: 10_529_294},
		{RouteNumber: "19", RouteMasterID: 11_909_907},
		{RouteNumber: "21", RouteMasterID: 11_146_200},
		{RouteNumber: "22", RouteMasterID: 0},
		{RouteNumber: "23", RouteMasterID: 5_926_221},
		{RouteNumber: "24", RouteMasterID: 946_201},
		{RouteNumber: "25", RouteMasterID: 7_650_283},
		{RouteNumber: "26", RouteMasterID: 11_074_556},
		{RouteNumber: "X26", RouteMasterID: 0},
		{RouteNumber: "27", RouteMasterID: 1_358_434},
		{RouteNumber: "29", RouteMasterID: 12_561_097},
		{RouteNumber: "X29", RouteMasterID: 0},
		{RouteNumber: "30", RouteMasterID: 0},
		{RouteNumber: "31", RouteMasterID: 7_797_069},
		{RouteNumber: "X31", RouteMasterID: 0},
		{RouteNumber: "33", RouteMasterID: 0},
		{RouteNumber: "X33", RouteMasterID: 0},
		{RouteNumber: "34", RouteMasterID: 15_786_785},
		{RouteNumber: "35", RouteMasterID: 8_593_878},
		{RouteNumber: "36", RouteMasterID: 2_247_597},
		{RouteNumber: "37", RouteMasterID: 11_300_139},
		{RouteNumber: "X37", RouteMasterID: 0},
		{RouteNumber: "38", RouteMasterID: 0},
		{RouteNumber: "41", RouteMasterID: 11_014_193}, //Not listed on Lothian buses??
		{RouteNumber: "44", RouteMasterID: 11_375_367},
		{RouteNumber: "X44", RouteMasterID: 0},
		{RouteNumber: "45", RouteMasterID: 0},
		{RouteNumber: "46", RouteMasterID: 0},
		{RouteNumber: "47", RouteMasterID: 11_245_656},
		{RouteNumber: "47B", RouteMasterID: 0},
		{RouteNumber: "48", RouteMasterID: 0},
		{RouteNumber: "49", RouteMasterID: 7_654_545},
		{RouteNumber: "98", RouteMasterID: 0},
		//Airport services
		{RouteNumber: "100", RouteMasterID: 11_217_795},
		{RouteNumber: "200", RouteMasterID: 7_261_862},
		{RouteNumber: "N200", RouteMasterID: 0},
		{RouteNumber: "400", RouteMasterID: 8_631_405},
		{RouteNumber: "N400", RouteMasterID: 0},
		//Country services
		{RouteNumber: "43", RouteMasterID: 11_105_014},
		{RouteNumber: "X18", RouteMasterID: 0},
		{RouteNumber: "X27", RouteMasterID: 0},
		{RouteNumber: "X28", RouteMasterID: 0},
		//East coast buses
		{RouteNumber: "X6", RouteMasterID: 4_222_037},
		{RouteNumber: "X7", RouteMasterID: 0},
		{RouteNumber: "106", RouteMasterID: 0},
		{RouteNumber: "113", RouteMasterID: 3_009_058},
		{RouteNumber: "124", RouteMasterID: 10_151_782},
		{RouteNumber: "X5", RouteMasterID: 10_151_806},
		{RouteNumber: "139", RouteMasterID: 0},
		{RouteNumber: "140", RouteMasterID: 10_613_957},
		{RouteNumber: "141", RouteMasterID: 0},
		//Night buses
		{RouteNumber: "N3", RouteMasterID: 0},
		{RouteNumber: "N11", RouteMasterID: 0},
		{RouteNumber: "N14", RouteMasterID: 0},
		{RouteNumber: "N16", RouteMasterID: 0},
		{RouteNumber: "N18", RouteMasterID: 0},
		{RouteNumber: "N22", RouteMasterID: 0},
		{RouteNumber: "N25", RouteMasterID: 0},
		{RouteNumber: "N26", RouteMasterID: 0},
		{RouteNumber: "N28", RouteMasterID: 0},
		{RouteNumber: "N30", RouteMasterID: 0},
		{RouteNumber: "N31", RouteMasterID: 0},
		{RouteNumber: "N37", RouteMasterID: 0},
		{RouteNumber: "N43", RouteMasterID: 0},
		{RouteNumber: "N44", RouteMasterID: 0},
		{RouteNumber: "N107", RouteMasterID: 0},
		{RouteNumber: "N113", RouteMasterID: 0},
		{RouteNumber: "N124", RouteMasterID: 0},
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
