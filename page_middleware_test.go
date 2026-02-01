package pages

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPageHandlerForMiddleware struct {
	mock.Mock
}

func (m *mockPageHandlerForMiddleware) Handle(w http.ResponseWriter, r *http.Request) error {
	args := m.Called(w, r)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

type mockPageDecoratorStrategyForMiddleware struct {
	mock.Mock
}

func (m *mockPageDecoratorStrategyForMiddleware) IsDecorable(ctx context.Context, pattern string, path string) (bool, error) {
	args := m.Called(ctx, pattern, path)
	return args.Bool(0), args.Error(1)
}

func (m *mockPageDecoratorStrategyForMiddleware) IsPatternDecorable(ctx context.Context, pattern string) (bool, error) {
	args := m.Called(ctx, pattern)
	return args.Bool(0), args.Error(1)
}

func (m *mockPageDecoratorStrategyForMiddleware) IsURIDecorable(ctx context.Context, uri string) (bool, error) {
	args := m.Called(ctx, uri)
	return args.Bool(0), args.Error(1)
}

type mockPageAuthorizerForMiddleware struct {
	mock.Mock
}

func (m *mockPageAuthorizerForMiddleware) Authorize(ctx context.Context, action PageAction) (Decision, error) {
	args := m.Called(ctx, action)
	return args.Get(0).(Decision), args.Error(1)
}

func TestNewPageMiddleware(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		logger := slog.New(slog.DiscardHandler)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, logger)

		assert.NotNil(t, middleware)
		assert.Equal(t, pageHandler, middleware.pageHandler)
		assert.Equal(t, authorizer, middleware.authorizer)
		assert.Equal(t, strategy, middleware.strategy)
		assert.NotNil(t, middleware.logger)
		assert.NotNil(t, middleware.pool)
	})

	t.Run("Nil pageHandler should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPageMiddleware(nil, &mockPageAuthorizerForMiddleware{}, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil authorizer should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPageMiddleware(&mockPageHandlerForMiddleware{}, nil, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil strategy should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPageMiddleware(&mockPageHandlerForMiddleware{}, &mockPageAuthorizerForMiddleware{}, nil, slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil logger uses discard handler", func(t *testing.T) {
		middleware := NewPageMiddleware(&mockPageHandlerForMiddleware{}, &mockPageAuthorizerForMiddleware{}, &mockPageDecoratorStrategyForMiddleware{}, nil)

		assert.NotNil(t, middleware.logger)
	})
}

func TestPageMiddleware_Middleware(t *testing.T) {
	t.Run("Strategy returns not decorable - calls next directly", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("Strategy returns error - calls next directly", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, assert.AnError)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			return nil
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("Nil context should panic", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error { return nil }

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			_ = middleware.Middleware(w, req, next)
		})
	})

	t.Run("No page in context", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error { return nil }

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
	})

	t.Run("Draft page - guest access denied", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Draft
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
	})

	t.Run("Private page - guest access denied", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Private
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
	})

	t.Run("Draft page - authenticated access allowed", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewDraftPage).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Draft
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
	})

	t.Run("Private page - authenticated access allowed", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewPrivatePage).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Published
		page.Visibility = Private
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
	})

	t.Run("Non-hybrid page - calls next directly", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		nextCalled := false
		next := func(w http.ResponseWriter, r *http.Request) error {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = PageCMS
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.True(t, nextCalled)
	})

	t.Run("Decorate true - sets X-Page-Decorable header", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, "1", w.Header().Get(HeaderXPageDecorable))
	})

	t.Run("Decorate false - sets X-Page-Not-Decorable header", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, "1", w.Header().Get(HeaderXPageNotDecorable))
	})

	t.Run("Next returns error", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			return assert.AnError
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Not decorable - writes 204 No Content without buffer", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			return nil
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		req.Header.Set(headerXRequestedWith, xmlHTTPRequest)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Not decorable with buffer - writes content with status", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		req.Header.Set(headerXRequestedWith, xmlHTTPRequest)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "content", w.Body.String())
	})

	t.Run("Decorable with content in buffer", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("decorated content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, template.HTML("decorated content"), c.Content())
	})

	t.Run("Authorizer error for draft page", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewDraftPage).Return(Allow, assert.AnError)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error { return nil }

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Draft
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPageNotFound)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Authorizer error for private page", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewPrivatePage).Return(Allow, assert.AnError)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error { return nil }

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetGuest(false)
		page := NewPage()
		page.Status = Published
		page.Visibility = Private
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPrivatePage)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Not decorable with custom status - writes content", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("not found"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = false
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		req.Header.Set(headerXRequestedWith, xmlHTTPRequest)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "not found", w.Body.String())
	})

	t.Run("Decorable - calls pageHandler", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategyForMiddleware{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, mock.Anything).Return(Allow, nil)

		middleware := NewPageMiddleware(pageHandler, authorizer, strategy, slog.New(slog.DiscardHandler))

		next := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("content"))
			return err
		}

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		site := NewSite()
		c.SetSite(site)
		c.SetSite(site)
		page := NewPage()
		page.Status = Published
		page.Visibility = Public
		page.Pattern = "/test"
		page.Decorate = true
		c.SetPage(page)

		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		err := middleware.Middleware(w, req, next)

		assert.NoError(t, err)
	})
}

