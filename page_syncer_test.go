package pages

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRouter is a mock implementation of Router interface
type MockRouter struct {
	mock.Mock
	patterns []string
}

func NewMockRouter(patterns []string) *MockRouter {
	return &MockRouter{patterns: patterns}
}

func (m *MockRouter) Patterns() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, pattern := range m.patterns {
			if !yield(pattern) {
				return
			}
		}
	}
}

func TestNewDefaultPageSyncer(t *testing.T) {
	mockStore := &MockPageStore{}
	mockRouter := NewMockRouter([]string{})
	generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
	ignore := func(context.Context, string) bool { return false }

	t.Run("valid parameters", func(t *testing.T) {
		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		assert.NotNil(t, syncer, "NewDefaultPageSyncer() should not return nil")
		assert.Equal(t, mockStore, syncer.store, "store should be set correctly")
		assert.Equal(t, mockRouter, syncer.router, "router should be set correctly")
		assert.NotNil(t, syncer.generator, "generator should be set")
		assert.NotNil(t, syncer.ignore, "ignore should be set")
	})

	t.Run("nil page store", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(nil, generator, mockRouter, ignore)
		}, "Should panic when store is nil")
	})

	t.Run("nil generator", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(mockStore, nil, mockRouter, ignore)
		}, "Should panic when generator is nil")
	})

	t.Run("nil router", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(mockStore, generator, nil, ignore)
		}, "Should panic when router is nil")
	})

	t.Run("nil ignore function", func(t *testing.T) {
		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, nil)

		assert.NotNil(t, syncer.ignore, "ignore should be set to default function when nil")

		// Test default ignore function
		result := syncer.ignore(context.Background(), "/test")
		assert.False(t, result, "Default ignore function should return false")
	})
}

func TestDefaultPageSyncer_Sync(t *testing.T) {
	ctx := context.Background()
	site := &Site{
		ID:   ID("test-site"),
		Name: "Test Site",
	}

	t.Run("sync with existing root page and no new patterns", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{})
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		// Mock existing root page
		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save for internal pages
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			// Should save 3 internal pages
			return len(pages) == 3
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with home hybrid pattern and no existing root", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{HomeHybridPattern})
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{HomeHybridPattern, PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save calls - root page is saved separately via createRootPage, then internal pages
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1 && pages[0].Pattern == HomeHybridPattern
		})).Return(nil).Once()

		// Since HomeHybridPattern was not found by FindByPatterns, it's considered a new page too
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 4 // HomeHybrid + 3 internal pages
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with new patterns and existing pages", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/blog/{slug}", "/about", "/contact"}
		mockRouter := NewMockRouter(routerPatterns)
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		// Mock existing root page
		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		// Mock existing pages
		existingPage := &Page{
			ID:      ID("existing-id"),
			SiteID:  site.ID,
			Pattern: "/about",
		}

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/blog/{slug}", "/about", "/contact", PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			// Return existing /about page
			yield(existingPage, nil)
		}))

		// Mock save for new pages (blog/{slug}, contact, and internal pages)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			// Should save 5 new pages: /blog/{slug}, /contact, and 3 internal pages
			return len(pages) == 5
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with generator error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID(""), errors.New("generator error") }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when generator fails")
		assert.Contains(t, err.Error(), "generate root page id error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with store error during FindByPatterns", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID("test-id"), nil }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		// Mock FindByPatterns returning an error
		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			yield(nil, errors.New("store error"))
		}))

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when FindByPatterns fails")
		assert.Contains(t, err.Error(), "find page by pattern error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with store error during Save", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID("new-id"), nil }
		ignore := func(context.Context, string) bool { return false }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, ignore)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		// Mock FindByPatterns for internal pages
		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save error
		mockStore.On("Save", ctx, mock.AnythingOfType("[]*pages.Page")).Return(errors.New("save error"))

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when Save fails")
		assert.Contains(t, err.Error(), "save root page error")

		mockStore.AssertExpectations(t)
	})
}

