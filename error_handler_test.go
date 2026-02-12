package pages

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/keratin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestErrorHandlerConfig_SetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *ErrorHandlerConfig
		assert func(t *testing.T, cfg *ErrorHandlerConfig)
	}{
		{
			name: "empty config gets defaults",
			cfg:  &ErrorHandlerConfig{},
			assert: func(t *testing.T, cfg *ErrorHandlerConfig) {
				assert.Equal(t, "error.gohtml", cfg.FallbackTemplate)
				assert.NotNil(t, cfg.StatusFunc)
				assert.NotNil(t, cfg.JSONHandler)
				assert.NotNil(t, cfg.Logger)
			},
		},
		{
			name: "partial config gets defaults for missing fields",
			cfg: &ErrorHandlerConfig{
				FallbackTemplate: "custom.html",
			},
			assert: func(t *testing.T, cfg *ErrorHandlerConfig) {
				assert.Equal(t, "custom.html", cfg.FallbackTemplate)
				assert.NotNil(t, cfg.StatusFunc)
				assert.NotNil(t, cfg.JSONHandler)
				assert.NotNil(t, cfg.Logger)
			},
		},
		{
			name: "fully populated config keeps values",
			cfg: &ErrorHandlerConfig{
				FallbackTemplate: "custom.html",
				StatusFunc:       func(context.Context, error) int { return 200 },
				JSONHandler:      http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
				Logger:           slog.New(slog.DiscardHandler),
			},
			assert: func(t *testing.T, cfg *ErrorHandlerConfig) {
				assert.Equal(t, "custom.html", cfg.FallbackTemplate)
				assert.NotNil(t, cfg.StatusFunc)
				assert.NotNil(t, cfg.JSONHandler)
				assert.NotNil(t, cfg.Logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.SetDefaults()
			tt.assert(t, tt.cfg)
		})
	}
}

