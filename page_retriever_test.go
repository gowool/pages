package pages

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewPageRetriever(t *testing.T) {
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
					NewDefaultPageRetriever(tt.args.manager)
				}, "NewDefaultPageRetriever should panic when manager is nil")
			} else {
				assert.NotPanics(t, func() {
					var retriever PageRetriever = NewDefaultPageRetriever(tt.args.manager)
					assert.NotNil(t, retriever, "NewDefaultPageRetriever should return non-nil retriever")
					defaultRetriever, ok := retriever.(*DefaultPageRetriever)
					assert.True(t, ok, "Retriever should be of type *DefaultPageRetriever")
					assert.Equal(t, tt.args.manager, defaultRetriever.manager, "Manager should be set correctly")
				}, "NewDefaultPageRetriever should not panic when manager is provided")
			}
		})
	}
}

func TestDefaultPageRetriever_Retrieve(t *testing.T) {
	// Test basic panic scenarios
	t.Run("nil request should panic", func(t *testing.T) {
		mockManager := &MockPageManager{}
		retriever := NewDefaultPageRetriever(mockManager)
		site := NewSite()

		assert.Panics(t, func() {
			_, _ = retriever.Retrieve(nil, site)
		}, "Retrieve should panic with nil request")
	})

	t.Run("nil site should panic", func(t *testing.T) {
		mockManager := &MockPageManager{}
		retriever := NewDefaultPageRetriever(mockManager)

		// Create a request with Pattern field using custom test request
		req := &http.Request{
			URL:     &url.URL{Path: "/test"},
			Pattern: "/test",
		}

		assert.Panics(t, func() {
			_, _ = retriever.Retrieve(req, nil)
		}, "Retrieve should panic with nil site")
	})

	t.Run("retrieve cms page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		cmsPage := NewTestPage(PageCMS)
		cmsPage.URL = "/foo/boo"
		cmsPage.SiteID = cmsPage.Site.ID

		mockManager.On("GetByURL", mock.Anything, cmsPage.Site, cmsPage.URL).Return(cmsPage, nil)

		retriever := NewDefaultPageRetriever(mockManager)

		req := CreateTestRequest("GET", cmsPage.AbsURL(), nil)
		req.Pattern = PageCMSPattern

		require.Equal(t, req.URL.Path, cmsPage.URL, "Pattern should be set to URL path")

		page, err := retriever.Retrieve(req, cmsPage.Site)

		assert.NoError(t, err)
		assert.Equal(t, cmsPage, page)

		mockManager.AssertExpectations(t)
	})

	t.Run("retrieve dynamic hybrid page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		page1 := NewTestPage("/foo/{slug}")
		page1.SiteID = page1.Site.ID

		mockManager.On("GetByPattern", mock.Anything, page1.Site, page1.Pattern).Return(page1, nil)

		retriever := NewDefaultPageRetriever(mockManager)

		req := CreateTestRequest("GET", page1.AbsURL("{slug}", "boo"), nil)
		req.Pattern = "GET " + page1.Pattern

		page, err := retriever.Retrieve(req, page1.Site)

		assert.NoError(t, err)
		assert.Equal(t, page1, page)

		mockManager.AssertExpectations(t)
	})

	t.Run("retrieve hybrid page", func(t *testing.T) {
		mockManager := &MockPageManager{}

		page1 := NewTestPage("/foo/boo")
		page1.SiteID = page1.Site.ID

		mockManager.On("GetByPattern", mock.Anything, page1.Site, page1.Pattern).Return(page1, nil)

		retriever := NewDefaultPageRetriever(mockManager)

		req := CreateTestRequest("GET", page1.AbsURL(), nil)
		req.Pattern = page1.Pattern

		page, err := retriever.Retrieve(req, page1.Site)

		assert.NoError(t, err)
		assert.Equal(t, page1, page)

		mockManager.AssertExpectations(t)
	})
}

func TestDefaultPageRetriever_InterfaceCompliance(t *testing.T) {
	var _ PageRetriever = (*DefaultPageRetriever)(nil)

	mockManager := &MockPageManager{}
	retriever := NewDefaultPageRetriever(mockManager)

	// This compilation check ensures interface compliance
	var _ PageRetriever = retriever
}

// Test that confirms proper type initialization and manager assignment
func TestDefaultPageRetriever_Fields(t *testing.T) {
	mockManager := &MockPageManager{}
	var retrieverInterface PageRetriever = NewDefaultPageRetriever(mockManager)

	// Test that the retriever properly stores the manager
	defaultRetriever, ok := retrieverInterface.(*DefaultPageRetriever)
	assert.True(t, ok, "Retriever should be of type *DefaultPageRetriever")
	assert.Equal(t, mockManager, defaultRetriever.manager, "Manager should be properly assigned")
}
