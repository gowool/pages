package repository

import (
	"context"

	"github.com/gowool/pages/model"
)

type Template interface {
	Repository[model.Template, int64]
	FindByName(ctx context.Context, name string) (model.Template, error)
}
