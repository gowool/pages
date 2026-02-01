package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"iter"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPageCreateRequest_Validate(t *testing.T) {
	t.Run("Valid request with all fields", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
			Title:    "Test Page",
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})

	t.Run("Valid request without title", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})

	t.Run("Missing URL should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			Template: "page.html",
			Title:    "Test Page",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Empty URL should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "",
			Template: "page.html",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Missing template should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:   "/test-page",
			Title: "Test Page",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Empty template should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("URL too long should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/" + strings.Repeat("a", 255),
			Template: "page.html",
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Template too long should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: strings.Repeat("a", 255),
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Title too long should fail", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "page.html",
			Title:    strings.Repeat("a", 255),
		}

		err := dto.Validate()

		assert.Error(t, err)
	})

	t.Run("Title at max length should pass", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/test",
			Template: "page.html",
			Title:    strings.Repeat("a", 254),
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})

	t.Run("URL at max length should pass", func(t *testing.T) {
		dto := PageCreateRequest{
			URL:      "/" + strings.Repeat("a", 253),
			Template: "page.html",
		}

		err := dto.Validate()

		assert.NoError(t, err)
	})
}

func TestNewPageCreateHandler(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		store := &mockPageStore{}
		beforeSave := func(ctx context.Context, page *Page) error { return nil }

		handler := NewPageCreateHandler(store, beforeSave)

		assert.NotNil(t, handler)
		assert.Equal(t, store, handler.store)
		assert.NotNil(t, handler.beforeSave)
	})

	t.Run("Nil store should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPageCreateHandler(nil, func(ctx context.Context, page *Page) error { return nil })
		})
	})

	t.Run("Nil beforeSave uses default", func(t *testing.T) {
		store := &mockPageStore{}

		handler := NewPageCreateHandler(store, nil)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.beforeSave)
	})

	t.Run("Default beforeSave returns nil", func(t *testing.T) {
		store := &mockPageStore{}
		handler := NewPageCreateHandler(store, nil)

		err := handler.beforeSave(context.Background(), NewPage())

		assert.NoError(t, err)
	})
}

func TestPageCreateHandler_Handle(t *testing.T) {
	t.Run("Nil context should panic", func(t *testing.T) {
		handler := NewPageCreateHandler(&mockPageStore{}, nil)
		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = handler.Handle(w, req)
		})
	})

	t.Run("No site in context", func(t *testing.T) {
		handler := NewPageCreateHandler(&mockPageStore{}, nil)
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		req := httptest.NewRequest("POST", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("JSON content type - valid request", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
			Title:    "Test Page",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
	})

	t.Run("JSON content type - decode error", func(t *testing.T) {
		handler := NewPageCreateHandler(&mockPageStore{}, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())

		req := httptest.NewRequest("POST", "/", strings.NewReader("invalid json")).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json decode")
	})

	t.Run("Form content type - valid request", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		form := url.Values{}
		form.Add("url", "/test-page")
		form.Add("template", "page.html")
		form.Add("title", "Test Page")
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode())).WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
	})

	t.Run("URL trimming - removes trailing slash", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page/",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
	})

	t.Run("URL without leading slash - adds prefix", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "test-page",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
	})

	t.Run("Parent page found", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("FindByURL", mock.Anything, mock.Anything, "/parent").Return(NewPage(), nil)
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/parent/child",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
	})

	t.Run("Parent page not found", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("FindByURL", mock.Anything, mock.Anything, "/parent").Return(nil, ErrPageNotFound)
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/parent/child",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
	})

	t.Run("Root page - no parent lookup", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
	})

	t.Run("BeforeSave callback returns error", func(t *testing.T) {
		store := &mockPageStore{}
		beforeSave := func(ctx context.Context, page *Page) error {
			return assert.AnError
		}
		handler := NewPageCreateHandler(store, beforeSave)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "before save")
	})

	t.Run("Store Save returns error", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(ErrUniqueViolation)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save")
	})

	t.Run("Success - page created and redirect", func(t *testing.T) {
		store := &mockPageStore{}
		store.On("Save", mock.Anything, mock.Anything).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
			Title:    "Test Page",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "https://localhost/test-page", w.Header().Get("Location"))
	})

	t.Run("Page properties set correctly", func(t *testing.T) {
		store := &mockPageStore{}
		var savedPage *Page
		store.On("Save", mock.Anything, mock.MatchedBy(func(p []*Page) bool {
			if len(p) > 0 {
				savedPage = p[0]
			}
			return true
		})).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "page.html",
			Title:    "Test Page",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		require.NotNil(t, savedPage)
		assert.Equal(t, PageCMS, savedPage.Pattern)
		assert.Equal(t, "Test Page", savedPage.Title)
		assert.Equal(t, "/test-page", savedPage.CustomURL)
		assert.Equal(t, "page.html", savedPage.Template)
		assert.True(t, savedPage.Decorate)
		assert.Equal(t, site.ID, savedPage.SiteID)
		assert.Equal(t, site, savedPage.Site)
	})
}

