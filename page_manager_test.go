package pages

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPageManager(t *testing.T) {
	tests := []struct {
		name    string
		storage PageStorage
		want    *DefaultPageManager
		wantPan bool
	}{
		{
			name:    "Valid storage",
			storage: &MockPageStorage{},
			want:    &DefaultPageManager{storage: &MockPageStorage{}},
			wantPan: false,
		},
		{
			name:    "Nil storage should panic",
			storage: nil,
			want:    nil,
			wantPan: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPan {
				assert.Panics(t, func() {
					NewPageManager(tt.storage)
				}, "NewPageManager() should panic with nil storage")
			} else {
				got := NewPageManager(tt.storage)
				assert.NotNil(t, got, "NewPageManager() should return non-nil manager")
				assert.Equal(t, tt.want.storage, got.storage, "NewPageManager() should set storage correctly")
			}
		})
	}
}

func TestDefaultPageManager_GetByID(t *testing.T) {
	ctx := context.Background()
	pageID := ID("test-page-id")
	expectedPage := &Page{ID: pageID, Name: "Test Page"}

	tests := []struct {
		name      string
		setupMock func(*MockPageStorage)
		pageID    ID
		wantPage  *Page
		wantError bool
		errorType error
	}{
		{
			name: "Successfully find page by ID",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByID", ctx, pageID).Return(expectedPage, nil)
			},
			pageID:    pageID,
			wantPage:  expectedPage,
			wantError: false,
		},
		{
			name: "Page not found by ID",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByID", ctx, pageID).Return(nil, ErrPageNotFound)
			},
			pageID:    pageID,
			wantPage:  nil,
			wantError: true,
			errorType: ErrPageNotFound,
		},
		{
			name: "Storage returns unexpected error",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByID", ctx, pageID).Return(nil, assert.AnError)
			},
			pageID:    pageID,
			wantPage:  nil,
			wantError: true,
			errorType: assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockPageStorage{}
			tt.setupMock(mockStorage)

			manager := &DefaultPageManager{storage: mockStorage}
			got, err := manager.GetByID(ctx, tt.pageID)

			if tt.wantError {
				assert.Error(t, err, "GetByID() should return an error")
				assert.Equal(t, tt.errorType, err, "GetByID() should return the expected error type")
				assert.Nil(t, got, "GetByID() should return nil page on error")
			} else {
				assert.NoError(t, err, "GetByID() should not return an error")
				assert.Equal(t, tt.wantPage, got, "GetByID() should return the expected page")
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestDefaultPageManager_GetByURL(t *testing.T) {
	ctx := context.Background()
	siteID := ID("test-site-id")
	site := &Site{ID: siteID, Name: "Test Site"}
	url := "/test/page"
	expectedPage := &Page{ID: ID("test-page-id"), URL: url, SiteID: siteID}

	tests := []struct {
		name      string
		setupMock func(*MockPageStorage)
		site      *Site
		url       string
		wantPage  *Page
		wantError bool
		errorType error
		wantPan   bool
	}{
		{
			name: "Successfully find page by URL",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByURL", ctx, siteID, url).Return(expectedPage, nil)
			},
			site:      site,
			url:       url,
			wantPage:  expectedPage,
			wantError: false,
			wantPan:   false,
		},
		{
			name: "Page not found by URL",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByURL", ctx, siteID, url).Return(nil, ErrPageNotFound)
			},
			site:      site,
			url:       url,
			wantPage:  nil,
			wantError: true,
			errorType: ErrPageNotFound,
			wantPan:   false,
		},
		{
			name:      "Nil site should panic",
			setupMock: func(m *MockPageStorage) {},
			site:      nil,
			url:       url,
			wantPage:  nil,
			wantError: false,
			wantPan:   true,
		},
		{
			name: "Storage returns unexpected error",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByURL", ctx, siteID, url).Return(nil, assert.AnError)
			},
			site:      site,
			url:       url,
			wantPage:  nil,
			wantError: true,
			errorType: assert.AnError,
			wantPan:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPan {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				assert.Panics(t, func() {
					manager.GetByURL(ctx, tt.site, tt.url)
				}, "GetByURL() should panic with nil site")
			} else {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				got, err := manager.GetByURL(ctx, tt.site, tt.url)

				if tt.wantError {
					assert.Error(t, err, "GetByURL() should return an error")
					assert.Equal(t, tt.errorType, err, "GetByURL() should return the expected error type")
					assert.Nil(t, got, "GetByURL() should return nil page on error")
				} else {
					assert.NoError(t, err, "GetByURL() should not return an error")
					assert.Equal(t, tt.wantPage, got, "GetByURL() should return the expected page")
				}

				mockStorage.AssertExpectations(t)
			}
		})
	}
}

