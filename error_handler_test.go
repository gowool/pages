package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testRedirectURIErr struct {
	uri string
}

func (e *testRedirectURIErr) Error() string {
	return "test redirect uri error: " + e.uri
}

func (e *testRedirectURIErr) URI() string {
	return e.uri
}

type testRedirectURLErr struct {
	url string
}

func (e *testRedirectURLErr) Error() string {
	return "test redirect url error: " + e.url
}

func (e *testRedirectURLErr) URL() string {
	return e.url
}

func TestHTTPErrorHandlerConfig_SetDefaults(t *testing.T) {
	t.Run("sets default values for all fields", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		assert.Equal(t, "error.gohtml", cfg.FallbackTemplate)
		assert.NotNil(t, cfg.StatusFunc)
		assert.NotNil(t, cfg.JSONHandler)
		assert.NotNil(t, cfg.Logger)
	})

	t.Run("preserves existing FallbackTemplate", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{FallbackTemplate: "custom.html"}
		cfg.SetDefaults()

		assert.Equal(t, "custom.html", cfg.FallbackTemplate)
	})

	t.Run("preserves existing StatusFunc", func(t *testing.T) {
		customStatusFunc := func(context.Context, error) int {
			return 418
		}
		cfg := HTTPErrorHandlerConfig{StatusFunc: customStatusFunc}
		cfg.SetDefaults()

		status := cfg.StatusFunc(context.Background(), errors.New("test"))
		assert.Equal(t, 418, status)
	})

	t.Run("preserves existing JSONHandler", func(t *testing.T) {
		customHandlerCalled := false
		customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customHandlerCalled = true
			w.WriteHeader(http.StatusOK)
		})
		cfg := HTTPErrorHandlerConfig{JSONHandler: customHandler}
		cfg.SetDefaults()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		cfg.JSONHandler.ServeHTTP(w, req)

		assert.True(t, customHandlerCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("wraps logger with group name", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		logHandler := cfg.Logger.Handler()
		assert.NotNil(t, logHandler)
	})
}

func TestHTTPErrorHandlerConfig_jsonHandler(t *testing.T) {
	t.Run("returns JSON response with status", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetStatus(http.StatusNotFound)
		c.SetError(ErrPageNotFound)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		req.Pattern = "/test"
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		cfg.jsonHandler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, MIMEApplicationJSON, w.Header().Get(HeaderContentType))

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		assert.Equal(t, float64(http.StatusNotFound), data["status"])
		assert.Equal(t, "Not Found", data["message"])
	})

	t.Run("includes site information when available", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.Name = "Test Site"
		c.SetSite(site)
		c.SetStatus(http.StatusInternalServerError)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		cfg.jsonHandler(w, req)

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		assert.NotNil(t, data["site"])
		siteData := data["site"].(map[string]any)
		assert.Equal(t, "Test Site", siteData["name"])
	})

	t.Run("includes error data for 422 status", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		testErr := errors.New("validation error: name is required")
		c.SetError(testErr)
		c.SetStatus(http.StatusUnprocessableEntity)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		cfg.jsonHandler(w, req)

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		assert.NotNil(t, data["data"])
	})

	t.Run("includes error when debug is enabled", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		testErr := errors.New("internal error details")
		c.SetError(testErr)
		c.SetStatus(http.StatusInternalServerError)
		c.SetDebug(true)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		cfg.jsonHandler(w, req)

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		assert.Equal(t, "internal error details", data["error"])
	})

	t.Run("excludes error when debug is disabled", func(t *testing.T) {
		cfg := HTTPErrorHandlerConfig{}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		testErr := errors.New("internal error details")
		c.SetError(testErr)
		c.SetStatus(http.StatusInternalServerError)
		c.SetDebug(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		cfg.jsonHandler(w, req)

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		_, exists := data["error"]
		assert.False(t, exists)
	})

	t.Run("logs error on JSON encoding failure", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		cfg := HTTPErrorHandlerConfig{Logger: logger}
		cfg.SetDefaults()

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetStatus(http.StatusInternalServerError)
		c.SetError(errors.New("test error"))

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := &errorResponseWriter{httptest.NewRecorder(), errors.New("write failed")}

		cfg.jsonHandler(w, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "write response error")
	})
}

