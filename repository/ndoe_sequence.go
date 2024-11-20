package repository

import "context"

type SequenceNode interface {
	Create(ctx context.Context) (int64, error)
	Delete(ctx context.Context, ids ...int64) error
}
