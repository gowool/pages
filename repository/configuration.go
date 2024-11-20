package repository

import (
	"context"

	"github.com/gowool/pages/model"
)

type Configuration interface {
	Load(ctx context.Context) (model.Configuration, error)
	Save(ctx context.Context, m *model.Configuration) error
}
