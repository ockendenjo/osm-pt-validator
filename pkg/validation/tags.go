package validation

import "fmt"

func checkTagsPresent(t Taggable, tags ...string) []ValidationError {
	validationErrors := []ValidationError{}

	tagMap := t.GetTags()
	for _, key := range tags {
		_, found := tagMap[key]
		if !found {
			ve := ValidationError{URL: t.GetElementURL(), Message: fmt.Sprintf("missing tag '%s'", key)}
			validationErrors = append(validationErrors, ve)
		}
	}

	return validationErrors
}

func checkTagValue(t Taggable, key string, expVal string) *ValidationError {
	val, found := t.GetTags()[key]
	if !found {
		return &ValidationError{URL: t.GetElementURL(), Message: fmt.Sprintf("missing tag '%s'", key)}
	}
	if val != expVal {
		return &ValidationError{URL: t.GetElementURL(), Message: fmt.Sprintf("tag '%s' should have value '%s'", key, expVal)}
	}
	return nil
}

type Taggable interface {
	GetTags() map[string]string
	GetElementURL() string
}
