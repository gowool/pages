package pages

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageHandler(t *testing.T) {
	t.Run("panics when theme is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPageHandler(nil)
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = handler.ServeHTTP(w, req)
		})
	})

	t.Run("returns error when site is not found", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetPage(NewPage())

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("returns error when page is not found", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(NewSite())

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("sets page headers", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Header = map[string][]string{
			"X-Custom":   {"value1"},
			"X-Multiple": {"value1", "value2"},
		}
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "value1", w.Header().Get("X-Custom"))
		assert.Equal(t, "value1", w.Header().Get("X-Multiple"))
		assert.Equal(t, []string{"value1", "value2"}, w.Header().Values("X-Multiple"))
	})

	t.Run("returns error when template is empty", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = ""
		c.SetStatus(http.StatusCreated)

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.ErrorIs(t, err, ErrTemplateEmpty)
	})

	t.Run("writes template output", func(t *testing.T) {
		theme := &MockTheme{content: "test content"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test content", w.Body.String())
		assert.Equal(t, "text/html; charset=UTF-8", w.Header().Get("Content-Type"))
	})

	t.Run("uses existing content type header", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"
		page.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("returns error on theme write failure", func(t *testing.T) {
		theme := &MockTheme{err: errors.New("write error")}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "write error")
	})

	t.Run("uses default pageCtx when nil", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "test.html", theme.template)
	})

	t.Run("uses custom pageCtx", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		customData := "custom context data"
		handler := NewPageHandlerWithConfig(theme, PageHandlerConfig{
			VarsFunc: func(*http.Request, *Context) any {
				return customData
			},
		})

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, customData, theme.data)
	})

	t.Run("uses discard logger when nil", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := NewPageHandler(theme)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("logs error on write response error", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		handler := NewPageHandlerWithConfig(theme, PageHandlerConfig{Logger: logger})

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		page := NewPage()
		page.Template = "test.html"

		c.SetSite(site)
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := &errorResponseWriter{httptest.NewRecorder(), errors.New("write failed")}

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Contains(t, logBuf.String(), "write response error")
	})
}
