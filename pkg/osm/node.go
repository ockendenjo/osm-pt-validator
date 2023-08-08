package osm

type Node struct {
	Elements []NodeElement `json:"elements"`
}

type NodeElement struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Lat     float32           `json:"lat"`
	Lon     float32           `json:"lon"`
	Version int32             `json:"version"`
	Tags    map[string]string `json:"tags"`
}

func (ne NodeElement) GetTags() map[string]string {
	return ne.Tags
}
