package pages

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gowool/gor"
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

func TestScheme(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func(*http.Request)
		want     string
	}{
		{
			name: "TLS connection",
			setupReq: func(r *http.Request) {
				r.TLS = &tls.ConnectionState{}
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Proto header",
			setupReq: func(r *http.Request) {
				r.Header.Set(gor.HeaderXForwardedProto, "https")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Protocol header",
			setupReq: func(r *http.Request) {
				r.Header.Set(gor.HeaderXForwardedProtocol, "https")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Ssl header with on",
			setupReq: func(r *http.Request) {
				r.Header.Set(gor.HeaderXForwardedSsl, "on")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Ssl header with off",
			setupReq: func(r *http.Request) {
				r.Header.Set(gor.HeaderXForwardedSsl, "off")
			},
			want: "http",
		},
		{
			name: "X-Url-Scheme header",
			setupReq: func(r *http.Request) {
				r.Header.Set(gor.HeaderXUrlScheme, "https")
			},
			want: "https",
		},
		{
			name:     "default to http",
			setupReq: func(r *http.Request) {},
			want:     "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL:    &url.URL{Path: "/test/path"},
				Host:   "localhost",
				Header: make(http.Header),
			}
			tt.setupReq(req)

			got := Scheme(req)

			assert.Equal(t, tt.want, got, "Scheme should return expected scheme")
		})
	}
}

func TestCheckMethod(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		skip         string
		expectedPath string
		expectedOK   bool
	}{
		{
			name:         "no method specified in skip",
			method:       "GET",
			skip:         "/api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "matching method",
			method:       "GET",
			skip:         "GET /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "non-matching method",
			method:       "POST",
			skip:         "GET /api/users",
			expectedPath: "",
			expectedOK:   false,
		},
		{
			name:         "case sensitive method matching",
			method:       "get",
			skip:         "GET /api/users",
			expectedPath: "",
			expectedOK:   false,
		},
		{
			name:         "different matching method",
			method:       "POST",
			skip:         "POST /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "PUT method matching",
			method:       "PUT",
			skip:         "PUT /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "DELETE method matching",
			method:       "DELETE",
			skip:         "DELETE /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "PATCH method matching",
			method:       "PATCH",
			skip:         "PATCH /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "HEAD method matching",
			method:       "HEAD",
			skip:         "HEAD /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "OPTIONS method matching",
			method:       "OPTIONS",
			skip:         "OPTIONS /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "method with empty path",
			method:       "GET",
			skip:         "GET ",
			expectedPath: "",
			expectedOK:   true,
		},
		{
			name:         "malformed pattern - missing path",
			method:       "GET",
			skip:         "GET",
			expectedPath: "GET",
			expectedOK:   true,
		},
		{
			name:         "malformed pattern - extra spaces",
			method:       "GET",
			skip:         "GET   /api/users",
			expectedPath: "/api/users",
			expectedOK:   true,
		},
		{
			name:         "empty skip string",
			method:       "GET",
			skip:         "",
			expectedPath: "",
			expectedOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, ok := CheckMethod(tt.method, tt.skip)
			assert.Equal(t, tt.expectedPath, path)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "pattern without method",
			pattern: "/",
			want:    "/",
		},
		{
			name:    "pattern with GET method",
			pattern: "GET /",
			want:    "/",
		},
		{
			name:    "pattern with POST method",
			pattern: "POST /blog/posts",
			want:    "/blog/posts",
		},
		{
			name:    "pattern with PUT method",
			pattern: "PUT /api/users/123",
			want:    "/api/users/123",
		},
		{
			name:    "pattern with DELETE method",
			pattern: "DELETE /api/users/123",
			want:    "/api/users/123",
		},
		{
			name:    "pattern with PATCH method",
			pattern: "PATCH /api/users/123",
			want:    "/api/users/123",
		},
		{
			name:    "pattern with HEAD method",
			pattern: "HEAD /api/health",
			want:    "/api/health",
		},
		{
			name:    "pattern with OPTIONS method",
			pattern: "OPTIONS /api/cors",
			want:    "/api/cors",
		},
		{
			name:    "pattern with dynamic path",
			pattern: "GET /posts/{id}",
			want:    "/posts/{id}",
		},
		{
			name:    "pattern with multiple dynamic segments",
			pattern: "GET /posts/{year}/{month}/{slug}",
			want:    "/posts/{year}/{month}/{slug}",
		},
		{
			name:    "pattern with rest parameter",
			pattern: "GET /api/{...rest}",
			want:    "/api/{...rest}",
		},
		{
			name:    "pattern with lowercase method",
			pattern: "get /test",
			want:    "/test",
		},
		{
			name:    "pattern with mixed case method",
			pattern: "Post /test",
			want:    "/test",
		},
		{
			name:    "complex api path",
			pattern: "GET /api/v1/users/{id}/posts/{postId}",
			want:    "/api/v1/users/{id}/posts/{postId}",
		},
		{
			name:    "pattern with trailing slash",
			pattern: "GET /blog/posts/",
			want:    "/blog/posts/",
		},
		{
			name:    "pattern with query parameters in pattern",
			pattern: "GET /search?q=test",
			want:    "/search?q=test",
		},
		{
			name:    "empty pattern",
			pattern: "",
			want:    "",
		},
		{
			name:    "pattern with only method",
			pattern: "GET",
			want:    "GET",
		},
		{
			name:    "pattern with multiple spaces",
			pattern: "GET  /test/path",
			want:    " /test/path",
		},
		{
			name:    "pattern with method and space only",
			pattern: "GET ",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{Pattern: tt.pattern}
			got := Pattern(req)
			assert.Equal(t, tt.want, got)
		})
	}
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
		w.Header().Set(gor.HeaderContentType, "application/json")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.code = http.StatusOK
		w.Header().Set(gor.HeaderContentType, gor.MIMETextHTML)

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
		req.Header.Set(gor.HeaderXRequestedWith, gor.XMLHTTPRequest)

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
