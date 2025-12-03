package pages

import (
	"context"
	"iter"
)

var _ SiteStorage = (*LocalhostSiteStorage)(nil)

type SiteStorage interface {
	FindEnabled(ctx context.Context) (iter.Seq2[*Site, error], error)
}

type LocalhostSiteStorage struct{}

func NewLocalhostSiteStorage() *LocalhostSiteStorage {
	return &LocalhostSiteStorage{}
}

func (s *LocalhostSiteStorage) FindEnabled(context.Context) (iter.Seq2[*Site, error], error) {
	return func(yield func(*Site, error) bool) {
		site := NewSite()
		site.Enabled = true

		yield(site, nil)
	}, nil
}
