package pages_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/gowool/pages"
)

// MockPageManager implements PageManager interface for testing
type MockPageManager struct {
	pages     map[pages.ID]*pages.Page
	byURL     map[string]*pages.Page
	byPattern map[string]*pages.Page
	byAlias   map[string]*pages.Page
}

func NewMockPageManager() *MockPageManager {
	return &MockPageManager{
		pages:     make(map[pages.ID]*pages.Page),
		byURL:     make(map[string]*pages.Page),
		byPattern: make(map[string]*pages.Page),
		byAlias:   make(map[string]*pages.Page),
	}
}

func (m *MockPageManager) AddPage(page *pages.Page) {
	if page.ID != "" {
		m.pages[page.ID] = page
	}
	if page.URL != "" {
		m.byURL[page.URL] = page
	}
	if page.Pattern != "" {
		m.byPattern[page.Pattern] = page
	}
	if page.Alias != "" {
		m.byAlias[page.Alias] = page
	}
}

func (m *MockPageManager) GetByID(ctx context.Context, id pages.ID) (*pages.Page, error) {
	if page, exists := m.pages[id]; exists {
		return page, nil
	}
	return nil, errors.New("page not found")
}

func (m *MockPageManager) GetByURL(ctx context.Context, site *pages.Site, url string) (*pages.Page, error) {
	if page, exists := m.byURL[url]; exists {
		return page, nil
	}
	return nil, errors.New("page not found")
}

func (m *MockPageManager) GetByPattern(ctx context.Context, site *pages.Site, pattern string) (*pages.Page, error) {
	if page, exists := m.byPattern[pattern]; exists {
		return page, nil
	}
	return nil, errors.New("page not found")
}

func (m *MockPageManager) GetByAlias(ctx context.Context, site *pages.Site, alias string) (*pages.Page, error) {
	if !strings.HasPrefix(alias, pages.PageAliasPrefix) {
		alias = pages.PageAliasPrefix + alias
	}
	if page, exists := m.byAlias[alias]; exists {
		return page, nil
	}
	return nil, errors.New("page not found")
}

// Helper functions for testing
func NewTestSite() *pages.Site {
	return &pages.Site{
		ID:     pages.ID(uuid.New().String()),
		Scheme: "https",
		Host:   "example.com",
	}
}

func NewTestPage(pattern string) *pages.Page {
	page := pages.NewPage()
	page.ID = pages.ID(uuid.New().String())
	page.Pattern = pattern
	page.Name = "Test Page"
	page.Site = NewTestSite()
	page.Status = pages.PublishStatus
	return page
}

func TestNewPageURLGenerator(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)

	// Test that generator implements interface
	var _ pages.URLGenerator = generator
	if generator == nil {
		t.Errorf("NewPageURLGenerator() should return non-nil URLGenerator")
	}

	// Test with nil site
	ctx := context.Background()
	if _, err := generator.Generate(ctx, nil, ""); err == nil {
		t.Errorf("Generate() with nil site should return error")
	}

	// Test with nil page
	site := NewTestSite()
	if _, err := generator.GenerateByPage(site, nil); err == nil {
		t.Errorf("GenerateByPage() with nil page should return error")
	}
}

func TestPageManager(t *testing.T) {
	manager := NewMockPageManager()

	// Test with nil context
	ctx := context.Background()
	if _, err := manager.GetByID(ctx, ""); err == nil {
		t.Errorf("GetByID() with empty ID should return error")
	}

	if _, err := manager.GetByURL(ctx, nil, ""); err == nil {
		t.Errorf("GetByURL() with nil site should return error")
	}

	site := NewTestSite()
	if _, err := manager.GetByURL(ctx, site, ""); err == nil {
		t.Errorf("GetByURL() with empty URL should return error")
	}
}

func TestPageURLGenerator_GetByID(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)

	// Test simple ID lookup
	ctx := context.Background()
	site := NewTestPage("/test")
	mockManager.AddPage(site)

	if _, err := generator.GenerateByID(ctx, site.Site, site.ID); err != nil {
		t.Errorf("GenerateByID() with valid ID = %v, want no error", err)
	}

	// Test with invalid ID
	if _, err := generator.GenerateByID(ctx, site.Site, "invalid-id"); err == nil {
		t.Errorf("GenerateByID() with invalid ID should return error")
	}
}

