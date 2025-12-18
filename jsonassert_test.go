package jsonassert_test

import (
	"fmt"
	"testing"

	"go-jsonassert"

	"github.com/stretchr/testify/assert"
)

const (
	jsonDoesNotMatch            = "JSON doesn't match:"
	jsonShouldMatch             = "JSON should match"
	jsonShouldNotMatch          = "JSON should not match"
	errofShouldHaveBeenCalled   = "ErrorF should have been called"
	failNowShouldHaveBeenCalled = "FailNow should have been called"
	invalidJsonError            = "unexpected end of JSON input"
)

func TestMatchJSON(t *testing.T) {

	testsCases := []struct {
		name     string
		actual   string
		expected string
	}{
		{
			name:     "exact match",
			actual:   `{"name":"John","age":30,"active":true}`,
			expected: `{"name":"John","age":30,"active":true}`,
		},
		{
			name:     "exact match, keys order mismatch",
			actual:   `{"active":true,"name":"John","age":30}`,
			expected: `{"name":"John","age":30,"active":true}`,
		},
		{
			name:     "exact match, with array",
			actual:   `{"users":[{"name":"bob"},{"name":"alice"}]}`,
			expected: `{"users":[{"name":"alice"},{"name":"bob"}]}`,
		},
		{
			name:     "with any uuid",
			actual:   `{"uuid":"16e489e9-c4d1-5393-a48f-d2071cbebe90"}`,
			expected: `{"uuid":"<AnyUUID>"}`,
		},
		{
			name:     "with any RFC3339",
			actual:   `{"time":"2025-10-31T15:00:46Z"}`,
			expected: `{"time":"<AnyRFC3339Time>"}`,
		},
		{
			name:     "with any array",
			actual:   `{"slice":[1,2,3]}`,
			expected: `{"slice":"<AnyArray>"}`,
		},
	}

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonassert.MatchJSON(t, tc.expected, tc.actual)
		})
	}
}

