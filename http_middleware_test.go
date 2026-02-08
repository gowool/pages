package pages

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type errorResponseWriter struct {
	*httptest.ResponseRecorder
	writeErr error
}

func (e *errorResponseWriter) Write(data []byte) (int, error) {
	return 0, e.writeErr
}

func TestDelayedWriter(t *testing.T) {
	t.Run("Reset initializes state", func(t *testing.T) {
		w := &delayedWriter{}
		rw := httptest.NewRecorder()

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("test"))

		w.reset(rw)

		assert.Equal(t, rw, w.ResponseWriter)
		assert.Equal(t, 0, w.buffer.Len())
		assert.False(t, w.committed)
		assert.Equal(t, http.StatusOK, w.code)
	})

	t.Run("WriteHeader tracks code", func(t *testing.T) {
		w := &delayedWriter{}

		w.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, w.code)
		assert.True(t, w.committed)
	})

	t.Run("Write buffers data", func(t *testing.T) {
		w := &delayedWriter{}
		data := []byte("test data")

		n, err := w.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "test data", w.buffer.String())
		assert.True(t, w.committed)
	})

	t.Run("Write commits if not already committed", func(t *testing.T) {
		w := &delayedWriter{}

		n, err := w.Write([]byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, http.StatusOK, w.code)
		assert.True(t, w.committed)
	})

	t.Run("Unwrap returns ResponseWriter", func(t *testing.T) {
		rw := httptest.NewRecorder()
		w := &delayedWriter{ResponseWriter: rw}

		unwrapped := w.Unwrap()

		assert.Equal(t, rw, unwrapped)
	})
}

func TestSelectSiteMiddleware(t *testing.T) {
	t.Run("panic when retriever is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			SelectSiteMiddleware(nil)
		})
	})

	t.Run("skip when skipper returns true", func(t *testing.T) {
		siteRetriever := NewMockSiteRetriever(NewSite(), "", nil)

		skipper := func(r *http.Request) bool {
			return true
		}

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		nextCalled := false
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever, skipper)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Nil(t, FromContext(req.Context()).Site())
	})

	t.Run("panic when context is nil", func(t *testing.T) {
		siteRetriever := NewMockSiteRetriever(NewSite(), "", nil)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)

		assert.Panics(t, func() {
			_ = handler.ServeHTTP(w, req)
		})
	})

	t.Run("site found successfully", func(t *testing.T) {
		site := NewSite()
		siteRetriever := NewMockSiteRetriever(site, "", nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		c := FromContext(req.Context())
		assert.NotNil(t, c.Site())
		assert.Equal(t, "http", c.Site().Scheme)
		assert.Equal(t, "example.com", req.Host)
	})

	t.Run("site not found with ErrSiteNotFound", func(t *testing.T) {
		siteRetriever := NewMockSiteRetriever(nil, "", ErrSiteNotFound)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("site not found with other error", func(t *testing.T) {
		otherErr := errors.New("database error")
		siteRetriever := NewMockSiteRetriever(nil, "", otherErr)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, otherErr)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("site is nil", func(t *testing.T) {
		siteRetriever := NewMockSiteRetriever(nil, "", nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("site scheme and host are set from request", func(t *testing.T) {
		site := NewSite()
		siteRetriever := NewMockSiteRetriever(site, "", nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Host = "custom.example.com"
		req.URL.Scheme = "https"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		c := FromContext(req.Context())
		assert.Equal(t, "custom.example.com", c.Site().Host)
		assert.Equal(t, "http", c.Site().Scheme)
	})

	t.Run("site is root when path is /", func(t *testing.T) {
		site := NewSite()
		siteRetriever := NewMockSiteRetriever(site, "", nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, FromContext(req.Context()).Site().IsRoot)
	})

	t.Run("site is not root when path is not /", func(t *testing.T) {
		site := NewSite()
		siteRetriever := NewMockSiteRetriever(site, "", nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.False(t, FromContext(req.Context()).Site().IsRoot)
	})

	t.Run("pathInfo updates request URL path", func(t *testing.T) {
		site := NewSite()
		pathInfo := "/updated/path"
		siteRetriever := NewMockSiteRetriever(site, pathInfo, nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/original/path", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectSiteMiddleware(siteRetriever)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, pathInfo, req.URL.Path)
	})
}

func TestSelectPageMiddleware(t *testing.T) {
	t.Run("panic when manager is nil", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		patternArgs := PatternArgs()

		assert.Panics(t, func() {
			SelectPageMiddleware(nil, authorizer, patternArgs)
		})
	})

	t.Run("panic when authorizer is nil", func(t *testing.T) {
		manager := NewMockPageManager()
		patternArgs := PatternArgs()

		assert.Panics(t, func() {
			SelectPageMiddleware(manager, nil, patternArgs)
		})
	})

	t.Run("default patternArgs when nil", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Published
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("skip when skipper returns true", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		skipper := func(r *http.Request) bool {
			return true
		}

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		nextCalled := false
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs(), skipper)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
		assert.Nil(t, FromContext(req.Context()).Page())
	})

	t.Run("panic when context is nil", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)

		assert.Panics(t, func() {
			_ = handler.ServeHTTP(w, req)
		})
	})

	t.Run("error when site not found in context", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSiteNotFound)
	})

	t.Run("find page by pattern (not PageCMSPattern)", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", req.Context(), site, "/test").Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.NotNil(t, FromContext(req.Context()).Page())
	})

	t.Run("find page by URL (PageCMSPattern)", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Published
		page.Visibility = Private

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/cms-page", nil)
		req.Pattern = "GET /{_page_cms...}"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByURL", req.Context(), site, "/cms-page").Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.NotNil(t, FromContext(req.Context()).Page())
	})

	t.Run("page error joined with ErrPageNotFound", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		otherErr := errors.New("database error")
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(nil, otherErr)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, otherErr)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("page is nil error", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("set page Site and SiteID if nil", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Site = nil
		page.SiteID = ""

		site := NewSite()
		site.ID = "site-123"

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, site, FromContext(req.Context()).Page().Site)
		assert.Equal(t, site.ID, FromContext(req.Context()).Page().SiteID)
	})

	t.Run("patternArgs for dynamic page", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Published
		page.Visibility = Private

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		req.Pattern = "GET /test/{id}"
		req.SetPathValue("{id}", "123")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("draft page guest not allowed", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Draft

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(true)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("draft page access denied", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Deny)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Draft

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("draft page authorized successfully", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Draft

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("private page guest not allowed", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Published
		page.Visibility = Private

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(true)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageForbidden)
	})

	t.Run("private page access denied", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Deny)

		page := NewPage()
		page.Pattern = "/test"
		page.Status = Published
		page.Visibility = Private

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageForbidden)
	})

	t.Run("private page authorized successfully", func(t *testing.T) {
		manager := NewMockPageManager()
		authorizer := NewMockPageAuthorizer()
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow)

		page := NewPage()
		page.Pattern = "/test"
		page.Visibility = Private

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(page, nil)

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := SelectPageMiddleware(manager, authorizer, PatternArgs())
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})
}

