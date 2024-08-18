package snsEvents

import "github.com/ockendenjo/osm-pt-validator/pkg/validation"

type InvalidRelationEvent struct {
	RelationID       int64                        `json:"relationID"`
	RelationURL      string                       `json:"relationURL"`
	RelationName     string                       `json:"name"`
	ValidationErrors []validation.ValidationError `json:"validationErrors"`
}
