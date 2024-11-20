package cache

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gowool/pages"
	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type SiteRepository struct {
	repository.Site
	repo[model.Site, int64]
}

func NewSiteRepository(inner repository.Site, c pages.Cache) SiteRepository {
	return SiteRepository{
		Site: inner,
		repo: repo[model.Site, int64]{inner: inner, cache: c, prefix: "cms::site"},
	}
}

func (r SiteRepository) FindByHosts(ctx context.Context, hosts []string, now time.Time) (sites []model.Site, err error) {
	key := fmt.Sprintf("%s:host:%s", r.prefix, strings.Join(hosts, "|"))

	if err = r.cache.Get(ctx, key, &sites); err == nil {
		if slices.ContainsFunc(sites, func(site model.Site) bool { return !site.IsEnabled(now) }) {
			_ = r.cache.DelByKey(ctx, key)

			goto INNER
		}
		return
	}

INNER:

	if sites, err = r.Site.FindByHosts(ctx, hosts, now); err != nil {
		return
	}

	r.set(ctx, key, internal.Map(sites, func(item model.Site) int64 {
		return item.ID
	}))
	return
}

func (r SiteRepository) FindByID(ctx context.Context, id int64) (model.Site, error) {
	return r.findByID(ctx, id)
}

func (r SiteRepository) Delete(ctx context.Context, ids ...int64) error {
	return r.delete(ctx, ids...)
}

func (r SiteRepository) Update(ctx context.Context, m *model.Site) error {
	if m == nil {
		return fmt.Errorf("cache: site repository update called with %w", repository.ErrNil)
	}

	defer r.del(ctx, m.ID)

	return r.Site.Update(ctx, m)
}