func TestDefaultPageManager_GetByPattern(t *testing.T) {
	ctx := context.Background()
	siteID := ID("test-site-id")
	site := &Site{ID: siteID, Name: "Test Site"}
	pattern := "/test/{slug}"
	expectedPage := &Page{ID: ID("test-page-id"), Pattern: pattern, SiteID: siteID}

	tests := []struct {
		name      string
		setupMock func(*MockPageStorage)
		site      *Site
		pattern   string
		wantPage  *Page
		wantError bool
		errorType error
		wantPan   bool
	}{
		{
			name: "Successfully find page by pattern",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByPattern", ctx, siteID, pattern).Return(expectedPage, nil)
			},
			site:      site,
			pattern:   pattern,
			wantPage:  expectedPage,
			wantError: false,
			wantPan:   false,
		},
		{
			name: "Page not found by pattern",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByPattern", ctx, siteID, pattern).Return(nil, ErrPageNotFound)
			},
			site:      site,
			pattern:   pattern,
			wantPage:  nil,
			wantError: true,
			errorType: ErrPageNotFound,
			wantPan:   false,
		},
		{
			name:      "Nil site should panic",
			setupMock: func(m *MockPageStorage) {},
			site:      nil,
			pattern:   pattern,
			wantPage:  nil,
			wantError: false,
			wantPan:   true,
		},
		{
			name: "Storage returns unexpected error",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByPattern", ctx, siteID, pattern).Return(nil, assert.AnError)
			},
			site:      site,
			pattern:   pattern,
			wantPage:  nil,
			wantError: true,
			errorType: assert.AnError,
			wantPan:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPan {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				assert.Panics(t, func() {
					manager.GetByPattern(ctx, tt.site, tt.pattern)
				}, "GetByPattern() should panic with nil site")
			} else {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				got, err := manager.GetByPattern(ctx, tt.site, tt.pattern)

				if tt.wantError {
					assert.Error(t, err, "GetByPattern() should return an error")
					assert.Equal(t, tt.errorType, err, "GetByPattern() should return the expected error type")
					assert.Nil(t, got, "GetByPattern() should return nil page on error")
				} else {
					assert.NoError(t, err, "GetByPattern() should not return an error")
					assert.Equal(t, tt.wantPage, got, "GetByPattern() should return the expected page")
				}

				mockStorage.AssertExpectations(t)
			}
		})
	}
}

