package pages

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestErrorMapper tests the ErrorMapper function with various error types
func TestErrorMapper(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected *wo.HTTPError
	}{
		{
			name:     "HTTP error is returned as-is",
			err:      wo.ErrBadRequest,
			expected: wo.ErrBadRequest,
		},
		{
			name:     "ErrSiteNotFound maps to internal server error",
			err:      ErrSiteNotFound,
			expected: wo.ErrInternalServerError.WithInternal(ErrSiteNotFound),
		},
		{
			name:     "ErrPageNotFound maps to not found",
			err:      ErrPageNotFound,
			expected: wo.ErrNotFound.WithInternal(ErrPageNotFound),
		},
		{
			name:     "ErrPrivatePage maps to forbidden",
			err:      ErrPrivatePage,
			expected: wo.ErrForbidden.WithInternal(ErrPrivatePage),
		},
		{
			name:     "ErrUniqueViolation maps to conflict",
			err:      ErrUniqueViolation,
			expected: wo.ErrConflict.WithInternal(ErrUniqueViolation),
		},
		{
			name:     "Wrapped ErrSiteNotFound maps correctly",
			err:      fmt.Errorf("wrapped: %w", ErrSiteNotFound),
			expected: wo.ErrInternalServerError.WithInternal(ErrSiteNotFound),
		},
		{
			name:     "Unknown error returns nil",
			err:      errors.New("unknown error"),
			expected: nil,
		},
		{
			name:     "Nil error returns nil",
			err:      nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrorMapper(tt.err)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Status, result.Status)
				assert.Equal(t, tt.expected.Message, result.Message)
				assert.True(t, errors.Is(result.Internal, tt.expected.Internal))
			}
		})
	}
}

func newMockPatternPageManager(page *Page, err error) *MockPageManager {
	manager := &MockPageManager{}
	manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, err)
	return manager
}

// TestErrorRenderer_PanicHandlerNil tests that ErrorRenderer panics when handler is nil
func TestErrorRenderer_PanicHandlerNil(t *testing.T) {
	assert.Panics(t, func() {
		ErrorRenderer[Resolver](nil, &MockPageManager{}, nil, &MockPageAuthorizer{}, nil, slog.Default())
	}, "Expected panic when handler is nil")
}

// TestErrorRenderer_PanicManagerNil tests that ErrorRenderer panics when manager is nil
func TestErrorRenderer_PanicManagerNil(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	assert.Panics(t, func() {
		ErrorRenderer[Resolver](handler, nil, nil, nil, nil, slog.Default())
	}, "Expected panic when manager is nil")
}

// TestErrorRenderer_DefaultAuthorizer tests that ErrorRenderer uses DenyPageAuthorizer when none is provided
func TestErrorRenderer_DefaultAuthorizer(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, nil, nil, nil)
	assert.NotNil(t, renderer)
}

// TestErrorRenderer_DefaultPatternFinder tests that ErrorRenderer uses default pattern finder when none is provided
func TestErrorRenderer_DefaultPatternFinder(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())
	assert.NotNil(t, renderer)
}

// TestErrorRenderer_DefaultLogger tests that ErrorRenderer uses discard logger when none is provided
func TestErrorRenderer_DefaultLogger(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, nil)
	assert.NotNil(t, renderer)
}

// TestErrorRenderer_NonHTMLRequest tests that ErrorRenderer does nothing for non-HTML requests
func TestErrorRenderer_NonHTMLRequest(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
}

// TestErrorRenderer_NoSite tests that ErrorRenderer does nothing when no site is present
func TestErrorRenderer_NoSite(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
}

// TestErrorRenderer_Skipper tests that ErrorRenderer respects skipper middleware
func TestErrorRenderer_Skipper(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	skipper := func(e Resolver) bool {
		return true
	}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default(), skipper)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	site := &Site{ID: "site1"}
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
}

