package pages

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
			NewPageCreateHandler(nil, authorizer)
		})
	})

	t.Run("panics when authorizer is nil", func(t *testing.T) {
		store := &MockPageStore{}

		assert.Panics(t, func() {
			NewPageCreateHandler(store, nil)
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)
		handler := NewPageCreateHandler(store, authorizer)

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
		handler := NewPageCreateHandler(store, authorizer)

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
		handler := NewPageCreateHandler(store, authorizer)

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
		handler := NewPageCreateHandler(store, authorizer)

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
		handler := NewPageCreateHandler(store, authorizer)

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

	t.Run("returns error when generator func failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandlerWithConfig(store, authorizer, PageCreateHandlerConfig{
			GeneratorFunc: func(context.Context) (ID, error) {
				return "", errors.New("generator error")
			},
		})

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

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generator error")
	})

	t.Run("handles JSON content type", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "http://example.com/test", redirect.url)
		store.AssertExpectations(t)
	})

	t.Run("handles form content type", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "http://example.com/test", redirect.url)
		store.AssertExpectations(t)
	})

	t.Run("trims trailing slash from URL", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "https://localhost/test", redirect.url)
	})

	t.Run("prepends slash to URL if missing", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "https://localhost/test", redirect.url)
	})

	t.Run("returns error on JSON decode failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		handler := NewPageCreateHandler(store, authorizer)

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

		handler := NewPageCreateHandler(store, authorizer)

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

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "https://localhost/parent/child", redirect.url)
		store.AssertExpectations(t)
	})

	t.Run("does not set parent when URL is root", func(t *testing.T) {
		store := &MockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		handler := NewPageCreateHandler(store, authorizer)

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

		assert.NotNil(t, err)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "https://localhost", redirect.url)
		store.AssertExpectations(t)
	})

	t.Run("returns error on before save failure", func(t *testing.T) {
		store := &MockPageStore{}
		authorizer := &MockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		beforeSave := func(ctx context.Context, page *Page) error {
			return errors.New("before save error")
		}

		handler := NewPageCreateHandlerWithConfig(store, authorizer, PageCreateHandlerConfig{
			BeforeSaveFunc: beforeSave,
		})

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

		handler := NewPageCreateHandler(store, authorizer)

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

		handler := NewPageCreateHandlerWithConfig(store, authorizer, PageCreateHandlerConfig{
			BeforeSaveFunc: beforeSave,
		})

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

		assert.NotNil(t, err)
		assert.True(t, beforeSaveCalled)

		var redirect *RedirectError
		assert.ErrorAs(t, err, &redirect)

		assert.Equal(t, http.StatusFound, redirect.status)
		assert.Equal(t, "https://localhost/test", redirect.url)
		store.AssertExpectations(t)
	})
}
