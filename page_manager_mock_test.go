package pages

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockPageManager implements PageManager interface using testify/mock
type MockPageManager struct {
	mock.Mock
}

func (m *MockPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	args := m.Called(ctx, site, url)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	args := m.Called(ctx, site, pattern)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	args := m.Called(ctx, site, alias)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}
