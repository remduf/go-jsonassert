# jsonassert

A flexible Go package for asserting JSON equality in tests with support for dynamic value matching.

## Features

- Deep JSON comparison with detailed error messages
- Built-in matchers for common patterns (UUIDs, timestamps, types)
- Custom matcher support via the `JSONMatcher` interface
- Order-insensitive array comparison
- Works with any testing framework (standard `testing`, testify, etc.)


## Quick Start

```go
import (
    "testing"
    "github.com/go-jsonassert"
)

func TestAPI(t *testing.T) {
    expected := `{
        "id": "<AnyUUID>",
        "name": "John Doe",
        "age": 30
    }`
    
    actual := `{
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "John Doe",
        "age": 30
    }`
    
    jsonassert.AssertJSON(t, expected, actual)
}
```

## Built-in Matchers

Use these tags in your expected JSON to match dynamic values:

| Matcher | Description | Example |
|---------|-------------|---------|
| `<AnyValue>` | Matches any value | `"field": "<AnyValue>"` |
| `<AnyString>` | Matches any non-empty string | `"name": "<AnyString>"` |
| `<AnyNumber>` | Matches any numeric type | `"age": "<AnyNumber>"` |
| `<AnyBool>` | Matches any boolean | `"active": "<AnyBool>"` |
| `<AnyObject>` | Matches any JSON object | `"data": "<AnyObject>"` |
| `<AnyArray>` | Matches any JSON array | `"items": "<AnyArray>"` |
| `<AnyUUID>` | Matches valid UUID strings | `"id": "<AnyUUID>"` |
| `<AnyRFC3339Time>` | Matches RFC3339 timestamps | `"createdAt": "<AnyRFC3339Time>"` |
| `<StringContain(substring string)>` | Matches string contain `substring` | `"name": "<StringContain('foo')>"` |

## Usage Examples

### Basic Assertion

```go
func TestBasicJSON(t *testing.T) {
    expected := `{"status": "success", "code": 200}`
    actual := `{"status": "success", "code": 200}`
    
    jsonassert.AssertJSON(t, expected, actual)
}
```

### Using Matchers

```go
func TestWithMatchers(t *testing.T) {
    expected := `{
        "user_id": "<AnyUUID>",
        "username": "<AnyString>",
        "score": "<AnyNumber>",
        "verified": "<AnyBool>",
        "created_at": "<AnyRFC3339Time>"
    }`
    
    actual := `{
        "user_id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "alice",
        "score": 42,
        "verified": true,
        "created_at": "2025-11-03T10:30:00Z"
    }`
    
    jsonassert.AssertJSON(t, expected, actual)
}
```

### Order-Insensitive Arrays

Arrays are compared regardless of order:

```go
func TestArrayComparison(t *testing.T) {
    expected := `{"items": [1, 2, 3]}`
    actual := `{"items": [3, 1, 2]}`
    
    jsonassert.AssertJSON(t, expected, actual) // Passes!
}
```

### Custom Matchers

Create your own matchers by using `jsonassert.NewMatcher`:

```go
// Custom matcher for email addresses
var AnyEmailMatcher = jsonassert.NewMatcher("AnyEmail", func(value any, param *string) bool {
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

// Use in tests
func TestCustomMatcher(t *testing.T) {
    expected := `{"email": "<Email>"}`
    actual := `{"email": "user@example.com"}`
    
    jsonassert.AssertJSON(t, expected, actual,
        AnyEmailMatcher{},
    )
}
```

### Regex Matcher

Match strings against patterns:

```go
func TestRegexMatcher(t *testing.T) {
    phonePattern := jsonassert.NewMustRegexMatcher(`^\+\d{1,3}-\d{3}-\d{4}$`, "<Phone>")
    
    expected := `{"phone": "<Phone>"}`
    actual := `{"phone": "+1-555-1234"}`
    
    jsonassert.AssertJSON(t, expected, actual, phonePattern)
}
```

### Non-Fatal Matching

Use `MatchJSON` instead of `AssertJSON` to continue test execution on failure:

```go
func TestMultipleAssertions(t *testing.T) {
    if !jsonassert.MatchJSON(t, expected1, actual1) {
        t.Log("First assertion failed, but continuing...")
    }
    
    if !jsonassert.MatchJSON(t, expected2, actual2) {
        t.Log("Second assertion also failed")
    }
}
```

## Error Messages

The package provides detailed error messages showing exactly what doesn't match:

```
JSON doesn't match:
Path 'user.id': expected type string, actual float64
Path 'user.email': value 'invalid' doesn't match '<AnyUUID>'
Path 'items[2]': not found in actual array
Path 'extra_field': unexpected in actual JSON
```

## API Reference

### Functions

#### `MatchJSON(t TestingT, expected string, actual string, additionalMatchers ...JSONMatcher) bool`

Compares two JSON strings and returns `true` if they match. Reports errors via `t.Errorf()` but doesn't stop test execution.

#### `AssertJSON(t TestingT, expected string, actual string, additionalMatchers ...JSONMatcher)`

Like `MatchJSON` but calls `t.FailNow()` on mismatch, stopping test execution immediately.

### Interfaces

#### `TestingT`

Minimal interface compatible with Go's standard `testing.T`:

```go
type TestingT interface {
    Errorf(format string, args ...any)
    Helper()
    FailNow()
}
```

#### `JSONMatcher`

Interface for creating custom matchers:

```go
type JSONMatcher interface {
    Match(value any) bool  // Test if value matches
    Tag() string           // Return tag string (e.g., "<MyMatcher>")
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details