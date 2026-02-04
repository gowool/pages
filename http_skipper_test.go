package pages

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChainSkipper(t *testing.T) {
	tests := []struct {
		name     string
		skippers []Skipper
		path     string
		method   string
		want     bool
	}{
		{
			name:     "no skippers",
			skippers: []Skipper{},
			path:     "/test",
			method:   http.MethodGet,
			want:     false,
		},
		{
			name: "first skipper returns true",
			skippers: []Skipper{
				PrefixPathSkipper("/api"),
				PrefixPathSkipper("/admin"),
			},
			path:   "/api/users",
			method: http.MethodGet,
			want:   true,
		},
		{
			name: "second skipper returns true",
			skippers: []Skipper{
				PrefixPathSkipper("/api"),
				PrefixPathSkipper("/admin"),
			},
			path:   "/admin/users",
			method: http.MethodGet,
			want:   true,
		},
		{
			name: "no skipper matches",
			skippers: []Skipper{
				PrefixPathSkipper("/api"),
				PrefixPathSkipper("/admin"),
			},
			path:   "/public/page",
			method: http.MethodGet,
			want:   false,
		},
		{
			name: "multiple skippers all return false",
			skippers: []Skipper{
				EqualPathSkipper("/exact"),
				SuffixPathSkipper(".json"),
			},
			path:   "/public/page",
			method: http.MethodGet,
			want:   false,
		},
		{
			name:     "nil skipper list",
			skippers: nil,
			path:     "/test",
			method:   http.MethodGet,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := ChainSkipper(tt.skippers...)
			req := createRequest(tt.method, tt.path)
			got := chain(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrefixPathSkipper(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		path     string
		method   string
		want     bool
	}{
		{
			name:     "matches single prefix",
			prefixes: []string{"/api"},
			path:     "/api/users",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "case insensitive match",
			prefixes: []string{"/API"},
			path:     "/api/users",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "case insensitive path",
			prefixes: []string{"/api"},
			path:     "/API/users",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "matches one of multiple prefixes",
			prefixes: []string{"/api", "/admin", "/public"},
			path:     "/admin/settings",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "no prefix match",
			prefixes: []string{"/api", "/admin"},
			path:     "/public/page",
			method:   http.MethodGet,
			want:     false,
		},
		{
			name:     "exact match with prefix",
			prefixes: []string{"/api"},
			path:     "/api",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "partial path after prefix",
			prefixes: []string{"/api/v1"},
			path:     "/api/v1/users",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "empty prefixes list",
			prefixes: []string{},
			path:     "/api/users",
			method:   http.MethodGet,
			want:     false,
		},
		{
			name:     "prefix with method specifier",
			prefixes: []string{"GET /api", "POST /api"},
			path:     "/api/users",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "prefix with wrong method",
			prefixes: []string{"GET /api"},
			path:     "/api/users",
			method:   http.MethodPost,
			want:     false,
		},
		{
			name:     "case insensitive method",
			prefixes: []string{"get /api"},
			path:     "/api/users",
			method:   http.MethodGet,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper := PrefixPathSkipper(tt.prefixes...)
			req := createRequest(tt.method, tt.path)
			got := skipper(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSuffixPathSkipper(t *testing.T) {
	tests := []struct {
		name     string
		suffixes []string
		path     string
		method   string
		want     bool
	}{
		{
			name:     "matches single suffix",
			suffixes: []string{".json"},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "case insensitive match",
			suffixes: []string{".JSON"},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "case insensitive path",
			suffixes: []string{".json"},
			path:     "/api/data.JSON",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "matches one of multiple suffixes",
			suffixes: []string{".json", ".xml", ".html"},
			path:     "/api/data.xml",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "no suffix match",
			suffixes: []string{".json", ".xml"},
			path:     "/api/data.txt",
			method:   http.MethodGet,
			want:     false,
		},
		{
			name:     "exact match with suffix",
			suffixes: []string{".json"},
			path:     "/.json",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "path before suffix",
			suffixes: []string{"/data.json"},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "empty suffixes list",
			suffixes: []string{},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     false,
		},
		{
			name:     "suffix with method specifier",
			suffixes: []string{"GET .json", "POST .json"},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     true,
		},
		{
			name:     "suffix with wrong method",
			suffixes: []string{"GET .json"},
			path:     "/api/data.json",
			method:   http.MethodPost,
			want:     false,
		},
		{
			name:     "case insensitive method",
			suffixes: []string{"get .json"},
			path:     "/api/data.json",
			method:   http.MethodGet,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper := SuffixPathSkipper(tt.suffixes...)
			req := createRequest(tt.method, tt.path)
			got := skipper(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEqualPathSkipper(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		path   string
		method string
		want   bool
	}{
		{
			name:   "matches single path",
			paths:  []string{"/health"},
			path:   "/health",
			method: http.MethodGet,
			want:   true,
		},
		{
			name:   "case insensitive match",
			paths:  []string{"/HEALTH"},
			path:   "/health",
			method: http.MethodGet,
			want:   true,
		},
		{
			name:   "case insensitive path",
			paths:  []string{"/health"},
			path:   "/HEALTH",
			method: http.MethodGet,
			want:   true,
		},
		{
			name:   "matches one of multiple paths",
			paths:  []string{"/health", "/metrics", "/ready"},
			path:   "/metrics",
			method: http.MethodGet,
			want:   true,
		},
		{
			name:   "no path match",
			paths:  []string{"/health", "/metrics"},
			path:   "/status",
			method: http.MethodGet,
			want:   false,
		},
		{
			name:   "prefix does not match",
			paths:  []string{"/api"},
			path:   "/api/users",
			method: http.MethodGet,
			want:   false,
		},
		{
			name:   "suffix does not match",
			paths:  []string{".json"},
			path:   "/data.json",
			method: http.MethodGet,
			want:   false,
		},
		{
			name:   "empty paths list",
			paths:  []string{},
			path:   "/health",
			method: http.MethodGet,
			want:   false,
		},
		{
			name:   "path with method specifier",
			paths:  []string{"GET /health", "POST /health"},
			path:   "/health",
			method: http.MethodGet,
			want:   true,
		},
		{
			name:   "path with wrong method",
			paths:  []string{"GET /health"},
			path:   "/health",
			method: http.MethodPost,
			want:   false,
		},
		{
			name:   "case insensitive method doesn't work for EqualPathSkipper",
			paths:  []string{"get /health"},
			path:   "/health",
			method: http.MethodGet,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipper := EqualPathSkipper(tt.paths...)
			req := createRequest(tt.method, tt.path)
			got := skipper(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPageSkipper(t *testing.T) {
	t.Run("panics with nil decorator strategy", func(t *testing.T) {
		assert.Panics(t, func() {
			PageSkipper(nil)
		})
	})

	tests := []struct {
		name               string
		method             string
		path               string
		pattern            string
		isPatternDecorable bool
		isURIDecorable     bool
		want               bool
	}{
		{
			name:               "non-GET method always skips",
			method:             http.MethodPost,
			path:               "/page",
			pattern:            "/page",
			isPatternDecorable: true,
			isURIDecorable:     true,
			want:               true,
		},
		{
			name:               "PUT method skips",
			method:             http.MethodPut,
			path:               "/page",
			pattern:            "/page",
			isPatternDecorable: true,
			isURIDecorable:     true,
			want:               true,
		},
		{
			name:               "DELETE method skips",
			method:             http.MethodDelete,
			path:               "/page",
			pattern:            "/page",
			isPatternDecorable: true,
			isURIDecorable:     true,
			want:               true,
		},
		{
			name:               "GET method with decorable pattern",
			method:             http.MethodGet,
			path:               "/page",
			pattern:            "/page",
			isPatternDecorable: true,
			isURIDecorable:     true,
			want:               false,
		},
		{
			name:               "GET method with non-decorable pattern",
			method:             http.MethodGet,
			path:               "/page",
			pattern:            "/page",
			isPatternDecorable: false,
			isURIDecorable:     true,
			want:               true,
		},
		{
			name:               "GET with PageCMSPattern calls IsURIDecorable",
			method:             http.MethodGet,
			path:               "/some/page",
			pattern:            PageCMSPattern,
			isPatternDecorable: false,
			isURIDecorable:     false,
			want:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStrategy := NewMockPageDecoratorStrategy(tt.isPatternDecorable)
			skipper := PageSkipper(mockStrategy)
			req := createRequestWithMethodPattern(tt.method, tt.path, tt.pattern)
			got := skipper(req)
			assert.Equal(t, tt.want, got)

			if tt.method == http.MethodGet {
				if tt.pattern == PageCMSPattern {
					mockStrategy.AssertCalled(t, "IsURIDecorable", mock.Anything, tt.path)
				} else {
					mockStrategy.AssertCalled(t, "IsPatternDecorable", mock.Anything, tt.pattern)
				}
			} else {
				mockStrategy.AssertNotCalled(t, "IsPatternDecorable", mock.Anything, mock.Anything)
				mockStrategy.AssertNotCalled(t, "IsURIDecorable", mock.Anything, mock.Anything)
			}
		})
	}
}

func createRequest(method, path string) *http.Request {
	return &http.Request{
		Method: method,
		URL:    &url.URL{Path: path},
	}
}

func createRequestWithMethodPattern(method, path, pattern string) *http.Request {
	return &http.Request{
		Method:  method,
		URL:     &url.URL{Path: path},
		Pattern: pattern,
	}
}
