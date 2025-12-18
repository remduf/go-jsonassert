package jsonassert

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type JSONMatcher interface {
	Match(value any, param *string) bool
	// string used in expected JSON (e.g., "<AnyUUID>")
	Tag() string
}

type jsonMatcher struct {
	tagValue  string
	matchFunc func(value any, param *string) bool
}

func (m jsonMatcher) Match(value any, param *string) bool {
	return m.matchFunc(value, param)
}

func (m jsonMatcher) Tag() string {
	return m.tagValue
}

func NewMatcher(
	tag string,
	matcherFunc func(value any, param *string) bool,
) JSONMatcher {
	return jsonMatcher{
		tagValue:  tag,
		matchFunc: matcherFunc,
	}
}

// matches any valid UUID string
var anyUUIDMatcher = NewMatcher("AnyUUID", func(value any, param *string) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	_, err := uuid.Parse(str)
	return err == nil
})

// matches any RFC3339 formatted time string
var anyRFC3339TimeMatcher = NewMatcher("AnyRFC3339Time", func(value any, param *string) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
})

// matches any non-empty string
var anyStringMatcher = NewMatcher("AnyString", func(value any, param *string) bool {
	str, ok := value.(string)
	return ok && str != ""
})

// matches any numeric type
var anyNumberMatcher = NewMatcher("AnyNumber", func(value any, param *string) bool {
	switch value.(type) {
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return true
	default:
		return false
	}
})

// matches any value (always returns true)
var anyValueMatcher = NewMatcher("AnyValue", func(value any, param *string) bool {
	return true
})

// matches any JSON object (map)
var anyObjectMatcher = NewMatcher("AnyObject", func(value any, param *string) bool {
	if value == nil {
		return false
	}
	_, ok := value.(map[string]any)
	return ok
})

// matches any JSON array
var anyArrayMatcher = NewMatcher("AnyArray", func(value any, param *string) bool {
	_, ok := value.([]any)
	return ok
})

// matches any boolean value
var anyBoolMatcher = NewMatcher("AnyBool", func(value any, param *string) bool {
	_, ok := value.(bool)
	return ok
})

// creates a new RegexMatcher with the given pattern and tag.
// Returns an error if the pattern is invalid.
func NewRegexMatcher(pattern string, tag string) (JSONMatcher, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return NewMatcher(tag, func(value any, param *string) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return re.MatchString(str)
	}), nil
}

// creates a new RegexMatcher and panics if the pattern is invalid.
// Use this when you're certain the pattern is valid (e.g., with constants).
func NewMustRegexMatcher(pattern string, tag string) JSONMatcher {
	matcher, err := NewRegexMatcher(pattern, tag)
	if err != nil {
		panic(err)
	}
	return matcher
}

// StringContainMatcher matches strings that contain a specific substring
var stringContainMatcher = NewMatcher("StringContain", func(value any, param *string) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return strings.Contains(str, *param)
})

// matches strings that match a specific regex pattern
// The regex pattern is provided as a parameter to the matcher
// Example usage in expected JSON: "<StringMatch('^foo.*bar$')>"
// if the regex is invalid, it panics at runtime
var stringMatchMatcher = NewMatcher("StringMatch", func(value any, param *string) bool {
	re, err := regexp.Compile(*param)
	if err != nil {
		panic(err)
	}
	str, ok := value.(string)
	if !ok {
		return false
	}
	return re.MatchString(str)
})
