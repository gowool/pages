package internal

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
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
				r.Header.Set(headerXForwardedProto, "https")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Protocol header",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedProtocol, "https")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Ssl header with on",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedSsl, "on")
			},
			want: "https",
		},
		{
			name: "X-Forwarded-Ssl header with off",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXForwardedSsl, "off")
			},
			want: "http",
		},
		{
			name: "X-Url-Scheme header",
			setupReq: func(r *http.Request) {
				r.Header.Set(headerXUrlScheme, "https")
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
