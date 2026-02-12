package pages

import (
	"context"
	"iter"
)

var _ SiteStore = (*LocalhostSiteStore)(nil)

type SiteStore interface {
	FindPublished(ctx context.Context) iter.Seq2[*Site, error]
}

type LocalhostSiteStore struct {
	site *Site
}

func NewLocalhostSiteStore() *LocalhostSiteStore {
	site := NewSite()
	site.Status = Published

	return &LocalhostSiteStore{site: site}
}

func (s *LocalhostSiteStore) FindPublished(context.Context) iter.Seq2[*Site, error] {
	return func(yield func(*Site, error) bool) {
		yield(s.site, nil)
	}
}
