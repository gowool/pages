package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexpPath(t *testing.T) {
	t.Run("Valid path pattern", func(t *testing.T) {
		path := "/blog/([a-z0-9-]+)"
		re, err := regexpPath(path)

		assert.NoError(t, err)
		assert.NotNil(t, re)

		// Test the compiled regex
		match, err := re.FindStringMatch("/blog/test-post")
		assert.NoError(t, err)
		assert.NotNil(t, match)
	})

	t.Run("Empty path", func(t *testing.T) {
		re, err := regexpPath("")

		assert.NoError(t, err)
		assert.NotNil(t, re)
	})

	t.Run("Root path", func(t *testing.T) {
		re, err := regexpPath("/")

		assert.NoError(t, err)
		assert.NotNil(t, re)
	})

	t.Run("Invalid regex pattern", func(t *testing.T) {
		path := "/blog/([a-z0-9-+" // Invalid regex
		re, err := regexpPath(path)

		assert.Error(t, err)
		assert.Nil(t, re)
	})

	t.Run("Cached result", func(t *testing.T) {
		path := "/test"

		// First call
		re1, err1 := regexpPath(path)
		assert.NoError(t, err1)
		assert.NotNil(t, re1)

		// Second call should return cached result
		re2, err2 := regexpPath(path)
		assert.NoError(t, err2)
		assert.Equal(t, re1, re2)
	})

	t.Run("Cached error result", func(t *testing.T) {
		path := "/invalid/([a-z0-9-+"

		// First call should cache error
		re1, err1 := regexpPath(path)
		assert.Error(t, err1)
		assert.Nil(t, re1)

		// Second call should return cached error
		re2, err2 := regexpPath(path)
		assert.Error(t, err2)
		assert.Nil(t, re2)
	})
}

func TestMatchRequest(t *testing.T) {
	t.Run("Empty relative path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)

		pathInfo, err := MatchRequest(req, "")

		assert.NoError(t, err)
		assert.Equal(t, "/test", pathInfo)
	})

	t.Run("Root relative path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)

		pathInfo, err := MatchRequest(req, "/")

		assert.NoError(t, err)
		assert.Equal(t, "/test", pathInfo)
	})

	t.Run("Matching pattern", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/blog/test-post", nil)

		pathInfo, err := MatchRequest(req, "/blog")

		assert.NoError(t, err)
		assert.Equal(t, "/test-post", pathInfo)
	})

	t.Run("Complex matching pattern", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/blog/2023/test-post", nil)

		pathInfo, err := MatchRequest(req, "/blog/([0-9]{4})")

		assert.NoError(t, err)
		assert.Equal(t, "2023", pathInfo)
	})

	t.Run("No match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/other/path", nil)

		pathInfo, err := MatchRequest(req, "/blog")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("Invalid regex pattern", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)

		pathInfo, err := MatchRequest(req, "/invalid/([a-z0-9-+")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("Root path request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/", nil)

		pathInfo, err := MatchRequest(req, "/")

		assert.NoError(t, err)
		assert.Equal(t, "/", pathInfo)
	})
}

func TestMatchRequest_InvalidMatchGroups(t *testing.T) {
	t.Run("invalid groups length error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)

		pathInfo, err := MatchRequest(req, "(")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("FindStringMatch error path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)

		pathInfo, err := MatchRequest(req, "^(")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("empty matched path returns /", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api", nil)

		pathInfo, err := MatchRequest(req, "/api")

		assert.NoError(t, err)
		assert.Equal(t, "/", pathInfo)
	})

	t.Run("relative path returns matched content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/blog/test-post", nil)

		pathInfo, err := MatchRequest(req, "/blog")

		assert.NoError(t, err)
		assert.Equal(t, "/test-post", pathInfo)
	})

	t.Run("complex path with regex", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/blog/2023/test", nil)

		pathInfo, err := MatchRequest(req, "/blog/([0-9]{4})")

		assert.NoError(t, err)
		assert.Equal(t, "2023", pathInfo)
	})
}

func TestHost(t *testing.T) {
	t.Run("Host without port", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Host = "example.com"

		host := Host(req)
		assert.Equal(t, "example.com", host)
	})

	t.Run("Host with port", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Host = "example.com:8080"

		host := Host(req)
		assert.Equal(t, "example.com", host)
	})

	t.Run("IPv6 host with port", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://[::1]", nil)
		req.Host = "[::1]:8080"

		host := Host(req)
		assert.Equal(t, "::1", host)
	})

	t.Run("Invalid host format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com", nil)
		req.Host = "invalid-host-format"

		host := Host(req)
		assert.Equal(t, "invalid-host-format", host)
	})
}

func BenchmarkRegexpPath(b *testing.B) {
	path := "/blog/([a-z0-9-]+)/([0-9]{4})"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = regexpPath(path)
	}
}

func BenchmarkMatchRequest(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com/blog/test-post/2023", nil)
	path := "/blog/([a-z0-9-]+)/([0-9]{4})"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MatchRequest(req, path)
	}
}
