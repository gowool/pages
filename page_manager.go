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
	store PageStore
}

func NewPageManager(store PageStore) *DefaultPageManager {
	if store == nil {
		panic("page manager: store is required")
	}

	return &DefaultPageManager{
		store: store,
	}
}

func (m *DefaultPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	return m.store.FindByID(ctx, id)
}

func (m *DefaultPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	if site == nil {
		panic("page manager: get by url: site is required")
	}

	return m.store.FindByURL(ctx, site.ID, url)
}

func (m *DefaultPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	if site == nil {
		panic("page manager: get by pattern: site is required")
	}

	return m.store.FindByPattern(ctx, site.ID, pattern)
}

func (m *DefaultPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	if site == nil {
		panic("page manager: get by alias: site is required")
	}

	if !strings.HasPrefix(alias, PageAliasPrefix) {
		alias = PageAliasPrefix + alias
	}

	return m.store.FindByAlias(ctx, site.ID, alias)
}
