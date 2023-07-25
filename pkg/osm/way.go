package osm

type Way struct {
	Elements []WayElement `json:"elements"`
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

func (we WayElement) GetFirstNode() int64 {
	return we.Nodes[0]
}

func (we WayElement) GetLastNode() int64 {
	return we.Nodes[len(we.Nodes)-1]
}

func (we WayElement) IsCircular() bool {
	return we.GetFirstNode() == we.GetLastNode()
}
