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

// MockPatterns is a mock implementation of Patterns interface
type MockPatterns struct {
	mock.Mock
	patterns []string
}

func NewMockPatterns(patterns []string) *MockPatterns {
	return &MockPatterns{patterns: patterns}
}

func (m *MockPatterns) Patterns() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, pattern := range m.patterns {
			if !yield(pattern) {
				return
			}
		}
	}
}

// MockPageDecoratorStrategy is a mock implementation of PageDecoratorStrategy interface
type MockPageDecoratorStrategy struct {
	mock.Mock
}

func NewMockPageDecoratorStrategy(alwaysDecorable bool) *MockPageDecoratorStrategy {
	strategy := &MockPageDecoratorStrategy{}
	if alwaysDecorable {
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		strategy.On("IsPatternDecorable", mock.Anything, mock.Anything).Return(true, nil)
		strategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(true, nil)
	} else {
		strategy.On("IsDecorable", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
		strategy.On("IsPatternDecorable", mock.Anything, mock.Anything).Return(false, nil)
		strategy.On("IsURIDecorable", mock.Anything, mock.Anything).Return(false, nil)
	}
	return strategy
}

func (m *MockPageDecoratorStrategy) IsDecorable(ctx context.Context, pattern, uri string) (bool, error) {
	args := m.Called(ctx, pattern, uri)
	return args.Bool(0), args.Error(1)
}

func (m *MockPageDecoratorStrategy) IsPatternDecorable(ctx context.Context, pattern string) (bool, error) {
	args := m.Called(ctx, pattern)
	return args.Bool(0), args.Error(1)
}

func (m *MockPageDecoratorStrategy) IsURIDecorable(ctx context.Context, uri string) (bool, error) {
	args := m.Called(ctx, uri)
	return args.Bool(0), args.Error(1)
}

func TestNewDefaultPageSyncer(t *testing.T) {
	mockStore := &MockPageStore{}
	mockPatterns := NewMockPatterns([]string{})
	generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
	strategy := NewMockPageDecoratorStrategy(true)

	t.Run("valid parameters", func(t *testing.T) {
		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		assert.NotNil(t, syncer, "NewDefaultPageSyncer() should not return nil")
		assert.Equal(t, mockStore, syncer.store, "store should be set correctly")
		assert.Equal(t, mockPatterns, syncer.patterns, "patterns should be set correctly")
		assert.NotNil(t, syncer.generator, "generator should be set")
		assert.NotNil(t, syncer.strategy, "strategy should be set")
	})

	t.Run("nil page store", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(PageSyncerConfig{}, nil, generator, mockPatterns, strategy)
		}, "Should panic when store is nil")
	})

	t.Run("nil generator", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, nil, mockPatterns, strategy)
		}, "Should panic when generator is nil")
	})

	t.Run("nil patterns", func(t *testing.T) {
		assert.Panics(t, func() {
			NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, nil, strategy)
		}, "Should panic when patterns is nil")
	})

	t.Run("nil strategy", func(t *testing.T) {
		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		assert.NotNil(t, syncer.strategy, "strategy should be set to default when nil")

		// Test default strategy
		result, _ := syncer.strategy.IsDecorable(context.Background(), "/test", "/test")
		assert.True(t, result, "Default strategy should return true")
	})
}

