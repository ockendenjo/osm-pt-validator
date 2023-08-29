package snsEvents

type InvalidRelationEvent struct {
	RelationID       int64    `json:"relationID"`
	RelationName     string   `json:"name"`
	ValidationErrors []string `json:"validationErrors"`
}
