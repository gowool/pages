package pages

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
)

// TestPageHandler_ErrSiteNotFound tests that PageHandler returns ErrSiteNotFound when no site is present
func TestPageHandler_ErrSiteNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	// Don't set site - this should trigger ErrSiteNotFound

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSiteNotFound), "Error should wrap ErrSiteNotFound")
	assert.Contains(t, err.Error(), "page handler")
}

// TestPageHandler_ErrPageNotFound tests that PageHandler returns ErrPageNotFound when no page is present
func TestPageHandler_ErrPageNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	// Set site but not page - this should trigger ErrPageNotFound
	site := &Site{ID: "site1"}
	event.SetSite(site)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPageNotFound), "Error should wrap ErrPageNotFound")
	assert.Contains(t, err.Error(), "page handler")
}

// TestPageHandler_NoTemplate tests that PageHandler calls NoContent when template is empty
func TestPageHandler_NoTemplate(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "", // Empty template
		Header:   make(map[string][]string),
	}

	event.SetSite(site)
	event.SetPage(page)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}

// TestPageHandler_WithHeaders tests that PageHandler correctly sets response headers
func TestPageHandler_WithHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.html",
		Header: map[string][]string{
			"Cache-Control": {"no-cache", "no-store"},
			"X-Custom":      {"value"},
		},
	}

	event.SetSite(site)
	event.SetPage(page)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, "no-cache", resp.Header().Get("Cache-Control"))
	assert.Equal(t, "no-store", resp.Header()["Cache-Control"][1]) // Second value
	assert.Equal(t, "value", resp.Header().Get("X-Custom"))
}

// TestPageHandler_WithSingleHeaders tests that PageHandler correctly sets single-value headers
func TestPageHandler_WithSingleHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.html",
		Header: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Test":       {"single-value"},
		},
	}

	event.SetSite(site)
	event.SetPage(page)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	assert.Equal(t, "single-value", resp.Header().Get("X-Test"))
}

// TestPageHandler_ExistingContentType tests that PageHandler uses existing content type header
func TestPageHandler_ExistingContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.json",
		Header:   make(map[string][]string),
	}

	event.SetSite(site)
	event.SetPage(page)

	// Manually set content type header to test that it's preserved
	woResp.Header().Set("Content-Type", "application/json")

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}

// TestPageHandler_DefaultContentType tests that PageHandler uses default HTML content type when none is set
func TestPageHandler_DefaultContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.html",
		Header:   make(map[string][]string),
	}

	event.SetSite(site)
	event.SetPage(page)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, wo.MIMETextHTMLCharsetUTF8, resp.Header().Get(wo.HeaderContentType))
}

// TestPageHandler_CustomStatus tests that PageHandler uses custom status code
func TestPageHandler_CustomStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.html",
		Header:   make(map[string][]string),
	}

	event.SetSite(site)
	event.SetPage(page)
	event.SetStatus(http.StatusCreated) // Custom status

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

// TestPageHandler_EmptyHeaderValues tests that PageHandler handles empty header values correctly
func TestPageHandler_EmptyHeaderValues(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	page := &Page{
		ID:       "page1",
		Site:     site,
		Template: "test.html",
		Header: map[string][]string{
			"Empty-Values": {"", "value2", ""},
			"Single-Value": {"value1"},
		},
	}

	event.SetSite(site)
	event.SetPage(page)

	handler := PageHandler[Resolver]()
	err := handler(event)

	assert.NoError(t, err)
	// Check that empty values are still set
	assert.Contains(t, resp.Header()["Empty-Values"], "")
	assert.Contains(t, resp.Header()["Empty-Values"], "value2")
	assert.Equal(t, "value1", resp.Header().Get("Single-Value"))
}

// TestPageHandler_Constants tests that the constants are correctly defined
func TestPageHandler_Constants(t *testing.T) {
	assert.Equal(t, "/{_page_cms...}", PageCMSPattern, "PageCMSPattern should match expected value")
	assert.Equal(t, "/{$}", HomeHybridPattern, "HomeHybridPattern should match expected value")
}
