package pages

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPageHandler(t *testing.T) {
	t.Run("panics when theme is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			PageHandler(nil, PageCtx(), slog.Default())
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := PageHandler(theme, PageCtx(), slog.Default())
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = handler.ServeHTTP(w, req)
		})
	})

	t.Run("returns error when site is not found", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := PageHandler(theme, PageCtx(), slog.Default())

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
		handler := PageHandler(theme, PageCtx(), slog.Default())

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
		handler := PageHandler(theme, PageCtx(), slog.Default())

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

	t.Run("writes status when template is empty", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := PageHandler(theme, PageCtx(), slog.Default())

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

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("writes template output", func(t *testing.T) {
		theme := &MockTheme{content: "test content"}
		handler := PageHandler(theme, PageCtx(), slog.Default())

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
		handler := PageHandler(theme, PageCtx(), slog.Default())

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
		handler := PageHandler(theme, PageCtx(), slog.Default())

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
		assert.Contains(t, err.Error(), "theme write error")
	})

	t.Run("uses default pageCtx when nil", func(t *testing.T) {
		theme := &MockTheme{content: "test"}
		handler := PageHandler(theme, nil, slog.Default())

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
		pageCtx := func(r *http.Request, c *Context) any {
			return customData
		}
		handler := PageHandler(theme, pageCtx, slog.Default())

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
		handler := PageHandler(theme, PageCtx(), nil)

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
		handler := PageHandler(theme, PageCtx(), logger)

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

type errorResponseWriter struct {
	*httptest.ResponseRecorder
	writeErr error
}

func (e *errorResponseWriter) Write(data []byte) (int, error) {
	return 0, e.writeErr
}

func TestPageCreateRequest_Validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "template.html",
			Title:    "Test Title",
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})

	t.Run("missing URL", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "",
			Template: "template.html",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("URL too long", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/" + strings.Repeat("a", 255),
			Template: "template.html",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("missing template", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("template too long", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: strings.Repeat("a", 255),
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("title too long", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "template.html",
			Title:    strings.Repeat("a", 255),
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("valid empty title", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "template.html",
			Title:    "",
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})
}

func TestPageCreateHandler(t *testing.T) {
	t.Run("panics when store is nil", func(t *testing.T) {
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		assert.Panics(t, func() {
			PageCreateHandler(nil, authorizer, nil)
		})
	})

	t.Run("panics when authorizer is nil", func(t *testing.T) {
		store := &MockPageStore{}

		assert.Panics(t, func() {
			PageCreateHandler(store, nil, nil)
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)
		handler := PageCreateHandler(store, authorizer, nil)

		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = handler.ServeHTTP(w, req)
		})
	})

	t.Run("returns error when site is not found", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)
		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetGuest(false)

		req := httptest.NewRequest("POST", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("returns error when guest", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)
		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("POST", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageUnauthorized)
	})

	t.Run("returns error when authorization denied", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)
		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest("POST", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageForbidden)
	})

	t.Run("returns error when method is not POST", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)
		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("handles JSON content type", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		site.Scheme = "http"
		site.Host = "example.com"
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/test","template":"test.html","title":"Test"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "http://example.com/test", w.Header().Get("Location"))
		store.AssertExpectations(t)
	})

	t.Run("handles form content type", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		site.Scheme = "http"
		site.Host = "example.com"
		c.SetSite(site)
		c.SetGuest(false)

		form := "url=/test&template=test.html&title=Test"
		req := httptest.NewRequest("POST", "/", strings.NewReader(form)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		store.AssertExpectations(t)
	})

	t.Run("trims trailing slash from URL", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/test//","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("prepends slash to URL if missing", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"test","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("returns error on JSON decode failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		body := `{invalid json`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json decode")
	})

	t.Run("returns error on parse form failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest("POST", "/", strings.NewReader("%invalid")).WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse form")
	})

	t.Run("returns error on validation failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"a","template":""}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate")
	})

	t.Run("sets parent page when URL has parent", func(t *testing.T) {
		store := &MockPageStore{}
		parentPage := NewPage()
		parentPage.ID = "parent1"
		parentPage.URL = "/parent"

		store.On("FindByURL", mock.Anything, mock.Anything, "/parent").Return(parentPage, nil)
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/parent/child","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("does not set parent when URL is root", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		store.AssertExpectations(t)
	})

	t.Run("returns error on before save failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		beforeSave := func(ctx context.Context, page *Page) error {
			return errors.New("before save error")
		}

		handler := PageCreateHandler(store, authorizer, beforeSave)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/test","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "before save")
	})

	t.Run("returns error on save failure", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(errors.New("save error"))

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := PageCreateHandler(store, authorizer, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/test","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save")
		store.AssertExpectations(t)
	})

	t.Run("calls beforeSave when provided", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		beforeSaveCalled := false
		beforeSave := func(ctx context.Context, page *Page) error {
			beforeSaveCalled = true
			assert.Equal(t, "/test", page.CustomURL)
			return nil
		}

		handler := PageCreateHandler(store, authorizer, beforeSave)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		body := `{"url":"/test","template":"test.html"}`
		req := httptest.NewRequest("POST", "/", strings.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, beforeSaveCalled)
		store.AssertExpectations(t)
	})
}

