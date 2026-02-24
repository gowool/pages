package pages

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gowool/keratin/middleware"
)

var _ PageSyncer = (*DefaultPageSyncer)(nil)

const (
	homeTemplate       = "page/home.gohtml"
	homeHybridTemplate = "page/home_hybrid.gohtml"
	hybridTemplate     = "page/hybrid.gohtml"
	createTemplate     = "internal/create.gohtml"
	error401Template   = "internal/error/401.gohtml"
	error403Template   = "internal/error/403.gohtml"
	error404Template   = "internal/error/404.gohtml"
	error4xxTemplate   = "internal/error/4xx.gohtml"
	error5xxTemplate   = "internal/error/5xx.gohtml"
)

type PageSyncer interface {
	Sync(ctx context.Context, site *Site) error
}

type Patterns interface {
	Patterns() iter.Seq[string]
}

type PageConfig struct {
	ParentID   *ID                 `json:"parentID,omitempty" yaml:"parentID,omitempty"`
	Template   *string             `json:"template,omitempty" yaml:"template,omitempty"`
	Position   *int                `json:"position,omitempty" yaml:"position,omitempty"`
	Decorate   *bool               `json:"decorate,omitempty" yaml:"decorate,omitempty"`
	Status     *Status             `json:"status,omitempty" yaml:"status,omitempty"`
	Visibility *Visibility         `json:"visibility,omitempty" yaml:"visibility,omitempty"`
	DOM        *DOM                `json:"dom,omitempty" yaml:"dom,omitempty"`
	Meta       Meta                `json:"meta,omitempty" yaml:"meta,omitempty"`
	Header     map[string][]string `json:"header,omitempty" yaml:"header,omitempty"`
}

type PageSyncerConfig struct {
	DefaultPage     *PageConfig            `json:"defaultPage,omitempty" yaml:"defaultPage,omitempty"`
	DefaultPatterns map[string]*PageConfig `json:"defaultPatterns,omitempty" yaml:"defaultPatterns,omitempty"`
}

func (c *PageSyncerConfig) SetDefaults() {
	if c.DefaultPage == nil {
		c.DefaultPage = new(PageConfig)
	}

	if c.DefaultPage.Template == nil {
		c.DefaultPage.Template = new(hybridTemplate)
	}

	if c.DefaultPage.Position == nil {
		c.DefaultPage.Position = new(1)
	}

	if c.DefaultPage.Decorate == nil {
		c.DefaultPage.Decorate = new(true)
	}

	if c.DefaultPage.Status == nil {
		c.DefaultPage.Status = new(Draft)
	}

	if c.DefaultPage.Visibility == nil {
		c.DefaultPage.Visibility = new(Public)
	}

	if c.DefaultPage.Meta == nil {
		c.DefaultPage.Meta = NewMeta(nil)
	}

	if c.DefaultPage.Header == nil {
		c.DefaultPage.Header = make(map[string][]string)
	}

	if c.DefaultPatterns == nil {
		c.DefaultPatterns = make(map[string]*PageConfig)
	}

	if p, ok := c.DefaultPatterns[HomeHybridPattern]; !ok || p == nil {
		c.DefaultPatterns[HomeHybridPattern] = new(PageConfig)
	}
	if c.DefaultPatterns[HomeHybridPattern].Template == nil {
		c.DefaultPatterns[HomeHybridPattern].Template = new(homeHybridTemplate)
	}
	if c.DefaultPatterns[HomeHybridPattern].Position == nil {
		c.DefaultPatterns[HomeHybridPattern].Position = new(0)
	}

	if p, ok := c.DefaultPatterns[PageInternalCreate]; !ok || p == nil {
		c.DefaultPatterns[PageInternalCreate] = new(PageConfig)
	}
	if c.DefaultPatterns[PageInternalCreate].Template == nil {
		c.DefaultPatterns[PageInternalCreate].Template = new(createTemplate)
	}

	if p, ok := c.DefaultPatterns[PageErrorUnauthorized]; !ok || p == nil {
		c.DefaultPatterns[PageErrorUnauthorized] = new(PageConfig)
	}
	if c.DefaultPatterns[PageErrorUnauthorized].Template == nil {
		c.DefaultPatterns[PageErrorUnauthorized].Template = new(error401Template)
	}

	if p, ok := c.DefaultPatterns[PageErrorForbidden]; !ok || p == nil {
		c.DefaultPatterns[PageErrorForbidden] = new(PageConfig)
	}
	if c.DefaultPatterns[PageErrorForbidden].Template == nil {
		c.DefaultPatterns[PageErrorForbidden].Template = new(error403Template)
	}

	if p, ok := c.DefaultPatterns[PageErrorNotFound]; !ok || p == nil {
		c.DefaultPatterns[PageErrorNotFound] = new(PageConfig)
	}
	if c.DefaultPatterns[PageErrorNotFound].Template == nil {
		c.DefaultPatterns[PageErrorNotFound].Template = new(error404Template)
	}

	if p, ok := c.DefaultPatterns[PageError4xx]; !ok || p == nil {
		c.DefaultPatterns[PageError4xx] = new(PageConfig)
	}
	if c.DefaultPatterns[PageError4xx].Template == nil {
		c.DefaultPatterns[PageError4xx].Template = new(error4xxTemplate)
	}

	if p, ok := c.DefaultPatterns[PageError5xx]; !ok || p == nil {
		c.DefaultPatterns[PageError5xx] = new(PageConfig)
	}
	if c.DefaultPatterns[PageError5xx].Template == nil {
		c.DefaultPatterns[PageError5xx].Template = new(error5xxTemplate)
	}
}

