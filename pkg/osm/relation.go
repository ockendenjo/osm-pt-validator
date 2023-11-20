package osm

import "fmt"

type Relation struct {
	Type    string            `json:"type"`
	ID      int64             `json:"id"`
	Version int32             `json:"version"`
	Members []Member          `json:"members"`
	Tags    map[string]string `json:"tags"`
}

func (r Relation) GetTags() map[string]string {
	return r.Tags
}

func (r Relation) GetElementURL() string {
	return fmt.Sprintf("https://www.openstreetmap.org/relation/%d", r.ID)
}

func (r Relation) IsPTv2() bool {
	v, found := r.GetTags()["public_transport:version"]
	if !found {
		return false
	}
	return v == "2"
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

func (m Member) RoleIsStop() bool {
	roles := []string{"stop", "stop_entry_only", "stop_exit_only"}
	for _, role := range roles {
		if m.Role == role {
			return true
		}
	}
	return false
}

func (m Member) RoleIsPlatform() bool {
	roles := []string{"platform", "platform_entry_only", "platform_exit_only"}
	for _, role := range roles {
		if m.Role == role {
			return true
		}
	}
	return false
}
