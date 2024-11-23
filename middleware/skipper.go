package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/gowool/pages"
	"github.com/gowool/pages/repository"
)

func SiteSkipper(cfgRepository repository.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cfg, err := cfgRepository.Load(c.Request().Context())
			if err != nil {
				return err
			}
			if cfg.SiteSkippers != nil && cfg.SiteSkippers.Skipper(c) {
				ctx := c.Request().Context()
				ctx = pages.SetSkipSelectSite(ctx)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}

func PageSkipper(cfgRepository repository.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cfg, err := cfgRepository.Load(c.Request().Context())
			if err != nil {
				return err
			}
			if cfg.PageSkippers != nil && cfg.PageSkippers.Skipper(c) {
				ctx := c.Request().Context()
				ctx = pages.SetSkipSelectPage(ctx)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}
