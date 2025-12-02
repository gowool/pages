package pages

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPageManager implements PageManager interface using testify/mock
type MockPageManager struct {
	mock.Mock
}

func (m *MockPageManager) GetByID(ctx context.Context, id ID) (*Page, error) {
	args := m.Called(ctx, id)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByURL(ctx context.Context, site *Site, url string) (*Page, error) {
	args := m.Called(ctx, site, url)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByPattern(ctx context.Context, site *Site, pattern string) (*Page, error) {
	args := m.Called(ctx, site, pattern)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPageManager) GetByAlias(ctx context.Context, site *Site, alias string) (*Page, error) {
	args := m.Called(ctx, site, alias)
	if page := args.Get(0); page != nil {
		return page.(*Page), args.Error(1)
	}
	return nil, args.Error(1)
}

// Helper functions for testing
func NewTestSite() *Site {
	return &Site{
		ID:     ID(uuid.New().String()),
		Scheme: "https",
		Host:   "example.com",
	}
}

func NewTestPage(pattern string) *Page {
	page := NewPage()
	page.ID = ID(uuid.New().String())
	page.Pattern = pattern
	page.Name = "Test Page"
	page.Site = NewTestSite()
	page.Status = PublishStatus
	return page
}

func TestNewPageURLGenerator(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)

	// Test that generator implements interface
	assert.NotNil(t, generator, "NewPageURLGenerator() should return non-nil URLGenerator")
	var _ URLGenerator = generator

	// Test with nil site
	ctx := context.Background()
	_, err := generator.Generate(ctx, nil, "")
	assert.Error(t, err, "Generate() with nil site should return error")

	// Test with nil page
	site := NewTestSite()
	_, err = generator.GenerateByPage(site, nil)
	assert.Error(t, err, "GenerateByPage() with nil page should return error")
}

func TestPageManager(t *testing.T) {
	manager := &MockPageManager{}

	// Test with empty ID
	ctx := context.Background()
	manager.On("GetByID", ctx, ID("")).Return(nil, assert.AnError)
	_, err := manager.GetByID(ctx, "")
	assert.Error(t, err, "GetByID() with empty ID should return error")

	// Test with nil site
	manager.On("GetByURL", ctx, (*Site)(nil), "").Return(nil, assert.AnError)
	_, err = manager.GetByURL(ctx, nil, "")
	assert.Error(t, err, "GetByURL() with nil site should return error")

	site := NewTestSite()
	manager.On("GetByURL", ctx, site, "").Return(nil, assert.AnError)
	_, err = manager.GetByURL(ctx, site, "")
	assert.Error(t, err, "GetByURL() with empty URL should return error")
}

func TestPageURLGenerator_GetByID(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)

	// Test simple ID lookup
	ctx := context.Background()
	site := NewTestPage("/test")
	mockManager.On("GetByID", ctx, site.ID).Return(site, nil)

	result, err := generator.GenerateByID(ctx, site.Site, site.ID)
	assert.NoError(t, err, "GenerateByID() with valid ID should not return error")
	assert.Equal(t, "https://example.com/test", result, "GenerateByID() should return correct URL")

	// Test with invalid ID
	invalidID := ID("invalid-id")
	mockManager.On("GetByID", ctx, invalidID).Return(nil, assert.AnError)
	_, err = generator.GenerateByID(ctx, site.Site, invalidID)
	assert.Error(t, err, "GenerateByID() with invalid ID should return error")
}

func TestPageURLGenerator_GetByAlias(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)

	// Test simple alias lookup
	ctx := context.Background()
	site := NewTestPage("/test")
	site.SetAlias("test-alias")
	mockManager.On("GetByAlias", ctx, site.Site, "test-alias").Return(site, nil)

	result, err := generator.GenerateByAlias(ctx, site.Site, "test-alias")
	assert.NoError(t, err, "GenerateByAlias() with valid alias should not return error")
	assert.Equal(t, "https://example.com/test", result, "GenerateByAlias() should return correct URL")

	// Test with invalid alias
	mockManager.On("GetByAlias", ctx, site.Site, "invalid-alias").Return(nil, assert.AnError)
	_, err = generator.GenerateByAlias(ctx, site.Site, "invalid-alias")
	assert.Error(t, err, "GenerateByAlias() with invalid alias should return error")

	// Test with empty alias
	mockManager.On("GetByAlias", ctx, site.Site, "").Return(nil, assert.AnError)
	_, err = generator.GenerateByAlias(ctx, site.Site, "")
	assert.Error(t, err, "GenerateByAlias() with empty alias should return error")
}

