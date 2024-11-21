package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type PageRepository struct {
	repository.Page
	repo[model.Page, int64]
}

func NewPageRepository(inner repository.Page, c pages.Cache) PageRepository {
	return PageRepository{
		Page: inner,
		repo: repo[model.Page, int64]{inner: inner, cache: c, prefix: "cms::page"},
	}
}

func (r PageRepository) FindByID(ctx context.Context, id int64) (m model.Page, err error) {
	key := fmt.Sprintf("%s:id:%v", r.prefix, id)

	if err = r.cache.Get(ctx, key, &m); err == nil {
		return
	}

	if m, err = r.findByID(ctx, id); err != nil {
		return
	}

	r.set(ctx, key, m)
	return
}

func (r PageRepository) FindByParentID(ctx context.Context, parentID int64, now time.Time) (pages []model.Page, err error) {
	key := fmt.Sprintf("%s:parent:%d", r.prefix, parentID)

	if err = r.cache.Get(ctx, key, &pages); err == nil {
		for _, p := range pages {
			if now.IsZero() || !p.IsEnabled(now) {
				_ = r.cache.DelByKey(ctx, key)
				goto INNER
			}
		}
		return
	}

INNER:
	if pages, err = r.Page.FindByParentID(ctx, parentID, now); err != nil {
		return
	}

	tags := make([]string, 0, len(pages)+1)
	tags = append(tags, fmt.Sprintf("%s:tag:%d", r.prefix, parentID))

	for _, p := range pages {
		tags = append(tags, fmt.Sprintf("%s:tag:%d", r.prefix, p.ID))
	}

	_ = r.cache.Set(ctx, key, pages, tags...)
	return
}

func (r PageRepository) FindByPattern(ctx context.Context, siteID int64, pattern string, now time.Time) (m model.Page, err error) {
	key := fmt.Sprintf("%s:pattern:%d:%s", r.prefix, siteID, pattern)

	if r.get(ctx, key, now, &m) {
		return
	}

	if m, err = r.Page.FindByPattern(ctx, siteID, pattern, now); err != nil {
		return
	}

	r.set(ctx, key, m)
	return
}

func (r PageRepository) FindByAlias(ctx context.Context, siteID int64, alias string, now time.Time) (m model.Page, err error) {
	key := fmt.Sprintf("%s:alias:%d:%s", r.prefix, siteID, alias)

	if r.get(ctx, key, now, &m) {
		return
	}

	if m, err = r.Page.FindByAlias(ctx, siteID, alias, now); err != nil {
		return
	}

	r.set(ctx, key, m)
	return
}

func (r PageRepository) FindByURL(ctx context.Context, siteID int64, url string, now time.Time) (m model.Page, err error) {
	key := fmt.Sprintf("%s:url:%d:%s", r.prefix, siteID, url)

	if r.get(ctx, key, now, &m) {
		return
	}

	if m, err = r.Page.FindByURL(ctx, siteID, url, now); err != nil {
		return
	}

	r.set(ctx, key, m)
	return
}

func (r PageRepository) Delete(ctx context.Context, ids ...int64) error {
	return r.delete(ctx, ids...)
}

func (r PageRepository) Update(ctx context.Context, m *model.Page) error {
	if m == nil {
		return errors.New("cache: page repository update called with nil model")
	}

	defer r.del(ctx, m.ID)

	return r.Page.Update(ctx, m)
}

func (r PageRepository) set(ctx context.Context, key string, m model.Page) {
	tags := []string{
		fmt.Sprintf("%s:tag:%d", r.prefix, m.ID),
		fmt.Sprintf("cms:site:tag:%d", m.SiteID),
	}
	if m.ParentID != nil {
		tags = append(tags, fmt.Sprintf("%s:tag:%d", r.prefix, m.ParentID))
	}

	_ = r.cache.Set(ctx, key, m, tags...)
}

func (r PageRepository) get(ctx context.Context, key string, now time.Time, m *model.Page) bool {
	if err := r.cache.Get(ctx, key, m); err == nil {
		if !now.IsZero() && m.IsEnabled(now) {
			return true
		}

		_ = r.cache.DelByKey(ctx, key)
	}
	return false
}