func TestPageURLGenerator_GetByAlias(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)

	// Test simple alias lookup
	ctx := context.Background()
	site := NewTestPage("/test")
	site.SetAlias("test-alias")
	mockManager.AddPage(site)

	if _, err := generator.GenerateByAlias(ctx, site.Site, "test-alias"); err != nil {
		t.Errorf("GenerateByAlias() with valid alias = %v, want no error", err)
	}

	// Test with invalid alias
	if _, err := generator.GenerateByAlias(ctx, site.Site, "invalid-alias"); err == nil {
		t.Errorf("GenerateByAlias() with invalid alias should return error")
	}

	// Test with empty alias
	if _, err := generator.GenerateByAlias(ctx, site.Site, ""); err == nil {
		t.Errorf("GenerateByAlias() with empty alias should return error")
	}
}

func TestPageURLGenerator_GetByPattern(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)

	// Test simple pattern lookup
	ctx := context.Background()
	site := NewTestPage("/test/{name}")
	mockManager.AddPage(site)

	if _, err := generator.GenerateByPattern(ctx, site.Site, "/test/{name}"); err != nil {
		t.Errorf("GenerateByPattern() with valid pattern = %v, want no error", err)
	}

	// Test with invalid pattern
	if _, err := generator.GenerateByPattern(ctx, site.Site, "/invalid/{pattern}"); err == nil {
		t.Errorf("GenerateByPattern() with invalid pattern should return error")
	}

	// Test with empty pattern
	if _, err := generator.GenerateByPattern(ctx, site.Site, ""); err == nil {
		t.Errorf("GenerateByPattern() with empty pattern should return error")
	}
}

func TestPageURLGenerator_GenerateByURL(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create a page with URL
	page := NewTestPage("/cms-page")
	page.Pattern = pages.PageCMS
	page.URL = "/test/page"
	mockManager.AddPage(page)

	// Test direct URL generation
	result, err := generator.GenerateByURL(ctx, site, "/test/page")
	if err != nil {
		t.Errorf("GenerateByURL() error = %v, want no error", err)
	}
	if result != "https://example.com/test/page" {
		t.Errorf("GenerateByURL() result = %v, want %v", result, "https://example.com/test/page")
	}

	// Test with invalid URL
	if _, err := generator.GenerateByURL(ctx, site, "/invalid/page"); err == nil {
		t.Errorf("GenerateByURL() with invalid URL should return error")
	}
}

func TestPageURLGenerator_GenerateByPattern(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create a page with pattern
	page := NewTestPage("/test/{name}")
	mockManager.AddPage(page)

	// Test direct pattern-based generation
	result, err := generator.GenerateByPattern(ctx, site, "/test/{name}", "{name}", "test-value")
	if err != nil {
		t.Errorf("GenerateByPattern() error = %v, want no error", err)
	}
	if result != "https://example.com/test/test-value" {
		t.Errorf("GenerateByPattern() result = %v, want %v", result, "https://example.com/test/test-value")
	}
}

func TestPageURLGenerator_GenerateByPageID(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	ctx := context.Background()

	// Create a page with ID
	page := NewTestPage("/test/{id}/test/{value}")
	mockManager.AddPage(page)

	// Test ID-based generation using the main Generate method
	result, err := generator.Generate(ctx, page.Site, page.ID, "{id}", "test-id", "{value}", "test-value")
	if err != nil {
		t.Errorf("Generate() with ID = %v, want no error", err)
	}
	if result != "https://example.com/test/test-id/test/test-value" {
		t.Errorf("Generate() result = %v, want %v", result, "https://example.com/test/test-id/test/test-value")
	}
}

