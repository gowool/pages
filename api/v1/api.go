package v1

import (
	"context"
	"database/sql"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/echox/api"

	"github.com/gowool/pages/repository"
)

var Info = api.CRUDInfo{Area: "admin", Version: "v1"}

func ErrorTransformer(_ context.Context, err error) error {
	var statusErr huma.StatusError
	if errors.As(err, &statusErr) {
		return statusErr
	}

	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repository.ErrNotFound) ||
		errors.Is(err, repository.ErrSiteNotFound) || errors.Is(err, repository.ErrPageNotFound) {
		return huma.Error404NotFound("Not Found", err)
	}

	if errors.Is(err, repository.ErrUniqueViolation) {
		return huma.Error409Conflict("Conflict", err)
	}

	return huma.Error500InternalServerError("Internal Server Error", err)
}
