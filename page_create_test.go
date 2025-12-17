package pages

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPageCreateRequest_Validate tests the validation of PageCreateRequest
func TestPageCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     PageCreateRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request with all fields",
			request: PageCreateRequest{
				URL:      "/test-page",
				Template: "test.html",
				Title:    "Test Page",
			},
			expectError: false,
		},
		{
			name: "Valid request with minimal fields",
			request: PageCreateRequest{
				URL:      "/min",
				Template: "min.html",
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			request: PageCreateRequest{
				Template: "test.html",
				Title:    "Test Page",
			},
			expectError: true,
			errorMsg:    "URL",
		},
		{
			name: "Missing Template",
			request: PageCreateRequest{
				URL:   "/test-page",
				Title: "Test Page",
			},
			expectError: true,
			errorMsg:    "Template",
		},
		{
			name: "Empty URL",
			request: PageCreateRequest{
				URL:      "",
				Template: "test.html",
			},
			expectError: true,
			errorMsg:    "URL",
		},
		{
			name: "Empty Template",
			request: PageCreateRequest{
				URL:      "/test-page",
				Template: "",
			},
			expectError: true,
			errorMsg:    "Template",
		},
		{
			name: "URL too long",
			request: PageCreateRequest{
				URL:      "/" + strings.Repeat("a", 255),
				Template: "test.html",
			},
			expectError: true,
			errorMsg:    "URL",
		},
		{
			name: "Template too long",
			request: PageCreateRequest{
				URL:      "/test-page",
				Template: strings.Repeat("a", 255),
			},
			expectError: true,
			errorMsg:    "Template",
		},
		{
			name: "Title too long",
			request: PageCreateRequest{
				URL:      "/test-page",
				Template: "test.html",
				Title:    strings.Repeat("a", 255),
			},
			expectError: true,
			errorMsg:    "Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					// Convert to lowercase for case-insensitive matching
					errorMsg := strings.ToLower(err.Error())
					expectedMsg := strings.ToLower(tt.errorMsg)
					assert.Contains(t, errorMsg, expectedMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewPageCreate tests the NewPageCreate constructor
func TestNewPageCreate(t *testing.T) {
	mockStore := &MockPageStore{}
	beforeSave := func(ctx context.Context, page *Page) error {
		return nil
	}

	handler := NewPageCreate[Resolver](mockStore, beforeSave)

	assert.NotNil(t, handler)
	assert.Equal(t, mockStore, handler.store)
	assert.NotNil(t, handler.beforeSave)
}

// TestNewPageCreateWithNilStore tests that NewPageCreate works with nil store
func TestNewPageCreateWithNilStore(t *testing.T) {
	handler := NewPageCreate[Resolver](nil, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.store)
	assert.Nil(t, handler.beforeSave)
}

// TestPageCreate_Handle_BindError tests handling of binding errors
func TestPageCreate_Handle_BindError(t *testing.T) {
	mockStore := &MockPageStore{}
	handler := NewPageCreate[Resolver](mockStore, nil)

	// Create event with invalid JSON to trigger bind error
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	err := handler.Handle(event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

// TestPageCreate_Handle_ValidationError tests handling of validation errors
func TestPageCreate_Handle_ValidationError(t *testing.T) {
	mockStore := &MockPageStore{}
	handler := NewPageCreate[Resolver](mockStore, nil)

	// Create event with invalid data to trigger validation error
	requestBody := `{"url": "", "template": ""}`
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	err := handler.Handle(event)

	assert.Error(t, err)
	errorMsg := strings.ToLower(err.Error())
	assert.Contains(t, errorMsg, "url")
	assert.Contains(t, errorMsg, "template")
}

// TestPageCreate_Handle_URLProcessing tests URL processing logic
func TestPageCreate_Handle_URLProcessing(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		expectedURL  string
		expectedName string
	}{
		{
			name:         "URL with leading slash",
			inputURL:     "/test-page",
			expectedURL:  "/test-page",
			expectedName: "TEST-PAGE",
		},
		{
			name:         "URL without leading slash",
			inputURL:     "test-page",
			expectedURL:  "/test-page",
			expectedName: "TEST-PAGE",
		},
		{
			name:         "URL with trailing slash",
			inputURL:     "/test-page/",
			expectedURL:  "/test-page",
			expectedName: "TEST-PAGE",
		},
		{
			name:         "URL with multiple slashes",
			inputURL:     "/blog/test-post/",
			expectedURL:  "/blog/test-post",
			expectedName: "BLOG TEST-POST",
		},
		{
			name:         "Root URL",
			inputURL:     "/",
			expectedURL:  "/",
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockPageStore{}
			beforeSaveCalled := false
			var processedPage *Page
			beforeSave := func(ctx context.Context, page *Page) error {
				beforeSaveCalled = true
				processedPage = page
				return nil
			}
			handler := NewPageCreate[Resolver](mockStore, beforeSave)

			// Create request body
			requestBody := `{"url": "` + tt.inputURL + `", "template": "test.html", "title": "Test Title"}`
			req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			event := &Event{}
			woResp := &wo.Response{ResponseWriter: resp}
			event.Reset(woResp, req, &MockPageTheme{})

			site := &Site{ID: "site1"}
			event.SetSite(site)

			// Mock parent page lookup (should return not found) - allow multiple calls
			mockStore.On("FindByURL", mock.Anything, ID("site1"), mock.Anything).Return(nil, errors.New("not found")).Maybe().Maybe()

			// Mock store Save
			mockStore.On("Save", mock.Anything, mock.AnythingOfType("[]*pages.Page")).Return(nil).Run(func(args mock.Arguments) {
				pages := args.Get(1).([]*Page)
				pages[0].ID = "test-id"
			})

			err := handler.Handle(event)

			assert.True(t, beforeSaveCalled)
			assert.NotNil(t, processedPage)
			assert.Equal(t, tt.expectedURL, processedPage.CustomURL)
			assert.Equal(t, tt.expectedName, processedPage.Name)

			mockStore.AssertExpectations(t)

			// Should return redirect error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "302")
		})
	}
}

// TestPageCreate_Handle_ParentPage tests parent page assignment
func TestPageCreate_Handle_ParentPage(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		expectParent bool
		parentURL    string
	}{
		{
			name:         "Root URL - no parent",
			inputURL:     "/",
			expectParent: false,
		},
		{
			name:         "Top level page - no parent",
			inputURL:     "/about",
			expectParent: false,
		},
		{
			name:         "Nested page - has parent",
			inputURL:     "/blog/post",
			expectParent: true,
			parentURL:    "/blog",
		},
		{
			name:         "Deeply nested page - has parent",
			inputURL:     "/category/subcategory/post",
			expectParent: true,
			parentURL:    "/category/subcategory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &MockPageStore{}
			handler := NewPageCreate[Resolver](mockStore, nil)

			// Create request body
			requestBody := `{"url": "` + tt.inputURL + `", "template": "test.html", "title": "Test Title"}`
			req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			event := &Event{}
			woResp := &wo.Response{ResponseWriter: resp}
			event.Reset(woResp, req, &MockPageTheme{})

			site := &Site{ID: "site1"}
			event.SetSite(site)

			// Mock parent page lookup
			if tt.expectParent {
				parentPage := &Page{ID: "parent1", URL: tt.parentURL}
				mockStore.On("FindByURL", mock.Anything, ID("site1"), tt.parentURL).Return(parentPage, nil)
			} else {
				mockStore.On("FindByURL", mock.Anything, ID("site1"), mock.Anything).Return(nil, errors.New("not found")).Maybe().Maybe()
			}

			var savedPage *Page
			// Mock store Save
			mockStore.On("Save", mock.Anything, mock.AnythingOfType("[]*pages.Page")).Return(nil).Run(func(args mock.Arguments) {
				pages := args.Get(1).([]*Page)
				savedPage = pages[0]
				savedPage.ID = "test-id"
			})

			err := handler.Handle(event)

			mockStore.AssertExpectations(t)

			assert.NotNil(t, savedPage)
			if tt.expectParent {
				assert.NotNil(t, savedPage.ParentID)
				assert.Equal(t, ID("parent1"), *savedPage.ParentID)
			} else {
				assert.Nil(t, savedPage.ParentID)
			}

			// Should return redirect error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "302")
		})
	}
}

// TestPageCreate_Handle_BeforeSaveError tests beforeSave callback error handling
func TestPageCreate_Handle_BeforeSaveError(t *testing.T) {
	mockStore := &MockPageStore{}
	beforeSaveErr := errors.New("before save error")
	beforeSave := func(ctx context.Context, page *Page) error {
		return beforeSaveErr
	}
	handler := NewPageCreate[Resolver](mockStore, beforeSave)

	// Create request body
	requestBody := `{"url": "/test", "template": "test.html", "title": "Test Title"}`
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	// Mock parent page lookup (should return not found)
	mockStore.On("FindByURL", mock.Anything, ID("site1"), mock.Anything).Return(nil, errors.New("not found")).Maybe()

	err := handler.Handle(event)

	assert.Error(t, err)
	assert.Equal(t, beforeSaveErr, err)
	mockStore.AssertExpectations(t)
}

// TestPageCreate_Handle_SaveError tests store save error handling
func TestPageCreate_Handle_SaveError(t *testing.T) {
	mockStore := &MockPageStore{}
	saveErr := errors.New("save error")
	handler := NewPageCreate[Resolver](mockStore, nil)

	// Create request body
	requestBody := `{"url": "/test", "template": "test.html", "title": "Test Title"}`
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1"}
	event.SetSite(site)

	// Mock parent page lookup (should return not found)
	mockStore.On("FindByURL", mock.Anything, ID("site1"), mock.Anything).Return(nil, errors.New("not found")).Maybe()

	// Mock store Save error
	mockStore.On("Save", mock.Anything, mock.AnythingOfType("[]*pages.Page")).Return(saveErr)

	err := handler.Handle(event)

	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
	mockStore.AssertExpectations(t)
}

// TestPageCreate_Handle_Success tests successful page creation
func TestPageCreate_Handle_Success(t *testing.T) {
	mockStore := &MockPageStore{}
	handler := NewPageCreate[Resolver](mockStore, nil)

	// Create request body
	requestBody := `{"url": "/success-page", "template": "success.html", "title": "Success Page"}`
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{ID: "site1", Host: "example.com"}
	event.SetSite(site)

	// Mock parent page lookup (should return not found)
	mockStore.On("FindByURL", mock.Anything, ID("site1"), mock.Anything).Return(nil, errors.New("not found")).Maybe()

	// Mock successful store Save
	mockStore.On("Save", mock.Anything, mock.AnythingOfType("[]*pages.Page")).Return(nil).Run(func(args mock.Arguments) {
		pages := args.Get(1).([]*Page)
		page := pages[0]
		page.ID = "created-page-id" // Simulate ID assignment
	})

	err := handler.Handle(event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "302")
	mockStore.AssertExpectations(t)
}

// TestPageCreate_Handle_Comprehensive tests the complete flow with realistic data
func TestPageCreate_Handle_Comprehensive(t *testing.T) {
	mockStore := &MockPageStore{}

	var savedPage *Page
	beforeSave := func(ctx context.Context, page *Page) error {
		// Add some custom processing
		page.Metadata = map[string]any{
			"processed": true,
		}
		savedPage = page
		return nil
	}
	handler := NewPageCreate[Resolver](mockStore, beforeSave)

	// Create request body
	requestBody := `{"url": "/comprehensive/test", "template": "comprehensive.html", "title": "Comprehensive Test Page"}`
	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	event := &Event{}
	woResp := &wo.Response{ResponseWriter: resp}
	event.Reset(woResp, req, &MockPageTheme{})

	site := &Site{
		ID:   "site1",
		Host: "example.com",
	}
	event.SetSite(site)

	// Mock parent page
	parentPage := &Page{
		ID:   "parent1",
		URL:  "/comprehensive",
		Name: "Comprehensive",
	}
	mockStore.On("FindByURL", mock.Anything, ID("site1"), "/comprehensive").Return(parentPage, nil)

	// Mock successful store Save
	mockStore.On("Save", mock.Anything, mock.AnythingOfType("[]*pages.Page")).Return(nil).Run(func(args mock.Arguments) {
		pages := args.Get(1).([]*Page)
		page := pages[0]
		page.ID = "comprehensive-page-id"
	})

	err := handler.Handle(event)

	// Verify the saved page has all expected properties
	assert.NotNil(t, savedPage)
	assert.Equal(t, PageCMS, savedPage.Pattern)
	assert.Equal(t, "COMPREHENSIVE TEST", savedPage.Name)
	assert.Equal(t, site, savedPage.Site)
	assert.Equal(t, ID("site1"), savedPage.SiteID)
	assert.Equal(t, "Comprehensive Test Page", savedPage.Title)
	assert.Equal(t, "/comprehensive/test", savedPage.CustomURL)
	assert.Equal(t, "comprehensive.html", savedPage.Template)
	assert.True(t, savedPage.Decorate)
	assert.Equal(t, ID("parent1"), *savedPage.ParentID)
	assert.Equal(t, parentPage, savedPage.Parent)
	assert.True(t, savedPage.Metadata["processed"].(bool))

	// Check redirect response
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "302")

	mockStore.AssertExpectations(t)
}
