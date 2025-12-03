package pages

import "github.com/stretchr/testify/mock"

// MockPageAuthorizer implements the PageAuthorizer interface
type MockPageAuthorizer struct {
	mock.Mock
}

func NewMockPageAuthorizer(decision Decision, err error) *MockPageAuthorizer {
	authorizer := &MockPageAuthorizer{}
	authorizer.On("Authorize", mock.Anything, mock.Anything).Return(decision, err)
	return authorizer
}

func (m *MockPageAuthorizer) Authorize(resolver Resolver, action PageAction) (Decision, error) {
	args := m.Called(resolver, action)
	decision, ok := args.Get(0).(Decision)
	if !ok {
		return Deny, args.Error(1)
	}
	return decision, args.Error(1)
}
