package osm

import "fmt"

func ValidateRelation(r Relation) ([]string, error) {
	validationErrors := []string{}
	for _, relationElement := range r.Elements {
		ve, err := validationRelationElement(relationElement)
		if err != nil {
			return []string{}, err
		}
		validationErrors = append(validationErrors, ve...)
	}
	return validationErrors, nil
}

func validationRelationElement(re RelationElement) ([]string, error) {
	return validateRETags(re), nil
}

func validateRETags(re RelationElement) []string {
	validationErrors := []string{}

	for _, s := range []string{"from", "to", "name", "network", "operator", "ref"} {
		ve := checkTagPresent(re, s)
		if ve != "" {
			validationErrors = append(validationErrors, ve)
		}
	}

	for k, v := range map[string]string{
		"type":                     "route",
		"public_transport:version": "2",
	} {
		ve := checkTagValue(re, k, v)
		if ve != "" {
			validationErrors = append(validationErrors, ve)
		}
	}

	return validationErrors
}

func checkTagPresent(t Taggable, key string) string {
	_, found := t.GetTags()[key]
	if !found {
		return fmt.Sprintf("missing tag '%s'", key)
	}
	return ""
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
}