func TestDefaultPageSyncer_Sync(t *testing.T) {
	ctx := context.Background()
	site := &Site{
		ID:   ID("test-site"),
		Name: "Test Site",
	}

	t.Run("sync with existing root page found by FindByURL", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}

		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 6
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with existing root page and no new patterns", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock existing root page
		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save for internal pages
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			// Should save 6 internal pages
			return len(pages) == 6
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with home hybrid pattern and no existing root", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{HomeHybridPattern})
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{HomeHybridPattern, PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save calls - root page is saved separately via createRootPage, then internal pages
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1 && pages[0].Pattern == HomeHybridPattern
		})).Return(nil).Once()

		// Since HomeHybridPattern was not found by FindByPatterns, it's considered a new page too
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 7 // HomeHybrid + 6 internal pages
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with new patterns and existing pages", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/blog/{slug}", "/about", "/contact"}
		mockPatterns := NewMockPatterns(routerPatterns)
		generator := func(context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

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
			return assert.ElementsMatch(t, ps, []string{"/blog/{slug}", "/about", "/contact", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			// Return existing /about page
			yield(existingPage, nil)
		}))

		// Mock save for new pages (blog/{slug}, contact, and internal pages)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			// Should save 8 new pages: /blog/{slug}, /contact, and 6 internal pages
			return len(pages) == 8
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should not return error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with generator error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID(""), errors.New("generator error") }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when generator fails")
		assert.Contains(t, err.Error(), "generate root page id error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with store error during FindByPatterns", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID("test-id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		// Mock FindByPatterns returning an error
		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			yield(nil, errors.New("store error"))
		}))

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when FindByPatterns fails")
		assert.Contains(t, err.Error(), "find page by pattern error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with generator error when creating new pages", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/blog/{slug}", "/about"}
		mockPatterns := NewMockPatterns(routerPatterns)

		// First call succeeds (root page), second call fails (new pages)
		callCount := 0
		generator := func(context.Context) (ID, error) {
			callCount++
			if callCount == 1 {
				return ID("root-id"), nil
			}
			return ID(""), errors.New("generator error for new page")
		}

		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/blog/{slug}", "/about", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save for root page (should succeed)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.Error(t, err, "Sync should return error when generator fails for new pages")
		assert.Contains(t, err.Error(), "generate page id error")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with no root page found and no home hybrid pattern", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/test"}
		mockPatterns := NewMockPatterns(routerPatterns)

		generator := func(context.Context) (ID, error) { return ID("root-id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock FindByURL returns nil (no root page found)
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		// Mock save for root page (since root is nil, it creates one)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1
		})).Return(nil).Once()

		// Mock save for new pages (/test + 6 internal pages = 7 pages)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 7
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed when creating root page")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with existing HomeHybridPattern page found", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{HomeHybridPattern, "/test"}
		mockPatterns := NewMockPatterns(routerPatterns)

		generator := func(context.Context) (ID, error) { return ID("id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		homeHybridPage := &Page{
			ID:      ID("home-hybrid-id"),
			SiteID:  site.ID,
			Pattern: HomeHybridPattern,
		}

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{HomeHybridPattern, "/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			yield(homeHybridPage, nil)
		}))

		// Mock save for new pages (/test + 6 internal pages)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 7
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed when existing home hybrid page is found")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with no new pages needed (all exist)", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/test"}
		mockPatterns := NewMockPatterns(routerPatterns)

		generator := func(context.Context) (ID, error) { return ID("id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}

		testPage := &Page{
			ID:      ID("test-id"),
			SiteID:  site.ID,
			Pattern: "/test",
		}

		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {
			yield(testPage, nil)
		}))

		// Mock save for internal pages only (6 pages, since /test exists)
		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 6
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed when no new pages needed")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with custom DefaultPage config", func(t *testing.T) {
		mockStore := &MockPageStore{}
		routerPatterns := []string{"/test"}
		mockPatterns := NewMockPatterns(routerPatterns)

		customTemplate := "custom/template.gohtml"
		customPosition := 10

		cfg := PageSyncerConfig{}
		cfg.DefaultPage = &PageConfig{
			Template: &customTemplate,
			Position: &customPosition,
		}

		generator := func(context.Context) (ID, error) { return ID("id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, strategy)

		rootPage := &Page{
			ID:      ID("root-id"),
			SiteID:  site.ID,
			Pattern: PageCMS,
			URL:     "/",
		}

		mockStore.On("FindByURL", ctx, site.ID, "/").Return(rootPage, nil)

		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
		})).Return(iter.Seq2[*Page, error](func(yield func(*Page, error) bool) {}))

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 7
		})).Return(nil).Once()

		err := syncer.Sync(ctx, site)
		assert.NoError(t, err, "Sync should succeed with custom default page config")

		mockStore.AssertExpectations(t)
	})

	t.Run("sync with store error during Save", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{"/test"})
		generator := func(context.Context) (ID, error) { return ID("new-id"), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, strategy)

		// Mock no existing root page
		mockStore.On("FindByURL", ctx, site.ID, "/").Return(nil, ErrPageNotFound)

		// Mock FindByPatterns for internal pages
		mockStore.On("FindByPatterns", ctx, site.ID, mock.MatchedBy(func(ps []string) bool {
			return assert.ElementsMatch(t, ps, []string{"/test", PageInternalCreate, PageErrorUnauthorized, PageErrorForbidden, PageErrorNotFound, PageError4xx, PageError5xx})
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
		mockPatterns := NewMockPatterns([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		cfg := PageSyncerConfig{}
		cfg.SetDefaults()
		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		t.Logf("homeHybridTemplate=%s, DefaultPatterns[HomeHybridPattern].Template=%v", homeHybridTemplate, cfg.DefaultPatterns[HomeHybridPattern].Template)
		if cfg.DefaultPatterns[HomeHybridPattern].Template != nil {
			t.Logf("Template value: %s", *cfg.DefaultPatterns[HomeHybridPattern].Template)
		}

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			if len(pages) != 1 {
				return false
			}
			p := pages[0]
			t.Logf("Page: Name=%s, Pattern=%s, Template=%s, Position=%d", p.Name, p.Pattern, p.Template, p.Position)
			return p.Pattern == HomeHybridPattern && p.Template == homeHybridTemplate
		})).Return(nil)

		rootPage, err := syncer.createRootPage(ctx, site, true)
		assert.NoError(t, err, "createRootPage should not return error")
		assert.NotNil(t, rootPage, "Root page should not be nil")
		assert.Equal(t, ID("root-id"), rootPage.ID, "Root page ID should be set")
		assert.Equal(t, site.ID, rootPage.SiteID, "Root page SiteID should be set")
		assert.Equal(t, HomeHybridPattern, rootPage.Pattern, "Root page pattern should be HomeHybridPattern")
		assert.Equal(t, homeHybridTemplate, rootPage.Template, "Root page template should be homeHybridTemplate")
		assert.Equal(t, 0, rootPage.Position, "Root page position should be 0")
		assert.Equal(t, site.Name+": Home Hybrid", rootPage.Name, "Root page name should include site name")

		mockStore.AssertExpectations(t)
	})

	t.Run("create regular CMS root page", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		mockStore.On("Save", ctx, mock.MatchedBy(func(pages []*Page) bool {
			return len(pages) == 1 && pages[0].Pattern == PageCMS && pages[0].URL == "/" && pages[0].Name == site.Name+": Home"
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
		assert.Equal(t, site.Name+": Home", rootPage.Name, "Root page name should include site name")

		mockStore.AssertExpectations(t)
	})

	t.Run("generator error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID(""), errors.New("generator error") }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		rootPage, err := syncer.createRootPage(ctx, site, true)
		assert.Error(t, err, "createRootPage should return error")
		assert.Nil(t, rootPage, "Root page should be nil")
		assert.Contains(t, err.Error(), "generate root page id error")
	})

	t.Run("store save error", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(ctx context.Context) (ID, error) { return ID("root-id"), nil }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

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

	t.Run("get patterns from patterns with valid GET patterns", func(t *testing.T) {
		routerPatterns := []string{
			"/blog/{slug}",
			"/about",
			"POST /contact", // Should be filtered out (POST method)
			"GET /products", // Should be kept (GET method)
			PageCMSPattern,  // Should be filtered out
		}

		mockPatterns := NewMockPatterns(routerPatterns)
		strategy := DefaultPageDecoratorStrategy

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, &MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockPatterns, strategy)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 9, "Should have 9 patterns (3 patterns + 6 internal)")
		assert.Contains(t, patterns, "/blog/{slug}", "Should include /blog/{slug}")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, "/products", "Should include /products")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.NotContains(t, patterns, "POST /contact", "Should not include POST /contact")
		assert.NotContains(t, patterns, PageCMSPattern, "Should not include CMS pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})

	t.Run("get patterns with home hybrid pattern", func(t *testing.T) {
		routerPatterns := []string{
			HomeHybridPattern,
			"/about",
		}

		mockPatterns := NewMockPatterns(routerPatterns)
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, &MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockPatterns, strategy)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 8, "Should have 8 patterns (2 patterns + 6 internal)")
		assert.Contains(t, patterns, HomeHybridPattern, "Should include HomeHybridPattern")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageErrorUnauthorized, "Should include internal 401 error pattern")
		assert.Contains(t, patterns, PageErrorForbidden, "Should include internal 403 error pattern")
		assert.Contains(t, patterns, PageErrorNotFound, "Should include internal 404 error pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.True(t, homeHybrid, "homeHybrid should be true")
	})

	t.Run("get patterns with home hybrid pattern", func(t *testing.T) {
		routerPatterns := []string{
			HomeHybridPattern,
			"/about",
		}

		mockPatterns := NewMockPatterns(routerPatterns)
		strategy := DefaultPageDecoratorStrategy

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, &MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockPatterns, strategy)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 8, "Should have 8 patterns (2 patterns + 6 internal)")
		assert.Contains(t, patterns, HomeHybridPattern, "Should include HomeHybridPattern")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageErrorUnauthorized, "Should include internal 401 error pattern")
		assert.Contains(t, patterns, PageErrorForbidden, "Should include internal 403 error pattern")
		assert.Contains(t, patterns, PageErrorNotFound, "Should include internal 404 error pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.True(t, homeHybrid, "homeHybrid should be true")
	})

	t.Run("get patterns with duplicate patterns", func(t *testing.T) {
		routerPatterns := []string{
			"/about",
			"/about", // Duplicate
			"/blog/{slug}",
			"/about", // Another duplicate
		}

		mockPatterns := NewMockPatterns(routerPatterns)
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, &MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockPatterns, strategy)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 8, "Should have 8 patterns (2 unique patterns + 6 internal)")
		assert.Contains(t, patterns, "/about", "Should include /about")
		assert.Contains(t, patterns, "/blog/{slug}", "Should include /blog/{slug}")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageErrorUnauthorized, "Should include internal 401 error pattern")
		assert.Contains(t, patterns, PageErrorForbidden, "Should include internal 403 error pattern")
		assert.Contains(t, patterns, PageErrorNotFound, "Should include internal 404 error pattern")
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
		mockPatterns := NewMockPatterns(routerPatterns)
		generator := func(ctx context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, store, generator, mockPatterns, strategy)

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
		mockPatterns := NewMockPatterns(routerPatterns)
		generator := func(ctx context.Context) (ID, error) { return ID(uuid.NewString()), nil }
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, store, generator, mockPatterns, strategy)

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

		// Should have: existing root, existing about, new contact, and 6 internal pages
		expectedNewPages := 7 // contact + 6 internal pages
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
		assert.Equal(t, existingAbout.Name, aboutPages[0].Name, "Should preserve existing page name")
	})
}