func TestPageURLGenerator_GetByPattern(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)

	// Test simple pattern lookup
	ctx := context.Background()
	site := NewTestPage("/test/{name}")
	mockManager.On("GetByPattern", ctx, site.Site, "/test/{name}").Return(site, nil)

	result, err := generator.GenerateByPattern(ctx, site.Site, "/test/{name}")
	assert.NoError(t, err, "GenerateByPattern() with valid pattern should not return error")
	assert.Equal(t, "https://example.com/test/{name}", result, "GenerateByPattern() should return correct URL")

	// Test with invalid pattern
	mockManager.On("GetByPattern", ctx, site.Site, "/invalid/{pattern}").Return(nil, assert.AnError)
	_, err = generator.GenerateByPattern(ctx, site.Site, "/invalid/{pattern}")
	assert.Error(t, err, "GenerateByPattern() with invalid pattern should return error")

	// Test with empty pattern
	mockManager.On("GetByPattern", ctx, site.Site, "").Return(nil, assert.AnError)
	_, err = generator.GenerateByPattern(ctx, site.Site, "")
	assert.Error(t, err, "GenerateByPattern() with empty pattern should return error")
}

func TestPageURLGenerator_GenerateByURL(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create a page with URL
	page := NewTestPage("/cms-page")
	page.Pattern = PageCMS
	page.URL = "/test/page"
	mockManager.On("GetByURL", ctx, site, "/test/page").Return(page, nil)

	// Test direct URL generation
	result, err := generator.GenerateByURL(ctx, site, "/test/page")
	assert.NoError(t, err, "GenerateByURL() with valid URL should not return error")
	assert.Equal(t, "https://example.com/test/page", result, "GenerateByURL() should return correct URL")

	// Test with invalid URL
	mockManager.On("GetByURL", ctx, site, "/invalid/page").Return(nil, assert.AnError)
	_, err = generator.GenerateByURL(ctx, site, "/invalid/page")
	assert.Error(t, err, "GenerateByURL() with invalid URL should return error")
}

func TestPageURLGenerator_GenerateByPattern(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create a page with pattern
	page := NewTestPage("/test/{name}")
	mockManager.On("GetByPattern", ctx, site, "/test/{name}").Return(page, nil)

	// Test direct pattern-based generation
	result, err := generator.GenerateByPattern(ctx, site, "/test/{name}", "{name}", "test-value")
	assert.NoError(t, err, "GenerateByPattern() with valid pattern should not return error")
	assert.Equal(t, "https://example.com/test/test-value", result, "GenerateByPattern() should return correct URL")
}

func TestPageURLGenerator_GenerateByPageID(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	ctx := context.Background()

	// Create a page with ID
	page := NewTestPage("/test/{id}/test/{value}")
	mockManager.On("GetByID", ctx, page.ID).Return(page, nil)

	// Test ID-based generation using the main Generate method
	result, err := generator.Generate(ctx, page.Site, page.ID, "{id}", "test-id", "{value}", "test-value")
	assert.NoError(t, err, "Generate() with valid ID should not return error")
	assert.Equal(t, "https://example.com/test/test-id/test/test-value", result, "Generate() should return correct URL")
}

