package pages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

func TestPageCtx(t *testing.T) {
	t.Run("Returns PageCtxFunc", func(t *testing.T) {
		viewCtx := PageCtx()

		assert.NotNil(t, viewCtx, "PageCtx should return non-nil function")

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		req := httptest.NewRequest("GET", "/", nil)

		result := viewCtx(req, c)

		pc, ok := result.(PageContext)
		assert.True(t, ok, "Result should be PageContext")
		assert.Equal(t, c, pc.Context)
		assert.Equal(t, req, pc.Request)
	})

	t.Run("PageContext Value returns context value", func(t *testing.T) {
		type testKey struct{}
		value := "testValue"
		parentCtx := context.WithValue(context.Background(), testKey{}, value)

		req := httptest.NewRequest("GET", "/", nil).WithContext(parentCtx)
		pc := PageContext{Request: req}

		result := pc.Value(testKey{})

		assert.Equal(t, value, result)
	})

	t.Run("PageContext Value returns nil for missing key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		pc := PageContext{Request: req}

		result := pc.Value("nonexistent")

		assert.Nil(t, result)
	})
}

func TestErrorPattern(t *testing.T) {
	finder := ErrorPattern()

	t.Run("StatusUnauthorized", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusUnauthorized)
		assert.Equal(t, PageErrorUnauthorized, pattern)
	})

	t.Run("StatusForbidden", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusForbidden)
		assert.Equal(t, PageErrorForbidden, pattern)
	})

	t.Run("StatusNotFound", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusNotFound)
		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("4xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusBadRequest)
		assert.Equal(t, PageError4xx, pattern)

		pattern = finder(context.Background(), http.StatusPaymentRequired)
		assert.Equal(t, PageError4xx, pattern)

		pattern = finder(context.Background(), http.StatusConflict)
		assert.Equal(t, PageError4xx, pattern)
	})

	t.Run("5xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusInternalServerError)
		assert.Equal(t, PageError5xx, pattern)

		pattern = finder(context.Background(), http.StatusBadGateway)
		assert.Equal(t, PageError5xx, pattern)

		pattern = finder(context.Background(), http.StatusServiceUnavailable)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("Other 5xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusNotImplemented)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("2xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusOK)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("3xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusMovedPermanently)
		assert.Equal(t, PageError5xx, pattern)
	})
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
		w.Header().Set(headerContentType, "application/json")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.status = http.StatusOK
		w.Header().Set(headerContentType, mimeTextHTML)

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
		w.status = http.StatusOK
		w.WriteHeader(http.StatusNotFound)

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("XMLHttpRequest header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(headerXRequestedWith, xmlHTTPRequest)

		isDecorable := IsDecorable(w, req)

		assert.False(t, isDecorable)
	})

	t.Run("All conditions allow decoration", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.status = http.StatusOK
		req := httptest.NewRequest("GET", "/", nil)

		isDecorable := IsDecorable(w, req)

		assert.True(t, isDecorable)
	})
}
