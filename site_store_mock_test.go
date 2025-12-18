package pages

import (
	"context"
	"iter"

	"github.com/stretchr/testify/mock"
)

// MockSiteStore is a mock implementation of SiteStore
type MockSiteStore struct {
	mock.Mock
}

func (m *MockSiteStore) FindEnabled(ctx context.Context) iter.Seq2[*Site, error] {
	args := m.Called(ctx)
	return args.Get(0).(iter.Seq2[*Site, error])
}
