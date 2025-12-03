package pages

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockPageSelector implements the PageSelector interface
type MockPageSelector struct {
	mock.Mock
}

func NewMockPageSelector(page *Page, err error) *MockPageSelector {
	selector := &MockPageSelector{}
	selector.On("Retrieve", mock.Anything, mock.Anything).Return(page, err)
	return selector
}

func (m *MockPageSelector) Retrieve(req *http.Request, site *Site) (*Page, error) {
	args := m.Called(req, site)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}
