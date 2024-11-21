package v1

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/echox/api"

	"github.com/gowool/pages"
)

var Info = api.CRUDInfo{Area: "admin", Version: "v1"}

func ErrorTransformer(_ context.Context, err error) error {
	var statusErr huma.StatusError
	if errors.As(err, &statusErr) {
		return statusErr
	}

	if pages.IsOneOfNotFound(err) {
		return huma.Error404NotFound("Not Found", err)
	}

	return huma.Error500InternalServerError("Internal Server Error", err)
}
