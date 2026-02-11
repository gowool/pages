# AGENTS.md

This file provides guidance for agentic coding assistants working in this repository.

## Development Commands

### Testing
```bash
# Clear test cache
go clean -testcache

# Run all tests with race detection
go test -race ./...

# Run tests with verbose output
go test -race -v ./...

# Run specific test function
go test -race -run TestPage_AbsURL

# Run tests for specific package
go test -race -v ./page_test.go

# Run tests with coverage
go test -race -cover ./...
```

### Building & Code Quality
```bash
# Build the package
go build ./...

# Format code (use go fmt)
go fmt ./...

# Run go vet
go vet ./...

# Clean up dependencies
go mod tidy

# Run static analysis (if available)
golangci-lint run -v --timeout=5m --build-tags=race --output.code-climate.path gl-code-quality-report.json
```

## Code Style Guidelines

### Imports
- Use standard library imports first, then third-party imports
- Group imports with blank lines between groups
- No unused imports - always run `go mod tidy` and `go vet`
- Import only what you need from the standard library

### Formatting
- Use `go fmt` for all code - no manual formatting
- Use tabs for indentation (Go standard)
- Maximum line length is not strictly enforced but keep it readable (~120 chars)
- Use `omitzero` JSON/YAML tags for zero-value omitempty behavior on time fields

### Types & Structs
- Define types using `type T baseType` pattern
- Exported types use PascalCase, unexported types use camelCase
- Pointer receiver methods for methods that modify state
- Value receiver methods for methods that don't modify state
- Always implement `String()` method for custom types with meaningful string representation

### Naming Conventions
- Constants: PascalCase (e.g., `PageCMS`, `DefaultCharset`)
- Interfaces: Simple nouns describing capability (e.g., `PageStore`, `PageManager`)
- Structs: Nouns, PascalCase for exported
- Functions/Methods: PascalCase for exported, camelCase for unexported
- Variables: camelCase
- Private fields: camelCase, unexported (lowercase first letter)
- Package-level variables: PascalCase if exported, camelCase if not
- Errors: `ErrXxx` format (e.g., `ErrPageNotFound`, `ErrSiteNotFound`)

### Constructors
- Use `NewTypeName()` pattern for constructors
- Use `NewTypeName()` prefix for all factory functions
- Constructors should initialize default values and validate required parameters
- Panic with descriptive message for missing required parameters (e.g., `panic("page manager: store is required")`)

### Error Handling
- Define errors as package-level variables using `errors.New()`
- Use `errors.Is()` and `errors.AsType[E error]()` for error checking
- Wrap errors with context using `fmt.Errorf("msg: %w", err)`
- Return nil pointers for "not found" cases, don't wrap in error
- Use `fmt.Errorf` for errors with dynamic messages

### Interfaces
- Define interfaces with minimal necessary methods
- Use interface composition for combining related behaviors
- Implement interface compliance check: `var _ Interface = (*Concrete)(nil)`
- Keep interfaces small and focused (single responsibility)

### Testing
- Use table-driven tests with `[]struct` pattern
- Name test functions as `TestTypeName_MethodName`
- Use `github.com/stretchr/testify/assert` for assertions
- Create mock implementations using `github.com/stretchr/testify/mock`
- Mock files should be in `mock_test.go` file
- Use `go test -race` to catch data races
- Test both success and error paths

### Generics
- Use type constraints like `[T Resolver]` for generic types
- Keep generic type names concise (T, K, V)
- Document generic type constraints with comments

### Constants
- Define related constants together as a block
- Use `iota` for enumerations
- Implement `String()` method for custom enum types
- Use `TypeNameFromString()` pattern for string-to-enum conversion
- Separate const blocks for different constant groups (e.g., link relations, policies)

### String Building
- Use `strings.Builder` for concatenating multiple strings
- Use `fmt.Sprintf` for simple string formatting
- Avoid string concatenation with `+` in loops

### Maps & Slices
- Use `maps.Clone()` and `slices.Clone()` for copying
- Initialize maps and slices with `make()` when capacity is known
- Check for nil before accessing map/slice elements

### ID Type
- Use `ID` type (string alias) for all entity identifiers
- Implement `IsZero()` and `String()` methods on ID type
- Generate IDs using `github.com/google/uuid`

### JSON/YAML Tags
- Use `json:"field,omitempty"` for optional fields
- Use `json:"field,omitzero"` for zero-value omitempty on time.Time
- Use both `json` and `yaml` tags for consistency

### Validation
- Use `github.com/invopop/validation` for struct validation
- Implement `Validate() error` method on request/response DTOs
- Use `validation.ValidateStruct()` with field-level rules

### Context
- Pass `context.Context` as first parameter to store/manager methods
- Use `context.Background()` for tests when no context is needed
- Store and retrieve domain objects from context using `FromContext()` or `MustContext()`

### Logging
- Use `log/slog` for structured logging
- Create loggers with `logger.WithGroup("component_name")` for context
- Use slog.DiscardHandler for nil logger defaults
- Log errors with context: `logger.Error("message", "key", value)`

### Visibility
- Export symbols only when needed by external packages
- Keep implementation details unexported
- Private fields lowercase, exported fields PascalCase

### Comments
- No inline comments for obvious code
- Document exported functions, types, and package behavior
- Keep comments concise and focused on "why" not "what"