func TestErrorHandlerConfig_jsonHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func(*Context)
		accept       string
		wantCode     int
		wantContains []string
	}{
		{
			name: "basic error without site",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusNotFound)
				c.SetError(ErrPageNotFound)
			},
			wantCode: http.StatusNotFound,
			wantContains: []string{
				`"code":404`,
				`"message":"Not Found"`,
			},
		},
		{
			name: "error with site",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusNotFound)
				c.SetError(ErrPageNotFound)
				c.SetSite(&Site{ID: "site-1", Name: "Test Site"})
			},
			wantCode: http.StatusNotFound,
			wantContains: []string{
				`"code":404`,
				`"message":"Not Found"`,
				`"site":{"id":"site-1","name":"Test Site"`,
			},
		},
		{
			name: "HTTPError with custom message",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusBadRequest)
				c.SetError(&keratin.HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Custom error message",
				})
			},
			wantCode: http.StatusBadRequest,
			wantContains: []string{
				`"code":400`,
				`"message":"Custom error message"`,
			},
		},
		{
			name: "unprocessable entity error includes error data",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusUnprocessableEntity)
				c.SetError(&keratin.HTTPError{
					Code: http.StatusUnprocessableEntity,
				})
			},
			wantCode: http.StatusUnprocessableEntity,
			wantContains: []string{
				`"code":422`,
			},
		},
		{
			name: "debug mode includes error details",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusInternalServerError)
				c.SetError(assert.AnError)
				c.SetDebug(true)
			},
			wantCode: http.StatusInternalServerError,
			wantContains: []string{
				`"code":500`,
				`"message":"Internal Server Error"`,
				`"error":"assert.AnError general error for testing"`,
			},
		},
		{
			name: "non-debug mode excludes error details",
			setupContext: func(c *Context) {
				c.SetStatus(http.StatusInternalServerError)
				c.SetError(assert.AnError)
				c.SetDebug(false)
			},
			wantCode: http.StatusInternalServerError,
			wantContains: []string{
				`"code":500`,
				`"message":"Internal Server Error"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
			cfg.SetDefaults()

			ctx, cancel := NewContext(context.Background())
			defer cancel()

			c := MustContext(ctx)
			tt.setupContext(c)

			req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			cfg.jsonHandler(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			body := w.Body.String()
			for _, s := range tt.wantContains {
				assert.Contains(t, body, s)
			}
		})
	}
}

func TestErrorHandler_Panics(t *testing.T) {
	tests := []struct {
		name         string
		pageHandler  keratin.Handler
		manager      PageManager
		errPattern   ErrorPatternFunc
		wantPanic    bool
		panicMessage string
	}{
		{
			name:         "nil pageHandler panics",
			pageHandler:  nil,
			manager:      &MockPageManager{},
			errPattern:   func(r *http.Request, status int, err error) string { return "" },
			wantPanic:    true,
			panicMessage: "http error handler: page handler is required",
		},
		{
			name:         "nil manager panics",
			pageHandler:  &mockHandler{},
			manager:      nil,
			errPattern:   func(r *http.Request, status int, err error) string { return "" },
			wantPanic:    true,
			panicMessage: "http error handler: page manager is required",
		},
		{
			name:         "nil errPattern panics",
			pageHandler:  &mockHandler{},
			manager:      &MockPageManager{},
			errPattern:   nil,
			wantPanic:    true,
			panicMessage: "http error handler: error pattern is required",
		},
		{
			name:        "valid parameters does not panic",
			pageHandler: &mockHandler{},
			manager:     &MockPageManager{},
			errPattern:  func(r *http.Request, status int, err error) string { return "" },
			wantPanic:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ErrorHandlerConfig{}

			if tt.wantPanic {
				assert.PanicsWithValue(t, tt.panicMessage, func() {
					ErrorHandler(cfg, tt.pageHandler, tt.manager, tt.errPattern)
				})
			} else {
				assert.NotPanics(t, func() {
					ErrorHandler(cfg, tt.pageHandler, tt.manager, tt.errPattern)
				})
			}
		})
	}
}

func TestHandler_ErrorHandler_Committed(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	errPattern := func(r *http.Request, status int, err error) string { return "" }

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := &mockResponseRecorder{ResponseRecorder: httptest.NewRecorder(), committed: true}

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusOK, w.Code, "Response should remain committed without changes")
}

type mockResponseRecorder struct {
	*httptest.ResponseRecorder
	committed bool
}

func (m *mockResponseRecorder) Header() http.Header {
	return m.ResponseRecorder.Header()
}

func (m *mockResponseRecorder) Write(b []byte) (int, error) {
	if m.committed {
		return 0, nil
	}
	return m.ResponseRecorder.Write(b)
}

func (m *mockResponseRecorder) WriteHeader(statusCode int) {
	if m.committed {
		return
	}
	m.ResponseRecorder.WriteHeader(statusCode)
}

func TestHandler_ErrorHandler_HEAD(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	c := MustContext(ctx)
	c.SetSite(&Site{ID: "site-1"})

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	manager := &MockPageManager{}
	errPattern := func(r *http.Request, status int, err error) string { return "" }

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodHead, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestHandler_ErrorHandler_JSON(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	c := MustContext(ctx)
	c.SetSite(&Site{ID: "site-1"})

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, mock.Anything, mock.Anything).Return(&Page{}, nil)
	errPattern := func(r *http.Request, status int, err error) string { return "" }

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	req.Header.Set(keratin.HeaderAccept, keratin.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
}

func TestHandler_ErrorHandler_NoSite(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	errPattern := func(r *http.Request, status int, err error) string { return "" }

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusOK, w.Code)
	handler.AssertCalled(t, "ServeHTTP", w, req)
}

func TestHandler_ErrorHandler_SuccessfulFlow(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	site := &Site{ID: "site-1"}
	c := MustContext(ctx)
	c.SetSite(site)

	errorPage := &Page{ID: "error-page", Pattern: PageError5xx}

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, site, PageError5xx).Return(errorPage, nil)

	errPattern := func(r *http.Request, status int, err error) string {
		return PageError5xx
	}

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusOK, w.Code)
	handler.AssertCalled(t, "ServeHTTP", w, req)
	manager.AssertCalled(t, "GetByPattern", ctx, site, PageError5xx)
}

func TestHandler_ErrorHandler_ManagerError(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	site := &Site{ID: "site-1"}
	c := MustContext(ctx)
	c.SetSite(site)

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, site, mock.Anything).Return(nil, assert.AnError)

	errPattern := func(r *http.Request, status int, err error) string {
		return PageError5xx
	}

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusOK, w.Code)
	handler.AssertCalled(t, "ServeHTTP", w, req)
}

func TestHandler_ErrorHandler_RequestID(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	c := MustContext(ctx)
	c.SetSite(&Site{ID: "site-1"})

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, mock.Anything, mock.Anything).Return(&Page{}, nil)

	errPattern := func(r *http.Request, status int, err error) string { return "" }

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	tests := []struct {
		name      string
		setupReq  func(*http.Request)
		setupResp func(*httptest.ResponseRecorder)
	}{
		{
			name: "request ID from request header",
			setupReq: func(r *http.Request) {
				r.Header.Set(keratin.HeaderXRequestID, "req-id-123")
			},
			setupResp: func(w *httptest.ResponseRecorder) {
			},
		},
		{
			name: "request ID from response header",
			setupReq: func(r *http.Request) {
			},
			setupResp: func(w *httptest.ResponseRecorder) {
				w.Header().Set(keratin.HeaderXRequestID, "resp-id-456")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
			tt.setupReq(req)
			w := httptest.NewRecorder()
			tt.setupResp(w)

			handlerFunc(w, req, assert.AnError)
		})
	}
}

func TestHandler_ErrorHandler_StatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		pattern  string
	}{
		{
			name:     "404 error",
			err:      ErrPageNotFound,
			wantCode: http.StatusNotFound,
			pattern:  PageErrorNotFound,
		},
		{
			name:     "403 error",
			err:      ErrPageForbidden,
			wantCode: http.StatusForbidden,
			pattern:  PageErrorForbidden,
		},
		{
			name:     "401 error",
			err:      ErrPageUnauthorized,
			wantCode: http.StatusUnauthorized,
			pattern:  PageErrorUnauthorized,
		},
		{
			name:     "409 error",
			err:      ErrUniqueViolation,
			wantCode: http.StatusConflict,
			pattern:  PageError4xx,
		},
		{
			name:     "500 error",
			err:      assert.AnError,
			wantCode: http.StatusInternalServerError,
			pattern:  PageError5xx,
		},
		{
			name:     "sql.ErrNoRows",
			err:      sql.ErrNoRows,
			wantCode: http.StatusNotFound,
			pattern:  PageErrorNotFound,
		},
		{
			name:     "keratin.ErrFileNotFound",
			err:      keratin.ErrFileNotFound,
			wantCode: http.StatusNotFound,
			pattern:  PageErrorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := NewContext(context.Background())
			defer cancel()

			site := &Site{ID: "site-1"}
			c := MustContext(ctx)
			c.SetSite(site)

			errorPage := &Page{ID: "error-page", Pattern: tt.pattern}

			cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
			handler := &mockHandler{}
			handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
			manager := &MockPageManager{}
			manager.On("GetByPattern", ctx, site, tt.pattern).Return(errorPage, nil)

			errPattern := func(r *http.Request, status int, err error) string {
				return tt.pattern
			}

			handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

			req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
			w := httptest.NewRecorder()

			handlerFunc(w, req, tt.err)

			handler.AssertCalled(t, "ServeHTTP", w, req)
			manager.AssertCalled(t, "GetByPattern", ctx, site, tt.pattern)
		})
	}
}

func TestHandler_ErrorHandler_EmptyPattern(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	site := &Site{ID: "site-1"}
	c := MustContext(ctx)
	c.SetSite(site)

	errorPage := &Page{ID: "error-page", Pattern: PageError5xx}

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, site, PageError5xx).Return(errorPage, nil)

	errPattern := func(r *http.Request, status int, err error) string {
		return ""
	}

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	manager.AssertCalled(t, "GetByPattern", ctx, site, PageError5xx)
}

func TestHandler_ErrorHandler_ContextSettings(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	site := &Site{ID: "site-1"}
	errorPage := &Page{ID: "error-page", Pattern: PageError5xx, Template: "error.html"}

	cfg := ErrorHandlerConfig{
		Logger:           slog.New(slog.DiscardHandler),
		FallbackTemplate: "fallback.html",
		StatusFunc:       ErrorStatus,
	}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)
	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, site, PageError5xx).Return(errorPage, nil)

	errPattern := func(r *http.Request, status int, err error) string {
		return PageError5xx
	}

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handlerFunc(w, req, assert.AnError)

	assert.Equal(t, http.StatusOK, w.Code)
	handler.AssertCalled(t, "ServeHTTP", w, req)
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	args := m.Called(w, r)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func BenchmarkErrorHandler_Flow(b *testing.B) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()

	site := &Site{ID: "site-1"}
	c := MustContext(ctx)
	c.SetSite(site)

	errorPage := &Page{ID: "error-page", Pattern: PageError5xx}

	cfg := ErrorHandlerConfig{Logger: slog.New(slog.DiscardHandler)}
	handler := &mockHandler{}
	handler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)

	manager := &MockPageManager{}
	manager.On("GetByPattern", ctx, site, PageError5xx).Return(errorPage, nil)

	errPattern := func(r *http.Request, status int, err error) string {
		return PageError5xx
	}

	handlerFunc := ErrorHandler(cfg, handler, manager, errPattern)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handlerFunc(w, req, assert.AnError)
	}
}
