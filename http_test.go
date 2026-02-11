package pages

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gowool/keratin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRequestWithPattern(pattern, path string) *http.Request {
	req := &http.Request{
		Pattern: pattern,
		URL:     &url.URL{Path: path},
	}
	return req
}

func TestPatternArgs(t *testing.T) {
	t.Run("empty pattern", func(t *testing.T) {
		req := createRequestWithPattern("/test", "/test")
		fn := PatternArgs()
		args := fn(req)

		assert.Nil(t, args, "PatternArgs should return nil for static pattern")
	})

	t.Run("single parameter", func(t *testing.T) {
		req := createRequestWithPattern("/posts/{id}", "/posts/123")
		req.SetPathValue("id", "123")

		fn := PatternArgs()
		args := fn(req)

		require.NotNil(t, args, "PatternArgs should return args for dynamic pattern")
		require.Len(t, args, 2, "Should have key-value pair")
		assert.Equal(t, "{id}", args[0], "First arg should be parameter key")
		assert.Equal(t, "123", args[1], "Second arg should be parameter value")
	})

	t.Run("multiple parameters", func(t *testing.T) {
		req := createRequestWithPattern("/posts/{year}/{month}/{slug}", "/posts/2024/01/test-post")
		req.SetPathValue("year", "2024")
		req.SetPathValue("month", "01")
		req.SetPathValue("slug", "test-post")

		fn := PatternArgs()
		args := fn(req)

		require.NotNil(t, args, "PatternArgs should return args for dynamic pattern")
		require.Len(t, args, 6, "Should have 3 key-value pairs")
		assert.Equal(t, "{year}", args[0])
		assert.Equal(t, "2024", args[1])
		assert.Equal(t, "{month}", args[2])
		assert.Equal(t, "01", args[3])
		assert.Equal(t, "{slug}", args[4])
		assert.Equal(t, "test-post", args[5])
	})

	t.Run("parameter with dots", func(t *testing.T) {
		req := createRequestWithPattern("/api/{...rest}", "/api/foo/bar/baz")
		req.SetPathValue("rest", "foo/bar/baz")

		fn := PatternArgs()
		args := fn(req)

		require.NotNil(t, args)
		require.Len(t, args, 2)
		assert.Equal(t, "{rest}", args[0], "Dots should be removed from parameter name")
		assert.Equal(t, "foo/bar/baz", args[1])
	})

	t.Run("malformed pattern without closing brace", func(t *testing.T) {
		req := createRequestWithPattern("/posts/{id", "/posts/123")

		fn := PatternArgs()
		args := fn(req)

		assert.NotNil(t, args, "PatternArgs returns empty slice for malformed pattern")
		assert.Empty(t, args, "PatternArgs should return empty slice for malformed pattern")
	})

	t.Run("pattern with mixed static and dynamic", func(t *testing.T) {
		req := createRequestWithPattern("/api/v1/users/{id}/posts/{postId}", "/api/v1/users/123/posts/456")
		req.SetPathValue("id", "123")
		req.SetPathValue("postId", "456")

		fn := PatternArgs()
		args := fn(req)

		require.NotNil(t, args)
		require.Len(t, args, 4)
		assert.Equal(t, "{id}", args[0])
		assert.Equal(t, "123", args[1])
		assert.Equal(t, "{postId}", args[2])
		assert.Equal(t, "456", args[3])
	})

	t.Run("closing brace without opening brace", func(t *testing.T) {
		req := createRequestWithPattern("/}{id}", "/test")

		fn := PatternArgs()
		assert.Panics(t, func() {
			fn(req)
		}, "PatternArgs should panic on invalid pattern with closing brace without opening")
	})
}

func TestIsDecorable(t *testing.T) {
	t.Run("Non-HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(keratin.HeaderContentType, "application/json")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.code = http.StatusOK
		w.Header().Set(keratin.HeaderContentType, keratin.MIMETextHTML)

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.True(t, isDecorable)
	})

	t.Run("X-Page-Not-Decorable header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(HeaderXPageNotDecorable, "1")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("X-Page-Decorable header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(HeaderXPageDecorable, "1")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.True(t, isDecorable)
	})

	t.Run("Status not OK", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.code = http.StatusOK
		w.WriteHeader(http.StatusNotFound)

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("XMLHttpRequest header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(keratin.HeaderXRequestedWith, keratin.XMLHTTPRequest)

		isDecorable := IsDecorable(w, req)

		assert.False(t, isDecorable)
	})

	t.Run("All conditions allow decoration", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.code = http.StatusOK
		req := httptest.NewRequest("GET", "/", nil)

		isDecorable := IsDecorable(w, req)

		assert.True(t, isDecorable)
	})
}