func TestDefaultPageSyncer_createRootPage(t *testing.T) {
	ctx := context.Background()
	site := &Site{
		ID:   ID("test-site"),
		Name: "Test Site",
	}

	t.Run("create home hybrid root page", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, nil)

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1 && pages[0].Pattern == HomeHybridPattern && pages[0].Template == homeHybridTemplate
		})).Return(nil)

		rootPage, err := syncer.createRootPage(ctx, site, true)
		assert.NoError(t, err, "createRootPage should not return error")
		assert.NotNil(t, rootPage, "Root page should not be nil")
		assert.Equal(t, ID("root-id"), rootPage.ID, "Root page ID should be set")
		assert.Equal(t, site.ID, rootPage.SiteID, "Root page SiteID should be set")
		assert.Equal(t, HomeHybridPattern, rootPage.Pattern, "Root page pattern should be HomeHybridPattern")
		assert.Equal(t, homeHybridTemplate, rootPage.Template, "Root page template should be homeHybridTemplate")
		assert.Equal(t, 0, rootPage.Position, "Root page position should be 0")

		mockStore.AssertExpectations(t)
	})

	t.Run("create regular CMS root page", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, nil)

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1 && pages[0].Pattern == PageCMS && pages[0].URL == "/"
		})).Return(nil)

		rootPage, err := syncer.createRootPage(ctx, site, false)
		assert.NoError(t, err, "createRootPage should not return error")
		assert.NotNil(t, rootPage, "Root page should not be nil")
		assert.Equal(t, ID("root-id"), rootPage.ID, "Root page ID should be set")
		assert.Equal(t, site.ID, rootPage.SiteID, "Root page SiteID should be set")
		assert.Equal(t, PageCMS, rootPage.Pattern, "Root page pattern should be PageCMS")
		assert.Equal(t, homeTemplate, rootPage.Template, "Root page template should be homeTemplate")
		assert.Equal(t, "/", rootPage.URL, "Root page URL should be /")
		assert.Equal(t, 0, rootPage.Position, "Root page position should be 0")

		mockStore.AssertExpectations(t)
	})

	t.Run("generator error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID(""), errors.New("generator error") }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, nil)

		rootPage, err := syncer.createRootPage(ctx, site, true)
		assert.Error(t, err, "createRootPage should return error")
		assert.Nil(t, rootPage, "Root page should be nil")
		assert.Contains(t, err.Error(), "generate root page id error")
	})

	t.Run("store save error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockRouter := NewMockRouter([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		syncer := NewDefaultPageSyncer(mockStore, generator, mockRouter, nil)

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1
		})).Return(errors.New("save error"))

		rootPage, err := syncer.createRootPage(ctx, site, true)
		assert.Error(t, err, "createRootPage should return error")
		assert.Nil(t, rootPage, "Root page should be nil")
		assert.Contains(t, err.Error(), "save root page error")

		mockStore.AssertExpectations(t)
	})
}

func TestDefaultPageSyncer_getPatterns(t *testing.T) {
	ctx := context.Background()

	t.Run("get patterns from router with valid GET patterns", func(t *testing.T) {
		routerPatterns := []string{
			"/blog/{slug}",
			"/about",
			"POST /contact", // Should be filtered out (POST method)
			"GET /products", // Should be kept (GET method)
			PageCMSPattern,  // Should be filtered out
			"/ignored",      // Should be filtered out by ignore function
		}

		mockRouter := NewMockRouter(routerPatterns)
		ignore := func(ctx context.Context, pattern string) bool {
			return pattern == "/ignored"
		}

		syncer := NewDefaultPageSyncer(&MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockRouter, ignore)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 6, "Should have 6 patterns (3 router + 3 internal)")
		assert.Contains(t, patterns, "/blog/{slug}", "Should include /blog/{slug}")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, "/products", "Should include /products")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.NotContains(t, patterns, "POST /contact", "Should not include POST /contact")
		assert.NotContains(t, patterns, PageCMSPattern, "Should not include CMS pattern")
		assert.NotContains(t, patterns, "/ignored", "Should not include ignored pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})

	t.Run("get patterns with home hybrid pattern", func(t *testing.T) {
		routerPatterns := []string{
			HomeHybridPattern,
			"/about",
		}

		mockRouter := NewMockRouter(routerPatterns)
		ignore := func(ctx context.Context, pattern string) bool { return false }

		syncer := NewDefaultPageSyncer(&MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockRouter, ignore)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 5, "Should have 5 patterns (2 router + 3 internal)")
		assert.Contains(t, patterns, HomeHybridPattern, "Should include HomeHybridPattern")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.True(t, homeHybrid, "homeHybrid should be true")
	})

	t.Run("get patterns with all patterns filtered out", func(t *testing.T) {
		routerPatterns := []string{
			"/blog/{slug}",
			"/about",
			PageCMSPattern,
		}

		mockRouter := NewMockRouter(routerPatterns)
		ignore := func(ctx context.Context, pattern string) bool { return true } // Ignore all

		syncer := NewDefaultPageSyncer(&MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockRouter, ignore)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 3, "Should have 3 internal patterns")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})

	t.Run("get patterns with duplicate patterns", func(t *testing.T) {
		routerPatterns := []string{
			"/about",
			"/about", // Duplicate
			"/blog/{slug}",
			"/about", // Another duplicate
		}

		mockRouter := NewMockRouter(routerPatterns)
		ignore := func(ctx context.Context, pattern string) bool { return false }

		syncer := NewDefaultPageSyncer(&MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockRouter, ignore)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 5, "Should have 5 patterns (2 unique router + 3 internal)")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, "/blog/{slug}", "Should include /blog/{slug}")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})
}

