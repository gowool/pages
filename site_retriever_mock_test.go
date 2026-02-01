package pages

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockSiteRetriever is a mock implementation of SiteRetriever
type MockSiteRetriever struct {
	mock.Mock
}

func NewMockSiteRetriever(site *Site, pathInfo string, err error) *MockSiteRetriever {
	retriever := &MockSiteRetriever{}
	retriever.On("Retrieve", mock.Anything).Return(site, pathInfo, err)
	return retriever
}

func (m *MockSiteRetriever) Retrieve(r *http.Request) (*Site, string, error) {
	args := m.Called(r)
	if args.Get(0) == nil && args.Get(2) == nil {
		return nil, "", nil
	}
	if args.Get(2) != nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*Site), args.String(1), nil
}
