package pages

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewPageSelector(t *testing.T) {
	type args struct {
		manager PageManager
	}
	tests := []struct {
		name        string
		args        args
		expectPanic bool
	}{
		{
			name:        "nil manager should panic",
			args:        args{manager: nil},
			expectPanic: true,
		},
		{
			name:        "valid manager should not panic",
			args:        args{manager: &MockPageManager{}},
			expectPanic: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewPageSelector(tt.args.manager)
				}, "NewPageSelector should panic when manager is nil")
			} else {
				assert.NotPanics(t, func() {
					var selector PageSelector = NewPageSelector(tt.args.manager)
					assert.NotNil(t, selector, "NewPageSelector should return non-nil selector")
					defaultSelector, ok := selector.(*DefaultPageSelector)
					assert.True(t, ok, "Selector should be of type *DefaultPageSelector")
					assert.Equal(t, tt.args.manager, defaultSelector.manager, "Manager should be set correctly")
				}, "NewPageSelector should not panic when manager is provided")
			}
		})
	}
}

func TestDefaultPageSelector_Retrieve(t *testing.T) {
	// Test basic panic scenarios
	t.Run("nil request should panic", func(t *testing.T) {
		mockManager := &MockPageManager{}
		selector := NewPageSelector(mockManager)
		site := NewSite()

		assert.Panics(t, func() {
			selector.Retrieve(nil, site)
		}, "Retrieve should panic with nil request")
	})

	t.Run("nil site should panic", func(t *testing.T) {
		mockManager := &MockPageManager{}
		selector := NewPageSelector(mockManager)

		// Create a request with Pattern field using custom test request
		req := &http.Request{
			URL:     &url.URL{Path: "/test"},
			Pattern: "/test",
		}

		assert.Panics(t, func() {
			selector.Retrieve(req, nil)
		}, "Retrieve should panic with nil site")
	})

	t.Run("retrieve cms page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		cmsPage := NewTestPage(PageCMS)
		cmsPage.URL = "/foo/boo"
		cmsPage.SiteID = cmsPage.Site.ID

		mockManager.On("GetByURL", mock.Anything, cmsPage.Site, cmsPage.URL).Return(cmsPage, nil)

		selector := NewPageSelector(mockManager)

		req := CreateTestRequest("GET", cmsPage.AbsURL(), nil)
		req.Pattern = PageCMSPattern

		require.Equal(t, req.URL.Path, cmsPage.URL, "Pattern should be set to URL path")

		page, err := selector.Retrieve(req, cmsPage.Site)

		assert.NoError(t, err)
		assert.Equal(t, cmsPage, page)

		mockManager.AssertExpectations(t)
	})

	t.Run("retrieve dynamic hybrid page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		page1 := NewTestPage("/foo/{slug}")
		page1.SiteID = page1.Site.ID

		mockManager.On("GetByPattern", mock.Anything, page1.Site, page1.Pattern).Return(page1, nil)

		selector := NewPageSelector(mockManager)

		req := CreateTestRequest("GET", page1.AbsURL("{slug}", "boo"), nil)
		req.Pattern = "GET " + page1.Pattern

		page, err := selector.Retrieve(req, page1.Site)

		assert.NoError(t, err)
		assert.Equal(t, page1, page)

		mockManager.AssertExpectations(t)
	})

	t.Run("retrieve hybrid page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		page1 := NewTestPage("/foo/boo")
		page1.SiteID = page1.Site.ID

		mockManager.On("GetByPattern", mock.Anything, page1.Site, page1.Pattern).Return(page1, nil)

		selector := NewPageSelector(mockManager)

		req := CreateTestRequest("GET", page1.AbsURL(), nil)
		req.Pattern = page1.Pattern

		page, err := selector.Retrieve(req, page1.Site)

		assert.NoError(t, err)
		assert.Equal(t, page1, page)

		mockManager.AssertExpectations(t)
	})
}

func TestDefaultPageSelector_InterfaceCompliance(t *testing.T) {
	var _ PageSelector = (*DefaultPageSelector)(nil)

	mockManager := &MockPageManager{}
	selector := NewPageSelector(mockManager)

	// This compilation check ensures interface compliance
	var _ PageSelector = selector
}

// Test that confirms proper type initialization and manager assignment
func TestDefaultPageSelector_Fields(t *testing.T) {
	mockManager := &MockPageManager{}
	var selectorInterface PageSelector = NewPageSelector(mockManager)

	// Test that the selector properly stores the manager
	defaultSelector, ok := selectorInterface.(*DefaultPageSelector)
	assert.True(t, ok, "Selector should be of type *DefaultPageSelector")
	assert.Equal(t, mockManager, defaultSelector.manager, "Manager should be properly assigned")
}