func TestDefaultPageSyncer_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("handle patterns pattern iteration stop", func(t *testing.T) {
		// Create a mock patterns that yields only one pattern then stops
		mockPatterns := NewMockPatterns([]string{"/pattern1"})
		strategy := NewMockPageDecoratorStrategy(true)

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, &MockPageStore{}, func(ctx context.Context) (ID, error) { return ID("test"), nil }, mockPatterns, strategy)

		patterns, homeHybrid := syncer.getPatterns(ctx)

		assert.Len(t, patterns, 7, "Should have 7 patterns (1 patterns + 6 internal)")
		assert.Contains(t, patterns, "/pattern1", "Should include the yielded pattern")
		assert.Contains(t, patterns, PageInternalCreate, "Should include internal create pattern")
		assert.Contains(t, patterns, PageErrorUnauthorized, "Should include internal 401 error pattern")
		assert.Contains(t, patterns, PageErrorForbidden, "Should include internal 403 error pattern")
		assert.Contains(t, patterns, PageErrorNotFound, "Should include internal 404 error pattern")
		assert.Contains(t, patterns, PageError4xx, "Should include internal 4xx error pattern")
		assert.Contains(t, patterns, PageError5xx, "Should include internal 5xx error pattern")
		assert.False(t, homeHybrid, "homeHybrid should be false")
	})
}

