package pages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
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

// MockThemeForSite is a simple mock implementation of Theme for testing
type MockThemeForSite struct {
}

func (m *MockThemeForSite) Write(ctx context.Context, w io.Writer, template string, data any) error {
	return nil
}

func TestSiteMiddleware_SuccessfulSiteResolution(t *testing.T) {
	t.Run("valid site should be set and next called", func(t *testing.T) {
		// Setup
		site := &Site{ID: "test-site", Scheme: "", Host: ""}
		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		resp := httptest.NewRecorder()

		mockSelector := NewMockSiteSelector(site, "/test", nil)
		event := &Event{}
		event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockThemeForSite{})

		// Execute
		middleware := SiteMiddleware[Resolver](mockSelector)
		err := middleware(event)

		// Assert
		assert.NoError(t, err)

		// Verify the resolver has the site set
		assert.Equal(t, site, event.Site(), "Site should be set on resolver")

		// Verify the site was configured correctly
		assert.Equal(t, "http", site.Scheme, "Site scheme should be set from resolver")
		assert.Equal(t, "example.com", site.Host, "Site host should be set from request")

		mockSelector.AssertExpectations(t)
	})
}

func TestSiteMiddleware_PathHandling(t *testing.T) {
	tests := []struct {
		name         string
		requestPath  string
		expectedPath string
	}{
		{
			name:         "empty path should default to root",
			requestPath:  "",
			expectedPath: "/",
		},
		{
			name:         "non-empty path should remain unchanged",
			requestPath:  "/test",
			expectedPath: "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			site := &Site{ID: "test-site"}
			req := httptest.NewRequest("GET", "http://example.com"+tt.requestPath, nil)
			if tt.requestPath == "" {
				req.URL.Path = ""
			}
			resp := httptest.NewRecorder()

			mockSelector := NewMockSiteSelector(site, "", nil)
			event := &Event{}
			event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockThemeForSite{})

			// Execute
			middleware := SiteMiddleware[Resolver](mockSelector)
			err := middleware(event)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPath, req.URL.Path, "Request path should be correct")
			assert.Equal(t, tt.expectedPath, req.URL.RawPath, "RawPath should match Path")

			mockSelector.AssertExpectations(t)
		})
	}
}

func TestSiteMiddleware_ErrorHandling(t *testing.T) {
	customErr := errors.New("custom error")

	tests := []struct {
		name          string
		site          *Site
		pathInfo      string
		err           error
		expectedError error
	}{
		{
			name:          "ErrSiteNotFound should be returned as-is",
			site:          nil,
			pathInfo:      "",
			err:           ErrSiteNotFound,
			expectedError: ErrSiteNotFound,
		},
		{
			name:          "custom error should be wrapped with ErrSiteNotFound",
			site:          nil,
			pathInfo:      "",
			err:           customErr,
			expectedError: errors.Join(customErr, ErrSiteNotFound),
		},
		{
			name:          "nil site should return ErrSiteNotFound",
			site:          nil,
			pathInfo:      "",
			err:           nil,
			expectedError: ErrSiteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			req := httptest.NewRequest("GET", "http://example.com/test", nil)
			resp := httptest.NewRecorder()

			mockSelector := NewMockSiteSelector(tt.site, tt.pathInfo, tt.err)
			event := &Event{}
			event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockThemeForSite{})

			// Execute
			middleware := SiteMiddleware[Resolver](mockSelector)
			err := middleware(event)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err, "Expected error")
				assert.ErrorIs(t, tt.expectedError, err, "Error should match expected")
			} else {
				assert.NoError(t, err, "Should not return error")
			}

			// Verify no site was set on error
			assert.Nil(t, event.Site(), "No site should be set on resolver when error occurs")

			mockSelector.AssertExpectations(t)
		})
	}
}

func TestSiteMiddleware_PathOverride(t *testing.T) {
	// Setup
	site := &Site{ID: "test-site"}
	pathInfo := "/overridden/path"
	req := httptest.NewRequest("GET", "http://example.com/original", nil)
	resp := httptest.NewRecorder()

	mockSelector := NewMockSiteSelector(site, pathInfo, nil)
	event := &Event{}
	event.Reset(&wo.Response{ResponseWriter: resp}, req, &MockThemeForSite{})

	// Execute
	middleware := SiteMiddleware[Resolver](mockSelector)
	err := middleware(event)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, pathInfo, req.URL.Path, "Request path should be overridden by pathInfo")
	assert.Equal(t, site, event.Site(), "Site should be set on resolver")

	mockSelector.AssertExpectations(t)
}

func TestSiteMiddleware_PanicOnNilSelector(t *testing.T) {
	assert.Panics(t, func() {
		SiteMiddleware[Resolver](nil)
	}, "Should panic when selector is nil")
}
