# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go package called `pages` that provides a comprehensive content management system for web pages. It's part of the larger `gowool` ecosystem and integrates with the `gowool/wo` web framework.

## Development Commands

### Testing
```bash
# Clear test cache
go clean -testcache

# Run all tests
go test -race ./...

# Run tests with verbose output
go test -race -v ./...

# Run specific test
go test -race -run TestPage_AbsURL

# Run tests with coverage
go test -race -cover ./...
```

### Building
```bash
# Build the package
go build ./...

# Build with verbose output
go build -v ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run go vet
go vet ./...

# Run go mod tidy to clean dependencies
go mod tidy
```

## Architecture Overview

### Core Components

1. **Page Management System**
   - `Page` struct: Core page entity with metadata, routing, and content management
   - `PageManager` interface: Abstraction for page operations (get by ID, URL, pattern, alias)
   - `DefaultPageManager`: Concrete implementation using store backends
   - `PageStore` interface: Data access layer for pages

2. **Site Management**
   - `Site` struct: Multi-site support with localization, timezone, and configuration
   - `SiteManager` and `SiteSelector`: Site resolution and management
   - Support for multiple sites with different hosts, locales, and countries

3. **Page Types**
   - **CMS Pages**: Content-managed pages (`_page_cms`)
   - **Hybrid Pages**: Static routing with dynamic patterns (e.g., `/blog/{slug}`)
   - **Internal Pages**: System pages for errors and special functionality
   - **Error Pages**: Custom 4xx/5xx error pages

4. **Security & Authorization**
   - `PageAuthorizer` interface: Authorization system for page operations
   - `DenyPageAuthorizer`: Deny all page operations
   - Permission levels: Allow, Deny for Create, Read, Update, Delete operations

5. **SEO & Metadata**
   - `MetaTags` struct: Comprehensive SEO metadata management
   - `SEO` struct: Advanced SEO features with JSON-LD, Open Graph, Twitter Cards
   - Automatic sitemap and robots.txt generation
   - URL generation and slug management

### Key Design Patterns

1. **Interface-Driven Design**: All major components use interfaces for flexibility
2. **Middleware Architecture**: HTTP middleware for site detection, page loading, authorization
3. **Error Handling**: Centralized error mapping and custom error page rendering
4. **Template Integration**: Function maps for template rendering with page helpers

### File Organization

- `page*.go`: Page management and related functionality
- `site*.go`: Site management and configuration
- `seo*.go`: SEO and metadata features
- `errors.go`: Error handling and HTTP error mapping
- `event.go`: Event system for page lifecycle events
- `*_test.go`: Comprehensive test coverage

## Dependencies

Key external dependencies:
- `github.com/gowool/wo`: Main web framework
- `github.com/google/uuid`: UUID generation
- `github.com/gosimple/slug`: URL slug generation
- `golang.org/x/text`: Internationalization support
- `github.com/invopop/validation`: Data validation
- `github.com/stretchr/testify`: Testing framework

## Testing Strategy

- Comprehensive unit tests with table-driven test patterns using `testify`
- URL generation and page routing tests
- Authorization and security tests
- Mock implementations for store and external dependencies using `testify`

## Common Development Patterns

1. **Creating New Pages**: Use `NewPage()` constructor, set properties, then `FixURL()` to generate proper URLs
2. **Multi-Site Support**: Sites are detected via middleware and automatically injected into contexts
3. **Dynamic Routing**: Use `{parameter}` syntax in page patterns for dynamic URLs
4. **Error Pages**: Customize by creating pages with patterns `_page_internal_error_4xx` and `_page_internal_error_5xx`
5. **Template Functions**: Access page data and helper functions through provided function maps in templates

## Store Interface

Pages require a `PageStore` implementation. The package includes a memory-based store for development, but production typically requires database or other persistent store implementations.