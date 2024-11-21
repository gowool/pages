package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type TemplateRepository struct {
	repository.Template
	repo[model.Template, int64]
}

func NewTemplateRepository(inner repository.Template, c pages.Cache) TemplateRepository {
	return TemplateRepository{
		Template: inner,
		repo:     repo[model.Template, int64]{inner: inner, cache: c, prefix: "cms::template"},
	}
}

func (r TemplateRepository) FindByName(ctx context.Context, name string) (m model.Template, err error) {
	key := fmt.Sprintf("%s:name:%s", r.prefix, name)

	if err = r.cache.Get(ctx, key, &m); err == nil {
		return
	}

	if m, err = r.Template.FindByName(ctx, name); err != nil {
		return
	}

	r.set(ctx, key, m, m.ID)
	return
}

func (r TemplateRepository) FindByID(ctx context.Context, id int64) (model.Template, error) {
	return r.findByID(ctx, id)
}

func (r TemplateRepository) Delete(ctx context.Context, ids ...int64) error {
	return r.delete(ctx, ids...)
}

func (r TemplateRepository) Update(ctx context.Context, m *model.Template) error {
	if m == nil {
		return errors.New("cache: template repository update called with nil model")
	}

	defer r.del(ctx, m.ID)

	return r.Template.Update(ctx, m)
}
