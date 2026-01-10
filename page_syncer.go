package pages

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"net/http"
	"slices"

	"github.com/gowool/wo/middleware"
)

var _ PageSyncer = (*DefaultPageSyncer)(nil)

var defaultPages = map[string]*Page{
	PageInternalCreate: newPage("Create Page", PageInternalCreate, "internal/create.gohtml"),
	PageError4xx:       newPage("Error 4xx", PageError4xx, "internal/error/4xx.gohtml"),
	PageError5xx:       newPage("Error 5xx", PageError5xx, "internal/error/5xx.gohtml"),
	HomeHybridPattern:  newPage("Home Hybrid", HomeHybridPattern, "page/home_hybrid.gohtml"),
}

const (
	homeTemplate       = "page/home.gohtml"
	homeHybridTemplate = "page/home_hybrid.gohtml"
	hybridTemplate     = "page/hybrid.gohtml"
)

type PageSyncer interface {
	Sync(ctx context.Context, site *Site) error
}

type Router interface {
	Patterns() iter.Seq[string]
}

type IDGenerator func(ctx context.Context) (ID, error)

type IgnorePattern func(ctx context.Context, pattern string) bool

type DefaultPageSyncer struct {
	router    Router
	store     PageStore
	generator IDGenerator
	ignore    IgnorePattern
}

func NewDefaultPageSyncer(
	store PageStore,
	generator IDGenerator,
	router Router,
	ignore IgnorePattern,
) *DefaultPageSyncer {
	if store == nil {
		panic("page syncer: page store is required")
	}

	if generator == nil {
		panic("page syncer: id generator is required")
	}

	if router == nil {
		panic("page syncer: router is required")
	}

	if ignore == nil {
		ignore = func(_ context.Context, _ string) bool { return false }
	}

	return &DefaultPageSyncer{
		store:     store,
		generator: generator,
		router:    router,
		ignore:    ignore,
	}
}

func (s *DefaultPageSyncer) Sync(ctx context.Context, site *Site) error {
	var root *Page

	pages := make(map[string]*Page)

	patterns, homeHybrid := s.getPatterns(ctx)
	if !homeHybrid {
		root, _ = s.store.FindByURL(ctx, site.ID, "/")
	}

	for page, err := range s.store.FindByPatterns(ctx, site.ID, patterns...) {
		if err != nil {
			return fmt.Errorf("page syncer: find page by pattern error: %w", err)
		}

		if page.Pattern == HomeHybridPattern {
			root = page
		}

		pages[page.Pattern] = page
	}

	if root == nil {
		var err error
		root, err = s.createRootPage(ctx, site, homeHybrid)
		if err != nil {
			return fmt.Errorf("page syncer: create root page error: %w", err)
		}
	}

	newPages := make([]*Page, 0, len(patterns))
	for _, pattern := range patterns {
		if _, ok := pages[pattern]; ok {
			continue
		}

		id, err := s.generator(ctx)
		if err != nil {
			return fmt.Errorf("page syncer: generate page id error: %w", err)
		}

		var page *Page
		if p, ok := defaultPages[pattern]; ok {
			page = p.Copy()
		} else {
			page = newPage(pattern, pattern, hybridTemplate)
		}

		page.ID = id
		page.SiteID = site.ID
		page.Site = site
		page.ParentID = &root.ID
		page.Parent = root

		newPages = append(newPages, page)
	}

	if err := s.store.Save(ctx, newPages...); err != nil {
		return fmt.Errorf("page syncer: save pages error: %w", err)
	}

	return nil
}

func (s *DefaultPageSyncer) createRootPage(ctx context.Context, site *Site, homeHybrid bool) (*Page, error) {
	var root *Page

	if homeHybrid {
		root = newPage("Home Hybrid", HomeHybridPattern, homeHybridTemplate)
	} else {
		root = newPage("Home", PageCMS, homeTemplate)
		root.URL = "/"
	}

	id, err := s.generator(ctx)
	if err != nil {
		return nil, fmt.Errorf("page syncer: generate root page id error: %w", err)
	}

	root.ID = id
	root.SiteID = site.ID
	root.Site = site
	root.Position = 0

	if err := s.store.Save(ctx, root); err != nil {
		return nil, fmt.Errorf("page syncer: save root page error: %w", err)
	}

	return root, nil
}

func (s *DefaultPageSyncer) getPatterns(ctx context.Context) ([]string, bool) {
	patterns := map[string]struct{}{
		PageInternalCreate: {},
		PageError4xx:       {},
		PageError5xx:       {},
	}

	for pattern := range s.router.Patterns() {
		var ok bool
		if pattern, ok = middleware.CheckMethod(http.MethodGet, pattern); !ok {
			continue
		}

		if pattern == PageCMSPattern {
			continue
		}

		if s.ignore(ctx, pattern) {
			continue
		}

		patterns[pattern] = struct{}{}
	}

	var homeHybrid bool
	_, homeHybrid = patterns[HomeHybridPattern]

	return slices.Collect(maps.Keys(patterns)), homeHybrid
}

func newPage(name, pattern, template string) *Page {
	page := NewPage()
	page.Name = name
	page.Pattern = pattern
	page.Template = template
	page.Position = 1
	page.Decorate = true
	return page
}
