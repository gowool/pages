package fx

import (
	"io/fs"

	"github.com/gowool/echox"
	"github.com/gowool/echox/api"
	"github.com/gowool/theme"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/gowool/pages"
	v1 "github.com/gowool/pages/api/v1"
	"github.com/gowool/pages/repository"
	cacherepo "github.com/gowool/pages/repository/cache"
	"github.com/gowool/pages/repository/fallback"
	fsrepo "github.com/gowool/pages/repository/fs"
)

var (
	OptionDecorateCacheConfigurationRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewConfigurationRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateFallbackConfigurationRepository = fx.Decorate(fallback.NewConfigurationRepository)
	OptionDecorateCacheSiteRepository             = fx.Decorate(
		fx.Annotate(
			cacherepo.NewSiteRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateCachePageRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewPageRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateCacheTemplateRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewTemplateRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateFSTemplateRepository = fx.Decorate(
		fx.Annotate(
			func(r repository.Template, fss []fs.FS) repository.Template {
				for _, fsys := range fss {
					r = fsrepo.NewTemplateRepository(r, fsys)
				}
				return r
			},
			fx.ParamTags("", `name:"template-fs"`),
		),
	)
	OptionDecorateCacheMenuRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewMenuRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateCacheNodeRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewNodeRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)

	OptionSeeder  = fx.Provide(fx.Annotate(pages.NewDefaultSeeder, fx.As(new(pages.Seeder))))
	OptionMenu    = fx.Provide(fx.Annotate(pages.NewDefaultMenu, fx.As(new(pages.Menu))))
	OptionMatcher = fx.Provide(
		fx.Annotate(
			pages.NewDefaultMatcher,
			fx.As(new(pages.Matcher)),
			fx.ParamTags(`group:"menu-voter"`),
		),
	)
	OptionURLVoter = fx.Provide(
		fx.Annotate(
			pages.NewURLVoter,
			fx.As(new(pages.Voter)),
			fx.ResultTags(`group:"menu-voter"`),
		),
	)

	OptionSiteSelector = fx.Provide(
		fx.Annotate(
			pages.NewDefaultSiteSelector,
			fx.As(new(pages.SiteSelector)),
		),
	)
	OptionPageHandler = fx.Provide(
		fx.Annotate(
			pages.NewDefaultPageHandler,
			fx.As(new(pages.PageHandler)),
		),
	)
	OptionPageCreateHandler = fx.Provide(pages.NewPageCreateHandler)
	OptionErrorHandler      = fx.Provide(pages.NewErrorHandler)
	OptionHTTPErrorHandler  = fx.Provide(func(h *pages.ErrorHandler) echo.HTTPErrorHandler { return h.Handle })
	OptionErrorResolver     = fx.Provide(pages.ErrorResolver)
	OptionRenderer          = fx.Provide(fx.Annotate(pages.NewRenderer, fx.As(new(echo.Renderer))))

	OptionThemeFuncMap     = fx.Provide(AsFuncMap(FuncMap))
	OptionThemeFuncMapMenu = fx.Provide(AsFuncMap(FuncMapMenu))
	OptionThemeFuncMapPage = fx.Provide(AsFuncMap(FuncMapPage))
	OptionThemeLoader      = fx.Provide(fx.Annotate(theme.NewRepositoryLoader, fx.As(new(theme.Loader))))

	OptionSiteSelectorMiddleware = fx.Provide(echox.AsMiddleware(SiteSelectorMiddleware))
	OptionPageSelectorMiddleware = fx.Provide(echox.AsMiddleware(PageSelectorMiddleware))
	OptionHybridPageMiddleware   = fx.Provide(echox.AsMiddleware(HybridPageMiddleware))
	OptionSiteSkipperMiddleware  = fx.Provide(echox.AsMiddleware(SiteSkipperMiddleware))
	OptionPageSkipperMiddleware  = fx.Provide(echox.AsMiddleware(PageSkipperMiddleware))
	OptionLoggerMiddleware       = fx.Provide(echox.AsMiddleware(LoggerMiddleware))

	OptionConfigurationAPI = fx.Provide(api.AsHandler(v1.NewConfiguration, fx.ParamTags("", "", `group:"api-option"`)))
	OptionMenuAPI          = fx.Provide(api.AsHandler(v1.NewMenu, fx.ParamTags("", "", `group:"api-option"`)))
	OptionNodeAPI          = fx.Provide(api.AsHandler(v1.NewNode, fx.ParamTags("", "", `group:"api-option"`)))
	OptionPageAPI          = fx.Provide(api.AsHandler(v1.NewPage, fx.ParamTags("", "", `group:"api-option"`)))
	OptionSiteAPI          = fx.Provide(api.AsHandler(v1.NewSite, fx.ParamTags("", "", `group:"api-option"`)))
	OptionTemplateAPI      = fx.Provide(api.AsHandler(v1.NewTemplate, fx.ParamTags("", "", `group:"api-option"`)))
)
