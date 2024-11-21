package cache

import (
	"context"
	"errors"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type ConfigurationRepository struct {
	repository.Configuration
	cache pages.Cache
	key   string
}

func NewConfigurationRepository(inner repository.Configuration, c pages.Cache) ConfigurationRepository {
	return ConfigurationRepository{
		Configuration: inner,
		cache:         c,
		key:           "cms::page:configuration",
	}
}

func (r ConfigurationRepository) Load(ctx context.Context) (m model.Configuration, err error) {
	if err = r.cache.Get(ctx, r.key, &m); err == nil {
		return
	}

	if m, err = r.Configuration.Load(ctx); err != nil {
		return
	}

	_ = r.cache.Set(ctx, r.key, m)
	return
}

func (r ConfigurationRepository) Save(ctx context.Context, m *model.Configuration) error {
	if m == nil {
		return errors.New("cache: configuration repository save called with nil model")
	}

	defer func() {
		_ = r.cache.DelByKey(ctx, r.key)
	}()

	return r.Configuration.Save(ctx, m)
}