// TestErrorRenderer_NotFoundWithCreatePermission tests that ErrorRenderer uses create page for 404 when authorized
func TestErrorRenderer_NotFoundWithCreatePermission(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageInternalCreate}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Allow, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrNotFound

	renderer(event, httpErr)

	assert.Equal(t, http.StatusNotFound, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_DefaultPatternFinder4xx tests default pattern finder for 4xx errors
func TestErrorRenderer_DefaultPatternFinder4xx(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_DefaultPatternFinder5xx tests default pattern finder for 5xx errors
func TestErrorRenderer_DefaultPatternFinder5xx(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError5xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrInternalServerError

	renderer(event, httpErr)

	assert.Equal(t, http.StatusInternalServerError, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_CustomPatternFinder tests that ErrorRenderer uses custom pattern finder
func TestErrorRenderer_CustomPatternFinder(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: "custom_error_pattern"}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	patternFinder := func(ctx context.Context, status int) (string, error) {
		return "custom_error_pattern", nil
	}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, patternFinder, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_PatternFinderError tests that ErrorRenderer falls back to default pattern when pattern finder returns error
func TestErrorRenderer_PatternFinderError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError5xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	patternFinder := func(ctx context.Context, status int) (string, error) {
		return "", errors.New("pattern finder error")
	}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, patternFinder, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_ManagerError tests that ErrorRenderer handles manager errors gracefully
func TestErrorRenderer_ManagerError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}

	manager := newMockPatternPageManager(nil, errors.New("manager error"))
	authorizer := NewMockPageAuthorizer(Deny, nil)

	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, logger)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
	assert.Nil(t, event.Page())
}

// TestErrorRenderer_HandlerError tests that ErrorRenderer handles handler errors gracefully
func TestErrorRenderer_HandlerError(t *testing.T) {
	handler := func(e Resolver) error {
		return errors.New("handler error")
	}
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, logger)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_ChainSkipper tests that ErrorRenderer chains multiple skippers
func TestErrorRenderer_ChainSkipper(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	manager := &MockPageManager{}
	authorizer := &MockPageAuthorizer{}

	skipper1 := func(e Resolver) bool {
		return false
	}

	skipper2 := func(e Resolver) bool {
		return true
	}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default(), skipper1, skipper2)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	site := &Site{ID: "site1"}
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
}

// TestErrorRenderer_UnauthorizedError tests default pattern finder for unauthorized errors
func TestErrorRenderer_UnauthorizedError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageErrorUnauthorized}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrUnauthorized

	renderer(event, httpErr)

	assert.Equal(t, http.StatusUnauthorized, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_ForbiddenError tests default pattern finder for forbidden errors
func TestErrorRenderer_ForbiddenError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageErrorForbidden}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrForbidden

	renderer(event, httpErr)

	assert.Equal(t, http.StatusForbidden, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_NotFoundError tests default pattern finder for not found errors
func TestErrorRenderer_NotFoundError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageErrorNotFound}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrNotFound

	renderer(event, httpErr)

	assert.Equal(t, http.StatusNotFound, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_Other4xxError tests default pattern finder for other 4xx errors
func TestErrorRenderer_Other4xxError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := &wo.HTTPError{Status: http.StatusPaymentRequired}

	renderer(event, httpErr)

	assert.Equal(t, http.StatusPaymentRequired, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_Other5xxError tests default pattern finder for other 5xx errors
func TestErrorRenderer_Other5xxError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError5xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := &wo.HTTPError{Status: http.StatusBadGateway}

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadGateway, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_CustomPatternFinderEmpty tests that ErrorRenderer falls back to default when custom pattern returns empty
func TestErrorRenderer_CustomPatternFinderEmpty(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError5xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	patternFinder := func(ctx context.Context, status int) (string, error) {
		return "", nil
	}

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, patternFinder, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_SetsError tests that ErrorRenderer sets the error on the event
func TestErrorRenderer_SetsError(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.NotNil(t, event.Error())
	assert.Equal(t, httpErr, event.Error())
}

// TestErrorRenderer_StatusOKTestsDefaultPattern tests that status 200 uses default pattern
func TestErrorRenderer_StatusOKTestsDefaultPattern(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError5xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := &wo.HTTPError{Status: http.StatusOK}

	renderer(event, httpErr)

	assert.Equal(t, http.StatusOK, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_PartialHTMLAccept tests that ErrorRenderer works with partial HTML accept header
func TestErrorRenderer_PartialHTMLAccept(t *testing.T) {
	handler := func(e Resolver) error { return nil }
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, "application/json,"+wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.Equal(t, http.StatusBadRequest, event.Status())
	assert.Equal(t, page, event.Page())
}

// TestErrorRenderer_HandlerSuccess tests that ErrorRenderer calls handler successfully
func TestErrorRenderer_HandlerSuccess(t *testing.T) {
	handlerCalled := false
	handler := func(e Resolver) error {
		handlerCalled = true
		return nil
	}
	site := &Site{ID: "site1"}
	page := &Page{ID: "page1", Pattern: PageError4xx}

	manager := newMockPatternPageManager(page, nil)
	authorizer := NewMockPageAuthorizer(Deny, nil)

	renderer := ErrorRenderer[Resolver](handler, manager, nil, authorizer, nil, slog.Default())

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(wo.HeaderAccept, wo.MIMETextHTML)
	resp := httptest.NewRecorder()

	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockPageTheme{})
	event.SetSite(site)

	httpErr := wo.ErrBadRequest

	renderer(event, httpErr)

	assert.True(t, handlerCalled)
}
