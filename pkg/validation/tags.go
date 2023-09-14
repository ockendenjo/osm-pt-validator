package validation

import "fmt"

func checkTagsPresent(t Taggable, tags ...string) []string {
	validationErrors := []string{}

	tagMap := t.GetTags()
	for _, key := range tags {
		_, found := tagMap[key]
		if !found {
			validationErrors = append(validationErrors, fmt.Sprintf("missing tag '%s' - %s", key, t.GetElementURL()))
		}
	}

	return validationErrors
}

func checkTagValue(t Taggable, key string, expVal string) string {
	val, found := t.GetTags()[key]
	if !found {
		return fmt.Sprintf("missing tag '%s'", key)
	}
	if val != expVal {
		return fmt.Sprintf("tag '%s' should have value '%s'", key, expVal)
	}
	return ""
}

type Taggable interface {
	GetTags() map[string]string
	GetElementURL() string
}
