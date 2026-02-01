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
	"github.com/stretchr/testify/mock"
)

func TestErrorPatternFinder(t *testing.T) {
	finder := ErrorPatternFinder()

	t.Run("StatusUnauthorized", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusUnauthorized)
		assert.Equal(t, PageErrorUnauthorized, pattern)
	})

	t.Run("StatusForbidden", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusForbidden)
		assert.Equal(t, PageErrorForbidden, pattern)
	})

	t.Run("StatusNotFound", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusNotFound)
		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("4xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusBadRequest)
		assert.Equal(t, PageError4xx, pattern)

		pattern = finder(context.Background(), http.StatusPaymentRequired)
		assert.Equal(t, PageError4xx, pattern)

		pattern = finder(context.Background(), http.StatusConflict)
		assert.Equal(t, PageError4xx, pattern)
	})

	t.Run("5xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusInternalServerError)
		assert.Equal(t, PageError5xx, pattern)

		pattern = finder(context.Background(), http.StatusBadGateway)
		assert.Equal(t, PageError5xx, pattern)

		pattern = finder(context.Background(), http.StatusServiceUnavailable)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("Other 5xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusNotImplemented)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("2xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusOK)
		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("3xx status", func(t *testing.T) {
		pattern := finder(context.Background(), http.StatusMovedPermanently)
		assert.Equal(t, PageError5xx, pattern)
	})
}

func TestNewDefaultErrorHandler(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		pageHandler := &mockPageHandler{}
		manager := &mockPageManager{}
		authorizer := &mockPageAuthorizer{}
		strategy := &mockPageDecoratorStrategy{}
		statusFinder := func(ctx context.Context, err error) int { return http.StatusNotFound }
		patternFinder := func(ctx context.Context, status int) string { return PageErrorNotFound }
		logger := slog.New(slog.DiscardHandler)

		handler := NewDefaultErrorHandler(pageHandler, manager, authorizer, strategy, statusFinder, patternFinder, logger)

		assert.NotNil(t, handler)
		assert.Equal(t, pageHandler, handler.pageHandler)
		assert.Equal(t, manager, handler.manager)
		assert.Equal(t, authorizer, handler.authorizer)
		assert.Equal(t, strategy, handler.strategy)
		assert.NotNil(t, handler.statusFinder)
		assert.NotNil(t, handler.patternFinder)
	})

	t.Run("Nil pageHandler should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(nil, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
				func(ctx context.Context, err error) int { return 0 },
				func(ctx context.Context, status int) string { return "" },
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil manager should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(&mockPageHandler{}, nil, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
				func(ctx context.Context, err error) int { return 0 },
				func(ctx context.Context, status int) string { return "" },
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil authorizer should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, nil, &mockPageDecoratorStrategy{},
				func(ctx context.Context, err error) int { return 0 },
				func(ctx context.Context, status int) string { return "" },
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil strategy should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, nil,
				func(ctx context.Context, err error) int { return 0 },
				func(ctx context.Context, status int) string { return "" },
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil statusFinder should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
				nil, func(ctx context.Context, status int) string { return "" },
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil patternFinder should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
				func(ctx context.Context, err error) int { return 0 }, nil,
				slog.New(slog.DiscardHandler))
		})
	})

	t.Run("Nil logger uses discard handler", func(t *testing.T) {
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			func(ctx context.Context, err error) int { return 0 },
			func(ctx context.Context, status int) string { return "" }, nil)

		assert.NotNil(t, handler.logger)
	})
}

