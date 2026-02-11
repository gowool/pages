package pages

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type patternError struct {
	pattern string
}

func (e *patternError) Error() string {
	return "pattern error"
}

func (e *patternError) Pattern() string {
	return e.pattern
}

type contextPatternError struct {
	pattern string
}

func (e *contextPatternError) Error() string {
	return "context pattern error"
}

func (e *contextPatternError) Pattern(*http.Request, int, error) string {
	return e.pattern
}

type wrappingError struct {
	inner error
}

func (e *wrappingError) Error() string {
	return e.inner.Error()
}

func (e *wrappingError) Unwrap() error {
	return e.inner
}

func TestErrorPattern_Panics(t *testing.T) {
	tests := []struct {
		name              string
		authorizer        PageAuthorizer
		decoratorStrategy PageDecoratorStrategy
		wantPanic         bool
		panicMessage      string
	}{
		{
			name:              "nil authorizer panics",
			authorizer:        nil,
			decoratorStrategy: &MockPageDecoratorStrategy{},
			wantPanic:         true,
			panicMessage:      "error pattern: authorizer is required",
		},
		{
			name:              "nil decoratorStrategy panics",
			authorizer:        &MockPageAuthorizer{},
			decoratorStrategy: nil,
			wantPanic:         true,
			panicMessage:      "error pattern: decorator strategy is required",
		},
		{
			name:              "valid parameters does not panic",
			authorizer:        &MockPageAuthorizer{},
			decoratorStrategy: &MockPageDecoratorStrategy{},
			wantPanic:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.PanicsWithValue(t, tt.panicMessage, func() {
					ErrorPattern(tt.authorizer, tt.decoratorStrategy)
				})
			} else {
				assert.NotPanics(t, func() {
					ErrorPattern(tt.authorizer, tt.decoratorStrategy)
				})
			}
		})
	}
}

func TestErrorPattern_CustomPattern(t *testing.T) {
	authorizer := &MockPageAuthorizer{}
	decoratorStrategy := &MockPageDecoratorStrategy{}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "error with Pattern() method",
			err:  &patternError{pattern: "custom-pattern"},
			want: "custom-pattern",
		},
		{
			name: "error with Pattern(*http.Request, int, error) method",
			err:  &contextPatternError{pattern: "context-pattern"},
			want: "context-pattern",
		},
		{
			name: "wrapped error with Pattern() method",
			err:  &wrappingError{inner: &patternError{pattern: "wrapped-pattern"}},
			want: "wrapped-pattern",
		},
		{
			name: "doubly wrapped error with Pattern() method",
			err:  &wrappingError{inner: &wrappingError{inner: &patternError{pattern: "deeply-wrapped-pattern"}}},
			want: "deeply-wrapped-pattern",
		},
		{
			name: "wrapped error with context Pattern() method",
			err:  &wrappingError{inner: &contextPatternError{pattern: "wrapped-context-pattern"}},
			want: "wrapped-context-pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := ErrorPattern(authorizer, decoratorStrategy)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			got := fn(req, http.StatusInternalServerError, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorPattern_ErrPageNotFound(t *testing.T) {
	tests := []struct {
		name         string
		authDecision Decision
		decorable    bool
		want         string
	}{
		{
			name:         "ErrPageNotFound with Allow and decorable returns PageInternalCreate",
			authDecision: Allow,
			decorable:    true,
			want:         PageInternalCreate,
		},
		{
			name:         "ErrPageNotFound with Deny and decorable returns PageErrorNotFound",
			authDecision: Deny,
			decorable:    true,
			want:         PageErrorNotFound,
		},
		{
			name:         "ErrPageNotFound with Allow and not decorable returns PageErrorNotFound",
			authDecision: Allow,
			decorable:    false,
			want:         PageErrorNotFound,
		},
		{
			name:         "ErrPageNotFound with Deny and not decorable returns PageErrorNotFound",
			authDecision: Deny,
			decorable:    false,
			want:         PageErrorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := &MockPageAuthorizer{}
			authorizer.On("Authorize", mock.Anything, CreatePage).Return(tt.authDecision)

			decoratorStrategy := &MockPageDecoratorStrategy{}
			decoratorStrategy.On("IsURIDecorable", mock.Anything, "/test").Return(tt.decorable)

			fn := ErrorPattern(authorizer, decoratorStrategy)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			got := fn(req, http.StatusNotFound, ErrPageNotFound)

			assert.Equal(t, tt.want, got)
			decoratorStrategy.AssertCalled(t, "IsURIDecorable", mock.Anything, "/test")

			if tt.decorable {
				authorizer.AssertCalled(t, "Authorize", mock.Anything, CreatePage)
			} else {
				authorizer.AssertNotCalled(t, "Authorize")
			}
		})
	}
}