func TestDefaultPageSyncer_newPage(t *testing.T) {
	site := &Site{
		ID:   ID("test-site"),
		Name: "Test Site",
	}

	t.Run("new page with default config", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/about", site)

		assert.Equal(t, site.ID, page.SiteID)
		assert.Equal(t, site.Name+": /about", page.Name)
		assert.Equal(t, "/about", page.Pattern)
		assert.NotNil(t, page.Created)
		assert.NotNil(t, page.Updated)
		assert.Equal(t, hybridTemplate, page.Template)
		assert.Equal(t, 1, page.Position)
		assert.True(t, page.Decorate)
		assert.Equal(t, Draft, page.Status)
		assert.Equal(t, Public, page.Visibility)
	})

	t.Run("new page with custom name", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("Custom Name", "/custom", site)

		assert.Equal(t, site.Name+": Custom Name", page.Name)
	})

	t.Run("new page with internal pattern name", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		syncer := NewDefaultPageSyncer(PageSyncerConfig{}, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "_page_internal_error", site)

		assert.Equal(t, site.Name+": internal error", page.Name)
	})

	t.Run("new page with custom config", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		customTemplate := "custom/template.gohtml"
		customPosition := 5
		customDecorate := false
		customStatus := Published
		customVisibility := Private
		customParentID := ID("parent-id")

		cfg := PageSyncerConfig{}
		cfg.DefaultPatterns = map[string]*PageConfig{
			"/custom": {
				ParentID:   &customParentID,
				Template:   &customTemplate,
				Position:   &customPosition,
				Decorate:   &customDecorate,
				Status:     &customStatus,
				Visibility: &customVisibility,
			},
		}

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/custom", site)

		assert.Equal(t, &customParentID, page.ParentID)
		assert.Equal(t, customTemplate, page.Template)
		assert.Equal(t, customPosition, page.Position)
		assert.Equal(t, customDecorate, page.Decorate)
		assert.Equal(t, customStatus, page.Status)
		assert.Equal(t, customVisibility, page.Visibility)
	})

	t.Run("new page with metadata and meta tags", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		customMetaTags := NewMetaTags("custom-charset")
		customMetadata := NewMetadata(map[string]any{"key": "value"})

		cfg := PageSyncerConfig{}
		cfg.DefaultPatterns = map[string]*PageConfig{
			"/custom": {
				MetaTags: customMetaTags,
				Metadata: customMetadata,
			},
		}

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/custom", site)

		assert.Equal(t, customMetaTags, page.MetaTags)
		assert.Equal(t, customMetadata, page.Metadata)
	})

	t.Run("new page with custom headers", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		customHeaders := map[string][]string{
			"X-Custom-Header": {"value1", "value2"},
		}

		cfg := PageSyncerConfig{}
		cfg.DefaultPatterns = map[string]*PageConfig{
			"/custom": {
				Header: customHeaders,
			},
		}

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/custom", site)

		assert.Equal(t, customHeaders, page.Header)
	})
}

