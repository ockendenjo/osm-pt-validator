package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
)

func main() {
	ctx := context.Background()

	var relationId int64
	flag.Int64Var(&relationId, "r", 0, "Relation ID")
	var npt bool
	flag.BoolVar(&npt, "npt", false, "Verify NaPTAN platform tags")
	flag.Parse()

	if relationId < 1 {
		panic(errors.New("relationID (-r) must be specified"))
	}

	osmClient := osm.NewClient()
	relation, err := osmClient.GetRelation(ctx, relationId)
	if err != nil {
		panic(err)
	}

	validator := validation.NewValidator(validation.Config{NaptanPlatformTags: npt}, osmClient)

	isValid, err := doValidation(ctx, validator, osmClient, relation)
	if err != nil {
		panic(err)
	}
	if !isValid {
		os.Exit(1)
	}
}

func doValidation(ctx context.Context, validator *validation.Validator, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {

	switch relation.Elements[0].Tags["type"] {
	case "route":
		return validateRoute(ctx, validator, relation)
	case "route_master":
		return validateRouteMaster(ctx, validator, osmClient, relation)
	default:
		return false, errors.New("unknown relation type")
	}
}

func validateRouteMaster(ctx context.Context, validator *validation.Validator, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: %s", relation.Elements[0].GetElementURL())

	validationErrors := validator.RouteMaster(relation)
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
				subIsValid, err := validateRoute(ctx, validator, subRelation)
				isValid = isValid && subIsValid
				if err != nil {
					return false, err
				}
			}
		}
	}

	return isValid, nil
}

func validateRoute(ctx context.Context, validator *validation.Validator, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: %s", relation.Elements[0].GetElementURL())
	validationErrors, err := validator.RouteRelation(ctx, relation)
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