func TestPageCreateHandler_Integration(t *testing.T) {
	t.Run("Full flow with parent page", func(t *testing.T) {
		store := &mockPageStore{}

		parentPage := NewPage()
		parentPage.ID = "parent1"
		store.On("FindByURL", mock.Anything, mock.Anything, "/parent").Return(parentPage, nil)

		var savedPage *Page
		store.On("Save", mock.Anything, mock.MatchedBy(func(p []*Page) bool {
			if len(p) > 0 {
				savedPage = p[0]
			}
			return true
		})).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/parent/child",
			Template: "page.html",
			Title:    "Child Page",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		require.NotNil(t, savedPage)
		assert.Equal(t, parentPage.ID, *savedPage.ParentID)
		assert.Equal(t, parentPage, savedPage.Parent)
	})

	t.Run("URL normalization edge cases", func(t *testing.T) {
		testCases := []struct {
			name        string
			inputURL    string
			expectedURL string
		}{
			{"Trailing slash", "/test/", "/test"},
			{"Multiple trailing slashes", "/test///", "/test"},
			{"No leading slash", "test", "/test"},
			{"Already correct", "/test", "/test"},
			{"Just slash", "/", "/"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				store := &mockPageStore{}
				var savedPage *Page
				store.On("Save", mock.Anything, mock.MatchedBy(func(p []*Page) bool {
					if len(p) > 0 {
						savedPage = p[0]
					}
					return true
				})).Return(nil)

				handler := NewPageCreateHandler(store, nil)

				parentCtx := context.Background()
				ctx, _ := NewContext(parentCtx)
				c := FromContext(ctx)
				site := NewSite()
				site.ID = "site1"
				c.SetSite(site)

				dto := PageCreateRequest{
					URL:      tc.inputURL,
					Template: "page.html",
				}
				body, _ := json.Marshal(dto)
				req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
				req.Header.Set("Content-Type", mimeApplicationJSON)
				w := httptest.NewRecorder()

				err := handler.Handle(w, req)

				assert.NoError(t, err)
				require.NotNil(t, savedPage)
				assert.Equal(t, tc.expectedURL, savedPage.CustomURL)
			})
		}
	})

	t.Run("Name generation from URL", func(t *testing.T) {
		store := &mockPageStore{}
		var savedPage *Page
		store.On("Save", mock.Anything, mock.MatchedBy(func(p []*Page) bool {
			if len(p) > 0 {
				savedPage = p[0]
			}
			return true
		})).Return(nil)

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/my-test-page",
			Template: "page.html",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		require.NotNil(t, savedPage)
		assert.Equal(t, "MY-TEST-PAGE", savedPage.Name)
	})

	t.Run("Page properties - all fields set correctly", func(t *testing.T) {
		store := &mockPageStore{}
		var savedPage *Page
		store.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			if len(args) >= 2 {
				if pages, ok := args[1].([]*Page); ok && len(pages) > 0 {
					savedPage = pages[0]
				}
			}
		})

		handler := NewPageCreateHandler(store, nil)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		site.ID = "site1"
		site.Title = "Test Site"
		site.Separator = " - "
		c.SetSite(site)

		dto := PageCreateRequest{
			URL:      "/test-page",
			Template: "test-template.html",
			Title:    "Test Page Title",
		}
		body, _ := json.Marshal(dto)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body)).WithContext(ctx)
		req.Header.Set("Content-Type", mimeApplicationJSON)
		w := httptest.NewRecorder()

		err := handler.Handle(w, req)

		assert.NoError(t, err)
		require.NotNil(t, savedPage)
		assert.Equal(t, PageCMS, savedPage.Pattern)
		assert.Equal(t, "Test Page Title", savedPage.Title)
		assert.Equal(t, "/test-page", savedPage.CustomURL)
		assert.Equal(t, "test-template.html", savedPage.Template)
		assert.True(t, savedPage.Decorate)
		assert.Equal(t, site.ID, savedPage.SiteID)
		assert.Equal(t, site, savedPage.Site)
	})
}

type mockPageStore struct {
	mock.Mock
}

func (m *mockPageStore) FindByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageStore) FindByURL(ctx context.Context, siteID ID, url string) (*Page, error) {
	args := m.Called(ctx, siteID, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageStore) FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error) {
	args := m.Called(ctx, siteID, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageStore) FindByPatterns(ctx context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error] {
	args := m.Called(ctx, siteID, patterns)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(iter.Seq2[*Page, error])
}

func (m *mockPageStore) FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error) {
	args := m.Called(ctx, siteID, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageStore) Save(ctx context.Context, pages ...*Page) error {
	args := m.Called(ctx, pages)
	return args.Error(0)
}
