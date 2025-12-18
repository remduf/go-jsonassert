# go-jsonassert AI Coding Instructions

## Project Overview

**go-jsonassert** is a flexible Go testing utility package for asserting JSON equality with support for dynamic value matching via custom matchers. It enables test developers to compare JSON responses without hardcoding dynamic values like UUIDs or timestamps.

Inspiration: jest snapshot testing and libraries like json-schema and jsonmatch.

## Architecture & Components

### Core Components

1. **[jsonassert.go](../jsonassert.go)** - Public API entry points
   - `MatchJSON()` - Returns bool on comparison result
   - `AssertJSON()` - Calls `t.FailNow()` on mismatch
   - `TestingT` interface - Works with any testing framework (supports `Helper()` for proper test output)

2. **[matcher.go](../matcher.go)** - Matcher system
   - `JSONMatcher` interface - All matchers implement this
   - Built-in matchers: `AnyUUID`, `AnyRFC3339Time`, `AnyString`, `AnyNumber`, `AnyValue`, `AnyBool`, `AnyObject`, `AnyArray`
   - `NewMatcher()` - Constructor for custom matchers
   - String matching matchers: `StringContain`, `StringMatch` (regex)

3. **[compare.go](../compare.go)** - Comparison engine
   - `matcherRegistry` - Central matcher registry (default + custom)
   - `compareValues()` - Recursive deep comparison with error collection
   - `parseMatcherTag()` - Parses tags like `<AnyUUID>`, `<StringMatch('pattern')>`
   - Error collection during traversal for detailed diff output

### Data Flow

1. User calls `MatchJSON(t, expected, actual, ...customMatchers)`
2. Both JSON strings are unmarshaled into generic `any` (interface{})
3. Matcher registry is created with default + custom matchers
4. `compareValues()` recursively traverses both structures:
   - Checks if expected value is a matcher tag (e.g., `"<AnyUUID>"`)
   - If matcher found, applies it; otherwise performs exact comparison
   - For arrays, uses order-insensitive comparison
5. Returns errors; formats detailed diff if mismatch found

### Key Design Decisions

- **Tag-based matchers** - Matchers are detected as specific string patterns (e.g., `"<AnyUUID>"`) during comparison, not parsed separately
- **Order-insensitive arrays** - Arrays are compared element-wise regardless of order
- **Generic interface{}** - JSON unmarshaling produces `map[string]any` and `[]any`; matchers handle type assertion
- **Pluggable matchers** - Additional matchers passed as variadic args; can override defaults

## Development Workflows

### Running Tests

```bash
go test ./...
go test -v ./...          # verbose output
go test -run TestMatchJSON  # specific test
```

### Building

```bash
go build ./...
```

## Project-Specific Conventions

### Matcher Implementation Pattern

All matchers follow this structure (see [matcher.go](../matcher.go) for examples):

```go
var customMatcher = NewMatcher("TagName", func(value any, param *string) bool {
    // Type assertion
    str, ok := value.(string)
    if !ok {
        return false
    }
    // Custom logic
    return validateString(str)
})
```

- **Tag name** (e.g., "AnyUUID") is case-sensitive
- Used in JSON as `"<TagName>"` or with params `"<TagName('param')>"`
- `param` is optional; nil if no parentheses in tag

### Error Message Format

`compareValues()` returns errors like:
```
at key 'user.email': expected <AnyEmail>, got 'invalid@email'
at index 2: expected object, got array
```

Error context includes JSON path for easier debugging.

### Testing Patterns

From [jsonassert_test.go](../jsonassert_test.go):
- Use subtests with descriptive names
- Test exact matches, key order independence, array order independence
- Test each built-in matcher type
- Test custom matcher registration
- Use both `MatchJSON()` (for conditional checks) and `AssertJSON()` (for immediate failure)

## Dependencies

- **github.com/google/uuid** - UUID validation in `anyUUIDMatcher`
- **github.com/stretchr/testify** - Test assertions (dev only)
- **time** (std) - RFC3339 timestamp parsing in `anyRFC3339TimeMatcher`

## Common Tasks

### Add a New Built-in Matcher

1. Define matcher in [matcher.go](../matcher.go) following the pattern
2. Register it in `newMatcherRegistry()` defaultMatchers slice
3. Add tests to [jsonassert_test.go](../jsonassert_test.go)
4. Document in README.md

### Fix Comparison Logic

Modify [compare.go](../compare.go):
- `compareValues()` - Main recursive comparison
- `compareObjects()` - Object-specific logic
- `compareArrays()` - Array comparison (order-insensitive)

### Extend Public API

Keep backward compatibility - `MatchJSON()` and `AssertJSON()` are the public surface.