type DefaultPageSyncer struct {
	cfg       PageSyncerConfig
	patterns  Patterns
	store     PageStore
	strategy  PageDecoratorStrategy
	generator IDGeneratorFunc
}

func NewDefaultPageSyncer(
	cfg PageSyncerConfig,
	store PageStore,
	patterns Patterns,
	strategy PageDecoratorStrategy,
	generator IDGeneratorFunc,
) *DefaultPageSyncer {
	if store == nil {
		panic("page syncer: page store is required")
	}

	if patterns == nil {
		panic("page syncer: patterns is required")
	}

	if strategy == nil {
		panic("page syncer: decorator strategy is required")
	}

	if generator == nil {
		generator = IDGenerator()
	}

	cfg.SetDefaults()

	return &DefaultPageSyncer{
		cfg:       cfg,
		store:     store,
		generator: generator,
		patterns:  patterns,
		strategy:  strategy,
	}
}

func (s *DefaultPageSyncer) Sync(ctx context.Context, site *Site) error {
	var root *Page

	pages := make(map[string]struct{})

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

		pages[site.ID.String()+page.Pattern] = struct{}{}
	}

	if root == nil {
		var err error
		if root, err = s.createRootPage(ctx, site, homeHybrid); err != nil {
			return fmt.Errorf("page syncer: create root page error: %w", err)
		}
		pages[site.ID.String()+root.Pattern] = struct{}{}
	}

	newPages := make([]*Page, 0, len(patterns))
	for _, pattern := range patterns {
		if _, ok := pages[site.ID.String()+pattern]; ok {
			continue
		}

		id, err := s.generator(ctx)
		if err != nil {
			return fmt.Errorf("page syncer: generate page id error: %w", err)
		}

		page := s.newPage("", pattern, site)
		page.ID = id
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
	id, err := s.generator(ctx)
	if err != nil {
		return nil, fmt.Errorf("page syncer: generate root page id error: %w", err)
	}

	var root *Page

	if homeHybrid {
		root = s.newPage("Home Hybrid", HomeHybridPattern, site)
	} else {
		root = s.newPage("Home", PageCMS, site)
		root.CustomURL = "/"
		root.Position = 0
		root.Template = homeTemplate
		root.FixURL()
	}

	root.ID = id

	if err := s.store.Save(ctx, root); err != nil {
		return nil, fmt.Errorf("page syncer: save root page error: %w", err)
	}

	return root, nil
}

func (s *DefaultPageSyncer) getPatterns(ctx context.Context) ([]string, bool) {
	patterns := make(map[string]struct{})

	for pattern := range s.patterns.Patterns() {
		var ok bool
		if pattern, ok = middleware.CheckMethod(http.MethodGet, pattern); !ok {
			continue
		}

		if pattern == PageCMSPattern {
			continue
		}

		if !s.strategy.IsPatternDecorable(ctx, pattern) {
			continue
		}

		patterns[pattern] = struct{}{}
	}

	var homeHybrid bool
	_, homeHybrid = patterns[HomeHybridPattern]

	for pattern := range s.cfg.DefaultPatterns {
		if strings.HasPrefix(pattern, PageInternalPrefix) {
			patterns[pattern] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(patterns)), homeHybrid
}

func (s *DefaultPageSyncer) newPage(name, pattern string, site *Site) *Page {
	if name == "" {
		name = pattern
	}

	name = site.Name + ": " + name

	t := time.Now().UTC()

	p := &Page{
		Site:    site,
		SiteID:  site.ID,
		Name:    name,
		Pattern: pattern,
		Created: t,
		Updated: t,
	}

	s.setPageConfig(p, s.cfg.DefaultPage)
	if patternPage, ok := s.cfg.DefaultPatterns[pattern]; ok && patternPage != nil {
		s.setPageConfig(p, patternPage)
	}

	return p
}

func (s *DefaultPageSyncer) setPageConfig(page *Page, config *PageConfig) {
	if config.ParentID != nil {
		page.ParentID = config.ParentID
	}
	if config.Template != nil {
		page.Template = *config.Template
	}
	if config.Position != nil {
		page.Position = *config.Position
	}
	if config.Decorate != nil {
		page.Decorate = *config.Decorate
	}
	if config.Status != nil {
		page.Status = *config.Status
	}
	if config.Visibility != nil {
		page.Visibility = *config.Visibility
	}
	if config.DOM != nil {
		page.DOM = config.DOM.Copy()
	}
	if config.Meta != nil {
		page.Meta = config.Meta
	}
	if config.Header != nil {
		page.Header = config.Header
	}
}
