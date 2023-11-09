package routes

import (
	"github.com/ockendenjo/osm-pt-validator/pkg/validation"
)

type RoutesFile struct {
	Config validation.Config  `json:"config"`
	Routes map[string][]Route `json:"routes"`
}

type Route struct {
	Name       string `json:"name"`
	RelationID int64  `json:"relation_id"`
}
