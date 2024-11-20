package fallback

import (
	"context"
	"sync"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type ConfigurationRepository struct {
	repository.Configuration
	cfg model.Configuration
	mu  sync.Mutex
}

func NewConfigurationRepository(inner repository.Configuration, cfg model.Configuration) *ConfigurationRepository {
	return &ConfigurationRepository{
		Configuration: inner,
		cfg:           cfg,
	}
}

func (r *ConfigurationRepository) Load(ctx context.Context) (model.Configuration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, err := r.Configuration.Load(ctx); err == nil {
		return r.cfg.With(m), nil
	}
	return r.cfg, nil
}
