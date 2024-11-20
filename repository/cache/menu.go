package cache

import (
	"context"
	"fmt"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type MenuRepository struct {
	repository.Menu
	repo[model.Menu, int64]
}

func NewMenuRepository(inner repository.Menu, c pages.Cache) MenuRepository {
	return MenuRepository{
		Menu: inner,
		repo: repo[model.Menu, int64]{inner: inner, cache: c, prefix: "cms::menu"},
	}
}

func (r MenuRepository) FindByID(ctx context.Context, id int64) (model.Menu, error) {
	return r.findByID(ctx, id)
}

func (r MenuRepository) Delete(ctx context.Context, ids ...int64) error {
	return r.delete(ctx, ids...)
}

func (r MenuRepository) Update(ctx context.Context, m *model.Menu) error {
	if m == nil {
		return fmt.Errorf("cache: menu repository update called with %w", repository.ErrNil)
	}

	defer r.del(ctx, m.ID)

	return r.Menu.Update(ctx, m)
}

func (r MenuRepository) FindByHandle(ctx context.Context, handle string) (m model.Menu, err error) {
	key := fmt.Sprintf("%s:handle:%s", r.prefix, handle)

	if err = r.cache.Get(ctx, key, &m); err == nil {
		return
	}

	if m, err = r.Menu.FindByHandle(ctx, handle); err != nil {
		return
	}

	r.set(ctx, key, m, m.ID)
	return
}
