package pages

import (
	"context"

	"github.com/google/uuid"
)

type ID string

func (id ID) IsZero() bool {
	return id == ""
}

func (id ID) String() string {
	return string(id)
}

type IDGeneratorFunc func(ctx context.Context) (ID, error)

func IDGenerator() IDGeneratorFunc {
	return func(context.Context) (ID, error) {
		return ID(uuid.NewString()), nil
	}
}
