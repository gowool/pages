package fx

import (
	"io/fs"

	"github.com/gowool/echox"
	"github.com/gowool/echox/api"
	"github.com/gowool/theme"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
	cacherepo "github.com/gowool/pages/repository/cache"
	"github.com/gowool/pages/repository/fallback"
	fsrepo "github.com/gowool/pages/repository/fs"
	"github.com/gowool/pages/repository/sql"
	"github.com/gowool/pages/repository/sql/pg"
	"github.com/gowool/pages/repository/sql/sqlite"
)

var (
	OptionSQLDriverPG              = fx.Provide(pg.NewDriver)
	OptionSQLConfigurationDriverPG = fx.Provide(pg.NewConfigurationDriver)
	OptionSQLSequenceDriverPG      = fx.Provide(pg.NewSequenceDriver)

	OptionSQLDriverSqlite              = fx.Provide(sqlite.NewDriver)
	OptionSQLConfigurationDriverSqlite = fx.Provide(sqlite.NewConfigurationDriver)
	OptionSQLSequenceDriverSqlite      = fx.Provide(sqlite.NewSequenceDriver)

	OptionConfigurationRepository = fx.Provide(fx.Annotate(sql.NewConfigurationRepository, fx.As(new(repository.Configuration))))
	OptionMenuRepository          = fx.Provide(fx.Annotate(sql.NewMenuRepository, fx.As(new(repository.Menu))))
	OptionNodeRepository          = fx.Provide(fx.Annotate(sql.NewNodeRepository, fx.As(new(repository.Node))))
	OptionNodeSequenceRepository  = fx.Provide(fx.Annotate(sql.NewSequenceNodeRepository, fx.As(new(repository.SequenceNode))))
	OptionPageRepository          = fx.Provide(fx.Annotate(sql.NewPageRepository, fx.As(new(repository.Page))))
	OptionSiteRepository          = fx.Provide(fx.Annotate(sql.NewSiteRepository, fx.As(new(repository.Site))))
	OptionTemplateRepository      = fx.Provide(fx.Annotate(sql.NewTemplateRepository, fx.As(new(repository.Template))))
	OptionThemeRepository         = fx.Provide(repository.NewThemeRepository)

	OptionDecorateCacheConfigurationRepository = fx.Decorate(
		fx.Annotate(
			cacherepo.NewConfigurationRepository,
			fx.ParamTags("", `name:"repository-cache"`),
		),
	)
	OptionDecorateFallbackConfigurationRepository = fx.Decorate(func(r repository.Configuration) repository.Configuration {
		return fallback.NewConfigurationRepository(r, model.NewConfiguration())
	})
	OptionDecorateCacheSiteRepository = fx.Decorate(
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

	OptionThemeFuncMap = fx.Provide(fx.Annotate(FuncMap, fx.ResultTags(`group:"theme-func-map"`)))
	OptionThemeLoader  = fx.Provide(fx.Annotate(theme.NewRepositoryLoader, fx.As(new(theme.Loader))))

	OptionSiteSelectorMiddleware = fx.Provide(echox.AsMiddleware(SiteSelectorMiddleware))
	OptionPageSelectorMiddleware = fx.Provide(echox.AsMiddleware(PageSelectorMiddleware))
	OptionHybridPageMiddleware   = fx.Provide(echox.AsMiddleware(HybridPageMiddleware))

	OptionConfigurationAPI = fx.Provide(api.AsHandler(NewConfigurationAPI, fx.ParamTags("", `group:"api-option"`)))
	OptionSiteAPI          = fx.Provide(api.AsHandler(NewSiteAPI, fx.ParamTags("", `group:"api-option"`)))
	OptionPageAPI          = fx.Provide(api.AsHandler(NewPageAPI, fx.ParamTags("", `group:"api-option"`)))
	OptionTemplateAPI      = fx.Provide(api.AsHandler(NewTemplateAPI, fx.ParamTags("", `group:"api-option"`)))
	OptionMenuAPI          = fx.Provide(api.AsHandler(NewMenuAPI, fx.ParamTags("", `group:"api-option"`)))
	OptionNodeAPI          = fx.Provide(api.AsHandler(NewNodeAPI, fx.ParamTags("", `group:"api-option"`)))
)
