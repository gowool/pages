package pages

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testPatternError struct {
	pattern string
}

func (e *testPatternError) Error() string {
	return "test error"
}

func (e *testPatternError) Pattern() string {
	return e.pattern
}

type testContextPatternError struct {
	pattern string
}

func (e *testContextPatternError) Error() string {
	return "test error with context"
}

func (e *testContextPatternError) Pattern(r *http.Request, status int, err error) string {
	return e.pattern
}

type testUnwrapError struct {
	base error
}

func (e *testUnwrapError) Error() string {
	return "wrapper error"
}

func (e *testUnwrapError) Unwrap() error {
	return e.base
}

func TestNewHTTPErrorPattern(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)

		pattern := NewHTTPErrorPattern(authorizer, strategy)

		assert.NotNil(t, pattern)
		assert.Equal(t, authorizer, pattern.authorizer)
		assert.Equal(t, strategy, pattern.decoratorStrategy)
	})

	t.Run("panics when authorizer is nil", func(t *testing.T) {
		strategy := NewMockPageDecoratorStrategy(true)

		assert.Panics(t, func() {
			NewHTTPErrorPattern(nil, strategy)
		})
	})

	t.Run("panics when decoratorStrategy is nil", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()

		assert.Panics(t, func() {
			NewHTTPErrorPattern(authorizer, nil)
		})
	})

	t.Run("panics when both are nil", func(t *testing.T) {
		assert.Panics(t, func() {
			NewHTTPErrorPattern(nil, nil)
		})
	})
}

func TestHTTPErrorPattern_Pattern(t *testing.T) {
	t.Run("error with Pattern() method", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := &testPatternError{pattern: "/custom/pattern"}

		result := pattern.Pattern(req, http.StatusNotFound, err)

		assert.Equal(t, "/custom/pattern", result)
	})

	t.Run("error with Pattern(r, status, err) method", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := &testContextPatternError{pattern: "/context/pattern"}

		result := pattern.Pattern(req, http.StatusInternalServerError, err)

		assert.Equal(t, "/context/pattern", result)
	})

	t.Run("unwraps error chain to find pattern", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		innerErr := &testPatternError{pattern: "/inner/pattern"}
		err := &testUnwrapError{base: innerErr}

		result := pattern.Pattern(req, http.StatusNotFound, err)

		assert.Equal(t, "/inner/pattern", result)
	})

	t.Run("multiple levels of error unwrapping", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		deepestErr := &testPatternError{pattern: "/deepest/pattern"}
		level2 := &testUnwrapError{base: deepestErr}
		level1 := &testUnwrapError{base: level2}

		result := pattern.Pattern(req, http.StatusNotFound, level1)

		assert.Equal(t, "/deepest/pattern", result)
	})

	t.Run("ErrPageNotFound with authorization allows CreatePage", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.Background())

		result := pattern.Pattern(req, http.StatusNotFound, ErrPageNotFound)

		assert.Equal(t, PageInternalCreate, result)
		authorizer.AssertCalled(t, "Authorize", mock.Anything, CreatePage)
	})

	t.Run("ErrPageNotFound with authorization denies CreatePage", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.Background())

		result := pattern.Pattern(req, http.StatusNotFound, ErrPageNotFound)

		assert.Equal(t, PageErrorNotFound, result)
		authorizer.AssertCalled(t, "Authorize", mock.Anything, CreatePage)
	})

	t.Run("ErrPageNotFound with URI not decorable", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(false)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.Background())

		result := pattern.Pattern(req, http.StatusNotFound, ErrPageNotFound)

		assert.Equal(t, PageErrorNotFound, result)
	})

	t.Run("ErrPageNotFound wrapped with authorization allows CreatePage", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.Background())
		err := errors.New("wrapper: page not found")

		result := pattern.Pattern(req, http.StatusNotFound, errors.Join(err, ErrPageNotFound))

		assert.Equal(t, PageInternalCreate, result)
	})

	t.Run("status 401 Unauthorized", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("unauthorized access")

		result := pattern.Pattern(req, http.StatusUnauthorized, err)

		assert.Equal(t, PageErrorUnauthorized, result)
	})

	t.Run("status 403 Forbidden", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("forbidden access")

		result := pattern.Pattern(req, http.StatusForbidden, err)

		assert.Equal(t, PageErrorForbidden, result)
	})

	t.Run("status 404 Not Found", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("not found")

		result := pattern.Pattern(req, http.StatusNotFound, err)

		assert.Equal(t, PageErrorNotFound, result)
	})

	t.Run("status 400 Bad Request", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("bad request")

		result := pattern.Pattern(req, http.StatusBadRequest, err)

		assert.Equal(t, PageError4xx, result)
	})

	t.Run("status 409 Conflict", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("conflict")

		result := pattern.Pattern(req, http.StatusConflict, err)

		assert.Equal(t, PageError4xx, result)
	})

	t.Run("status 422 Unprocessable Entity", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("unprocessable")

		result := pattern.Pattern(req, http.StatusUnprocessableEntity, err)

		assert.Equal(t, PageError4xx, result)
	})

	t.Run("status 500 Internal Server Error", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("internal server error")

		result := pattern.Pattern(req, http.StatusInternalServerError, err)

		assert.Equal(t, PageError5xx, result)
	})

	t.Run("status 502 Bad Gateway", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("bad gateway")

		result := pattern.Pattern(req, http.StatusBadGateway, err)

		assert.Equal(t, PageError5xx, result)
	})

	t.Run("status 503 Service Unavailable", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := errors.New("service unavailable")

		result := pattern.Pattern(req, http.StatusServiceUnavailable, err)

		assert.Equal(t, PageError5xx, result)
	})

	t.Run("pattern method takes precedence over status code", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := &testPatternError{pattern: "/custom/error"}

		result := pattern.Pattern(req, http.StatusInternalServerError, err)

		assert.Equal(t, "/custom/error", result)
		assert.NotEqual(t, PageError5xx, result)
	})

	t.Run("context pattern method is used instead of status code", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		err := &testContextPatternError{pattern: "/context/pattern"}

		result := pattern.Pattern(req, http.StatusNotFound, err)

		assert.Equal(t, "/context/pattern", result)
		assert.NotEqual(t, PageErrorNotFound, result)
	})

	t.Run("nil error returns status-based pattern", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		result := pattern.Pattern(req, http.StatusNotFound, nil)

		assert.Equal(t, PageErrorNotFound, result)
	})

	t.Run("wrapped ErrPageNotFound with authorization denied", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.Background())
		err := errors.New("wrapper: " + ErrPageNotFound.Error())

		result := pattern.Pattern(req, http.StatusNotFound, err)

		assert.Equal(t, PageErrorNotFound, result)
	})

	t.Run("error chain with both pattern and ErrPageNotFound", func(t *testing.T) {
		authorizer := NewMockPageAuthorizer()
		strategy := NewMockPageDecoratorStrategy(true)
		pattern := NewHTTPErrorPattern(authorizer, strategy)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		baseErr := &testUnwrapError{base: &testPatternError{pattern: "/custom/pattern"}}

		result := pattern.Pattern(req, http.StatusNotFound, baseErr)

		assert.Equal(t, "/custom/pattern", result)
	})
}

func TestHTTPErrorPattern_Interface(t *testing.T) {
	t.Run("implements ErrorPattern interface", func(t *testing.T) {
		var _ ErrorPattern = (*HTTPErrorPattern)(nil)
	})
}