func TestPageSyncerConfig_SetDefaults(t *testing.T) {
	t.Run("SetDefaults with nil DefaultPage", func(t *testing.T) {
		cfg := PageSyncerConfig{}
		cfg.SetDefaults()

		assert.NotNil(t, cfg.DefaultPage)
		assert.NotNil(t, cfg.DefaultPage.Template)
		assert.NotNil(t, cfg.DefaultPage.Position)
		assert.NotNil(t, cfg.DefaultPage.Decorate)
		assert.NotNil(t, cfg.DefaultPage.Status)
		assert.NotNil(t, cfg.DefaultPage.Visibility)
		assert.NotNil(t, cfg.DefaultPage.MetaTags)
		assert.NotNil(t, cfg.DefaultPage.Metadata)
		assert.NotNil(t, cfg.DefaultPage.Header)
	})

	t.Run("SetDefaults with existing DefaultPage", func(t *testing.T) {
		customTemplate := "existing.gohtml"
		cfg := PageSyncerConfig{
			DefaultPage: &PageConfig{
				Template: &customTemplate,
			},
		}
		cfg.SetDefaults()

		assert.Equal(t, customTemplate, *cfg.DefaultPage.Template)
		assert.NotNil(t, cfg.DefaultPage.Position)
	})

	t.Run("SetDefaults with nil DefaultPatterns", func(t *testing.T) {
		cfg := PageSyncerConfig{}
		cfg.SetDefaults()

		assert.NotNil(t, cfg.DefaultPatterns)
		assert.NotNil(t, cfg.DefaultPatterns[HomeHybridPattern])
		assert.NotNil(t, cfg.DefaultPatterns[PageInternalCreate])
		assert.NotNil(t, cfg.DefaultPatterns[PageErrorUnauthorized])
		assert.NotNil(t, cfg.DefaultPatterns[PageErrorForbidden])
		assert.NotNil(t, cfg.DefaultPatterns[PageErrorNotFound])
		assert.NotNil(t, cfg.DefaultPatterns[PageError4xx])
		assert.NotNil(t, cfg.DefaultPatterns[PageError5xx])
	})

	t.Run("SetDefaults with existing DefaultPatterns", func(t *testing.T) {
		customTemplate := "custom.gohtml"
		cfg := PageSyncerConfig{
			DefaultPatterns: map[string]*PageConfig{
				HomeHybridPattern: {
					Template: &customTemplate,
				},
			},
		}
		cfg.SetDefaults()

		assert.Equal(t, customTemplate, *cfg.DefaultPatterns[HomeHybridPattern].Template)
		assert.NotNil(t, cfg.DefaultPatterns[PageInternalCreate])
	})

	t.Run("SetDefaults with nil pattern in DefaultPatterns", func(t *testing.T) {
		cfg := PageSyncerConfig{
			DefaultPatterns: map[string]*PageConfig{
				HomeHybridPattern: nil,
			},
		}
		cfg.SetDefaults()

		assert.NotNil(t, cfg.DefaultPatterns[HomeHybridPattern])
		assert.NotNil(t, cfg.DefaultPatterns[HomeHybridPattern].Template)
	})

	t.Run("SetDefaults with partial fields set", func(t *testing.T) {
		customStatus := Published
		cfg := PageSyncerConfig{
			DefaultPage: &PageConfig{
				Status: &customStatus,
			},
		}
		cfg.SetDefaults()

		assert.Equal(t, customStatus, *cfg.DefaultPage.Status)
		assert.NotNil(t, cfg.DefaultPage.Template)
		assert.NotNil(t, cfg.DefaultPage.Position)
		assert.NotNil(t, cfg.DefaultPage.Decorate)
		assert.NotNil(t, cfg.DefaultPage.Visibility)
	})
}

