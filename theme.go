package pages

import (
	"context"
	"io"
)

type Theme interface {
	Write(ctx context.Context, w io.Writer, template string, data any) error
}
