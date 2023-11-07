package osm

import "fmt"

type Node struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Lat     float32           `json:"lat"`
	Lon     float32           `json:"lon"`
	Version int32             `json:"version"`
	Tags    map[string]string `json:"tags"`
}

func (n Node) GetTags() map[string]string {
	return n.Tags
}

func (n Node) GetElementURL() string {
	return fmt.Sprintf("https://www.openstreetmap.org/node/%d", n.ID)
}
