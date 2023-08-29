package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: main.go <relationId>")
	}
	ctx := context.Background()

	relationIdStr := os.Args[1]
	relationId, err := strconv.ParseInt(relationIdStr, 10, 64)
	if err != nil {
		panic(err)
	}

	osmClient := osm.NewClient()
	relation, err := osmClient.GetRelation(ctx, relationId)
	if err != nil {
		panic(err)
	}

	isValid, err := doValidation(ctx, osmClient, relation)
	if err != nil {
		panic(err)
	}
	if !isValid {
		os.Exit(1)
	}
}

func doValidation(ctx context.Context, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {

	switch relation.Elements[0].Tags["type"] {
	case "route":
		return validateRoute(ctx, osmClient, relation)
	case "route_master":
		return validateRouteMaster(ctx, osmClient, relation)
	default:
		return false, errors.New("unknown relation type")
	}
}

func validateRouteMaster(ctx context.Context, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: https://www.openstreetmap.org/relation/%d", relation.Elements[0].ID)
	validationErrors := osm.ValidateRouteMaster(relation)
	printErrors(validationErrors)
	isValid := len(validationErrors) < 1

	for _, element := range relation.Elements {
		for _, member := range element.Members {
			if member.Type == "relation" {
				fmt.Println("")
				subRelation, err := osmClient.GetRelation(ctx, member.Ref)
				if err != nil {
					return false, err
				}
				subIsValid, err := validateRoute(ctx, osmClient, subRelation)
				isValid = isValid && subIsValid
				if err != nil {
					return false, err
				}
			}
		}
	}

	return isValid, nil
}

func validateRoute(ctx context.Context, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: https://www.openstreetmap.org/relation/%d", relation.Elements[0].ID)
	validationErrors, err := osm.ValidateRelation(ctx, osmClient, relation)
	if err != nil {
		return false, err
	}
	printErrors(validationErrors)
	isValid := len(validationErrors) < 1
	return isValid, nil
}

func printErrors(validationErrors []string) {
	if len(validationErrors) < 1 {
		log.Println("relation is valid")
		return
	}

	for _, ve := range validationErrors {
		log.Println(ve)
	}
}
