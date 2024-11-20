package pages

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var errorPages = map[string]model.Page{
	model.PageErrorInternal: {
		Pattern:  model.PageErrorInternal,
		Title:    "Error Internal",
		Name:     "Error",
		Template: "@error/internal.gohtml",
	},
	model.PageError4xx: {
		Pattern:  model.PageError4xx,
		Title:    "Error 4xx",
		Name:     "Error",
		Template: "@error/4xx.gohtml",
	},
	model.PageError5xx: {
		Pattern:  model.PageError5xx,
		Title:    "Error 5xx",
		Name:     "Error",
		Template: "@error/5xx.gohtml",
	},
}

type Seeder interface {
	Boot(context.Context) error
}

type DefaultSeeder struct {
	siteRepository repository.Site
	pageRepository repository.Page
	logger         *zap.Logger
}

func NewDefaultSeeder(siteRepository repository.Site, pageRepository repository.Page, logger *zap.Logger) *DefaultSeeder {
	return &DefaultSeeder{
		siteRepository: siteRepository,
		pageRepository: pageRepository,
		logger:         logger,
	}
}

func (s *DefaultSeeder) Boot(ctx context.Context) error {
	sites, err := s.FindOrCreateLocalhost(ctx)
	if err != nil {
		return err
	}
	for _, site := range sites {
		if err = s.InternalCreatePage(ctx, site); err != nil {
			return err
		}
		if err = s.CreateErrorPages(ctx, site); err != nil {
			return err
		}
	}
	return nil
}

func (s *DefaultSeeder) FindOrCreateLocalhost(ctx context.Context) ([]model.Site, error) {
	sites, err := s.siteRepository.FindByHosts(ctx, []string{"localhost"}, time.Time{})
	if err != nil {
		return nil, err
	}
	if len(sites) > 0 {
		s.logger.Info("localhost site already exists")
		return sites, nil
	}

	now := time.Now()
	site := model.Site{
		ID:        -1,
		Name:      "Localhost",
		Host:      "localhost",
		Locale:    "en_US",
		Separator: " - ",
		Published: &now,
	}

	if err = s.siteRepository.Create(ctx, &site); err != nil {
		return nil, err
	}

	s.logger.Info("created localhost site", zap.Int64("id", site.ID))

	return []model.Site{site}, nil
}

func (s *DefaultSeeder) CreateErrorPages(ctx context.Context, site model.Site) error {
	for pattern, page := range errorPages {
		_, err := s.pageRepository.FindByPattern(ctx, site.ID, pattern, time.Time{})
		if err == nil {
			continue
		}

		now := time.Now()
		page.Published = &now
		page.SiteID = site.ID
		if err = s.pageRepository.Create(ctx, &page); err != nil {
			return err
		}
		s.logger.Info("created error page", zap.String("pattern", pattern))
	}
	return nil
}

func (s *DefaultSeeder) InternalCreatePage(ctx context.Context, site model.Site) error {
	_, err := s.pageRepository.FindByPattern(ctx, site.ID, model.PageInternalCreate, time.Time{})
	if err == nil {
		return nil
	}

	now := time.Now()
	if err = s.pageRepository.Create(ctx, &model.Page{
		SiteID:    site.ID,
		Pattern:   model.PageInternalCreate,
		Title:     "Create Not Found Page",
		Name:      "Create page",
		Template:  "@error/page_create.gohtml",
		Published: &now,
	}); err != nil {
		return err
	}

	s.logger.Info("created internal create page")
	return nil
}
