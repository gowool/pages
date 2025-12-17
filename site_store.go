package pages

import (
	"context"
	"iter"
)

var _ SiteStore = (*LocalhostSiteStore)(nil)

type SiteStore interface {
	FindEnabled(ctx context.Context) (iter.Seq2[*Site, error], error)
}

type LocalhostSiteStore struct{}

func NewLocalhostSiteStore() *LocalhostSiteStore {
	return &LocalhostSiteStore{}
}

func (s *LocalhostSiteStore) FindEnabled(context.Context) (iter.Seq2[*Site, error], error) {
	return func(yield func(*Site, error) bool) {
		site := NewSite()
		site.Enabled = true

		yield(site, nil)
	}, nil
}
