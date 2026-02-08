package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/got"
	"github.com/gowool/keratin"
	"github.com/gowool/keratin/adapter"
	"github.com/gowool/keratin/middleware"
	"github.com/gowool/pages"
	"github.com/gowool/pages/internal"
)

var (
	debug = false
	guest = false

	publicFS    = os.DirFS("public")
	templatesFS = os.DirFS("templates")
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&guest, "guest", false, "guest mode")
	flag.Parse()
}

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
	faviconSkip := middleware.EqualPathSkipper("/favicon.ico")
	apiSkip := middleware.PrefixPathSkipper("/api")

	selectSite := pages.SelectSiteMiddleware(siteRetriever, faviconSkip, apiSkip)
	selectPage := pages.SelectPageMiddleware(pageManager, authorizer, pages.PatternArgs(), faviconSkip, apiSkip, pageSkipper)
	hybridPage := pages.HybridPageMiddleware(pageHandler, logger, faviconSkip, apiSkip, pageSkipper)

	errorPattern := pages.NewHTTPErrorPattern(authorizer, strategy)
	errorHandler := pages.NewHTTPErrorHandlerWithConfig(pageHandler, pageManager, errorPattern, pages.HTTPErrorHandlerConfig{
		Logger: logger,
	})

	router := keratin.NewRouter(errorHandler)
	router.PreFunc(func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			c := pages.MustContext(r.Context())
			c.SetDebug(debug)
			c.SetGuest(guest)

			return next.ServeHTTP(w, r)
		})
	})
	router.PreFunc(middleware.RequestLogger(middleware.RequestLoggerConfig{
		Logger:          logger,
		ErrorStatusFunc: pages.ErrorStatus,
	}))
	router.PreFunc(selectSite)

	router.GET("/favicon.ico", keratin.FileFS(publicFS, "favicon.ico"))

	front := router.Group("")
	front.UseFunc(selectPage)
	front.UseFunc(hybridPage)

	front.GET(pages.PageCMSPattern, pageHandler)
	front.POST("/_/create", pageCreate)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := pages.NewContext(r.Context())
		defer cancel()

		router.ServeHTTP(w, r.WithContext(ctx))
	})

	humaConfig := huma.DefaultConfig("Wool Pages", "0.0.1")
	humaConfig.Servers = []*huma.Server{{URL: "/api"}}

	humaAPI := huma.NewAPI(humaConfig, adapter.NewAdapter(&humRouter{
		Handler:     handler,
		RouterGroup: router.Group("/api"),
	}))

	apiV1 := huma.NewGroup(humaAPI, "/v1")

	huma.Get(apiV1, "/patterns", func(ctx context.Context, _ *struct{}) (*humaResponse[[]string], error) {
		return &humaResponse[[]string]{
			Body: slices.Collect(router.Patterns()),
		}, nil
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

type humRouter struct {
	http.Handler
	*keratin.RouterGroup
}

type humaResponse[T any] struct {
	Body T
}
