package pages

import "context"

var _ SiteStore = (*LocalhostSiteStore)(nil)

type SiteStore interface {
	FindPublished(ctx context.Context) ([]*Site, error)
}

type LocalhostSiteStore struct{}

func NewLocalhostSiteStore() *LocalhostSiteStore {
	return &LocalhostSiteStore{}
}

func (s *LocalhostSiteStore) FindPublished(context.Context) ([]*Site, error) {
	site := NewSite()
	site.Status = Published

	return []*Site{site}, nil
}
