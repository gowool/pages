package pages

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/gowool/r"
	"github.com/invopop/validation"
	"github.com/stretchr/testify/assert"
)

type testStatusCodeError struct {
	code int
}

func (e *testStatusCodeError) Error() string {
	return "test status code error"
}

func (e *testStatusCodeError) StatusCode() int {
	return e.code
}

type testStatusError struct {
	code int
}

func (e *testStatusError) Error() string {
	return "test status error"
}

func (e *testStatusError) Status() int {
	return e.code
}

type testContextStatusError struct {
	code int
}

func (e *testContextStatusError) Error() string {
	return "test context status error"
}

func (e *testContextStatusError) Status(context.Context, error) int {
	return e.code
}

type testStatusWrapper struct {
	base error
}

func (e *testStatusWrapper) Error() string {
	return "wrapper error"
}

func (e *testStatusWrapper) Unwrap() error {
	return e.base
}

type testBothStatusError struct {
	testStatusCodeError
	testStatusError
}

func (e *testBothStatusError) Error() string {
	return e.testStatusCodeError.Error()
}

type testContextAwareStatusError struct {
	code int
}

func (e *testContextAwareStatusError) Error() string {
	return "context aware error"
}

func (e *testContextAwareStatusError) Status(ctx context.Context, _ error) int {
	if ctx.Value(testContextAwareStatusError{}) != nil {
		return http.StatusAccepted
	}
	return e.code
}

type testSimpleStatus struct {
	code int
}

func (e *testSimpleStatus) Error() string {
	return "simple status"
}

func (e *testSimpleStatus) Status() int {
	return e.code
}

