package pages

import (
	"context"
	"iter"

	"github.com/stretchr/testify/mock"
)

// MockPageStore is a mock implementation of PageStore interface
type MockPageStore struct {
	mock.Mock
}

func (m *MockPageStore) FindByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByURL(ctx context.Context, siteID ID, url string) (*Page, error) {
	args := m.Called(ctx, siteID, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error) {
	args := m.Called(ctx, siteID, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error) {
	args := m.Called(ctx, siteID, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStore) FindByPatterns(ctx context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error] {
	args := m.Called(ctx, siteID, patterns)
	return args.Get(0).(iter.Seq2[*Page, error])
}

func (m *MockPageStore) Save(ctx context.Context, pages ...*Page) error {
	args := m.Called(ctx, pages)
	return args.Error(0)
}
