package fx

import (
	"github.com/gowool/echox"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/gowool/pages"
	"github.com/gowool/pages/middleware"
	"github.com/gowool/pages/repository"
)

func SiteSelectorMiddleware(siteSelector pages.SiteSelector, cfgRepository repository.Configuration) echox.Middleware {
	return echox.NewMiddleware("site-selector", middleware.SiteSelector(middleware.SiteSelectorConfig{
		SiteSelector:  siteSelector,
		CfgRepository: cfgRepository,
	}))
}

type PageSelectorParams struct {
	fx.In
	PageHandler    pages.PageHandler
	CfgRepository  repository.Configuration
	PageRepository repository.Page
}

func PageSelectorMiddleware(params PageSelectorParams) echox.Middleware {
	return echox.NewMiddleware("page-selector", middleware.PageSelector(middleware.PageSelectorConfig{
		PageHandler:    params.PageHandler,
		CfgRepository:  params.CfgRepository,
		PageRepository: params.PageRepository,
	}))
}

func HybridPageMiddleware(pageHandler pages.PageHandler, cfgRepository repository.Configuration) echox.Middleware {
	return echox.NewMiddleware("hybrid-page", middleware.HybridPage(middleware.HybridPageConfig{
		PageHandler:   pageHandler,
		CfgRepository: cfgRepository,
	}))
}

func SiteSkipperMiddleware(cfgRepository repository.Configuration) echox.Middleware {
	return echox.NewMiddleware("site-skipper", middleware.SiteSkipper(cfgRepository))
}

func PageSkipperMiddleware(cfgRepository repository.Configuration) echox.Middleware {
	return echox.NewMiddleware("page-skipper", middleware.PageSkipper(cfgRepository))
}

func LoggerMiddleware(cfg echox.RequestLoggerConfig, cfgRepository repository.Configuration, logger *zap.Logger) echox.Middleware {
	skipper := cfg.Skipper
	cfg.Skipper = func(c echo.Context) bool {
		if skipper != nil && skipper(c) {
			return true
		}

		configuration, err := cfgRepository.Load(c.Request().Context())
		if err != nil || configuration.LoggerSkippers == nil {
			return false
		}
		return configuration.LoggerSkippers.Skipper(c)
	}

	additionalFieldsFunc := cfg.AdditionalFieldsFunc
	cfg.AdditionalFieldsFunc = func(c echo.Context) []zap.Field {
		var attributes []zap.Field
		if additionalFieldsFunc == nil {
			attributes = make([]zap.Field, 0, 4)
		} else {
			attributes = additionalFieldsFunc(c)
		}

		area := echox.CtxArea(c.Request().Context())
		attributes = append(attributes,
			zap.String("area", area),
			zap.String("pattern", c.Request().Pattern))

		if site := pages.CtxSite(c.Request().Context()); site != nil {
			attributes = append(attributes, zap.Dict("site",
				zap.Int64("id", site.ID),
				zap.String("name", site.Name),
				zap.String("host", site.Host),
				zap.String("locale", site.Locale),
				zap.String("relative-path", site.RelativePath),
			))
		}

		if page := pages.CtxPage(c.Request().Context()); page != nil {
			attributes = append(attributes, zap.Dict("page",
				zap.Int64("id", page.ID),
				zap.Int64("site-id", page.SiteID),
				zap.Int64p("parent-id", page.ParentID),
				zap.String("title", page.Title),
				zap.String("pattern", page.Pattern),
				zap.String("url", page.URL),
			))
		}
		return attributes
	}

	return echox.LoggerMiddleware(cfg, logger)
}
