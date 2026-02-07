package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gowool/got"
	"github.com/gowool/pages"
	"github.com/gowool/pages/internal"
	"github.com/gowool/r"
)

var (
	debug = true
	guest = false

	publicFS    = os.DirFS("public")
	templatesFS = os.DirFS("templates")
)

func main() {
	logger := slog.Default()

	strategy := NewPageDecoratorStrategy()
	authorizer := NewPageAuthorizer()

	generator := pages.IDGenerator()

	siteStore := pages.NewLocalhostSiteStore()
	pageStore := pages.NewMemoryPageStore()

	pageManager := pages.NewPageManager(pageStore)
	urlGenerator := pages.NewPageURLGenerator(pageManager)

	theme := newTheme("main", templatesFS)
	theme.SetParent(newTheme(".", pages.ErrorTemplateFS))
	theme.SetDebug(debug)
	theme.AddFuncMap(got.Funcs)
	theme.AddFuncMap(pages.SEOFuncMap)
	theme.AddFuncMap(pages.PageFuncMap(urlGenerator))

	siteRetriever := pages.NewHTTPSiteRetriever(siteStore)

	pageHandler := pages.NewPageHandlerWithConfig(theme, pages.PageHandlerConfig{
		Logger: logger,
	})
	pageCreate := pages.NewPageCreateHandlerWithConfig(pageStore, authorizer, pages.PageCreateHandlerConfig{
		GeneratorFunc: generator,
	})

	pageSkipper := pages.PageSkipper(strategy)
	faviconSkip := pages.EqualPathSkipper("/favicon.ico")

	selectSite := pages.SelectSiteMiddleware(siteRetriever, faviconSkip)
	selectPage := pages.SelectPageMiddleware(pageManager, authorizer, pages.PatternArgs(), faviconSkip, pageSkipper)
	hybridPage := pages.HybridPageMiddleware(pageHandler, logger, faviconSkip, pageSkipper)

	errorPattern := pages.NewHTTPErrorPattern(authorizer, strategy)
	errorHandler := pages.NewHTTPErrorHandlerWithConfig(pageHandler, pageManager, errorPattern, pages.HTTPErrorHandlerConfig{
		Logger: logger,
	})

	router := r.NewRouter(errorHandler)
	router.PreFunc(func(next r.Handler) r.Handler {
		return r.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			c := pages.MustContext(r.Context())
			c.SetDebug(debug)
			c.SetGuest(guest)

			return next.ServeHTTP(w, r)
		})
	})
	router.PreFunc(selectSite)
	router.UseFunc(selectPage)
	router.UseFunc(hybridPage)

	router.GET(pages.PageCMSPattern, pageHandler)
	router.POST("/_/create", pageCreate)
	router.GET("/favicon.ico", r.FileFS(publicFS, "favicon.ico"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := pages.NewContext(r.Context())
		defer cancel()

		router.ServeHTTP(w, r.WithContext(ctx))
	})

	syncer := pages.NewDefaultPageSyncer(
		pages.PageSyncerConfig{
			DefaultPage: &pages.PageConfig{
				Status: internal.Ref(pages.Published),
			},
		},
		pageStore,
		router,
		strategy,
		generator,
	)

	ctx := context.Background()

	for site, err := range siteStore.FindPublished(ctx) {
		if err != nil {
			panic(err)
		}

		if err = syncer.Sync(ctx, site); err != nil {
			panic(err)
		}
	}

	logger.Info("server started: http://localhost:8888")

	if err := http.ListenAndServe(":8888", handler); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		panic(err)
	}
}

func newTheme(name string, fsys fs.FS) *got.Theme {
	store := got.NewStoreFS(fsys)
	return got.NewTheme(name, store)
}

type PageDecoratorStrategy struct{}

func NewPageDecoratorStrategy() *PageDecoratorStrategy {
	return &PageDecoratorStrategy{}
}

func (PageDecoratorStrategy) IsPatternDecorable(ctx context.Context, pattern string) bool {
	return true
}

func (PageDecoratorStrategy) IsURIDecorable(ctx context.Context, uri string) bool {
	return true
}

type PageAuthorizer struct {
}

func NewPageAuthorizer() *PageAuthorizer {
	return &PageAuthorizer{}
}

func (PageAuthorizer) Authorize(ctx context.Context, action pages.PageAction) pages.Decision {
	if guest {
		return pages.Deny
	}
	return pages.Allow
}
