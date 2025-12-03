package pages

import (
	"context"
	"iter"

	"github.com/stretchr/testify/mock"
)

// MockSiteStorage is a mock implementation of SiteStorage
type MockSiteStorage struct {
	mock.Mock
}

func (m *MockSiteStorage) FindEnabled(ctx context.Context) (iter.Seq2[*Site, error], error) {
	args := m.Called(ctx)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	sites := args.Get(0).([]*Site)
	return func(yield func(*Site, error) bool) {
		for _, site := range sites {
			if !yield(site, nil) {
				break
			}
		}
	}, nil
}