func TestDefaultErrorHandler_Handle(t *testing.T) {
	t.Run("Nil context should panic", func(t *testing.T) {
		handler := createTestErrorHandler()
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			handler.Handle(w, req, errors.New("test error"))
		})
	})

	t.Run("No site in context", func(t *testing.T) {
		handler := createTestErrorHandler()
		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "500")
	})

	t.Run("Status from finder is less than 400", func(t *testing.T) {
		statusFinder := func(ctx context.Context, err error) int { return http.StatusOK }
		pageHandler := &mockPageHandler{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(NewPage(), nil)

		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			statusFinder, ErrorPatternFinder(), slog.New(slog.DiscardHandler))

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, errors.New("test error"))

		c2 := FromContext(ctx)
		assert.Equal(t, http.StatusInternalServerError, c2.Status())
	})

	t.Run("Manager GetByPattern returns error", func(t *testing.T) {
		loggerBuf := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(loggerBuf, nil))

		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(nil, ErrPageNotFound)

		handler := NewDefaultErrorHandler(&mockPageHandler{}, manager, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return http.StatusNotFound },
			ErrorPatternFinder(), logger)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, errors.New("test error"))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, loggerBuf.String(), "find page by pattern return error")
	})

	t.Run("PageHandler returns error", func(t *testing.T) {
		loggerBuf := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(loggerBuf, nil))

		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		pageHandler := &mockPageHandler{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(assert.AnError)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(NewPage(), nil)

		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return http.StatusNotFound },
			ErrorPatternFinder(), logger)

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, errors.New("test error"))

		assert.Contains(t, loggerBuf.String(), "page handler return error")
	})

	t.Run("Success path", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		pageHandler := &mockPageHandler{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, mock.Anything).Return(NewPage(), nil)

		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return http.StatusNotFound },
			ErrorPatternFinder(), slog.New(slog.DiscardHandler))

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		testErr := errors.New("test error")
		handler.Handle(w, req, testErr)

		c2 := FromContext(ctx)
		assert.Equal(t, testErr, c2.Error())
		assert.Equal(t, http.StatusNotFound, c2.Status())
		assert.True(t, c2.HasPage())
	})
}

func TestDefaultErrorHandler_getPattern(t *testing.T) {
	t.Run("Status is 404, IsDecorable true, Authorize Allow", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		authorizer := &mockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow, nil)

		handler := createTestErrorHandler()
		handler.strategy = strategy
		handler.authorizer = authorizer

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusNotFound)

		assert.Equal(t, PageInternalCreate, pattern)
	})

	t.Run("Status is 404, IsDecorable true, Authorize Deny", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		authorizer := &mockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny, nil)

		patternFinder := func(ctx context.Context, status int) string { return PageErrorNotFound }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, authorizer, strategy,
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusNotFound)

		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("Status is 404, IsDecorable false", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		patternFinder := func(ctx context.Context, status int) string { return PageErrorNotFound }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusNotFound)

		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("Status is 404, IsDecorable returns error", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, assert.AnError)

		patternFinder := func(ctx context.Context, status int) string { return PageErrorNotFound }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusNotFound)

		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("Status is 404, Authorize returns error", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		authorizer := &mockPageAuthorizer{}
		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny, assert.AnError)

		patternFinder := func(ctx context.Context, status int) string { return PageErrorNotFound }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, authorizer, strategy,
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusNotFound)

		assert.Equal(t, PageErrorNotFound, pattern)
	})

	t.Run("Status is not 404", func(t *testing.T) {
		handler := createTestErrorHandler()

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusInternalServerError)

		assert.Equal(t, PageError5xx, pattern)
	})

	t.Run("Status is not 404, patternFinder returns custom pattern", func(t *testing.T) {
		patternFinder := func(ctx context.Context, status int) string { return PageErrorForbidden }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusForbidden)

		assert.Equal(t, PageErrorForbidden, pattern)
	})

	t.Run("patternFinder returns empty string", func(t *testing.T) {
		patternFinder := func(ctx context.Context, status int) string { return "" }
		handler := NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			func(ctx context.Context, err error) int { return 0 }, patternFinder, slog.New(slog.DiscardHandler))

		req := httptest.NewRequest("GET", "/test", nil)

		pattern := handler.getPattern(req, http.StatusInternalServerError)

		assert.Equal(t, PageError5xx, pattern)
	})
}

func TestDefaultErrorHandler_internal(t *testing.T) {
	t.Run("Internal error page rendering", func(t *testing.T) {
		handler := createTestErrorHandler()

		w := httptest.NewRecorder()
		err := errors.New("test error")

		handler.internal(context.Background(), w, http.StatusInternalServerError, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
		assert.Contains(t, w.Body.String(), "500")
	})

	t.Run("Template execution error is logged", func(t *testing.T) {
		loggerBuf := bytes.NewBuffer(nil)
		logger := slog.New(slog.NewTextHandler(loggerBuf, nil))

		pageHandler := &mockPageHandler{}
		manager := &mockPageManager{}
		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			func(ctx context.Context, err error) int { return 0 },
			func(ctx context.Context, status int) string { return "" }, logger)

		w := &mockErrorResponseWriter{}
		err := errors.New("test error")

		handler.internal(context.Background(), w, http.StatusInternalServerError, err)

		assert.Contains(t, loggerBuf.String(), "write response error")
	})
}

