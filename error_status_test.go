package pages

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/gowool/keratin"
	"github.com/invopop/validation"
	"github.com/stretchr/testify/assert"
)

type testStatusCodeError struct {
	code int
}

func (e *testStatusCodeError) Error() string {
	return "test code code error"
}

func (e *testStatusCodeError) StatusCode() int {
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

	t.Run("keratin.ErrFileNotFound returns 404", func(t *testing.T) {
		ctx := context.Background()
		status := ErrorStatus(ctx, keratin.ErrFileNotFound)

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

	t.Run("unwraps error chain to find StatusCode()", func(t *testing.T) {
		ctx := context.Background()
		innerErr := &testStatusCodeError{code: http.StatusBadGateway}
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

	t.Run("wrapped error with no code methods returns 500", func(t *testing.T) {
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
		name     string
		code     int
		expected int
	}{
		{"Status 200", http.StatusOK, http.StatusInternalServerError},
		{"Status 201", http.StatusCreated, http.StatusInternalServerError},
		{"Status 400", http.StatusBadRequest, http.StatusBadRequest},
		{"Status 404", http.StatusNotFound, http.StatusNotFound},
		{"Status 500", http.StatusInternalServerError, http.StatusInternalServerError},
		{"Status 503", http.StatusServiceUnavailable, http.StatusServiceUnavailable},
		{"Status 418", http.StatusTeapot, http.StatusTeapot},
	}

	ctx := context.Background()

	t.Run("StatusCode() method", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := &testStatusCodeError{code: tt.code}
				got := ErrorStatus(ctx, err)
				assert.Equal(t, tt.expected, got)
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
