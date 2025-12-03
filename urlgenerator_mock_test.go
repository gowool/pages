package pages

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockURLGenerator implements URLGenerator interface using testify/mock
type MockURLGenerator struct {
	mock.Mock
}

func (m *MockURLGenerator) Generate(ctx context.Context, site *Site, arg any, args ...any) (string, error) {
	callArgs := m.Called(ctx, site, arg, args)
	if url := callArgs.Get(0); url != nil {
		return url.(string), callArgs.Error(1)
	}
	return "", callArgs.Error(1)
}