func TestDefaultPageSyncer_IntegrationScenarios(t *testing.T) {
	ctx := context.Background()
	site := &Site{
		ID:   ID("integration-site"),
		Name: "Integration Test Site",
	}

	t.Run("full sync workflow with new site", func(t *testing.T) {
		// Use real memory store for integration test
		store := NewMemoryPageStore()
		routerPatterns := []string{
			"/blog/{slug}",
			"/about",
			"/contact",
			"POST /api/contact", // Should be filtered out
		}
		mockRouter := NewMockRouter(routerPatterns)
		generator := func(ctx context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		ignore := func(ctx context.Context, pattern string) bool { return false }

		syncer := NewDefaultPageSyncer(store, generator, mockRouter, ignore)

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed for new site")

		// Verify root page was created
		rootPage, err := store.FindByURL(ctx, site.ID, "/")
		assert.NoError(t, err, "Should find root page")
		assert.Equal(t, PageCMS, rootPage.Pattern, "Root should be CMS page")

		// Verify hybrid pages were created
		blogPage, err := store.FindByPattern(ctx, site.ID, "/blog/{slug}")
		assert.NoError(t, err, "Should find blog page")
		assert.Equal(t, "/blog/{slug}", blogPage.Pattern)
		assert.Equal(t, hybridTemplate, blogPage.Template)
		assert.Equal(t, &rootPage.ID, blogPage.ParentID)

		aboutPage, err := store.FindByPattern(ctx, site.ID, "/about")
		assert.NoError(t, err, "Should find about page")
		assert.Equal(t, "/about", aboutPage.Pattern)
		assert.Equal(t, hybridTemplate, aboutPage.Template)

		contactPage, err := store.FindByPattern(ctx, site.ID, "/contact")
		assert.NoError(t, err, "Should find contact page")
		assert.Equal(t, "/contact", contactPage.Pattern)
		assert.Equal(t, hybridTemplate, contactPage.Template)

		// Verify internal pages were created
		_, err = store.FindByPattern(ctx, site.ID, PageInternalCreate)
		assert.NoError(t, err, "Should find create page")

		_, err = store.FindByPattern(ctx, site.ID, PageError4xx)
		assert.NoError(t, err, "Should find 4xx error page")

		_, err = store.FindByPattern(ctx, site.ID, PageError5xx)
		assert.NoError(t, err, "Should find 5xx error page")
	})

	t.Run("sync with existing pages should not create duplicates", func(t *testing.T) {
		store := NewMemoryPageStore()
		routerPatterns := []string{"/about", "/contact"}
		mockRouter := NewMockRouter(routerPatterns)
		generator := func(ctx context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		ignore := func(ctx context.Context, pattern string) bool { return false }

		syncer := NewDefaultPageSyncer(store, generator, mockRouter, ignore)

		// Pre-populate with some pages
		existingRoot := &Page{
			ID:      ID("existing-root"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
			Name:    "Home",
		}
		existingAbout := &Page{
			ID:       ID("existing-about"),
			SiteID:   site.ID,
			Pattern:  "/about",
			ParentID: &existingRoot.ID,
			Name:     "About",
		}

		err := store.Save(ctx, existingRoot, existingAbout)
		require.NoError(t, err, "Should save existing pages")

		// Get initial count
		dataBefore := store.GetData()
		initialCount := len(dataBefore)

		err = syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed")

		// Verify no duplicates were created
		dataAfter := store.GetData()
		finalCount := len(dataAfter)

		// Should have: existing root, existing about, new contact, and 3 internal pages
		expectedNewPages := 4 // contact + 3 internal pages
		assert.Equal(t, initialCount+expectedNewPages, finalCount, "Should create expected number of new pages")

		// Verify existing about page wasn't duplicated
		aboutPages := []*Page{}
		for _, page := range dataAfter {
			if page.Pattern == "/about" {
				aboutPages = append(aboutPages, page)
			}
		}
		assert.Len(t, aboutPages, 1, "Should have exactly one about page")
		assert.Equal(t, ID("existing-about"), aboutPages[0].ID, "Should be the existing about page")
	})
}

func TestDefaultPageSyncer_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("handle router pattern iteration stop", func(t *testing.T) {
		// Create a mock router that yields only one pattern then stops
		mockRouter := NewMockRouter([]string{"/pattern1"})
		ignore := func(ctx context.Context, pattern string) bool { return false }

		syncer := NewDefaultPageSyncer(&MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockRouter, ignore)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 4, "Should have 4 patterns (1 router + 3 internal)")
		assert.Contains(t, patterns, "/pattern1", "Should include the yielded pattern")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})
}

func TestDefaultPages(t *testing.T) {
	t.Run("verify default pages are properly defined", func(t *testing.T) {
		// Test that default pages exist and have correct properties
		createPage, ok := defaultPages[PageInternalCreate]
		require.True(t, ok, "Create page should exist in defaultPages")
		assert.Equal(t, "Create Page", createPage.Name, "Create page name should be correct")
		assert.Equal(t, PageInternalCreate, createPage.Pattern, "Create page pattern should be correct")
		assert.Equal(t, "internal/create.gohtml", createPage.Template, "Create page template should be correct")

		error4xxPage, ok := defaultPages[PageError4xx]
		require.True(t, ok, "4xx error page should exist in defaultPages")
		assert.Equal(t, "Error 4xx", error4xxPage.Name, "4xx error page name should be correct")
		assert.Equal(t, PageError4xx, error4xxPage.Pattern, "4xx error page pattern should be correct")
		assert.Equal(t, "internal/error/4xx.gohtml", error4xxPage.Template, "4xx error page template should be correct")

		error5xxPage, ok := defaultPages[PageError5xx]
		require.True(t, ok, "5xx error page should exist in defaultPages")
		assert.Equal(t, "Error 5xx", error5xxPage.Name, "5xx error page name should be correct")
		assert.Equal(t, PageError5xx, error5xxPage.Pattern, "5xx error page pattern should be correct")
		assert.Equal(t, "internal/error/5xx.gohtml", error5xxPage.Template, "5xx error page template should be correct")

		homeHybridPage, ok := defaultPages[HomeHybridPattern]
		require.True(t, ok, "Home hybrid page should exist in defaultPages")
		assert.Equal(t, "Home Hybrid", homeHybridPage.Name, "Home hybrid page name should be correct")
		assert.Equal(t, HomeHybridPattern, homeHybridPage.Pattern, "Home hybrid page pattern should be correct")
		assert.Equal(t, "page/home_hybrid.gohtml", homeHybridPage.Template, "Home hybrid page template should be correct")
	})

	t.Run("verify default pages can be copied safely", func(t *testing.T) {
		originalPage := defaultPages[PageInternalCreate]
		copiedPage := originalPage.Copy()

		assert.Equal(t, originalPage.Name, copiedPage.Name, "Copied page should have same name")
		assert.Equal(t, originalPage.Pattern, copiedPage.Pattern, "Copied page should have same pattern")
		assert.Equal(t, originalPage.Template, copiedPage.Template, "Copied page should have same template")

		// Modify the copy and verify original is unchanged
		copiedPage.Name = "Modified Name"
		assert.NotEqual(t, originalPage.Name, copiedPage.Name, "Original should be different from modified copy")
	})
}
