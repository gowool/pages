package pages

import (
	"context"
	"strings"
)

var _ PageManager = (*DefaultPageManager)(nil)

type PageManager interface {
	GetByID(ctx context.Context, id ID) (*Page, error)
	GetByURL(ctx context.Context, site *Site, url string) (*Page, error)
	GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error)
	GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error)
}

type DefaultPageManager struct {
	storage PageStorage
}

func NewPageManager(storage PageStorage) *DefaultPageManager {
	if storage == nil {
		panic("page manager: storage is required")
	}

	return &DefaultPageManager{
		storage: storage,
	}
}

func (m *DefaultPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	return m.storage.FindByID(ctx, id)
}

func (m *DefaultPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	if site == nil {
		panic("page manager: get by url: site is required")
	}

	return m.storage.FindByURL(ctx, site.ID, url)
}

func (m *DefaultPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	if site == nil {
		panic("page manager: get by pattern: site is required")
	}

	return m.storage.FindByPattern(ctx, site.ID, pattern)
}

func (m *DefaultPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	if site == nil {
		panic("page manager: get by alias: site is required")
	}

	if !strings.HasPrefix(alias, PageAliasPrefix) {
		alias = PageAliasPrefix + alias
	}

	return m.storage.FindByAlias(ctx, site.ID, alias)
}
