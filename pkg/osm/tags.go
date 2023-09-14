package osm

type Taggable interface {
	GetTags() map[string]string
	GetElementURL() string
}
