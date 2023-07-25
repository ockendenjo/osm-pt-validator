package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func main() {
	if len(os.Args) < 2 {
		panic("usage: main.go <relationId>")
	}

	relationIdStr := os.Args[1]
	relationId, err := strconv.ParseInt(relationIdStr, 10, 64)
	if err != nil {
		panic(err)
	}
	log.Printf("validating relation %d", relationId)

	osmClient := osm.NewClient()
	relation, err := osmClient.GetRelation(context.Background(), relationId)
	if err != nil {
		panic(err)
	}

	validationErrors, err := osm.ValidateRelation(relation)
	if err != nil {
		panic(err)
	}
	if len(validationErrors) < 1 {
		log.Println("relation is valid")
		return
	}

	for _, ve := range validationErrors {
		log.Println(ve)
	}
}
