package pages

import (
	"context"
	"iter"

	"github.com/stretchr/testify/mock"
)

// MockPageStorage is a mock implementation of PageStorage interface
type MockPageStorage struct {
	mock.Mock
}

func (m *MockPageStorage) FindByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStorage) FindByURL(ctx context.Context, siteID ID, url string) (*Page, error) {
	args := m.Called(ctx, siteID, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStorage) FindByPattern(ctx context.Context, siteID ID, pattern string) (*Page, error) {
	args := m.Called(ctx, siteID, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStorage) FindByAlias(ctx context.Context, siteID ID, alias string) (*Page, error) {
	args := m.Called(ctx, siteID, alias)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}

func (m *MockPageStorage) FindByPatterns(ctx context.Context, siteID ID, patterns ...string) iter.Seq2[*Page, error] {
	args := m.Called(ctx, siteID, patterns)
	return args.Get(0).(iter.Seq2[*Page, error])
}

func (m *MockPageStorage) Save(ctx context.Context, pages ...*Page) error {
	args := m.Called(ctx, pages)
	return args.Error(0)
}
