package pages

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func createRequestWithMethodPattern(method, path, pattern string) *http.Request {
	return &http.Request{
		Method:  method,
		URL:     &url.URL{Path: path},
		Pattern: pattern,
	}
}