func TestNewHTTPErrorHandler(t *testing.T) {
	t.Run("creates handler with default config", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		assert.NotNil(t, h)
		assert.Equal(t, handler, h.pageHandler)
		assert.Equal(t, manager, h.manager)
		assert.Equal(t, pattern, h.errPattern)
		assert.Equal(t, "error.gohtml", h.fallbackTemplate)
		assert.NotNil(t, h.jsonHandler)
		assert.NotNil(t, h.errStatusFunc)
		assert.NotNil(t, h.logger)
	})

	t.Run("panics when pageHandler is nil", func(t *testing.T) {
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		assert.Panics(t, func() {
			NewHTTPErrorHandler(nil, manager, pattern)
		})
	})

	t.Run("panics when manager is nil", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		pattern := &MockErrorPattern{}

		assert.Panics(t, func() {
			NewHTTPErrorHandler(handler, nil, pattern)
		})
	})

	t.Run("panics when errPattern is nil", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()

		assert.Panics(t, func() {
			NewHTTPErrorHandler(handler, manager, nil)
		})
	})
}

func TestNewHTTPErrorHandlerWithConfig(t *testing.T) {
	t.Run("creates handler with custom config", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		var logBuf bytes.Buffer
		customLogger := slog.New(slog.NewTextHandler(&logBuf, nil))

		config := HTTPErrorHandlerConfig{
			FallbackTemplate: "custom.html",
			StatusFunc: func(ctx context.Context, err error) int {
				return 418
			},
			Logger: customLogger,
		}

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, config)

		assert.NotNil(t, h)
		assert.Equal(t, "custom.html", h.fallbackTemplate)

		status := h.errStatusFunc(context.Background(), errors.New("test"))
		assert.Equal(t, 418, status)
	})

	t.Run("applies defaults to empty config", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, HTTPErrorHandlerConfig{})

		assert.NotNil(t, h)
		assert.Equal(t, "error.gohtml", h.fallbackTemplate)
		assert.NotNil(t, h.jsonHandler)
		assert.NotNil(t, h.errStatusFunc)
		assert.NotNil(t, h.logger)
	})
}

func TestHTTPErrorHandler_redirect(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantURL    string
		wantStatus int
	}{
		{
			name:       "redirects for 301 status",
			err:        NewRedirectError("/new-location", http.StatusMovedPermanently),
			wantURL:    "/new-location",
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "redirects for 302 status",
			err:        NewRedirectError("/temp", http.StatusFound),
			wantURL:    "/temp",
			wantStatus: http.StatusFound,
		},
		{
			name:       "redirects to / when error has no URL",
			err:        errors.New("test error"),
			wantURL:    "/",
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "redirects for 302 status when error has URI",
			err:        &testRedirectURIErr{uri: "/new-location"},
			wantURL:    "/new-location",
			wantStatus: http.StatusFound,
		},
		{
			name:       "redirects for 302 status when error has URL",
			err:        &testRedirectURLErr{url: "/new-location"},
			wantURL:    "/new-location",
			wantStatus: http.StatusFound,
		},
		{
			name:       "redirects for 302 status when error is wrapped",
			err:        fmt.Errorf("wrapped error: %w", NewRedirectError("/new-location", http.StatusFound)),
			wantURL:    "/new-location",
			wantStatus: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageHandler := &MockTheme{content: "test"}
			handler := NewPageHandler(pageHandler)
			manager := NewMockPageManager()
			pattern := &MockErrorPattern{}

			h := NewHTTPErrorHandler(handler, manager, pattern)

			req := httptest.NewRequest(http.MethodGet, "/old", nil)
			w := httptest.NewRecorder()

			redirected := h.redirect(w, req, tt.wantStatus, tt.err)

			assert.True(t, redirected)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantURL, w.Header().Get("Location"))
		})
	}

	t.Run("does not redirect for non-redirect status", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		redirected := h.redirect(w, req, http.StatusNotFound, ErrPageNotFound)

		assert.False(t, redirected)
		assert.Empty(t, w.Header().Get("Location"))
	})
}

func TestHTTPErrorHandler_serveHTTP(t *testing.T) {
	t.Run("calls page handler successfully", func(t *testing.T) {
		pageHandler := &MockTheme{content: "error page content"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "error.html"
		c.SetSite(site)
		c.SetPage(page)
		c.SetStatus(http.StatusNotFound)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.serveHTTP(w, req)

		assert.Equal(t, "error page content", w.Body.String())
	})

	t.Run("sets fallback template when not set", func(t *testing.T) {
		pageHandler := &MockTheme{content: "fallback content"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = ""
		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.serveHTTP(w, req)

		assert.Equal(t, "error.gohtml", c.Template())
		assert.Equal(t, "fallback content", w.Body.String())
	})

	t.Run("logs error when page handler fails", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		pageHandler := &MockTheme{err: errors.New("render error")}
		handler := NewPageHandler(pageHandler)

		config := HTTPErrorHandlerConfig{Logger: logger}
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, config)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "error.html"
		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.serveHTTP(w, req)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "page handler error")
	})
}

