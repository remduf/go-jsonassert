// Package jsonassert provides flexible JSON comparison with support for custom matchers.
// It enables test developers to assert JSON equality without hardcoding dynamic values
// like UUIDs or timestamps.
package jsonassert

import (
	"fmt"
	"reflect"
	"strings"
)

// holds all registered matchers indexed by their tags
type matcherRegistry struct {
	matchers map[string]JSONMatcher
}

// creates a new registry with default matchers and any additional ones
func newMatcherRegistry(additionalMatchers []JSONMatcher) *matcherRegistry {
	registry := &matcherRegistry{
		matchers: make(map[string]JSONMatcher),
	}

	// Register default matchers
	defaultMatchers := []JSONMatcher{
		anyValueMatcher,
		anyUUIDMatcher,
		anyRFC3339TimeMatcher,
		anyStringMatcher,
		anyNumberMatcher,
		anyObjectMatcher,
		anyBoolMatcher,
		anyArrayMatcher,
		stringContainMatcher,
		stringMatchMatcher,
	}

	for _, m := range defaultMatchers {
		registry.matchers[m.Tag()] = m
	}

	// Register additional matchers (can override defaults)
	for _, m := range additionalMatchers {
		registry.matchers[m.Tag()] = m
	}

	return registry
}

// retrieves a matcher by its tag
func (r *matcherRegistry) getMatcherFunc(expected any) (func(value any) (string, bool), bool) {
	expectedStr, ok := expected.(string)
	if !ok {
		return nil, false
	}
	tag, param := parseMatcherTag(expectedStr)
	matcher, ok := r.matchers[tag]
	if !ok {
		return nil, false
	}

	return func(value any) (string, bool) {
		return expectedStr, matcher.Match(value, param)
	}, true
}

// parseMatcherTag parses a matcher tag and extracts the name and parameters
// Examples: "<AnyUUID>" -> ("AnyUUID", "")
//
//	"<StringContain('foo')>" -> ("StringContain", "foo")
//	"<TimeFormat('2006-01-02')>" -> ("TimeFormat", "2006-01-02")
func parseMatcherTag(tag string) (string, *string) {
	if !strings.HasPrefix(tag, "<") || !strings.HasSuffix(tag, ">") {
		return "", nil
	}

	// Remove < and >
	content := tag[1 : len(tag)-1]

	// Check for parameters
	openParen := strings.Index(content, "(")
	if openParen == -1 {
		// No parameters
		return content, nil
	}

	closeParen := strings.LastIndex(content, ")")
	if closeParen == -1 || closeParen <= openParen {
		// Invalid format
		return "", nil
	}

	name := content[:openParen]
	paramContent := content[openParen+1 : closeParen]

	// Remove quotes if present
	if len(paramContent) >= 2 {
		if (strings.HasPrefix(paramContent, "'") && strings.HasSuffix(paramContent, "'")) ||
			(strings.HasPrefix(paramContent, "\"") && strings.HasSuffix(paramContent, "\"")) {
			paramContent = paramContent[1 : len(paramContent)-1]
		}
	}

	return name, &paramContent
}

// compareValues recursively compares two values (objects, arrays, primitives).
// It handles matcher tags in the expected value (e.g., "<AnyUUID>").
func (r *matcherRegistry) compareValues(expected, actual any, path string) []string {
	var errors []string

	// Check if expected value is a matcher tag (e.g., "<AnyUUID>")
	if matchFunc, found := r.getMatcherFunc(expected); found {
		tag, match := matchFunc(actual)
		if !match {
			errors = append(errors,
				fmt.Sprintf("Path '%s': value '%v' doesn't match '%s'", path, actual, tag))
		}
		return errors
	}

	// Compare types - they must match
	expectedType := reflect.TypeOf(expected)
	actualType := reflect.TypeOf(actual)

	if expectedType != actualType {
		errors = append(errors,
			fmt.Sprintf("Path '%s': expected type %v, actual %v", path, expectedType, actualType))
		return errors
	}

	switch expectedVal := expected.(type) {
	case map[string]any:
		errs := r.compareMap(expectedVal, actual.(map[string]any), path)
		errors = append(errors, errs...)

	case []any:
		errs := r.compageSlice(expectedVal, actual.([]any), path)
		errors = append(errors, errs...)

	default:
		// Compare primitive values
		if !reflect.DeepEqual(expected, actual) {
			errors = append(errors, fmt.Sprintf("Path '%s': expected '%v', actual '%v'", path, expected, actual))
		}
	}

	return errors
}

func (r *matcherRegistry) compareMap(expectedMap, actualMap map[string]any, path string) []string {
	var errors []string

	// Check all expected keys exist in actual
	for key, expectedValue := range expectedMap {
		currentPath := path + "." + key
		actualValue, exists := actualMap[key]
		if !exists {
			errors = append(errors, fmt.Sprintf("Path '%s': missing in actual JSON", currentPath))
		}

		errs := r.compareValues(expectedValue, actualValue, currentPath)
		match := len(errs) == 0
		if !match {
			errors = append(errors, errs...)
		}
	}

	// Check for unexpected keys in actual
	for key := range actualMap {
		if _, exists := expectedMap[key]; !exists {
			currentPath := path + "." + key
			errors = append(errors,
				fmt.Sprintf("Path '%s': unexpected in actual JSON", currentPath))
		}
	}

	return errors
}

func (r *matcherRegistry) compageSlice(expectedSlice, actualSlice []any, path string) []string {
	var errors []string

	if len(expectedSlice) != len(actualSlice) {
		errors = append(errors,
			fmt.Sprintf(
				"Path '%s': array length expected %d, actual %d",
				path,
				len(expectedSlice),
				len(actualSlice),
			),
		)
		return errors
	}

	// Track which actual items have been actualMatched
	actualMatched := make([]bool, len(actualSlice))

	// For each expected item, find a matching actual item
	for i, expectedItem := range expectedSlice {
		found := false
		var bestErrors []string

		for j, actualItem := range actualSlice {
			if actualMatched[j] {
				continue
			}

			errs := r.compareValues(expectedItem, actualItem, fmt.Sprintf("%s[%d]", path, i))
			match := len(errs) == 0
			if match {
				actualMatched[j] = true
				found = true
				break
			}

			// Keep first set of errors for reporting
			if !match && bestErrors == nil {
				bestErrors = errs
			}
		}

		if !found {
			errors = append(errors, bestErrors...)
			errors = append(errors,
				fmt.Sprintf("Path '%s[%d]': not found in actual array", path, i))
		}
	}

	return errors
}