func TestHybridPageMiddleware(t *testing.T) {
	t.Run("panic when pageHandler is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			HybridPageMiddleware(nil, nil)
		})
	})

	t.Run("default logger when nil", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		assert.NotNil(t, middleware)
	})

	t.Run("built-in skipper for non-hybrid pages", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = PageCMSPattern
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		nextCalled := false
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("skip when skipper returns true", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		skipper := func(r *http.Request) bool {
			return true
		}

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		nextCalled := false
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil, skipper)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("error when page not found in context", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("set X-Page-Decorable header", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "1", w.Header().Get(HeaderXPageDecorable))
	})

	t.Run("set X-Page-Not-Decorable header", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "1", w.Header().Get(HeaderXPageNotDecorable))
	})

	t.Run("code > 0 sets context code", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Status())
	})

	t.Run("context content is set from buffer", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		testContent := "test content"
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			_, err := w.Write([]byte(testContent))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, testContent, string(c.Content()))
	})

	t.Run("IsDecorable returns false - write response directly", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		testContent := "test content"
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(testContent))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, testContent, w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("IsDecorable returns true - call pageHandler", func(t *testing.T) {
		pageHandlerCalled := false
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			pageHandlerCalled = true
			_, err := w.Write([]byte("decorated content"))
			return err
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Content-Type", "text/html")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			_, err := w.Write([]byte("test content"))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.True(t, pageHandlerCalled)
	})

	t.Run("write response with buffer when not decorable", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		testContent := "buffered content"
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			_, err := w.Write([]byte(testContent))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, testContent, w.Body.String())
	})

	t.Run("logger error on write failure", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := &errorResponseWriter{
			ResponseRecorder: httptest.NewRecorder(),
			writeErr:         errors.New("write error"),
		}

		testContent := "test content"
		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			_, err := w.Write([]byte(testContent))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, logger)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
	})

	t.Run("next handler returns error", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return errors.New("handler error")
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.Error(t, err)
	})

	t.Run("empty buffer writes nothing", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Empty(t, w.Body.String())
	})

	t.Run("decorable with non-HTML content type", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("test content"))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "test content", w.Body.String())
	})

	t.Run("decorable with XMLHttpRequest", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			_, err := w.Write([]byte("test content"))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, "test content", w.Body.String())
	})

	t.Run("non-OK code writes response", func(t *testing.T) {
		pageHandler := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return nil
		})

		parentCtx, cancel := NewContext(context.Background())
		defer cancel()
		c := FromContext(parentCtx)

		page := NewPage()
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(parentCtx)
		w := httptest.NewRecorder()

		next := HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("not found"))
			return err
		})

		middleware := HybridPageMiddleware(pageHandler, nil)
		handler := middleware(next)
		err := handler.ServeHTTP(w, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "not found", w.Body.String())
		assert.Equal(t, http.StatusNotFound, c.Status())
	})
}

func TestPatternArgsFunc(t *testing.T) {
	t.Run("PatternArgs returns empty for no dynamic pattern", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Pattern = "GET /test"

		argsFunc := PatternArgs()
		args := argsFunc(req)

		assert.Nil(t, args)
	})

	t.Run("PatternArgs returns args for dynamic pattern", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
		req.Pattern = "GET /test/{id}"
		req.SetPathValue("{id}", "123")

		argsFunc := PatternArgs()
		args := argsFunc(req)

		assert.NotNil(t, args)
	})
}
