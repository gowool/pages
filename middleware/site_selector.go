package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gowool/pages"
	"github.com/gowool/pages/repository"
)

type SiteSelectorConfig struct {
	Skipper       middleware.Skipper
	SiteSelector  pages.SiteSelector
	CfgRepository repository.Configuration
}

func SiteSelector(cfg SiteSelectorConfig) echo.MiddlewareFunc {
	if cfg.CfgRepository == nil {
		panic("configuration repository is not specified")
	}
	if cfg.SiteSelector == nil {
		panic("site selector service is not specified")
	}
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			if r.Pattern == "" {
				r.Pattern = c.Path()
			}

			if cfg.Skipper(c) || pages.SkipSelectSite(r.Context()) {
				return next(c)
			}

			configuration, err := cfg.CfgRepository.Load(r.Context())
			if err != nil {
				return errors.Join(err, pages.ErrInternal)
			}

			if configuration.IgnoreURI(r.URL.Path) {
				return next(c)
			}

			site, urlPath, err := cfg.SiteSelector.Retrieve(r)
			if err != nil {
				var e pages.RedirectError
				if errors.As(err, &e) {
					return c.Redirect(e.Status, e.URL)
				}
				return errors.Join(err, pages.ErrInternal)
			}

			path := r.URL.Path
			rawPath := r.URL.RawPath
			defer func() {
				r.URL.Path = path
				r.URL.RawPath = rawPath
			}()

			r.URL.Path = urlPath
			r.URL.RawPath = ""

			ctx := pages.WithSite(r.Context(), site)
			c.SetRequest(r.WithContext(ctx))

			return next(c)
		}
	}
}
