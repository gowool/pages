package pages

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gowool/r"
	"github.com/invopop/validation"
)

// ErrorStatus returns an error status code.
func ErrorStatus(ctx context.Context, err error) int {
	if errors.Is(err, ErrPageNotFound) || errors.Is(err, sql.ErrNoRows) || errors.Is(err, r.ErrFileNotFound) {
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

	e := err
	for {
		switch t := e.(type) {
		case interface{ StatusCode() int }:
			return t.StatusCode()
		case interface{ Status() int }:
			return t.Status()
		case interface {
			Status(context.Context, error) int
		}:
			return t.Status(ctx, err)
		case interface{ Redirect() (string, int) }:
			_, status := t.Redirect()
			return status
		case interface{ Unwrap() error }:
			e = t.Unwrap()
			continue
		default:
			return http.StatusInternalServerError
		}
	}
}
