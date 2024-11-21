package repository

import (
	"context"

	"github.com/gowool/pages/model"
)

type Node interface {
	Repository[model.Node, int64]
	FindWithChildren(ctx context.Context, id int64) ([]model.Node, error)
}
