package repository

import (
	"context"
	"errors"
	"time"

	"github.com/gowool/pages/model"
)

var ErrSiteNotFound = errors.New("site not found")

type Site interface {
	repository[model.Site, int64]
	FindByHosts(ctx context.Context, hosts []string, now time.Time) ([]model.Site, error)
}