func TestPageErrorHandler(t *testing.T) {
	t.Run("panics when page handler is nil", func(t *testing.T) {
		manager := &MockPageManager{}
		authorizer := &MockPageAuthorizer{}
		strategy := &MockPageDecoratorStrategy{}
		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := func(context.Context, int) string { return "" }

		assert.Panics(t, func() {
			PageErrorHandler(nil, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())
		})
	})

	t.Run("panics when manager is nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
		authorizer := &MockPageAuthorizer{}
		strategy := &MockPageDecoratorStrategy{}
		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := func(context.Context, int) string { return "" }

		assert.Panics(t, func() {
			PageErrorHandler(pageHandler, nil, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())
		})
	})

	t.Run("panics when authorizer is nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
		manager := &MockPageManager{}
		strategy := &MockPageDecoratorStrategy{}
		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := func(context.Context, int) string { return "" }

		assert.Panics(t, func() {
			PageErrorHandler(pageHandler, manager, nil, strategy, errorStatusFunc, errorPatternFunc, slog.Default())
		})
	})

	t.Run("panics when strategy is nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
		manager := &MockPageManager{}
		authorizer := &MockPageAuthorizer{}
		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := func(context.Context, int) string { return "" }

		assert.Panics(t, func() {
			PageErrorHandler(pageHandler, manager, authorizer, nil, errorStatusFunc, errorPatternFunc, slog.Default())
		})
	})

	t.Run("panics when errorStatusFunc is nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
		manager := &MockPageManager{}
		authorizer := &MockPageAuthorizer{}
		strategy := &MockPageDecoratorStrategy{}
		errorPatternFunc := func(context.Context, int) string { return "" }

		assert.Panics(t, func() {
			PageErrorHandler(pageHandler, manager, authorizer, strategy, nil, errorPatternFunc, slog.Default())
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
		manager := &MockPageManager{}
		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)
		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()
		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			handler.ServeHTTP(w, req, errors.New("test"))
		})
	})

	t.Run("uses fallback when site is not found", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusNotFound }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(nil)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "text/html; charset=UTF-8", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})

	t.Run("normalizes low status codes to Internal Server Error", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusMovedPermanently }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageError5xx)
	})

	t.Run("uses internal create pattern for 404 when decorable and authorized", func(t *testing.T) {
		theme := &MockTheme{content: "create"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		createPage := NewPage()
		createPage.Template = "create.html"
		createPage.ID = "create1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageInternalCreate).Return(createPage, nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		strategy := NewMockPageDecoratorStrategy(true)

		errorStatusFunc := func(context.Context, error) int { return http.StatusNotFound }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/new-page", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("not found"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageInternalCreate)
	})

	t.Run("uses error pattern for 404 when not decorable", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		notFoundPage := NewPage()
		notFoundPage.Template = "404.html"
		notFoundPage.ID = "404"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(notFoundPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusNotFound }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("not found"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound)
	})

	t.Run("uses error pattern for 404 when not authorized", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		notFoundPage := NewPage()
		notFoundPage.Template = "404.html"
		notFoundPage.ID = "404"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(notFoundPage, nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)

		strategy := NewMockPageDecoratorStrategy(true)

		errorStatusFunc := func(context.Context, error) int { return http.StatusNotFound }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("not found"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound)
	})

	t.Run("uses custom error pattern func", func(t *testing.T) {
		theme := &MockTheme{content: "custom"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		customPage := NewPage()
		customPage.Template = "custom.html"
		customPage.ID = "custom1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, "custom_error").Return(customPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusBadRequest }
		errorPatternFunc := func(context.Context, int) string { return "custom_error" }

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("bad request"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, "custom_error")
	})

	t.Run("sets error, status, and page in context", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		testError := errors.New("test error")

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, testError)

		c = FromContext(ctx)
		assert.Equal(t, testError, c.Error())
		assert.Equal(t, http.StatusInternalServerError, c.Status())
		assert.Equal(t, errorPage, c.Page())
	})

	t.Run("uses fallback when manager returns error", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, logger)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, logBuf.String(), "find page by pattern return error")
	})

	t.Run("uses fallback when page handler returns error", func(t *testing.T) {
		theme := &MockTheme{err: errors.New("handler error")}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, logger)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, logBuf.String(), "page handler return error")
	})

	t.Run("uses default error pattern when errorPatternFunc is nil", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, nil, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageError5xx)
	})

	t.Run("uses discard logger when nil", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, nil, nil)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))
	})

	t.Run("logs error on template execute failure", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "error.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&logBuf, nil))
		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, logger)

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		c.SetSite(nil)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := &errorResponseWriter{httptest.NewRecorder(), errors.New("write failed")}

		handler.ServeHTTP(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, logBuf.String(), "write response error")
	})

	t.Run("successfully renders error page", func(t *testing.T) {
		theme := &MockTheme{content: "error page content"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "500.html"
		errorPage.ID = "error1"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := ErrorPattern()

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "error page content", w.Body.String())
	})

	t.Run("uses PageError5xx when errorPatternFunc returns empty", func(t *testing.T) {
		theme := &MockTheme{content: "error"}
		pageHandler := PageHandler(theme, PageCtx(), slog.Default())

		manager := &MockPageManager{}
		errorPage := NewPage()
		errorPage.Template = "5xx.html"
		errorPage.ID = "5xx"
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(errorPage, nil)

		authorizer := &MockPageAuthorizer{}
		strategy := NewMockPageDecoratorStrategy(false)

		errorStatusFunc := func(context.Context, error) int { return http.StatusInternalServerError }
		errorPatternFunc := func(context.Context, int) string { return "" }

		handler := PageErrorHandler(pageHandler, manager, authorizer, strategy, errorStatusFunc, errorPatternFunc, slog.Default())

		ctx, _ := NewContext(context.Background())
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req, errors.New("test error"))

		manager.AssertCalled(t, "GetByPattern", mock.Anything, mock.Anything, PageError5xx)
	})
}

func TestErrorTemplate(t *testing.T) {
	t.Run("renders error template", func(t *testing.T) {
		var buf bytes.Buffer
		data := map[string]any{
			"Title":  "Not Found",
			"Status": 404,
		}

		err := ErrorTemplate.Execute(&buf, data)

		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Not Found!")
		assert.Contains(t, output, "Code 404")
	})
}
