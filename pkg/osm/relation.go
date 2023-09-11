package osm

import "fmt"

type Relation struct {
	Elements []RelationElement `json:"elements"`
}

type RelationElement struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Version int32             `json:"version"`
	Members []Member          `json:"members"`
	Tags    map[string]string `json:"tags"`
}

func (re RelationElement) GetTags() map[string]string {
	return re.Tags
}

func (re RelationElement) GetElementURL() string {
	return fmt.Sprintf("https://www.openstreetmap.org/relation/%d", re.ID)
}

type Member struct {
	Type string `json:"type"`
	Ref  int64  `json:"ref"`
	Role string `json:"role"`
}

func (m Member) String() string {
	return fmt.Sprintf("%s %d (%s)", m.Type, m.Ref, m.Role)
}

func (m Member) GetElementURL() string {
	return fmt.Sprintf("https://www.openstreetmap.org/%s/%d", m.Type, m.Ref)
}
