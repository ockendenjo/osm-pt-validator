package osm

type Way struct {
	Elements []RelationElement `json:"elements"`
}

type WayElement struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Version int32             `json:"version"`
	Nodes   []int64           `json:"nodes"`
	Tags    map[string]string `json:"tags"`
}

func (we WayElement) GetTags() map[string]string {
	return we.Tags
}
