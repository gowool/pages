package main

import (
	"context"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/got"
	"github.com/gowool/keratin"
	"github.com/gowool/keratin/adapter"
	"github.com/gowool/keratin/middleware"
	"github.com/gowool/keratin/server"
	"github.com/gowool/pages"
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

type RouterWrapper struct {
	http.Handler
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := slog.Default()

	handler, sync := build(logger)

	if err := sync(ctx); err != nil {
		panic(err)
	}

	logger.Info("server started: http://localhost:8888")

	srv := server.New(server.Config{Address: ":8888"}, handler, logger.WithGroup("server"))

	go srv.Start(context.Background())

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("server stop error", "error", err)
	}
}

func build(logger *slog.Logger) (*RouterWrapper, func(context.Context) error) {
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
	theme.AddFuncMap(pages.FuncMap(urlGenerator))

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
	selectPage := pages.SelectPageMiddleware(pageManager, authorizer, faviconSkip, apiSkip, pageSkipper)
	hybridPage := pages.HybridPageMiddleware(pageHandler, logger, faviconSkip, apiSkip, pageSkipper)

	errorPattern := pages.ErrorPattern(authorizer, strategy)
	errorHandler := pages.ErrorHandler(pages.ErrorHandlerConfig{Logger: logger}, pageHandler, pageManager, errorPattern)

	router := keratin.NewRouter(
		keratin.WithErrorHandler(errorHandler),
		keratin.WithRequestInterceptor(func(r *http.Request) (*http.Request, func()) {
			ctx, cancel := pages.NewContext(r.Context())

			return r.WithContext(ctx), cancel
		}),
	)
	router.PreHTTPFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := pages.MustContext(r.Context())
			c.SetDebug(debug)
			c.SetGuest(guest)

			c.DOM().HTML.Attr = pages.NewAttr(
				"dir", "ltr",
				"lang", "en",
				"prefix", "og: https://ogp.me/ns#",
			)

			next.ServeHTTP(w, r)
		})
	})
	router.PreFunc(middleware.RequestLogger(middleware.RequestLoggerConfig{
		Logger:          logger,
		ErrorStatusFunc: pages.ErrorStatus,
	}))
	router.PreFunc(selectSite)
	router.PreFunc(func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			c := pages.MustContext(r.Context())
			if !c.HasSite() {
				return next.ServeHTTP(w, r)
			}

			c.Site().Title = "Wool Pages"

			if c.Site().Locale != "" {
				lang := strings.ReplaceAll(c.Site().Locale, "_", "-")
				c.DOM().HTML.Attr.Add("lang", lang)
				c.DOM().Head.Add(pages.PropertyMetaNode("og:locale", lang))
			}

			if c.Site().Title != "" {
				c.DOM().Head.Add(pages.PropertyMetaNode("og:site_name", c.Site().Title))
			}

			return next.ServeHTTP(w, r)
		})
	})

	router.GET("/favicon.ico", keratin.FileFS(publicFS, "favicon.ico"))

	front := router.Group("")
	front.UseFunc(selectPage)

	patternArgs := pages.PatternArgs()
	front.UseFunc(func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			c := pages.MustContext(r.Context())
			if !c.HasPage() {
				return next.ServeHTTP(w, r)
			}

			c.Page().Title = c.Page().Name

			var args []any
			if c.Page().IsDynamic() {
				args = patternArgs(r)
			}

			c.DOM().Head.Add(pages.PropertyMetaNode("og:url", c.Page().AbsURL(args...)))

			return next.ServeHTTP(w, r)
		})
	})

	front.UseFunc(hybridPage)

	front.Route("", pages.PageCMSPattern, pageHandler)
	front.Route(http.MethodPost, "/_/create", pageCreate)

	front.Any("/foo", func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("ALL FOO"))
		return err
	})

	front.GET("/foo", func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set(pages.HeaderXPageNotDecorable, "1")
		_, err := w.Write([]byte("GET FOO"))
		return err
	})

	front.GET("/boo", func(w http.ResponseWriter, r *http.Request) error {
		return keratin.HTML(w, http.StatusOK, "<h1>Page BOO</h1><p>This is hybrid wrapped page.</p>")
	})

	humaConfig := huma.DefaultConfig("Wool Pages", "0.0.1")
	humaConfig.Servers = []*huma.Server{{URL: "/api"}}

	wrapper := new(RouterWrapper)

	humaAPI := huma.NewAPI(humaConfig, adapter.NewAdapter(wrapper, router.Group("/api")))

	apiV1 := huma.NewGroup(humaAPI, "/v1")

	huma.Get(apiV1, "/patterns", func(ctx context.Context, _ *struct{}) (*humaResponse[[]string], error) {
		return &humaResponse[[]string]{
			Body: slices.Collect(router.Patterns()),
		}, nil
	})

	syncer := pages.NewDefaultPageSyncer(
		pages.PageSyncerConfig{
			DefaultPage: &pages.PageConfig{
				Status: new(pages.Published),
			},
		},
		pageStore,
		router,
		strategy,
		generator,
	)

	sync := func(ctx context.Context) error {
		sites, err := siteStore.FindPublished(ctx)
		if err != nil {
			return err
		}
		for _, site := range sites {
			if err = syncer.Sync(ctx, site); err != nil {
				return err
			}
		}
		return nil
	}

	wrapper.Handler = router.Build()

	return wrapper, sync
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
	if strings.HasPrefix(pattern, "/api") || strings.HasPrefix(pattern, "/favicon.ico") {
		return false
	}
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

type humaResponse[T any] struct {
	Body T
}
