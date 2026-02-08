package pages

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gowool/gor"
	"github.com/invopop/validation"
)

// ErrorStatus returns an error code code.
func ErrorStatus(_ context.Context, err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}

	if errors.Is(err, ErrPageNotFound) || errors.Is(err, sql.ErrNoRows) || errors.Is(err, gor.ErrFileNotFound) {
		return http.StatusNotFound
	}

	if errors.Is(err, ErrPageForbidden) {
		return http.StatusForbidden
	}

	if errors.Is(err, ErrPageUnauthorized) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrUniqueViolation) {
		return http.StatusConflict
	}

	var validErrs validation.Errors
	if errors.As(err, &validErrs) {
		return http.StatusUnprocessableEntity
	}

	return gor.HTTPErrorStatusCode(err)
}