func TestPageMiddleware_allow(t *testing.T) {
	t.Run("Guest access - returns error", func(t *testing.T) {
		pageHandler := &mockPageHandlerForMiddleware{}
		authorizer := &mockPageAuthorizerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))

		err := middleware.allow(context.Background(), ViewDraftPage, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guest is not allowed")
	})

	t.Run("Authenticated access - calls authorizer", func(t *testing.T) {
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewDraftPage).Return(Allow, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))

		err := middleware.allow(context.Background(), ViewDraftPage, false)

		assert.NoError(t, err)
		authorizer.AssertCalled(t, "Authorize", mock.Anything, ViewDraftPage)
	})

	t.Run("Authorizer returns error", func(t *testing.T) {
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewDraftPage).Return(Deny, assert.AnError)

		pageHandler := &mockPageHandlerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))

		err := middleware.allow(context.Background(), ViewDraftPage, false)

		assert.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("Authorizer returns Deny", func(t *testing.T) {
		authorizer := &mockPageAuthorizerForMiddleware{}
		authorizer.On("Authorize", mock.Anything, ViewDraftPage).Return(Deny, nil)

		pageHandler := &mockPageHandlerForMiddleware{}
		middleware := NewPageMiddleware(pageHandler, authorizer, &mockPageDecoratorStrategyForMiddleware{}, slog.New(slog.DiscardHandler))

		err := middleware.allow(context.Background(), ViewDraftPage, false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})
}

func TestIsDecorable(t *testing.T) {
	t.Run("Non-HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(headerContentType, "application/json")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("HTML content type", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.status = http.StatusOK
		w.Header().Set(headerContentType, mimeTextHTML)

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.True(t, isDecorable)
	})

	t.Run("X-Page-Not-Decorable header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(HeaderXPageNotDecorable, "1")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("X-Page-Decorable header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.Header().Set(HeaderXPageDecorable, "1")

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.True(t, isDecorable)
	})

	t.Run("Status not OK", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.status = http.StatusOK
		w.WriteHeader(http.StatusNotFound)

		isDecorable := IsDecorable(w, httptest.NewRequest("GET", "/", nil))

		assert.False(t, isDecorable)
	})

	t.Run("XMLHttpRequest header set", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(headerXRequestedWith, xmlHTTPRequest)

		isDecorable := IsDecorable(w, req)

		assert.False(t, isDecorable)
	})

	t.Run("All conditions allow decoration", func(t *testing.T) {
		w := &delayedWriter{ResponseWriter: httptest.NewRecorder()}
		w.status = http.StatusOK
		req := httptest.NewRequest("GET", "/", nil)

		isDecorable := IsDecorable(w, req)

		assert.True(t, isDecorable)
	})
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
		assert.False(t, w.commited)
		assert.Equal(t, http.StatusOK, w.status)
	})

	t.Run("WriteHeader tracks status", func(t *testing.T) {
		w := &delayedWriter{}

		w.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, w.status)
		assert.True(t, w.commited)
	})

	t.Run("Write buffers data", func(t *testing.T) {
		w := &delayedWriter{}
		data := []byte("test data")

		n, err := w.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "test data", w.buffer.String())
		assert.True(t, w.commited)
	})

	t.Run("Write commits if not already committed", func(t *testing.T) {
		w := &delayedWriter{}

		n, err := w.Write([]byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, http.StatusOK, w.status)
		assert.True(t, w.commited)
	})

	t.Run("Unwrap returns ResponseWriter", func(t *testing.T) {
		rw := httptest.NewRecorder()
		w := &delayedWriter{ResponseWriter: rw}

		unwrapped := w.Unwrap()

		assert.Equal(t, rw, unwrapped)
	})
}
