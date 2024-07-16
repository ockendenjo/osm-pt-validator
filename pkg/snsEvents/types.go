package snsEvents

type InvalidRelationEvent struct {
	RelationID       int64    `json:"relationID"`
	RelationURL      string   `json:"relationURL"`
	RelationName     string   `json:"name"`
	ValidationErrors []string `json:"validationErrors"`
}