func TestErrorStatus(t *testing.T) {
	t.Run("ErrPageNotFound returns 404", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, ErrPageNotFound)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("sql.ErrNoRows returns 404", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, sql.ErrNoRows)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("r.ErrFileNotFound returns 404", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, r.ErrFileNotFound)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("ErrPageForbidden returns 403", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, ErrPageForbidden)

		assert.Equal(t, http.StatusForbidden, status)
	})

	t.Run("ErrPageUnauthorized returns 401", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, ErrPageUnauthorized)

		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("ErrUniqueViolation returns 409", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, ErrUniqueViolation)

		assert.Equal(t, http.StatusConflict, status)
	})

	t.Run("validation errors return 422", func(t *testing.T) {
		ctx := context.Background()
		err := validation.Errors{
			"field": errors.New("validation failed"),
		}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusUnprocessableEntity, status)
	})

	t.Run("empty validation errors return 422", func(t *testing.T) {
		ctx := context.Background()
		err := validation.Errors{}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusUnprocessableEntity, status)
	})

	t.Run("error with StatusCode() method", func(t *testing.T) {
		ctx := context.Background()
		err := &testStatusCodeError{code: http.StatusBadRequest}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("error with Status() method", func(t *testing.T) {
		ctx := context.Background()
		err := &testStatusError{code: http.StatusBadRequest}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("error with Status(ctx, err) method", func(t *testing.T) {
		ctx := context.Background()
		err := &testContextStatusError{code: http.StatusBadRequest}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("RedirectError returns redirect status", func(t *testing.T) {
		ctx := context.Background()
		err := NewRedirectError("/new-location", http.StatusFound)
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusFound, status)
	})

	t.Run("RedirectError with 301 status", func(t *testing.T) {
		ctx := context.Background()
		err := NewRedirectError("/permanent", http.StatusMovedPermanently)
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusMovedPermanently, status)
	})

	t.Run("RedirectError with 307 status", func(t *testing.T) {
		ctx := context.Background()
		err := NewRedirectError("/temporary", http.StatusTemporaryRedirect)
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusTemporaryRedirect, status)
	})

	t.Run("unwraps error chain to find StatusCode()", func(t *testing.T) {
		ctx := context.Background()
		innerErr := &testStatusCodeError{code: http.StatusBadGateway}
		err := &testStatusWrapper{base: innerErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadGateway, status)
	})

	t.Run("unwraps error chain to find Status()", func(t *testing.T) {
		ctx := context.Background()
		innerErr := &testStatusError{code: http.StatusBadGateway}
		err := &testStatusWrapper{base: innerErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadGateway, status)
	})

	t.Run("unwraps error chain to find Status(ctx, err)", func(t *testing.T) {
		ctx := context.Background()
		innerErr := &testContextStatusError{code: http.StatusBadGateway}
		err := &testStatusWrapper{base: innerErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusBadGateway, status)
	})

	t.Run("multiple levels of error unwrapping", func(t *testing.T) {
		ctx := context.Background()
		deepestErr := &testStatusCodeError{code: http.StatusServiceUnavailable}
		level2 := &testStatusWrapper{base: deepestErr}
		level1 := &testStatusWrapper{base: level2}
		status := ErrorStatus(ctx, level1)

		assert.Equal(t, http.StatusServiceUnavailable, status)
	})

	t.Run("StatusCode() takes precedence over Status()", func(t *testing.T) {
		ctx := context.Background()
		err := &testBothStatusError{
			testStatusCodeError: testStatusCodeError{code: http.StatusTeapot},
			testStatusError:     testStatusError{code: http.StatusBadRequest},
		}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusTeapot, status)
	})

	t.Run("simple Status() method", func(t *testing.T) {
		ctx := context.Background()
		err := &testSimpleStatus{code: http.StatusTeapot}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusTeapot, status)
	})

	t.Run("ErrPageNotFound in wrapped error", func(t *testing.T) {
		ctx := context.Background()
		err := &testStatusWrapper{base: ErrPageNotFound}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("sql.ErrNoRows in wrapped error", func(t *testing.T) {
		ctx := context.Background()
		err := &testStatusWrapper{base: sql.ErrNoRows}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("generic error returns 500", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("something went wrong")
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("nil error returns 500", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, nil)

		assert.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("empty string error returns 500", func(t *testing.T) {
		ctx := context.Background()
		err := errors.New("")
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusInternalServerError, status)
	})

	t.Run("context is not used for basic errors", func(t *testing.T) {
		ctx1 := context.Background()
		ctx2 := context.TODO()

		err := ErrPageNotFound
		status1 := ErrorStatus(ctx1, err)
		status2 := ErrorStatus(ctx2, err)

		assert.Equal(t, status1, status2)
	})

	t.Run("context is passed to Status(ctx, err) method", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), testContextAwareStatusError{}, struct{}{})
		err := &testContextAwareStatusError{code: http.StatusBadRequest}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusAccepted, status)
	})

	t.Run("wrapped RedirectError", func(t *testing.T) {
		ctx := context.Background()
		redirectErr := NewRedirectError("/moved", http.StatusMovedPermanently)
		err := &testStatusWrapper{base: redirectErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusMovedPermanently, status)
	})

	t.Run("wrapped validation errors", func(t *testing.T) {
		ctx := context.Background()
		validErr := validation.Errors{
			"field": errors.New("invalid"),
		}
		err := &testStatusWrapper{base: validErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusUnprocessableEntity, status)
	})

	t.Run("predefined errors take precedence over StatusCode()", func(t *testing.T) {
		ctx := context.Background()
		innerErr := &testStatusCodeError{code: http.StatusPaymentRequired}
		err := errors.Join(innerErr, ErrPageNotFound)
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("wrapped error with no status methods returns 500", func(t *testing.T) {
		ctx := context.Background()
		innerErr := errors.New("inner error")
		err := &testStatusWrapper{base: innerErr}
		status := ErrorStatus(ctx, err)

		assert.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestErrorStatus_AllStatusCodeMethods(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"ErrPageNotFound", ErrPageNotFound, http.StatusNotFound},
		{"sql.ErrNoRows", sql.ErrNoRows, http.StatusNotFound},
		{"ErrPageForbidden", ErrPageForbidden, http.StatusForbidden},
		{"ErrPageUnauthorized", ErrPageUnauthorized, http.StatusUnauthorized},
		{"ErrUniqueViolation", ErrUniqueViolation, http.StatusConflict},
		{"validation error", validation.Errors{"field": errors.New("invalid")}, http.StatusUnprocessableEntity},
		{"generic error", errors.New("error"), http.StatusInternalServerError},
		{"nil", nil, http.StatusInternalServerError},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorStatus(ctx, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorStatus_InterfaceStatusMethods(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"Status 200", http.StatusOK},
		{"Status 201", http.StatusCreated},
		{"Status 400", http.StatusBadRequest},
		{"Status 404", http.StatusNotFound},
		{"Status 500", http.StatusInternalServerError},
		{"Status 503", http.StatusServiceUnavailable},
		{"Status 418", http.StatusTeapot},
	}

	ctx := context.Background()

	t.Run("StatusCode() method", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := &testStatusCodeError{code: tt.code}
				got := ErrorStatus(ctx, err)
				assert.Equal(t, tt.code, got)
			})
		}
	})

	t.Run("Status() method", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := &testStatusError{code: tt.code}
				got := ErrorStatus(ctx, err)
				assert.Equal(t, tt.code, got)
			})
		}
	})

	t.Run("Status(ctx, err) method", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := &testContextStatusError{code: tt.code}
				got := ErrorStatus(ctx, err)
				assert.Equal(t, tt.code, got)
			})
		}
	})
}

func TestErrorStatus_PredefinedErrorsTakePrecedence(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"ErrPageNotFound takes precedence", ErrPageNotFound, http.StatusNotFound},
		{"sql.ErrNoRows takes precedence", sql.ErrNoRows, http.StatusNotFound},
		{"ErrPageForbidden takes precedence", ErrPageForbidden, http.StatusForbidden},
		{"ErrPageUnauthorized takes precedence", ErrPageUnauthorized, http.StatusUnauthorized},
		{"ErrUniqueViolation takes precedence", ErrUniqueViolation, http.StatusConflict},
		{"validation error takes precedence", validation.Errors{"field": errors.New("invalid")}, http.StatusUnprocessableEntity},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorStatus(ctx, tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorStatus_RedirectErrorStatuses(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		status int
	}{
		{"301 Moved Permanently", "/permanent", http.StatusMovedPermanently},
		{"302 Found", "/temporary", http.StatusFound},
		{"303 See Other", "/other", http.StatusSeeOther},
		{"307 Temporary Redirect", "/temp-redirect", http.StatusTemporaryRedirect},
		{"308 Permanent Redirect", "/perm-redirect", http.StatusPermanentRedirect},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRedirectError(tt.url, tt.status)
			got := ErrorStatus(ctx, err)
			assert.Equal(t, tt.status, got)
		})
	}
}
