package pages

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockTheme is a mock implementation of the Theme interface
type MockTheme struct {
	mock.Mock
}

func (m *MockTheme) Write(ctx context.Context, w io.Writer, template string, data any) error {
	args := m.Called(ctx, w, template, data)
	return args.Error(0)
}

// MockPageTheme is a simple mock implementation of Theme for testing
type MockPageTheme struct {
}

func (m *MockPageTheme) Write(context.Context, io.Writer, string, any) error {
	return nil
}
