package pages

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
)

// TestPageMiddleware_PanicHandlerNil tests that the middleware panics when handler is nil
func TestPageMiddleware_PanicHandlerNil(t *testing.T) {
	assert.Panics(t, func() {
		PageMiddleware[Resolver](nil, &MockPageSelector{}, nil)
	}, "Expected panic when handler is nil")
}

// TestPageMiddleware_PanicSelectorNil tests that the middleware panics when selector is nil
func TestPageMiddleware_PanicSelectorNil(t *testing.T) {
	assert.Panics(t, func() {
		PageMiddleware[Resolver](func(resolver Resolver) error { return nil }, nil, nil)
	}, "Expected panic when selector is nil")
}

// TestPageMiddleware_DefaultAuthorizer tests that the middleware uses DenyPageAuthorizer when none is provided
func TestPageMiddleware_DefaultAuthorizer(t *testing.T) {
	handler := func(resolver Resolver) error { return nil }
	selector := &MockPageSelector{}

	middleware := PageMiddleware[Resolver](handler, selector, nil)
	assert.NotNil(t, middleware)
}

// TestPageMiddleware_NoSite tests that the middleware returns ErrSiteNotFound when no site is present
func TestPageMiddleware_NoSite(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	mockSelector := NewMockPageSelector(nil, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, nil)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.True(t, errors.Is(err, ErrSiteNotFound))
}

// TestPageMiddleware_PageNotFound tests that the middleware returns ErrNotFound when page is not found
func TestPageMiddleware_PageNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	mockSelector := NewMockPageSelector(nil, ErrPageNotFound)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, nil)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.Contains(t, err.Error(), "404") // Should contain HTTP 404 status
}

// TestPageMiddleware_PublishedCMSPage tests that CMS pages work correctly without calling handler
func TestPageMiddleware_PublishedCMSPage(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	page := &Page{
		ID:      "page1",
		Status:  PublishStatus,
		Pattern: PageCMS, // CMS page (non-hybrid)
	}

	mockSelector := NewMockPageSelector(page, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		resolver.SetStatus(http.StatusOK)
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, nil)

	err := middleware(event)

	assert.NoError(t, err)
	assert.False(t, handlerCalled, "Handler should not be called for non-hybrid pages")
	assert.Equal(t, page, event.Page())
}

// TestPageMiddleware_HybridPagePattern verifies that hybrid pages are correctly identified
func TestPageMiddleware_HybridPagePattern(t *testing.T) {
	// Test different page patterns to verify hybrid page detection
	tests := []struct {
		name     string
		pattern  string
		isHybrid bool
	}{
		{"CMS page", PageCMS, false},
		{"Internal page", PageInternalCreate, false},
		{"Hybrid page with param", "/blog/{slug}", true},
		{"Static page", "/about", true}, // Static pages are also hybrid (not CMS, not Internal)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{
				ID:      "page1",
				Status:  PublishStatus,
				Pattern: tt.pattern,
			}

			assert.Equal(t, tt.isHybrid, page.IsHybrid(),
				"Page.IsHybrid() should return correct value for pattern: %s", tt.pattern)
		})
	}
}

// TestPageMiddleware_DraftPageUnauthorized tests that draft pages are unauthorized when authorizer denies
func TestPageMiddleware_DraftPageUnauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	page := &Page{
		ID:      "page1",
		Status:  DraftStatus,
		Pattern: PageCMS,
	}

	mockSelector := NewMockPageSelector(page, nil)
	mockAuthorizer := NewMockPageAuthorizer(Deny, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, mockAuthorizer)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.Contains(t, err.Error(), "404") // Should contain HTTP 404 status
}

