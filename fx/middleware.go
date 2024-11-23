package fx

import (
	"github.com/gowool/echox"
	"go.uber.org/fx"

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
