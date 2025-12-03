package pages

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockSiteSelector is a mock implementation of SiteSelector
type MockSiteSelector struct {
	mock.Mock
}

func NewMockSiteSelector(site *Site, pathInfo string, err error) *MockSiteSelector {
	selector := &MockSiteSelector{}
	selector.On("Retrieve", mock.Anything).Return(site, pathInfo, err)
	return selector
}

func (m *MockSiteSelector) Retrieve(r *http.Request) (*Site, string, error) {
	args := m.Called(r)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*Site), args.String(1), nil
}
