package pages

import (
	"context"
	"iter"
)

var _ SiteStorage = (*DefaultSiteStorage)(nil)

type SiteStorage interface {
	FindEnabled(ctx context.Context) (iter.Seq2[*Site, error], error)
}

type DefaultSiteStorage struct{}

func NewSiteStorage() *DefaultSiteStorage {
	return &DefaultSiteStorage{}
}

func (s *DefaultSiteStorage) FindEnabled(context.Context) (iter.Seq2[*Site, error], error) {
	return func(yield func(*Site, error) bool) {
		site := NewSite()
		site.Enabled = true

		yield(site, nil)
	}, nil
}
