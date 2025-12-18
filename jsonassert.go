package jsonassert

import (
	"encoding/json"
	"strings"
)

// TestingT is the minimal interface required for testing frameworks
type TestingT interface {
	Errorf(format string, args ...any)
	FailNow()
}
type tHelper = interface {
	Helper()
}

// MatchJSON compares two JSON strings and returns true if they match.
// It supports custom matchers for flexible comparison (e.g., <AnyUUID>, <AnyString>).
// Additional matchers can be provided via the additionalMatchers parameter.
func MatchJSON(t TestingT, expected string, actual string, additionalMatchers ...JSONMatcher) bool {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	var expectedData, actualData any

	// Parse expected JSON
	if err := json.Unmarshal([]byte(expected), &expectedData); err != nil {
		t.Errorf(err.Error())
		return false
	}

	// Parse actual JSON
	if err := json.Unmarshal([]byte(actual), &actualData); err != nil {
		t.Errorf(err.Error())
		return false
	}

	// Create matcher registry with default and additional matchers
	registry := newMatcherRegistry(additionalMatchers)
	errors := registry.compareValues(expectedData, actualData, "")

	match := len(errors) == 0
	if !match {
		prettyActual, _ := json.MarshalIndent(actualData, "", "  ")

		t.Errorf(`JSON doesn't match:
expected:
%s
actual:
%s
diffs:
%s`,
			expected,
			string(prettyActual),
			strings.Join(errors, "\n"),
		)
	}

	return match
}

// AssertJSON is similar to MatchJSON but calls t.FailNow() if comparison fails.
// Use this when you want the test to stop immediately on mismatch.
func AssertJSON(t TestingT, expected string, actual string, additionalMatchers ...JSONMatcher) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}
	if !MatchJSON(t, expected, actual, additionalMatchers...) {
		t.FailNow()
	}
}
