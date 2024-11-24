package middleware

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

func SiteSkipper(cfgRepository repository.Configuration) echo.MiddlewareFunc {
	return skipperMiddleware(cfgRepository, pages.SetSkipSelectSite, func(cfg model.Configuration) *model.Skippers {
		return cfg.SiteSkippers
	})
}

func PageSkipper(cfgRepository repository.Configuration) echo.MiddlewareFunc {
	return skipperMiddleware(cfgRepository, pages.SetSkipSelectPage, func(cfg model.Configuration) *model.Skippers {
		return cfg.PageSkippers
	})
}

func skipperMiddleware(
	cfgRepository repository.Configuration,
	setter func(context.Context) context.Context,
	fn func(model.Configuration) *model.Skippers,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cfg, err := cfgRepository.Load(c.Request().Context())
			if err != nil {
				return err
			}
			if s := fn(cfg); s != nil && s.Skipper(c) {
				ctx := c.Request().Context()
				ctx = setter(ctx)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}
