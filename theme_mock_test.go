package pages

import (
	"context"
	"io"
)

// MockPageTheme is a simple mock implementation of Theme for testing
type MockPageTheme struct {
}

func (m *MockPageTheme) Write(context.Context, io.Writer, string, any) error {
	return nil
}