func TestDefaultErrorHandler_ErrorHandlerInterface(t *testing.T) {
	t.Run("Implements ErrorHandler", func(t *testing.T) {
		handler := createTestErrorHandler()

		var _ ErrorHandler = handler
	})
}

func TestErrorTemplate(t *testing.T) {
	t.Run("ErrorTemplate is not nil", func(t *testing.T) {
		assert.NotNil(t, ErrorTemplate)
	})

	t.Run("ErrorTemplate can be executed", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]any{
			"Title":  "Not Found",
			"Status": 404,
		}

		err := ErrorTemplate.Execute(w, data)

		assert.NoError(t, err)
		assert.Contains(t, w.Body.String(), "Not Found")
		assert.Contains(t, w.Body.String(), "404")
	})
}

func TestDefaultErrorHandler_IntegrationScenarios(t *testing.T) {
	t.Run("Full error handling flow with 404", func(t *testing.T) {
		strategy := &mockPageDecoratorStrategy{}
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		pageHandler := &mockPageHandler{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageErrorNotFound).Return(NewPage(), nil)

		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, strategy,
			func(ctx context.Context, err error) int { return http.StatusNotFound },
			ErrorPatternFinder(), slog.New(slog.DiscardHandler))

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/notfound", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, ErrPageNotFound)

		assert.True(t, c.HasError())
		assert.Equal(t, http.StatusNotFound, c.Status())
	})

	t.Run("Full error handling flow with 500", func(t *testing.T) {
		pageHandler := &mockPageHandler{}
		pageHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

		manager := &mockPageManager{}
		manager.On("GetByPattern", mock.Anything, mock.Anything, PageError5xx).Return(NewPage(), nil)

		handler := NewDefaultErrorHandler(pageHandler, manager, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
			func(ctx context.Context, err error) int { return http.StatusInternalServerError },
			ErrorPatternFinder(), slog.New(slog.DiscardHandler))

		parentCtx := context.Background()
		ctx, _ := NewContext(parentCtx)
		c := FromContext(ctx)
		c.SetSite(NewSite())
		req := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.Handle(w, req, errors.New("internal error"))

		assert.True(t, c.HasError())
		assert.Equal(t, http.StatusInternalServerError, c.Status())
	})
}

func createTestErrorHandler() *DefaultErrorHandler {
	return NewDefaultErrorHandler(&mockPageHandler{}, &mockPageManager{}, &mockPageAuthorizer{}, &mockPageDecoratorStrategy{},
		func(ctx context.Context, err error) int { return 0 },
		ErrorPatternFinder(), slog.New(slog.DiscardHandler))
}

type mockPageHandler struct {
	mock.Mock
}

func (m *mockPageHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	args := m.Called(w, r)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

type mockPageManager struct {
	mock.Mock
}

func (m *mockPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	args := m.Called(ctx, site, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	args := m.Called(ctx, site, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *mockPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	args := m.Called(ctx, site, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

type mockPageAuthorizer struct {
	mock.Mock
}

func (m *mockPageAuthorizer) Authorize(ctx context.Context, action PageAction) (Decision, error) {
	args := m.Called(ctx, action)
	return args.Get(0).(Decision), args.Error(1)
}

type mockPageDecoratorStrategy struct {
	mock.Mock
}

func (m *mockPageDecoratorStrategy) IsDecorable(ctx context.Context, pattern string, path string) (bool, error) {
	args := m.Called(ctx, pattern, path)
	return args.Bool(0), args.Error(1)
}

func (m *mockPageDecoratorStrategy) IsPatternDecorable(ctx context.Context, pattern string) (bool, error) {
	args := m.Called(ctx, pattern)
	return args.Bool(0), args.Error(1)
}

func (m *mockPageDecoratorStrategy) IsURIDecorable(ctx context.Context, uri string) (bool, error) {
	args := m.Called(ctx, uri)
	return args.Bool(0), args.Error(1)
}

type mockErrorResponseWriter struct {
	http.ResponseWriter
}

func (w *mockErrorResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *mockErrorResponseWriter) WriteHeader(statusCode int) {}

func (w *mockErrorResponseWriter) Write(b []byte) (int, error) {
	return 0, assert.AnError
}
