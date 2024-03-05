package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
	"github.com/ockendenjo/osm-pt-validator/pkg/routes"
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
)

func main() {
	ctx := context.Background()

	var relationId int64
	flag.Int64Var(&relationId, "r", 0, "Relation ID")
	var npt bool
	flag.BoolVar(&npt, "npt", false, "Verify NaPTAN platform tags")
	var inputFile string
	flag.StringVar(&inputFile, "f", "", "Routes file (validation config read from file too)")
	flag.Parse()

	if relationId < 1 && inputFile == "" {
		panic(errors.New("relationID (-r) or routes file (-f) must be specified"))
	}

	if relationId > 0 {
		validateSingleRelation(ctx, relationId, npt)
		return
	}
	validateFile(ctx, inputFile)
}

func validateFile(ctx context.Context, inputFile string) {
	file, err := os.Open(inputFile) // #nosec G304 -- File inclusion via variable is intentional
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var routesFile routes.RoutesFile
	err = json.Unmarshal(bytes, &routesFile)
	if err != nil {
		panic(err)
	}

	osmClient := osm.NewClient()
	validator := validation.NewValidator(routesFile.Config, osmClient)

	allValid := true
	for _, routeList := range routesFile.Routes {
		for i, r := range routeList {
			if i > 0 {
				fmt.Println("")
			}
			if r.RelationID < 1 {
				continue
			}

			relation, err := osmClient.GetRelation(ctx, r.RelationID)
			if err != nil {
				panic(err)
			}

			isValid, err := doValidation(ctx, validator, osmClient, relation)
			if err != nil {
				panic(err)
			}
			if !isValid {
				allValid = false
			}
		}
	}

	if !allValid {
		os.Exit(1)
	}
}

func validateSingleRelation(ctx context.Context, relationId int64, npt bool) {
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

	switch relation.Tags["type"] {
	case "route":
		return validateRoute(ctx, validator, relation)
	case "route_master":
		return validateRouteMaster(ctx, validator, osmClient, relation)
	default:
		return false, errors.New("unknown relation type")
	}
}

func validateRouteMaster(ctx context.Context, validator *validation.Validator, osmClient *osm.OSMClient, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: %s", relation.GetElementURL())

	validationErrors := validator.RouteMaster(relation)
	printErrors(validationErrors)
	isValid := len(validationErrors) < 1

	for _, member := range relation.Members {
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

	return isValid, nil
}

func validateRoute(ctx context.Context, validator *validation.Validator, relation osm.Relation) (bool, error) {
	log.Printf("validating relation: %s", relation.GetElementURL())
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
