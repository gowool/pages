package pages

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewCtx(t *testing.T) {
	t.Run("Returns ViewCtxFunc", func(t *testing.T) {
		viewCtx := ViewCtx()

		assert.NotNil(t, viewCtx, "ViewCtx should return non-nil function")

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		req := httptest.NewRequest("GET", "/", nil)

		result := viewCtx(c, req)

		pc, ok := result.(PageContext)
		assert.True(t, ok, "Result should be PageContext")
		assert.Equal(t, c, pc.Context)
		assert.Equal(t, req, pc.Request)
	})

	t.Run("PageContext Value returns context value", func(t *testing.T) {
		type testKey struct{}
		value := "testValue"
		parentCtx := context.WithValue(context.Background(), testKey{}, value)

		req := httptest.NewRequest("GET", "/", nil).WithContext(parentCtx)
		pc := PageContext{Request: req}

		result := pc.Value(testKey{})

		assert.Equal(t, value, result)
	})

	t.Run("PageContext Value returns nil for missing key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		pc := PageContext{Request: req}

		result := pc.Value("nonexistent")

		assert.Nil(t, result)
	})
}

func TestNewDefaultPageHandler(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		theme := &mockTheme{}
		viewCtxFunc := ViewCtx()
		logger := slog.New(slog.DiscardHandler)

		handler := NewDefaultPageHandler(theme, viewCtxFunc, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, theme, handler.theme)
		assert.NotNil(t, handler.viewCtxFunc)
	})

	t.Run("Nil theme should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageHandler(nil, ViewCtx(), slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil viewCtxFunc uses default", func(t *testing.T) {
		theme := &mockTheme{}
		logger := slog.New(slog.DiscardHandler)

		handler := NewDefaultPageHandler(theme, nil, logger)

		assert.NotNil(t, handler.viewCtxFunc)
	})

	t.Run("Nil logger uses discard handler", func(t *testing.T) {
		theme := &mockTheme{}

		handler := NewDefaultPageHandler(theme, ViewCtx(), nil)

		assert.NotNil(t, handler.logger)
	})

	t.Run("All nil except theme", func(t *testing.T) {
		theme := &mockTheme{}

		handler := NewDefaultPageHandler(theme, nil, nil)

		assert.NotNil(t, handler)
		assert.Equal(t, theme, handler.theme)
		assert.NotNil(t, handler.viewCtxFunc)
		assert.NotNil(t, handler.logger)
	})
}

func TestDefaultPageHandler_Handle(t *testing.T) {
	t.Run("Nil context should panic", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = handler.Handle(w, req)
		})
	})

	t.Run("No site in context", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("No page in context", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("Page with empty template", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = ""
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("Page with empty template and custom status", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = ""
		c.SetPage(page)
		c.SetStatus(http.StatusNotFound)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Page with template", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html; charset=UTF-8", w.Header().Get("Content-Type"))
		assert.Equal(t, expectedContent, w.Body.String())
		assert.Equal(t, "test.html", theme.template)
	})

	t.Run("Theme write error", func(t *testing.T) {
		theme := &mockTheme{err: assert.AnError}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "theme write error")
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Custom content type in page header", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		page.Header = http.Header{"Content-Type": []string{"application/json"}}
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("Custom content type overrides default", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		page.Header = http.Header{
			"Content-Type": []string{"text/plain"},
			"X-Custom":     []string{"custom-value"},
		}
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "custom-value", w.Header().Get("X-Custom"))
	})

	t.Run("Multiple values for same header", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		page.Header = http.Header{
			"Set-Cookie": []string{"cookie1=value1", "cookie2=value2"},
		}
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		cookies := w.Header().Values("Set-Cookie")
		assert.Len(t, cookies, 2)
		assert.Equal(t, "cookie1=value1", cookies[0])
		assert.Equal(t, "cookie2=value2", cookies[1])
	})

	t.Run("ResponseWriter write error is logged but not returned", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		loggerBuf := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(loggerBuf, nil))
		handler := NewDefaultPageHandler(theme, ViewCtx(), logger)
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)

		w := &errorResponseWriter{ResponseWriter: httptest.NewRecorder()}

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.NotEmpty(t, loggerBuf.String())
		assert.Contains(t, loggerBuf.String(), "write response error")
	})

	t.Run("Custom viewCtxFunc is called", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		viewCtxCalled := false
		theme := &mockTheme{content: expectedContent}
		viewCtxFunc := func(c *Context, r *http.Request) any {
			viewCtxCalled = true
			return "custom context"
		}
		handler := NewDefaultPageHandler(theme, viewCtxFunc, slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.True(t, viewCtxCalled, "viewCtxFunc should have been called")
		assert.Equal(t, "custom context", theme.data)
	})

	t.Run("Page with non-default status", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		c.SetPage(page)
		c.SetStatus(http.StatusNotFound)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Empty page headers map", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		page.Header = http.Header{}
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedContent, w.Body.String())
	})

	t.Run("Handler preserves existing content-type header", func(t *testing.T) {
		expectedContent := "<html><body>Test Content</body></html>"
		theme := &mockTheme{content: expectedContent}
		handler := NewDefaultPageHandler(theme, ViewCtx(), slog.New(slog.DiscardHandler))
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		page := NewPage()
		page.Template = "test.html"
		page.Header = http.Header{}
		c.SetPage(page)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		w.Header().Set("Content-Type", "application/xml")

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
	})
}

func TestDefaultPageHandler_PageHandlerInterface(t *testing.T) {
	t.Run("Implements PageHandler", func(t *testing.T) {
		handler := NewDefaultPageHandler(&mockTheme{}, ViewCtx(), slog.New(slog.DiscardHandler))

		var _ PageHandler = handler
	})
}

func TestConstants(t *testing.T) {
	t.Run("Constants are defined correctly", func(t *testing.T) {
		assert.Equal(t, "Content-Type", headerContentType)
		assert.Equal(t, "text/html", mimeTextHTML)
		assert.Equal(t, "text/html; charset=UTF-8", mimeTextHTMLCharsetUTF8)
	})
}

type mockTheme struct {
	content  string
	template string
	err      error
	data     any
}

func (m *mockTheme) Write(_ context.Context, w io.Writer, template string, data any) error {
	m.template = template
	m.data = data
	if m.err != nil {
		return m.err
	}
	_, err := w.Write([]byte(m.content))
	return err
}

type errorResponseWriter struct {
	http.ResponseWriter
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, assert.AnError
}
