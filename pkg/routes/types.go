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

type SearchFile struct {
	Searches map[string]SearchConfig `json:"searches"`
}

type SearchConfig struct {
	BBox         []float64 `json:"bbox"`
	Files        []string  `json:"files"`
	CheckMissing bool      `json:"check_missing"`
	CheckV1      bool      `json:"check_v1"`
}
