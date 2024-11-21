package repository

import (
	"context"
	"time"

	"github.com/gowool/pages/model"
)

type Site interface {
	Repository[model.Site, int64]
	FindByHosts(ctx context.Context, hosts []string, now time.Time) ([]model.Site, error)
}
