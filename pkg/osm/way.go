package osm

import "fmt"

type Way struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Version int32             `json:"version"`
	Nodes   []int64           `json:"nodes"`
	Tags    map[string]string `json:"tags"`
}

func (w *Way) GetTags() map[string]string {
	return w.Tags
}

func (w *Way) GetElementURL() string {
	return fmt.Sprintf("https://www.openstreetmap.org/way/%d", w.ID)
}

func (w *Way) GetFirstNode() int64 {
	return w.Nodes[0]
}

func (w *Way) GetLastNode() int64 {
	return w.Nodes[len(w.Nodes)-1]
}

func (w *Way) IsCircular() bool {
	return w.GetFirstNode() == w.GetLastNode()
}