func TestDefaultPageManager_GetByAlias(t *testing.T) {
	ctx := context.Background()
	siteID := ID("test-site-id")
	site := &Site{ID: siteID, Name: "Test Site"}
	alias := "test-alias"
	aliasWithPrefix := PageAliasPrefix + alias
	expectedPage := &Page{ID: ID("test-page-id"), Alias: aliasWithPrefix, SiteID: siteID}

	tests := []struct {
		name      string
		setupMock func(*MockPageStorage)
		site      *Site
		alias     string
		wantPage  *Page
		wantError bool
		errorType error
		wantPan   bool
	}{
		{
			name: "Successfully find page by alias without prefix",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByAlias", ctx, siteID, aliasWithPrefix).Return(expectedPage, nil)
			},
			site:      site,
			alias:     alias,
			wantPage:  expectedPage,
			wantError: false,
			wantPan:   false,
		},
		{
			name: "Successfully find page by alias with prefix",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByAlias", ctx, siteID, aliasWithPrefix).Return(expectedPage, nil)
			},
			site:      site,
			alias:     aliasWithPrefix,
			wantPage:  expectedPage,
			wantError: false,
			wantPan:   false,
		},
		{
			name: "Page not found by alias",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByAlias", ctx, siteID, aliasWithPrefix).Return(nil, ErrPageNotFound)
			},
			site:      site,
			alias:     alias,
			wantPage:  nil,
			wantError: true,
			errorType: ErrPageNotFound,
			wantPan:   false,
		},
		{
			name:      "Nil site should panic",
			setupMock: func(m *MockPageStorage) {},
			site:      nil,
			alias:     alias,
			wantPage:  nil,
			wantError: false,
			wantPan:   true,
		},
		{
			name: "Storage returns unexpected error",
			setupMock: func(m *MockPageStorage) {
				m.On("FindByAlias", ctx, siteID, aliasWithPrefix).Return(nil, assert.AnError)
			},
			site:      site,
			alias:     alias,
			wantPage:  nil,
			wantError: true,
			errorType: assert.AnError,
			wantPan:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPan {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				assert.Panics(t, func() {
					manager.GetByAlias(ctx, tt.site, tt.alias)
				}, "GetByAlias() should panic with nil site")
			} else {
				mockStorage := &MockPageStorage{}
				tt.setupMock(mockStorage)
				manager := &DefaultPageManager{storage: mockStorage}

				got, err := manager.GetByAlias(ctx, tt.site, tt.alias)

				if tt.wantError {
					assert.Error(t, err, "GetByAlias() should return an error")
					assert.Equal(t, tt.errorType, err, "GetByAlias() should return the expected error type")
					assert.Nil(t, got, "GetByAlias() should return nil page on error")
				} else {
					assert.NoError(t, err, "GetByAlias() should not return an error")
					assert.Equal(t, tt.wantPage, got, "GetByAlias() should return the expected page")
				}

				mockStorage.AssertExpectations(t)
			}
		})
	}
}

func TestDefaultPageManager_InterfaceCompliance(t *testing.T) {
	// Test that DefaultPageManager implements PageManager interface
	var _ PageManager = (*DefaultPageManager)(nil)

	mockStorage := &MockPageStorage{}
	manager := NewPageManager(mockStorage)

	assert.NotNil(t, manager, "DefaultPageManager should implement PageManager interface")
	assert.Implements(t, (*PageManager)(nil), manager, "DefaultPageManager should implement PageManager interface")
}

func TestDefaultPageManager_ContextPropagation(t *testing.T) {
	// Test that context is properly passed to storage methods
	ctx := context.Background()
	siteID := ID("test-site-id")
	site := &Site{ID: siteID}
	pageID := ID("test-page-id")

	tests := []struct {
		name     string
		testFunc func(*MockPageStorage, *DefaultPageManager)
	}{
		{
			name: "GetByID propagates context",
			testFunc: func(m *MockPageStorage, pm *DefaultPageManager) {
				m.On("FindByID", ctx, pageID).Return(&Page{}, nil)
				pm.GetByID(ctx, pageID)
			},
		},
		{
			name: "GetByURL propagates context",
			testFunc: func(m *MockPageStorage, pm *DefaultPageManager) {
				m.On("FindByURL", ctx, siteID, "/test").Return(&Page{}, nil)
				pm.GetByURL(ctx, site, "/test")
			},
		},
		{
			name: "GetByPattern propagates context",
			testFunc: func(m *MockPageStorage, pm *DefaultPageManager) {
				m.On("FindByPattern", ctx, siteID, "/test/{slug}").Return(&Page{}, nil)
				pm.GetByPattern(ctx, site, "/test/{slug}")
			},
		},
		{
			name: "GetByAlias propagates context",
			testFunc: func(m *MockPageStorage, pm *DefaultPageManager) {
				m.On("FindByAlias", ctx, siteID, PageAliasPrefix+"test").Return(&Page{}, nil)
				pm.GetByAlias(ctx, site, "test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockPageStorage{}
			manager := &DefaultPageManager{storage: mockStorage}

			tt.testFunc(mockStorage, manager)
			mockStorage.AssertExpectations(t)
		})
	}
}