func TestHTTPErrorHandler_ServeHTTP(t *testing.T) {
	t.Run("handles HEAD request with status only", func(t *testing.T) {
		pageHandler := &MockTheme{content: "should not see this"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetError(ErrPageNotFound)

		req := httptest.NewRequest(http.MethodHead, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, ErrPageNotFound)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

	t.Run("returns JSON for JSON accept header", func(t *testing.T) {
		pageHandler := &MockTheme{content: "HTML content"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetError(ErrPageNotFound)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		req.Header.Set(HeaderAccept, MIMEApplicationJSON)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, ErrPageNotFound)

		assert.Equal(t, MIMEApplicationJSON, w.Header().Get(HeaderContentType))

		var data map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &data)
		assert.NoError(t, err)
		assert.Equal(t, float64(http.StatusNotFound), data["status"])
	})

	t.Run("serves fallback template when no site in context", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		pageHandler := &MockTheme{content: "fallback error page"}
		handler := NewPageHandler(pageHandler)

		config := HTTPErrorHandlerConfig{Logger: logger}
		manager := NewMockPageManager()
		pattern := &MockErrorPattern{}

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, config)

		ctx, _ := NewContext(context.Background())

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, errors.New("test error"))

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "no site found in context")
		assert.Equal(t, "fallback error page", w.Body.String())
	})

	t.Run("retrieves error page by pattern from manager", func(t *testing.T) {
		pageHandler := &MockTheme{content: "404 page content"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		errorPage := NewPage()
		errorPage.Template = "error_404.html"

		pattern.On("Pattern", mock.Anything, http.StatusNotFound, ErrPageNotFound).Return(PageErrorNotFound)
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(errorPage, nil)

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, ErrPageNotFound)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "404 page content", w.Body.String())
		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound)
	})

	t.Run("uses PageError5xx when pattern returns empty", func(t *testing.T) {
		pageHandler := &MockTheme{content: "500 page content"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		errorPage := NewPage()
		errorPage.Template = "error_500.html"

		pattern.On("Pattern", mock.Anything, http.StatusInternalServerError, mock.Anything).Return("")
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, errors.New("internal error"))

		assert.Equal(t, "500 page content", w.Body.String())
		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageError5xx)
	})

	t.Run("logs error when GetByPattern fails", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		pageHandler := &MockTheme{content: "fallback content"}
		handler := NewPageHandler(pageHandler)

		config := HTTPErrorHandlerConfig{Logger: logger}
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		pattern.On("Pattern", mock.Anything, http.StatusNotFound, ErrPageNotFound).Return(PageErrorNotFound)
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(nil, errors.New("database error"))

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, config)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, ErrPageNotFound)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "find page by pattern return error")
		assert.Equal(t, "fallback content", w.Body.String())
	})

	t.Run("logs request details", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))

		pageHandler := &MockTheme{content: "error content"}
		handler := NewPageHandler(pageHandler)

		config := HTTPErrorHandlerConfig{Logger: logger}
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		errorPage := NewPage()
		errorPage.Template = "error.html"

		pattern.On("Pattern", mock.Anything, http.StatusNotFound, ErrPageNotFound).Return(PageErrorNotFound)
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(errorPage, nil)

		h := NewHTTPErrorHandlerWithConfig(handler, manager, pattern, config)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.Name = "Test Site"
		c.SetSite(site)
		c.SetDebug(true)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodPost, "/test/path?query=value", nil).WithContext(ctx)
		req.RemoteAddr = "127.0.0.1:8080"
		req.Header.Set("Referer", "https://example.com")
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req, ErrPageNotFound)

		logOutput := logBuf.String()
		assert.Contains(t, logOutput, "request failed")
		assert.Contains(t, logOutput, "status=404")
		assert.Contains(t, logOutput, "method=POST")
		assert.Contains(t, logOutput, "remote_addr=127.0.0.1:8080")
		assert.Contains(t, logOutput, "debug=true")
		assert.Contains(t, logOutput, "guest=false")
	})

	t.Run("redirects for redirect status codes", func(t *testing.T) {
		pageHandler := &MockTheme{content: "should not see this"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		h := NewHTTPErrorHandler(handler, manager, pattern)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/old", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		redirectErr := NewRedirectError("/new", http.StatusMovedPermanently)
		h.ServeHTTP(w, req, redirectErr)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/new", w.Header().Get("Location"))
	})
}

func TestHTTPErrorHandler_Interface(t *testing.T) {
	t.Run("implements ErrorHandler interface", func(t *testing.T) {
		pageHandler := &MockTheme{content: "test"}
		handler := NewPageHandler(pageHandler)
		manager := NewMockPageManager()
		pattern := NewMockErrorPattern()

		h := NewHTTPErrorHandler(handler, manager, pattern)

		var _ ErrorHandler = (*HTTPErrorHandler)(nil)
		assert.NotNil(t, h)
	})
}
