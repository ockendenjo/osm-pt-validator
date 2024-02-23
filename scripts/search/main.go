package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/routes"
	"github.com/serjvanilla/go-overpass"
	"io"
	"os"
)

func main() {
	knownIds, err := loadRelationIds([]string{
		"routes/edinburgh.json",
		"routes/fife.json",
	})
	if err != nil {
		panic(err)
	}

	client := overpass.New()

	//[bbox:30.618338,-96.323712,30.591028,-96.330826]
	res, err := client.Query(`[bbox:55.91554391887415,-3.4239578247070317,55.96457564403378,-2.945709228515625][out:json];relation["type"="route"]["route"="bus"]["public_transport:version"="2"];>>;out ids;`)
	if err != nil {
		panic(err)
	}

	osmClient := osm.NewClient()

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

	var exitError bool

	for i, _ := range ids {
		if !knownIds[i] {
			fmt.Printf("relation %d is not being monitored\n", i)
			exitError = true
		}
	}

	if exitError {
		os.Exit(1)
	}
}

func getRelationIds(file string) (map[int64]bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var routesFile routes.RoutesFile
	err = json.Unmarshal(bytes, &routesFile)
	if err != nil {
		return nil, err
	}

	ids := map[int64]bool{}
	for _, routeGroup := range routesFile.Routes {
		for _, route := range routeGroup {
			ids[route.RelationID] = true
		}
	}
	return ids, nil
}

func loadRelationIds(files []string) (map[int64]bool, error) {
	baseMap := map[int64]bool{}
	for _, file := range files {
		m, err := getRelationIds(file)
		if err != nil {
			return nil, err
		}
		for i, b := range m {
			baseMap[i] = b
		}
	}
	return baseMap, nil
}
