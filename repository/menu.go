package repository

import (
	"context"

	"github.com/gowool/pages/model"
)

type Menu interface {
	Repository[model.Menu, int64]
	FindByHandle(ctx context.Context, handle string) (model.Menu, error)
}