func TestDefaultPageSyncer_setPageConfig(t *testing.T) {
	t.Run("set page config with all fields", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		customParentID := ID("parent-id")
		customTemplate := "custom/template.gohtml"
		customPosition := 10
		customDecorate := false
		customStatus := Published
		customVisibility := Private
		customMetaTags := NewMetaTags("utf-8")
		customMetadata := NewMetadata(map[string]any{"key": "value"})
		customHeaders := map[string][]string{"X-Header": {"value"}}

		cfg := PageSyncerConfig{}
		cfg.DefaultPatterns = map[string]*PageConfig{
			"/test": {
				ParentID:   &customParentID,
				Template:   &customTemplate,
				Position:   &customPosition,
				Decorate:   &customDecorate,
				Status:     &customStatus,
				Visibility: &customVisibility,
				MetaTags:   customMetaTags,
				Metadata:   customMetadata,
				Header:     customHeaders,
			},
		}

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/test", &Site{ID: ID("site"), Name: "Site"})

		assert.Equal(t, &customParentID, page.ParentID)
		assert.Equal(t, customTemplate, page.Template)
		assert.Equal(t, customPosition, page.Position)
		assert.Equal(t, customDecorate, page.Decorate)
		assert.Equal(t, customStatus, page.Status)
		assert.Equal(t, customVisibility, page.Visibility)
		assert.Equal(t, customMetaTags, page.MetaTags)
		assert.Equal(t, customMetadata, page.Metadata)
		assert.Equal(t, customHeaders, page.Header)
	})

	t.Run("set page config with nil fields does not override", func(t *testing.T) {
		mockStore := &MockPageStore{}
		mockPatterns := NewMockPatterns([]string{})
		generator := func(context.Context) (ID, error) { return ID("test"), nil }

		site := &Site{ID: ID("site"), Name: "Site"}

		cfg := PageSyncerConfig{}
		cfg.DefaultPage = &PageConfig{
			Template:   ref("default.gohtml"),
			Position:   ref(5),
			Decorate:   ref(true),
			Status:     ref(Published),
			Visibility: ref(Public),
		}
		cfg.DefaultPatterns = map[string]*PageConfig{
			"/test": {}, // All nil
		}

		syncer := NewDefaultPageSyncer(cfg, mockStore, generator, mockPatterns, nil)

		page := syncer.newPage("", "/test", site)

		assert.Equal(t, "default.gohtml", page.Template)
		assert.Equal(t, 5, page.Position)
		assert.True(t, page.Decorate)
		assert.Equal(t, Published, page.Status)
		assert.Equal(t, Public, page.Visibility)
	})
}

func TestRef(t *testing.T) {
	t.Run("ref creates pointer to value", func(t *testing.T) {
		strValue := "test"
		strPtr := ref(strValue)
		assert.NotNil(t, strPtr)
		assert.Equal(t, "test", *strPtr)

		intValue := 42
		intPtr := ref(intValue)
		assert.NotNil(t, intPtr)
		assert.Equal(t, 42, *intPtr)

		boolValue := true
		boolPtr := ref(boolValue)
		assert.NotNil(t, boolPtr)
		assert.True(t, *boolPtr)

		statusValue := Published
		statusPtr := ref(statusValue)
		assert.NotNil(t, statusPtr)
		assert.Equal(t, Published, *statusPtr)
	})
}