func TestMatchJSON_WithCustomMatcher(t *testing.T) {
	var EmailMatcher = jsonassert.NewMatcher("AnyEmail", func(value any, param *string) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}

		contains := func(s, substr string) bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}

		return len(str) > 0 && contains(str, "@")
	})

	t.Run("CustomMatcher match", func(t *testing.T) {
		expected := `{
			"email": "<AnyEmail>"
		}`

		actual := `{
			"email": "user@example.com"
		}`

		jsonassert.MatchJSON(t, expected, actual, EmailMatcher)
	})

	t.Run("CustomMatcher not match", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"email": "<AnyEmail>"
		}`

		actual := `{
			"email": "user"
		}`

		jsonassert.MatchJSON(mockT, expected, actual, EmailMatcher)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)

		assert.Contains(t, mockT.errorFormat, jsonDoesNotMatch)
	})

}

func TestMatchJSON_WithRegexMatcher(t *testing.T) {
	postalCodeMatcher :=
		jsonassert.NewMustRegexMatcher(`^[A-Z]\d[A-Z] \d[A-Z]\d$`, "PostalCode")

	expected := `{
		"address": {
			"street": "123 Main St",
			"postalCode": "<PostalCode>"
		}
	}`

	actual := `{
		"address": {
			"street": "123 Main St",
			"postalCode": "H2X 1Y7"
		}
	}`

	if !jsonassert.MatchJSON(t, expected, actual, postalCodeMatcher) {
		t.Fatal(jsonShouldMatch)
	}
}

func TestMatchJSON_WithStringContainMatcher(t *testing.T) {

	expected := `{
		"title": "<StringContain('Senior')>"
	}`

	actual := `{
		"title": "Senior Associate"
	}`

	if !jsonassert.MatchJSON(t, expected, actual) {
		t.Fatal(jsonShouldMatch)
	}
}

func TestMatchJSON_WithStringMatchMatcher(t *testing.T) {

	expected := `{
		"message": "<StringMatch(consentValidFor(.*)is missing)>"
	}`

	actual := `{
		"message": "attribute \"consentValidFor\" is missing"
	}`

	if !jsonassert.MatchJSON(t, expected, actual) {
		t.Fatal(jsonShouldMatch)
	}
}

func TestMatchJSON_Mismatch(t *testing.T) {
	mockT := &mockTesting{}

	expected := `{
		"name": "John",
		"age": 30
	}`

	actual := `{
		"name": "Jane",
		"age": 30
	}`

	result := jsonassert.MatchJSON(mockT, expected, actual)

	if result {
		t.Fatal(jsonShouldNotMatch)
	}

	assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
}

func TestMatchJSON_MissingField(t *testing.T) {
	mockT := &mockTesting{}

	expected := `{
		"name": "John",
		"age": 30,
		"email": "john@example.com"
	}`

	actual := `{
		"name": "John",
		"age": 30
	}`

	result := jsonassert.MatchJSON(mockT, expected, actual)

	if result {
		t.Fatal(jsonShouldNotMatch)
	}

	assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
}

func TestMatchJSON_ExtraField(t *testing.T) {
	mockT := &mockTesting{}

	expected := `{
		"name": "John",
		"age": 30
	}`

	actual := `{
		"name": "John",
		"age": 30,
		"email": "john@example.com"
	}`

	result := jsonassert.MatchJSON(mockT, expected, actual)

	if result {
		t.Fatal(jsonShouldNotMatch)
	}

	assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
}

func TestMatchJSON_WithAnyValue(t *testing.T) {
	expected := `{
		"id": "<AnyValue>",
		"data": "<AnyValue>",
		"flag": "<AnyValue>"
	}`

	actual := `{
		"id": 12345,
		"data": {
			"nested": "object"
		},
		"flag": true
	}`

	if !jsonassert.MatchJSON(t, expected, actual) {
		t.Fatal(jsonShouldMatch)
	}
}

func TestMatchJSON_WithAnyObject(t *testing.T) {
	t.Run("AnyObject", func(t *testing.T) {
		expected := `{
			"metadata": "<AnyObject>",
			"settings": "<AnyObject>"
		}`

		actual := `{
			"metadata": {
				"version": "1.0",
				"author": "John"
			},
			"settings": {
				"theme": "dark"
			}
		}`

		if !jsonassert.MatchJSON(t, expected, actual) {
			t.Fatal(jsonShouldMatch)
		}
	})

	t.Run("ShouldFailOnNull", func(t *testing.T) {
		mockT := &mockTesting{}

		expected := `{
			"data": "<AnyObject>"
		}`

		actual := `{
			"data": null
		}`

		result := jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, false, result, jsonShouldNotMatch)
		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
	})

	t.Run("ShouldFailOnPrimitive", func(t *testing.T) {
		mockT := &mockTesting{}

		expected := `{
			"data": "<AnyObject>"
		}`

		actual := `{
			"data": "string value"
		}`

		result := jsonassert.MatchJSON(mockT, expected, actual)

		if result {
			t.Fatal("AnyObject should not accept string")
		}

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
	})

}

func TestMatchJSON_WithAnyArray(t *testing.T) {
	t.Run("AnyArray", func(t *testing.T) {
		expected := `{
			"items": "<AnyArray>",
			"tags": "<AnyArray>"
		}`

		actual := `{
			"items": [1, 2, 3],
			"tags": []
		}`

		if !jsonassert.MatchJSON(t, expected, actual) {
			t.Fatal(jsonShouldMatch)
		}
	})

	t.Run("ShouldFailOnObject", func(t *testing.T) {
		mockT := &mockTesting{}

		expected := `{
			"data": "<AnyArray>"
		}`

		actual := `{
			"data": {"key": "value"}
		}`

		result := jsonassert.MatchJSON(mockT, expected, actual)

		if result {
			t.Fatal("AnyArray should not accept object")
		}

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
	})
}

func TestMatchJSON_WithAnyBool(t *testing.T) {
	t.Run("AnyBool", func(t *testing.T) {
		expected := `{
			"active": "<AnyBool>",
			"verified": "<AnyBool>"
		}`

		actual := `{
			"active": true,
			"verified": false
		}`

		if !jsonassert.MatchJSON(t, expected, actual) {
			t.Fatal(jsonShouldMatch)
		}
	})
	t.Run("ShouldFailOnString", func(t *testing.T) {
		mockT := &mockTesting{}

		expected := `{
			"active": "<AnyBool>"
		}`

		actual := `{
			"active": "true"
		}`

		result := jsonassert.MatchJSON(mockT, expected, actual)

		if result {
			t.Fatal("AnyBool should not accept string")
		}
	})

}

func TestMatchJSON_CombiningMultipleMatchers(t *testing.T) {
	expected := `{
		"id": "<AnyUUID>",
		"user": "<AnyObject>",
		"roles": "<AnyArray>",
		"active": "<AnyBool>",
		"score": "<AnyNumber>",
		"metadata": "<AnyValue>",
		"updatedAt": "<AnyRFC3339Time>"
	}`

	actual := `{
		"id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"user": {
			"name": "Alice",
			"email": "alice@example.com"
		},
		"roles": ["admin", "user"],
		"active": true,
		"score": 95.5,
		"metadata": null,
		"updatedAt": "2025-10-31T15:00:46Z"
	}`

	if !jsonassert.MatchJSON(t, expected, actual) {
		t.Fatal(jsonShouldMatch)
	}
}

func TestMatchJSON_NoMatch(t *testing.T) {

	t.Run("NoMatch", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"name": "John"
		}`

		actual := `{
			"name": "Alice"
		}`

		jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Contains(t, mockT.errorFormat, jsonDoesNotMatch)
	})

	t.Run("ArrayLen", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"item": [{},{},{}]
		}`

		actual := `{
			"item": [{},{},{},{}]
		}`

		jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Contains(t, mockT.errorFormat, jsonDoesNotMatch)
	})

	t.Run("Invalid expected JSON", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"name": "John}`

		actual := `{
			"name": "Alice"
		}`

		jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Equal(t, invalidJsonError, mockT.errorFormat)
	})

	t.Run("Invalid expected nested JSON", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"products": [
				{
					"id": "P001",
					"price": 1.01
				}
			]
		}`

		actual := `{
			"products": [
				{
					"id": "P001",
					"price": 1.011
				}
			]
		}`

		jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Contains(t, mockT.errorFormat, jsonDoesNotMatch)
		assert.Contains(t, mockT.errorMsg, "products[0].price': expected '1.01', actual '1.011")
	})

	t.Run("Invalid actual JSON", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"name": "John"
		}`

		actual := `{
			"name": "Alice}`

		jsonassert.MatchJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Equal(t, invalidJsonError, mockT.errorFormat)
	})

	t.Run("AssertJSON", func(t *testing.T) {
		mockT := &mockTesting{}
		expected := `{
			"name": "John"
		}`

		actual := `{
			"name": "Alice}`

		jsonassert.AssertJSON(mockT, expected, actual)

		assert.Equal(t, true, mockT.errorCalled, errofShouldHaveBeenCalled)
		assert.Equal(t, true, mockT.failCalled, failNowShouldHaveBeenCalled)
		assert.Equal(t, invalidJsonError, mockT.errorFormat)
	})

}

type mockTesting struct {
	errorCalled bool
	failCalled  bool
	errorFormat string
	errorMsg    string
}

func (m *mockTesting) Errorf(format string, args ...any) {
	m.errorCalled = true
	m.errorFormat = format

	m.errorMsg = fmt.Sprintf(format, args...)
}
func (m *mockTesting) FailNow() {
	m.failCalled = true
}