func TestPageURLGenerator_ErrorCases(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	site := NewTestSite()

	// Test with nil page
	_, err := generator.GenerateByPage(site, nil)
	assert.Error(t, err, "GenerateByPage() with nil page should return error")

	// Test with CMS page (should work for CMS pages)
	cmsPage := NewTestPage(PageCMS)
	cmsPage.URL = "/cms-page"
	cmsPage.Site = site

	result, err := generator.GenerateByPage(site, cmsPage)
	assert.NoError(t, err, "GenerateByPage() with CMS page should not return error")
	assert.Equal(t, "https://example.com/cms-page", result, "GenerateByPage() with CMS page should return correct URL")

	// Test with internal page (should return empty string)
	internalPage := NewTestPage(PageInternalCreate)
	internalPage.Site = site

	result, err = generator.GenerateByPage(site, internalPage)
	assert.NoError(t, err, "GenerateByPage() with internal page should not return error")
	assert.Empty(t, result, "GenerateByPage() with internal page should return empty string")
}

func TestPageURLGenerator_PatternMatching(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create pages with patterns
	userPage := NewTestPage("/users/{username}")
	mockManager.On("GetByPattern", ctx, site, "/users/{username}").Return(userPage, nil)

	blogPage := NewTestPage("/blog/{category}/{slug}")
	mockManager.On("GetByPattern", ctx, site, "/blog/{category}/{slug}").Return(blogPage, nil)

	// Test exact pattern match
	result, err := generator.GenerateByPattern(ctx, site, "/users/{username}", "{username}", "john.doe")
	assert.NoError(t, err, "GenerateByPattern() with valid pattern should not return error")
	assert.Equal(t, "https://example.com/users/john.doe", result, "GenerateByPattern() should return correct URL")

	// Test pattern with multiple parameters
	result, err = generator.GenerateByPattern(ctx, site, "/blog/{category}/{slug}", "{category}", "technology", "{slug}", "golang-tips")
	assert.NoError(t, err, "GenerateByPattern() with valid pattern should not return error")
	assert.Equal(t, "https://example.com/blog/technology/golang-tips", result, "GenerateByPattern() should return correct URL")
}

func TestPageURLGenerator_GenerateMethod(t *testing.T) {
	mockManager := &MockPageManager{}
	generator := NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Test with Page struct
	page := NewTestPage("/test")
	result, err := generator.Generate(ctx, site, page)
	assert.NoError(t, err, "Generate() with Page should not return error")
	assert.Equal(t, "https://example.com/test", result, "Generate() with Page should return correct URL")

	// Test with *Page (same as above)
	result, err = generator.Generate(ctx, site, page)
	assert.NoError(t, err, "Generate() with *Page should not return error")
	assert.Equal(t, "https://example.com/test", result, "Generate() with *Page should return correct URL")

	// Test with ID
	mockManager.On("GetByID", ctx, page.ID).Return(page, nil)
	result, err = generator.Generate(ctx, site, page.ID)
	assert.NoError(t, err, "Generate() with ID should not return error")
	assert.Equal(t, "https://example.com/test", result, "Generate() with ID should return correct URL")

	// Test with pattern
	mockManager.On("GetByPattern", ctx, site, "/test").Return(page, nil)
	result, err = generator.Generate(ctx, site, "/test")
	assert.NoError(t, err, "Generate() with pattern should not return error")
	assert.Equal(t, "https://example.com/test", result, "Generate() with pattern should return correct URL")

	// Test with CMS pattern
	cmsPage := NewTestPage(PageCMS)
	cmsPage.URL = "/cms-page"
	cmsPage.Site = site
	mockManager.On("GetByURL", ctx, site, "/cms-page").Return(cmsPage, nil)
	result, err = generator.Generate(ctx, site, PageCMS, "url", "/cms-page")
	assert.NoError(t, err, "Generate() with CMS pattern should not return error")
	assert.Equal(t, "https://example.com/cms-page", result, "Generate() with CMS pattern should return correct URL")

	// Test with unsupported type
	_, err = generator.Generate(ctx, site, 123)
	assert.Error(t, err, "Generate() with unsupported type should return error")

	// Test with nil site
	_, err = generator.Generate(ctx, nil, page)
	assert.Error(t, err, "Generate() with nil site should return error")
}
