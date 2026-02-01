package pages

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockPageRetriever implements the PageRetriever interface
type MockPageRetriever struct {
	mock.Mock
}

func NewMockPageSelector(page *Page, err error) *MockPageRetriever {
	retriever := &MockPageRetriever{}
	retriever.On("Retrieve", mock.Anything, mock.Anything).Return(page, err)
	return retriever
}

func (m *MockPageRetriever) Retrieve(req *http.Request, site *Site) (*Page, error) {
	args := m.Called(req, site)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Page), args.Error(1)
}
