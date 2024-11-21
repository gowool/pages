package repository

import (
	"context"

	"github.com/gowool/cr"
)

type Repository[M any, ID any] interface {
	Find(ctx context.Context, criteria *cr.Criteria) ([]M, error)
	FindAndCount(ctx context.Context, criteria *cr.Criteria) ([]M, int, error)
	FindByID(ctx context.Context, id ID) (M, error)
	Delete(ctx context.Context, ids ...ID) error
	Create(ctx context.Context, m *M) error
	Update(ctx context.Context, m *M) error
}
