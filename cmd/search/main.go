package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/jsii-runtime-go"
	"github.com/ockendenjo/handler"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/routes"
	"github.com/serjvanilla/go-overpass"
	"io"
)

func main() {
	//queueUrl := handler.MustGetEnv("QUEUE_URL")
	bucketName := handler.MustGetEnv("S3_BUCKET_NAME")
	//topicArn := handler.MustGetEnv("TOPIC_ARN")

	handler.BuildAndStart(func(awsConfig aws.Config) handler.Handler[any, any] {
		//sqsClient := sqs.NewFromConfig(awsConfig)
		s3Client := s3.NewFromConfig(awsConfig)
		//listObjects := buildListObjectKeys(s3Client.ListObjectsV2, bucketName)
		//readFile := getFileReader(s3Client.GetObject, bucketName)

		//batchSend := util.NewSQSBatchSender(sqsClient.SendMessageBatch, queueUrl)
		//osmClient := osm.NewClient().WithXRay()

		loadSearches := getConfigLoader(s3Client.GetObject, bucketName)
		checker := getSearchChecker()
		return buildHandler(loadSearches, checker)

	})
}

func buildHandler(loadConfig ConfigLoader, checkArea Checker) handler.Handler[any, any] {

	return func(ctx context.Context, event any) (any, error) {

		configs, err := loadConfig(ctx)
		if err != nil {
			return nil, err
		}

		for _, c := range configs {
			_, err := checkArea(ctx, c)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	}
}

type ConfigLoader func(ctx context.Context) ([]routes.SearchConfig, error)
type GetObjectApi = func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
type Checker func(ctx context.Context, c routes.SearchConfig) (any, error)

func getConfigLoader(getObject GetObjectApi, bucket string) ConfigLoader {
	return func(ctx context.Context) ([]routes.SearchConfig, error) {
		res, err := getObject(ctx, &s3.GetObjectInput{
			Bucket: &bucket,
			Key:    jsii.String("search/search.json"),
		})
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var searchFile routes.SearchFile
		err = json.Unmarshal(bytes, &searchFile)
		if err != nil {
			return nil, err
		}

		searches := []routes.SearchConfig{}
		for _, config := range searchFile.Searches {
			searches = append(searches, config)
		}
		return searches, nil
	}
}

func getSearchChecker() Checker {

	return func(ctx context.Context, c routes.SearchConfig) (any, error) {
		if c.CheckMissing {

		}

		if c.CheckV1 {

		}

		return nil, nil
	}
}

func checkMissing(ctx context.Context, c routes.SearchConfig, osmClient *osm.OSMClient) (any, error) {

	//TODO: Load routes files from S3
	knownIds := map[int64]bool{}

	client := overpass.New() //FIXME add tracing

	queryStr := fmt.Sprintf(`[bbox:%f,%f,%f,%f][out:json];relation["type"="route"]["route"="bus"]["public_transport:version"="2"];>>;out ids;`, c.BBox[0], c.BBox[1], c.BBox[2], c.BBox[3])
	res, err := client.Query(queryStr)
	if err != nil {
		return nil, err
	}

	ids := map[int64]bool{}

	for i, _ := range res.Relations {
		parents, err := osmClient.GetRelationRelations(context.Background(), i)
		if err != nil {
			panic(err)
		}
		for _, parent := range parents {
			if parent.Tags["type"] == "route_master" {
				ids[parent.ID] = true
			}
		}
	}

	for i, _ := range ids {
		if !knownIds[i] {
			fmt.Printf("relation %d is not being monitored\n", i) //FIXME
		}
	}

	return nil, nil
}
