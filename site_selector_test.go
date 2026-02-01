package pages

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultSiteSelector(t *testing.T) {
	t.Run("valid retriever", func(t *testing.T) {
		mockRetriever := &MockSiteRetriever{}
		selector := NewDefaultSiteSelector(mockRetriever)

		assert.NotNil(t, selector, "Selector should not be nil")
		assert.Same(t, mockRetriever, selector.retriever, "Retriever should be set")
	})

	t.Run("nil retriever", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultSiteSelector(nil)
		}, "NewDefaultSiteSelector should panic with nil retriever")
	})
}

func TestDefaultSiteSelector_Select(t *testing.T) {
	t.Run("context is nil", func(t *testing.T) {
		mockRetriever := &MockSiteRetriever{}
		selector := NewDefaultSiteSelector(mockRetriever)

		req := &http.Request{}

		assert.Panics(t, func() {
			_ = selector.Select(req)
		}, "Select should panic when context is nil")
	})

	t.Run("site not found error", func(t *testing.T) {
		mockRetriever := NewMockSiteRetriever(nil, "", ErrSiteNotFound)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error")
		assert.ErrorIs(t, err, ErrSiteNotFound, "Error should be ErrSiteNotFound")
	})

	t.Run("site retrieval error", func(t *testing.T) {
		mockRetriever := NewMockSiteRetriever(nil, "", assert.AnError)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error")
		assert.ErrorIs(t, err, assert.AnError, "Error should wrap assert.AnError")
		assert.ErrorIs(t, err, ErrSiteNotFound, "Error should wrap ErrSiteNotFound")
	})

	t.Run("site returned is nil", func(t *testing.T) {
		mockRetriever := NewMockSiteRetriever(nil, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error when site is nil")
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("site found without pathInfo", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"
		site.Host = "localhost"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err, "Select should succeed")
		assert.True(t, c.HasSite(), "Context should have site")
		assert.Equal(t, site, c.Site(), "Site should be set in context")
		assert.Equal(t, req.Host, site.Host, "Host should be updated from request")
		assert.Equal(t, "http", site.Scheme, "Scheme should default to http")
		assert.Equal(t, "/test/path", req.URL.Path, "Path should not be modified when pathInfo is empty")
	})

	t.Run("site found with pathInfo", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"
		site.Host = "localhost"

		pathInfo := "/adjusted/path"
		mockRetriever := NewMockSiteRetriever(site, pathInfo, nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		defer cancel()

		req := createTestRequest()
		originalPath := req.URL.Path
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err, "Select should succeed")
		assert.Equal(t, pathInfo, req.URL.Path, "Request URL path should be updated")
		assert.NotEqual(t, originalPath, req.URL.Path, "Path should be different from original")
		assert.Equal(t, site, c.Site())
	})

	t.Run("site with nil Host in request", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		defer cancel()

		req := createTestRequest()
		req.Host = ""
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err, "Select should succeed with empty Host")
		assert.Equal(t, "", site.Host, "Site Host should be set to empty string")
		assert.Equal(t, site, c.Site())
	})
}

func TestDefaultSiteSelector_Select_SchemeDetection(t *testing.T) {
	t.Run("TLS connection", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.TLS = &tls.ConnectionState{}
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "Scheme should be https for TLS connection")
	})

	t.Run("X-Forwarded-Proto header", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXForwardedProto, "https")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "Scheme should use X-Forwarded-Proto header")
	})

	t.Run("X-Forwarded-Protocol header", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXForwardedProtocol, "https")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "Scheme should use X-Forwarded-Protocol header")
	})

	t.Run("X-Forwarded-Ssl header with on", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXForwardedSsl, "on")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "Scheme should be https when X-Forwarded-Ssl is on")
	})

	t.Run("X-Forwarded-Ssl header with off", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXForwardedSsl, "off")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "http", site.Scheme, "Scheme should be http when X-Forwarded-Ssl is off")
	})

	t.Run("X-Url-Scheme header", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXUrlScheme, "https")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "Scheme should use X-Url-Scheme header")
	})

	t.Run("default to http", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "http", site.Scheme, "Scheme should default to http")
	})

	t.Run("TLS takes precedence over headers", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.TLS = &tls.ConnectionState{}
		req.Header.Set(headerXForwardedProto, "http")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "TLS should take precedence over headers")
	})

	t.Run("X-Forwarded-Proto takes precedence over X-Url-Scheme", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := createTestRequest()
		req.Header.Set(headerXForwardedProto, "https")
		req.Header.Set(headerXUrlScheme, "http")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, "https", site.Scheme, "X-Forwarded-Proto should take precedence")
	})
}

func TestGetScheme(t *testing.T) {
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
			req := createTestRequest()
			tt.setupReq(req)

			got := getScheme(req)

			assert.Equal(t, tt.want, got, "getScheme should return expected scheme")
		})
	}
}

func TestDefaultSiteSelector_Select_Integration(t *testing.T) {
	t.Run("full flow with proxy headers", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"
		site.Host = "localhost"

		pathInfo := "/api/v1/adjusted"

		mockRetriever := NewMockSiteRetriever(site, pathInfo, nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		defer cancel()

		req := createTestRequest()
		req.Host = "example.com"
		req.Header.Set(headerXForwardedProto, "https")
		req.Header.Set(headerXForwardedSsl, "on")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		require.NoError(t, err, "Select should succeed")
		assert.Equal(t, site, c.Site(), "Site should be set")
		assert.Equal(t, "example.com", site.Host, "Host should be from request")
		assert.Equal(t, "https", site.Scheme, "Scheme should be https")
		assert.Equal(t, pathInfo, req.URL.Path, "Path should be adjusted")
	})

	t.Run("verify context receives site", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		mockRetriever := NewMockSiteRetriever(site, "", nil)
		selector := NewDefaultSiteSelector(mockRetriever)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		defer cancel()

		req := createTestRequest()
		req = req.WithContext(ctx)

		err := selector.Select(req)

		require.NoError(t, err)
		assert.True(t, c.HasSite())
		assert.Same(t, site, c.Site())
	})
}

func createTestRequest() *http.Request {
	return &http.Request{
		URL:    &url.URL{Path: "/test/path"},
		Host:   "localhost",
		Header: make(http.Header),
	}
}
