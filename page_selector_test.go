package pages

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNewDefaultPageSelector(t *testing.T) {
	t.Run("valid retriever and patternArgs", func(t *testing.T) {
		mockRetriever := &MockPageRetriever{}
		patternArgs := PatternArgs()

		selector := NewDefaultPageSelector(mockRetriever, patternArgs)

		assert.NotNil(t, selector, "Selector should not be nil")
		assert.Same(t, mockRetriever, selector.retriever, "Retriever should be set")
		assert.NotNil(t, selector.patternArgs, "PatternArgs should be set")
	})

	t.Run("valid retriever with nil patternArgs", func(t *testing.T) {
		mockRetriever := &MockPageRetriever{}

		selector := NewDefaultPageSelector(mockRetriever, nil)

		assert.NotNil(t, selector)
		assert.Same(t, mockRetriever, selector.retriever)
		assert.NotNil(t, selector.patternArgs, "Default PatternArgs should be set")
	})

	t.Run("nil retriever", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSelector(nil, PatternArgs())
		}, "NewDefaultPageSelector should panic with nil retriever")
	})
}

func TestDefaultPageSelector_Select(t *testing.T) {
	t.Run("context is nil", func(t *testing.T) {
		mockRetriever := &MockPageRetriever{}
		selector := NewDefaultPageSelector(mockRetriever, nil)

		req := &http.Request{}

		assert.Panics(t, func() {
			_ = selector.Select(req)
		}, "Select should panic when context is nil")
	})

	t.Run("context without site", func(t *testing.T) {
		mockRetriever := &MockPageRetriever{}
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		defer cancel()

		req := &http.Request{}
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error when no site")
		assert.ErrorIs(t, err, ErrSiteNotFound, "Error should be ErrSiteNotFound")
		assert.Contains(t, err.Error(), "page selector", "Error should have prefix")
	})

	t.Run("page not found", func(t *testing.T) {
		mockRetriever := NewMockPageSelector(nil, ErrPageNotFound)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(NewSite())
		defer cancel()

		req := createRequestWithPattern("/test", "/test")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error when page not found")
		assert.ErrorIs(t, err, ErrPageNotFound, "Error should be ErrPageNotFound")
		assert.Contains(t, err.Error(), "page selector")
	})

	t.Run("page retrieval error", func(t *testing.T) {
		testErr := &testError{msg: "database error"}
		mockRetriever := NewMockPageSelector(nil, testErr)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(NewSite())
		defer cancel()

		req := createRequestWithPattern("/test", "/test")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error on retrieval failure")
		assert.ErrorIs(t, err, ErrPageNotFound, "Error should wrap ErrPageNotFound")
		assert.ErrorIs(t, err, testErr, "Error should contain original error")
		assert.Contains(t, err.Error(), "page selector")
	})

	t.Run("page returned is nil", func(t *testing.T) {
		mockRetriever := NewMockPageSelector(nil, nil)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(NewSite())
		defer cancel()

		req := createRequestWithPattern("/test", "/test")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.Error(t, err, "Select should return error when page is nil")
		assert.ErrorIs(t, err, ErrPageNotFound)
		assert.Contains(t, err.Error(), "page selector")
	})

	t.Run("static page found", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/about"
		page.Site = site

		mockRetriever := NewMockPageSelector(page, nil)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(site)
		defer cancel()

		req := createRequestWithPattern("/about", "/about")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err, "Select should succeed for static page")
		assert.True(t, c.HasPage(), "Context should have page")
		assert.Equal(t, page, c.Page(), "Page should be set in context")
	})

	t.Run("dynamic page found with pattern args", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/posts/{id}"

		mockRetriever := NewMockPageSelector(page, nil)
		selector := NewDefaultPageSelector(mockRetriever, PatternArgs())

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(site)
		defer cancel()

		req := createRequestWithPattern("/posts/{id}", "/posts/123")
		req.SetPathValue("id", "123")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err, "Select should succeed for dynamic page")
		assert.True(t, c.HasPage(), "Context should have page")
		assert.Equal(t, page, c.Page())
	})

	t.Run("page with nil site", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/test"

		mockRetriever := NewMockPageSelector(page, nil)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(site)
		defer cancel()

		req := createRequestWithPattern("/test", "/test")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, site, page.Site, "Site should be set on page")
		assert.Equal(t, site.ID, page.SiteID, "SiteID should be set on page")
	})

	t.Run("page with site already set", func(t *testing.T) {
		siteInContext := NewSite()
		siteInContext.ID = "site1"

		siteOnPage := NewSite()
		siteOnPage.ID = "site2"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/test"
		page.Site = siteOnPage
		page.SiteID = siteOnPage.ID

		mockRetriever := NewMockPageSelector(page, nil)
		selector := NewDefaultPageSelector(mockRetriever, nil)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(siteInContext)
		defer cancel()

		req := createRequestWithPattern("/test", "/test")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.Equal(t, siteOnPage, page.Site, "Existing site should be preserved")
		assert.Equal(t, siteOnPage.ID, page.SiteID, "SiteID should be preserved")
	})

	t.Run("custom pattern args function", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/posts/{id}"

		mockRetriever := NewMockPageSelector(page, nil)

		customArgsCalled := false
		customPatternArgs := func(r *http.Request) []any {
			customArgsCalled = true
			return []any{"{id}", "custom-value"}
		}

		selector := NewDefaultPageSelector(mockRetriever, customPatternArgs)

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(site)
		defer cancel()

		req := createRequestWithPattern("/posts/{id}", "/posts/123")
		req.SetPathValue("id", "123")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.True(t, customArgsCalled, "Custom pattern args should be called")
	})
}

func TestDefaultPageSelector_Select_Integration(t *testing.T) {
	t.Run("full flow with dynamic page", func(t *testing.T) {
		site := NewSite()
		site.ID = "site1"
		site.Host = "example.com"

		page := NewPage()
		page.ID = "page1"
		page.Pattern = "/posts/{year}/{month}/{slug}"
		page.Name = "Post Detail"

		mockRetriever := NewMockPageSelector(page, nil)
		selector := NewDefaultPageSelector(mockRetriever, PatternArgs())

		ctx, cancel := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(site)
		defer cancel()

		req := createRequestWithPattern("/posts/{year}/{month}/{slug}", "/posts/2024/01/my-post")
		req.SetPathValue("year", "2024")
		req.SetPathValue("month", "01")
		req.SetPathValue("slug", "my-post")
		req = req.WithContext(ctx)

		err := selector.Select(req)

		assert.NoError(t, err)
		assert.True(t, c.HasPage())
		assert.Equal(t, page, c.Page())
		assert.Equal(t, site, page.Site)
		assert.Equal(t, site.ID, page.SiteID)
	})
}

func createRequestWithPattern(pattern, path string) *http.Request {
	req := &http.Request{
		Pattern: pattern,
		URL:     &url.URL{Path: path},
	}
	return req
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
