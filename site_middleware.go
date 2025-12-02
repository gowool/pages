package pages

import (
	"errors"

	"github.com/gowool/wo/middleware"
)

func SiteMiddleware[T Resolver](selector SiteSelector, skippers ...middleware.Skipper[T]) func(T) error {
	if selector == nil {
		panic("site middleware: selector is required")
	}

	skip := middleware.ChainSkipper[T](skippers...)

	return func(e T) error {
		r := e.Request()

		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		r.URL.RawPath = r.URL.Path

		if skip(e) {
			return e.Next()
		}

		site, pathInfo, err := selector.Retrieve(r)
		if err != nil {
			if errors.Is(err, ErrSiteNotFound) {
				return err
			}
			return errors.Join(err, ErrSiteNotFound)
		}

		if site == nil {
			return ErrSiteNotFound
		}

		if pathInfo != "" {
			r.URL.Path = pathInfo
		}

		site.Scheme = e.Scheme()
		site.Host = r.Host

		e.SetSite(site)

		return e.Next()
	}
}