func TestPageURLGenerator_ErrorCases(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	site := NewTestSite()

	// Test with nil page
	if _, err := generator.GenerateByPage(site, nil); err == nil {
		t.Errorf("GenerateByPage() with nil page should return error")
	}

	// Test with CMS page (should work for CMS pages)
	cmsPage := NewTestPage(pages.PageCMS)
	cmsPage.URL = "/cms-page"
	cmsPage.Site = site
	mockManager.AddPage(cmsPage)

	if result, err := generator.GenerateByPage(site, cmsPage); err != nil {
		t.Errorf("GenerateByPage() with CMS page error = %v, want no error", err)
	} else if result != "https://example.com/cms-page" {
		t.Errorf("GenerateByPage() with CMS page result = %v, want %v", result, "https://example.com/cms-page")
	}

	// Test with internal page (should return empty string)
	internalPage := NewTestPage(pages.PageInternalCreate)
	internalPage.Site = site

	if result, err := generator.GenerateByPage(site, internalPage); err != nil {
		t.Errorf("GenerateByPage() with internal page error = %v, want no error", err)
	} else if result != "" {
		t.Errorf("GenerateByPage() with internal page result = %v, want empty string", result)
	}
}

func TestPageURLGenerator_PatternMatching(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Create pages with patterns
	userPage := NewTestPage("/users/{username}")
	mockManager.AddPage(userPage)

	blogPage := NewTestPage("/blog/{category}/{slug}")
	mockManager.AddPage(blogPage)

	// Test exact pattern match
	result, err := generator.GenerateByPattern(ctx, site, "/users/{username}", "{username}", "john.doe")
	if err != nil {
		t.Errorf("GenerateByPattern() error = %v, want no error", err)
	}
	if result != "https://example.com/users/john.doe" {
		t.Errorf("GenerateByPattern() result = %v, want %v", result, "https://example.com/users/john.doe")
	}

	// Test pattern with multiple parameters
	result, err = generator.GenerateByPattern(ctx, site, "/blog/{category}/{slug}", "{category}", "technology", "{slug}", "golang-tips")
	if err != nil {
		t.Errorf("GenerateByPattern() error = %v, want no error", err)
	}
	if result != "https://example.com/blog/technology/golang-tips" {
		t.Errorf("GenerateByPattern() result = %v, want %v", result, "https://example.com/blog/technology/golang-tips")
	}
}

func TestPageURLGenerator_GenerateMethod(t *testing.T) {
	mockManager := NewMockPageManager()
	generator := pages.NewPageURLGenerator(mockManager)
	ctx := context.Background()
	site := NewTestSite()

	// Test with Page struct
	page := NewTestPage("/test")
	mockManager.AddPage(page)

	if result, err := generator.Generate(ctx, site, page); err != nil {
		t.Errorf("Generate() with Page error = %v, want no error", err)
	} else if result != "https://example.com/test" {
		t.Errorf("Generate() with Page result = %v, want %v", result, "https://example.com/test")
	}

	// Test with *Page
	if result, err := generator.Generate(ctx, site, page); err != nil {
		t.Errorf("Generate() with *Page error = %v, want no error", err)
	} else if result != "https://example.com/test" {
		t.Errorf("Generate() with *Page result = %v, want %v", result, "https://example.com/test")
	}

	// Test with ID
	if result, err := generator.Generate(ctx, site, page.ID); err != nil {
		t.Errorf("Generate() with ID error = %v, want no error", err)
	} else if result != "https://example.com/test" {
		t.Errorf("Generate() with ID result = %v, want %v", result, "https://example.com/test")
	}

	// Test with pattern
	if result, err := generator.Generate(ctx, site, "/test"); err != nil {
		t.Errorf("Generate() with pattern error = %v, want no error", err)
	} else if result != "https://example.com/test" {
		t.Errorf("Generate() with pattern result = %v, want %v", result, "https://example.com/test")
	}

	// Test with CMS pattern
	cmsPage := NewTestPage(pages.PageCMS)
	cmsPage.URL = "/cms-page"
	cmsPage.Site = site
	mockManager.AddPage(cmsPage)

	if result, err := generator.Generate(ctx, site, pages.PageCMS, "url", "/cms-page"); err != nil {
		t.Errorf("Generate() with CMS pattern error = %v, want no error", err)
	} else if result != "https://example.com/cms-page" {
		t.Errorf("Generate() with CMS pattern result = %v, want %v", result, "https://example.com/cms-page")
	}

	// Test with unsupported type
	if _, err := generator.Generate(ctx, site, 123); err == nil {
		t.Errorf("Generate() with unsupported type should return error")
	}

	// Test with nil site
	if _, err := generator.Generate(ctx, nil, page); err == nil {
		t.Errorf("Generate() with nil site should return error")
	}
}