func TestErrorPattern_StatusCodes(t *testing.T) {
	authorizer := &MockPageAuthorizer{}
	authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)

	decoratorStrategy := &MockPageDecoratorStrategy{}
	decoratorStrategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(false)

	fn := ErrorPattern(authorizer, decoratorStrategy)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	err := errors.New("some error")

	tests := []struct {
		name   string
		status int
		want   string
	}{
		{
			name:   "401 returns PageErrorUnauthorized",
			status: http.StatusUnauthorized,
			want:   PageErrorUnauthorized,
		},
		{
			name:   "403 returns PageErrorForbidden",
			status: http.StatusForbidden,
			want:   PageErrorForbidden,
		},
		{
			name:   "404 returns PageErrorNotFound",
			status: http.StatusNotFound,
			want:   PageErrorNotFound,
		},
		{
			name:   "400 returns PageError4xx",
			status: http.StatusBadRequest,
			want:   PageError4xx,
		},
		{
			name:   "402 returns PageError4xx",
			status: 402,
			want:   PageError4xx,
		},
		{
			name:   "499 returns PageError4xx",
			status: 499,
			want:   PageError4xx,
		},
		{
			name:   "500 returns PageError5xx",
			status: http.StatusInternalServerError,
			want:   PageError5xx,
		},
		{
			name:   "502 returns PageError5xx",
			status: http.StatusBadGateway,
			want:   PageError5xx,
		},
		{
			name:   "599 returns PageError5xx",
			status: 599,
			want:   PageError5xx,
		},
		{
			name:   "200 returns PageError5xx (2xx falls to 5xx)",
			status: http.StatusOK,
			want:   PageError5xx,
		},
		{
			name:   "300 returns PageError5xx (3xx falls to 5xx)",
			status: http.StatusMultipleChoices,
			want:   PageError5xx,
		},
		{
			name:   "600 returns PageError5xx (out of range falls to 5xx)",
			status: 600,
			want:   PageError5xx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(req, tt.status, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorPattern_Priority(t *testing.T) {
	authorizer := &MockPageAuthorizer{}
	authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

	decoratorStrategy := &MockPageDecoratorStrategy{}
	decoratorStrategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(true)

	fn := ErrorPattern(authorizer, decoratorStrategy)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tests := []struct {
		name   string
		status int
		err    error
		want   string
	}{
		{
			name:   "custom pattern takes priority over ErrPageNotFound",
			status: http.StatusNotFound,
			err:    &patternError{pattern: "custom"},
			want:   "custom",
		},
		{
			name:   "custom pattern takes priority over status 401",
			status: http.StatusUnauthorized,
			err:    &contextPatternError{pattern: "custom-401"},
			want:   "custom-401",
		},
		{
			name:   "ErrPageNotFound with Allow returns PageInternalCreate",
			status: http.StatusNotFound,
			err:    ErrPageNotFound,
			want:   PageInternalCreate,
		},
		{
			name:   "status 401 returns PageErrorUnauthorized when no custom pattern",
			status: http.StatusUnauthorized,
			err:    errors.New("other error"),
			want:   PageErrorUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(req, tt.status, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorPattern_WrappingUnwrapping(t *testing.T) {
	authorizer := &MockPageAuthorizer{}
	decoratorStrategy := &MockPageDecoratorStrategy{}
	decoratorStrategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(false)

	fn := ErrorPattern(authorizer, decoratorStrategy)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "pattern error wrapped in standard errors",
			err:  fmt.Errorf("wrapped: %w", &patternError{pattern: "inner-pattern"}),
			want: "inner-pattern",
		},
		{
			name: "multiple levels of wrapping",
			err:  &wrappingError{inner: &wrappingError{inner: &patternError{pattern: "deep-pattern"}}},
			want: "deep-pattern",
		},
		{
			name: "mixed wrapping with errors.Wrap and custom wrapper",
			err:  &wrappingError{inner: fmt.Errorf("standard: %w", &patternError{pattern: "mixed-pattern"})},
			want: "mixed-pattern",
		},
		{
			name: "wrapped ErrPageNotFound",
			err:  fmt.Errorf("wrapped: %w", ErrPageNotFound),
			want: PageErrorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(req, http.StatusNotFound, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorPattern_EmptyRequest(t *testing.T) {
	authorizer := &MockPageAuthorizer{}
	authorizer.On("Authorize", mock.Anything, CreatePage).Return(Deny)

	decoratorStrategy := &MockPageDecoratorStrategy{}
	decoratorStrategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(false)

	fn := ErrorPattern(authorizer, decoratorStrategy)

	tests := []struct {
		name   string
		req    *http.Request
		status int
		err    error
		want   string
	}{
		{
			name:   "nil request",
			req:    nil,
			status: http.StatusNotFound,
			err:    errors.New("test"),
			want:   PageErrorNotFound,
		},
		{
			name:   "empty request",
			req:    &http.Request{},
			status: http.StatusNotFound,
			err:    errors.New("test"),
			want:   PageErrorNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.req, tt.status, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkErrorPattern_CustomPattern(b *testing.B) {
	authorizer := &MockPageAuthorizer{}
	decoratorStrategy := &MockPageDecoratorStrategy{}
	fn := ErrorPattern(authorizer, decoratorStrategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	err := &patternError{pattern: "custom"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(req, http.StatusInternalServerError, err)
	}
}

func BenchmarkErrorPattern_ErrPageNotFound(b *testing.B) {
	authorizer := &MockPageAuthorizer{}
	authorizer.On("Authorize", mock.Anything, CreatePage).Return(Allow)

	decoratorStrategy := &MockPageDecoratorStrategy{}
	decoratorStrategy.On("IsURIDecorable", mock.Anything, "/test").Return(true)

	fn := ErrorPattern(authorizer, decoratorStrategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(req, http.StatusNotFound, ErrPageNotFound)
	}
}

func BenchmarkErrorPattern_StatusCodes(b *testing.B) {
	authorizer := &MockPageAuthorizer{}
	decoratorStrategy := &MockPageDecoratorStrategy{}
	fn := ErrorPattern(authorizer, decoratorStrategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	err := errors.New("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(req, http.StatusInternalServerError, err)
	}
}