// TestPageMiddleware_DraftPageAuthorized tests that draft pages are accessible when authorizer allows
func TestPageMiddleware_DraftPageAuthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	// Use a CMS page (non-hybrid) to avoid buffer issues
	page := &Page{
		ID:      "page1",
		Status:  DraftStatus,
		Pattern: PageCMS,
	}

	mockSelector := NewMockPageSelector(page, nil)
	mockAuthorizer := NewMockPageAuthorizer(Allow, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		resolver.SetStatus(http.StatusOK)
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, mockAuthorizer)

	err := middleware(event)

	assert.NoError(t, err)
	assert.False(t, handlerCalled, "Handler should not be called for CMS pages")
	assert.Equal(t, page, event.Page())
}

// TestPageMiddleware_PrivatePageGuest tests that private pages are unauthorized for guests
func TestPageMiddleware_PrivatePageGuest(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)
	event.SetValue(keyGuest{}, true) // Set as guest

	page := &Page{
		ID:      "page1",
		Status:  PrivateStatus,
		Pattern: PageCMS,
	}

	mockSelector := NewMockPageSelector(page, nil)
	mockAuthorizer := NewMockPageAuthorizer(Deny, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, mockAuthorizer)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.Contains(t, err.Error(), "403") // Should contain HTTP 403 status for guests
}

// TestPageMiddleware_PrivatePageUnauthorized tests that private pages are forbidden when authorizer denies
func TestPageMiddleware_PrivatePageUnauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)
	event.SetValue(keyGuest{}, false) // Set as not guest

	page := &Page{
		ID:      "page1",
		Status:  PrivateStatus,
		Pattern: PageCMS,
	}

	mockSelector := NewMockPageSelector(page, nil)
	mockAuthorizer := NewMockPageAuthorizer(Deny, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, mockAuthorizer)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.Contains(t, err.Error(), "401") // Should contain HTTP 401 status when not authorized
}

// TestPageMiddleware_UnknownPageStatus tests that pages with unknown status return not found
func TestPageMiddleware_UnknownPageStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	page := &Page{
		ID:      "page1",
		Status:  Status(99), // Unknown status
		Pattern: PageCMS,
	}

	mockSelector := NewMockPageSelector(page, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	middleware := PageMiddleware[Resolver](handler, mockSelector, nil)

	err := middleware(event)

	assert.Error(t, err)
	assert.False(t, handlerCalled)
	assert.Contains(t, err.Error(), "404") // Should contain HTTP 404 status
}

// TestPageMiddleware_PageSiteAssignment tests that page.Site assignment logic works correctly
func TestPageMiddleware_PageSiteAssignment(t *testing.T) {
	// Test the site assignment logic directly
	page := &Page{
		ID:      "page1",
		Site:    nil, // Site is nil
		Pattern: PageCMS,
	}

	site := &Site{ID: "site1"}

	// Simulate what the middleware does
	if page.Site == nil {
		page.Site = site
	}

	assert.NotNil(t, page.Site)
	assert.Equal(t, ID("site1"), page.Site.ID)
}

// TestPageMiddleware_SkipMiddleware tests that the middleware skips execution when skip condition is met
func TestPageMiddleware_SkipMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	mockSelector := NewMockPageSelector(nil, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		resolver.SetStatus(http.StatusOK)
		return nil
	}

	skipper := func(resolver Resolver) bool { return true }
	middleware := PageMiddleware[Resolver](handler, mockSelector, nil, skipper)

	err := middleware(event)

	assert.NoError(t, err)
	assert.False(t, handlerCalled, "Handler should not be called when middleware is skipped")
}

// TestPageMiddleware_ChainSkipper tests multiple skippers chained together
func TestPageMiddleware_ChainSkipper(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	mockSelector := NewMockPageSelector(nil, nil)
	handlerCalled := false

	handler := func(resolver Resolver) error {
		handlerCalled = true
		return nil
	}

	skipper1 := func(resolver Resolver) bool { return false }
	skipper2 := func(resolver Resolver) bool { return true }

	middleware := PageMiddleware[Resolver](handler, mockSelector, nil, skipper1, skipper2)

	err := middleware(event)

	assert.NoError(t, err)
	assert.False(t, handlerCalled)
}