func TestDefaultPages(t *testing.T) {
	t.Run("verify default page configs are properly defined", func(t *testing.T) {
		cfg := PageSyncerConfig{}
		cfg.SetDefaults()

		// Test that HomeHybridPattern config exists
		homeHybridConfig, ok := cfg.DefaultPatterns[HomeHybridPattern]
		require.True(t, ok, "Home hybrid pattern should exist in DefaultPatterns")
		require.NotNil(t, homeHybridConfig, "Home hybrid config should not be nil")
		assert.NotNil(t, homeHybridConfig.Template, "Home hybrid template should be set")
		assert.Equal(t, homeHybridTemplate, *homeHybridConfig.Template, "Home hybrid template should be correct")
		assert.NotNil(t, homeHybridConfig.Position, "Home hybrid position should be set")
		assert.Equal(t, 0, *homeHybridConfig.Position, "Home hybrid position should be 0")

		// Test that PageInternalCreate config exists
		createConfig, ok := cfg.DefaultPatterns[PageInternalCreate]
		require.True(t, ok, "Create page pattern should exist in DefaultPatterns")
		require.NotNil(t, createConfig, "Create page config should not be nil")
		assert.NotNil(t, createConfig.Template, "Create page template should be set")
		assert.Equal(t, createTemplate, *createConfig.Template, "Create page template should be correct")

		// Test that error page configs exist
		error4xxConfig, ok := cfg.DefaultPatterns[PageError4xx]
		require.True(t, ok, "4xx error page pattern should exist in DefaultPatterns")
		require.NotNil(t, error4xxConfig, "4xx error page config should not be nil")
		assert.NotNil(t, error4xxConfig.Template, "4xx error page template should be set")
		assert.Equal(t, error4xxTemplate, *error4xxConfig.Template, "4xx error page template should be correct")

		error5xxConfig, ok := cfg.DefaultPatterns[PageError5xx]
		require.True(t, ok, "5xx error page pattern should exist in DefaultPatterns")
		require.NotNil(t, error5xxConfig, "5xx error page config should not be nil")
		assert.NotNil(t, error5xxConfig.Template, "5xx error page template should be set")
		assert.Equal(t, error5xxTemplate, *error5xxConfig.Template, "5xx error page template should be correct")
	})

	t.Run("verify default page config is properly set", func(t *testing.T) {
		cfg := PageSyncerConfig{}
		cfg.SetDefaults()

		assert.NotNil(t, cfg.DefaultPage, "DefaultPage config should not be nil")
		assert.NotNil(t, cfg.DefaultPage.Template, "DefaultPage template should be set")
		assert.Equal(t, hybridTemplate, *cfg.DefaultPage.Template, "DefaultPage template should be hybrid")
		assert.NotNil(t, cfg.DefaultPage.Position, "DefaultPage position should be set")
		assert.Equal(t, 1, *cfg.DefaultPage.Position, "DefaultPage position should be 1")
		assert.NotNil(t, cfg.DefaultPage.Decorate, "DefaultPage decorate should be set")
		assert.True(t, *cfg.DefaultPage.Decorate, "DefaultPage decorate should be true")
		assert.NotNil(t, cfg.DefaultPage.Status, "DefaultPage status should be set")
		assert.Equal(t, Draft, *cfg.DefaultPage.Status, "DefaultPage status should be Draft")
		assert.NotNil(t, cfg.DefaultPage.Visibility, "DefaultPage visibility should be set")
		assert.Equal(t, Public, *cfg.DefaultPage.Visibility, "DefaultPage visibility should be Public")
		assert.NotNil(t, cfg.DefaultPage.MetaTags, "DefaultPage metaTags should be set")
	})

	t.Run("verify CMS patterns are removed from DefaultPatterns", func(t *testing.T) {
		cfg := PageSyncerConfig{}
		cfg.SetDefaults()

		_, ok := cfg.DefaultPatterns[PageCMS]
		assert.False(t, ok, "PageCMS should be removed from DefaultPatterns")

		_, ok = cfg.DefaultPatterns[PageCMSPattern]
		assert.False(t, ok, "PageCMSPattern should be removed from DefaultPatterns")
	})
}
